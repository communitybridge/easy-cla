// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package sign

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/communitybridge/easycla/cla-backend-go/github"
	"github.com/communitybridge/easycla/cla-backend-go/github_organizations"
	"github.com/communitybridge/easycla/cla-backend-go/project/common"
	"github.com/communitybridge/easycla/cla-backend-go/projects_cla_groups"
	"github.com/communitybridge/easycla/cla-backend-go/repositories"
	"github.com/communitybridge/easycla/cla-backend-go/signatures"
	"github.com/communitybridge/easycla/cla-backend-go/users"
	"github.com/communitybridge/easycla/cla-backend-go/v2/cla_groups"
	"github.com/communitybridge/easycla/cla-backend-go/v2/gitlab_organizations"
	"github.com/communitybridge/easycla/cla-backend-go/v2/store"
	"github.com/go-openapi/strfmt"
	"github.com/gofrs/uuid"

	"github.com/sirupsen/logrus"

	acsService "github.com/communitybridge/easycla/cla-backend-go/v2/acs-service"
	"github.com/communitybridge/easycla/cla-backend-go/v2/organization-service/client/organizations"

	organizationService "github.com/communitybridge/easycla/cla-backend-go/v2/organization-service"

	projectService "github.com/communitybridge/easycla/cla-backend-go/v2/project-service"
	userService "github.com/communitybridge/easycla/cla-backend-go/v2/user-service"

	log "github.com/communitybridge/easycla/cla-backend-go/logging"

	"github.com/communitybridge/easycla/cla-backend-go/company"
	v1Models "github.com/communitybridge/easycla/cla-backend-go/gen/v1/models"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	"github.com/communitybridge/easycla/cla-backend-go/utils"
)

// constants
const (
	DontLoadRepoDetails = false
	DocSignFalse        = "false"
)

// errors
var (
	ErrCCLANotEnabled        = errors.New("corporate license agreement is not enabled with this project")
	ErrTemplateNotConfigured = errors.New("cla template not configured for this project")
	ErrNotInOrg              error
)

// ProjectRepo contains project repo methods
type ProjectRepo interface {
	GetCLAGroupByID(ctx context.Context, claGroupID string, loadRepoDetails bool) (*v1Models.ClaGroup, error)
}

// Service interface defines the sign service methods
type Service interface {
	VoidEnvelope(ctx context.Context, envelopeID, message string) error
	PrepareSignRequest(ctx context.Context, signRequest *DocuSignEnvelopeRequest) (*DocusignEnvelopeResponse, error)
	GetSignURL(email, recipientID, userName, clientUserId, envelopeID, returnURL string) (string, error)
	createEnvelope(ctx context.Context, payload *DocuSignEnvelopeRequest) (string, error)
	addDocumentToEnvelope(ctx context.Context, envelopeID, documentName string, document []byte) error

	RequestCorporateSignature(ctx context.Context, lfUsername string, authorizationHeader string, input *models.CorporateSignatureInput) (*models.CorporateSignatureOutput, error)
	RequestIndividualSignature(ctx context.Context, input *models.IndividualSignatureInput, preferredEmail string) (*models.IndividualSignatureOutput, error)
	RequestIndividualSignatureGerrit(ctx context.Context, input *models.IndividualSignatureInput) (*models.IndividualSignatureOutput, error)
	SignedIndividualCallbackGithub(ctx context.Context, payload []byte, installationID, changeRequestID, repositoryID string) error
	SignedIndividualCallbackGitlab(ctx context.Context, payload []byte, userID, organizationID, mergeRequestID, repositoryID string) error
	SignedIndividualCallbackGerrit(ctx context.Context, payload []byte, userID string) error
	SignedCorporateCallback(ctx context.Context, payload []byte, companyID, projectID string) error
}

// service
type service struct {
	ClaV4ApiURL          string
	ClaV1ApiURL          string
	companyRepo          company.IRepository
	projectRepo          ProjectRepo
	projectClaGroupsRepo projects_cla_groups.Repository
	companyService       company.IService
	claGroupService      cla_groups.Service
	docsignPrivateKey    string
	userService          users.Service
	signatureService     signatures.SignatureService
	storeRepository      store.Repository
	repositoryService    repositories.Service
	githubOrgService     github_organizations.Service
	gitlabOrgService     gitlab_organizations.ServiceInterface
}

// NewService returns an instance of v2 project service
func NewService(apiURL, v1API string, compRepo company.IRepository, projectRepo ProjectRepo, pcgRepo projects_cla_groups.Repository, compService company.IService, claGroupService cla_groups.Service, docsignPrivateKey string, userService users.Service, signatureService signatures.SignatureService, storeRepository store.Repository,
	repositoryService repositories.Service, githubOrgService github_organizations.Service, gitlabOrgService gitlab_organizations.ServiceInterface) Service {
	return &service{
		ClaV4ApiURL:          apiURL,
		ClaV1ApiURL:          v1API,
		companyRepo:          compRepo,
		projectRepo:          projectRepo,
		projectClaGroupsRepo: pcgRepo,
		companyService:       compService,
		claGroupService:      claGroupService,
		docsignPrivateKey:    docsignPrivateKey,
		userService:          userService,
		signatureService:     signatureService,
		storeRepository:      storeRepository,
		githubOrgService:     githubOrgService,
		gitlabOrgService:     gitlabOrgService,
		repositoryService:    repositoryService,
	}
}

type requestCorporateSignatureInput struct {
	ProjectID         string `json:"project_id,omitempty"`
	CompanyID         string `json:"company_id,omitempty"`
	SendAsEmail       bool   `json:"send_as_email,omitempty"`
	SigningEntityName string `json:"signing_entity_name,omitempty"`
	AuthorityName     string `json:"authority_name,omitempty"`
	AuthorityEmail    string `json:"authority_email,omitempty"`
	ReturnURL         string `json:"return_url,omitempty"`
}

func validateCorporateSignatureInput(input *models.CorporateSignatureInput) error {
	if input.SendAsEmail {
		log.Debugf("input.AuthorityName validation %s", input.AuthorityName)
		if strings.TrimSpace(input.AuthorityName) == "" {
			log.Warn("error in input.AuthorityName ")
			return errors.New("require authority_name")
		}
		if input.AuthorityEmail == "" {
			return errors.New("require authority_email")
		}
	} else {
		if input.ReturnURL.String() == "" {
			return errors.New("require return_url")
		}
	}
	if input.ProjectSfid == nil || *input.ProjectSfid == "" {
		return errors.New("require project_sfid")
	}
	if input.CompanySfid == nil || *input.CompanySfid == "" {
		return errors.New("require company_sfid")
	}
	return nil
}

func (s *service) RequestCorporateSignature(ctx context.Context, lfUsername string, authorizationHeader string, input *models.CorporateSignatureInput) (*models.CorporateSignatureOutput, error) { // nolint
	f := logrus.Fields{
		"functionName":      "sign.RequestCorporateSignature",
		utils.XREQUESTID:    ctx.Value(utils.XREQUESTID),
		"lfUsername":        lfUsername,
		"projectSFID":       input.ProjectSfid,
		"companySFID":       input.CompanySfid,
		"signingEntityName": input.SigningEntityName,
		"authorityName":     input.AuthorityName,
		"authorityEmail":    input.AuthorityEmail.String(),
		"sendAsEmail":       input.SendAsEmail,
		"returnURL":         input.ReturnURL,
	}

	/**
		1. Ensure Company Exists
		2. Ensure this is a valid project
	**/

	usc := userService.GetClient()

	log.WithFields(f).Debug("validating input parameters...")
	err := validateCorporateSignatureInput(input)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("unable to validat corporate signature input")
		return nil, err
	}

	// 1. Ensure Company Exists
	var comp *v1Models.Company
	// Backwards compatible - if the signing entity name is not set, then we fall back to using the CompanySFID lookup
	// which will return the company record where the company name == signing entity name
	if input.SigningEntityName == "" {
		comp, err = s.companyRepo.GetCompanyByExternalID(ctx, utils.StringValue(input.CompanySfid))
		if err != nil {
			log.WithFields(f).WithError(err).Warn("unable to fetch company records by signing entity name value")
			return nil, err
		}
	} else {
		// Big change here - since we can have multiple EasyCLA Company records with the same external SFID, we now
		// switch over to query by the signing entity name.
		comp, err = s.companyRepo.GetCompanyBySigningEntityName(ctx, input.SigningEntityName)
		if err != nil {
			log.WithFields(f).WithError(err).Warn("unable to fetch company records by signing entity name value")
			return nil, err
		}
	}

	// 2. Ensure this is a valid project
	psc := projectService.GetClient()
	log.WithFields(f).Debug("looking up project by SFID...")
	project, err := psc.GetProject(utils.StringValue(input.ProjectSfid))
	if err != nil {
		log.WithFields(f).WithError(err).Warn("unable to fetch project SFID")
		return nil, err
	}

	var claGroupID string
	if !utils.IsProjectHaveParent(project) || utils.IsProjectHasRootParent(project) || utils.GetProjectParentSFID(project) == "" {
		// this is root project
		cgmlist, perr := s.projectClaGroupsRepo.GetProjectsIdsForFoundation(ctx, utils.StringValue(input.ProjectSfid))
		if perr != nil {
			log.WithFields(f).WithError(err).Warn("unable to lookup other projects associated with this project SFID")
			return nil, perr
		}
		if len(cgmlist) == 0 {
			// no cla group is link with root_project
			return nil, projects_cla_groups.ErrProjectNotAssociatedWithClaGroup
		}
		claGroups := utils.NewStringSet()
		for _, cg := range cgmlist {
			claGroup, claGroupErr := s.claGroupService.GetCLAGroup(ctx, cg.ClaGroupID)
			if err != nil {
				log.WithFields(f).WithError(claGroupErr).Warn("unable to lookup cla group")
				return nil, err
			}

			// ensure that cla group for project is a foundation level cla group
			if claGroup != nil && cg.ProjectSFID == utils.StringValue(input.ProjectSfid) {
				claGroups.Add(cg.ClaGroupID)
			}
		}

		if claGroups.Length() > 1 {
			// multiple cla group are linked with root_project
			// so we can not determine which cla-group to use
			return nil, errors.New("invalid project_sfid. multiple cla-groups are associated with this project_sfid")
		}
		claGroupID = (claGroups.List())[0]

	} else {
		cgm, perr := s.projectClaGroupsRepo.GetClaGroupIDForProject(ctx, utils.StringValue(input.ProjectSfid))
		if perr != nil {
			log.WithFields(f).WithError(err).Warn("unable to lookup CLA Group ID for this project SFID")
			return nil, perr
		}
		claGroupID = cgm.ClaGroupID
	}

	f["claGroupID"] = claGroupID
	log.WithFields(f).Debug("loading CLA Group by ID...")
	proj, err := s.projectRepo.GetCLAGroupByID(ctx, claGroupID, DontLoadRepoDetails)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("unable to lookup CLA Group by CLA Group ID")
		return nil, err
	}
	if !proj.ProjectCCLAEnabled {
		log.WithFields(f).Warn("unable to request corporate signature - CCLA is not enabled for this CLA Group")
		return nil, ErrCCLANotEnabled
	}
	if len(proj.ProjectCorporateDocuments) == 0 {
		log.WithFields(f).Warn("unable to request corporate signature - missing corporate documents in the CLA Group configuration")
		return nil, ErrTemplateNotConfigured
	}
	var currentUserEmail string
	// Email flow
	if input.SendAsEmail {
		log.WithFields(f).Debugf("Sending request as an email to: %s...", input.AuthorityEmail.String())
		// this would be used only in case of cla-signatory
		err = prepareUserForSigning(ctx, input.AuthorityEmail.String(), utils.StringValue(input.CompanySfid), utils.StringValue(input.ProjectSfid), input.SigningEntityName)
		if err != nil {
			// Ignore conflict - role has already been assigned
			if _, ok := err.(*organizations.CreateOrgUsrRoleScopesConflict); !ok {
				return nil, err
			}
		}
	} else {
		// Direct to DocuSign flow...

		log.WithFields(f).Debugf("Loading user by username: %s...", lfUsername)
		userModel, userErr := usc.GetUserByUsername(lfUsername)
		if userErr != nil {
			return nil, userErr
		}

		if userModel != nil {
			for _, email := range userModel.Emails {
				if email != nil && *email.IsPrimary {
					currentUserEmail = *email.EmailAddress
				}
			}
		}

		err = prepareUserForSigning(ctx, currentUserEmail, utils.StringValue(input.CompanySfid), utils.StringValue(input.ProjectSfid), input.SigningEntityName)
		if err != nil {
			// Ignore conflict - role has already been assigned
			if _, ok := err.(*organizations.CreateOrgUsrRoleScopesConflict); !ok {
				return nil, err
			}
		}
	}

	signature, err := s.requestCorporateSignature(ctx, s.ClaV4ApiURL, &requestCorporateSignatureInput{
		ProjectID:         proj.ProjectID,
		CompanyID:         comp.CompanyID,
		SigningEntityName: input.SigningEntityName,
		SendAsEmail:       input.SendAsEmail,
		AuthorityName:     input.AuthorityName,
		AuthorityEmail:    input.AuthorityEmail.String(),
		ReturnURL:         input.ReturnURL.String(),
	}, comp, proj, lfUsername, currentUserEmail)

	if err != nil {
		if input.AuthorityEmail.String() != "" {
			// remove role
			removeErr := removeSignatoryRole(ctx, input.AuthorityEmail.String(), utils.StringValue(input.CompanySfid), utils.StringValue(input.ProjectSfid))
			if removeErr != nil {
				log.WithFields(f).WithError(removeErr).Warnf("failed to remove signatory role. companySFID :%s, email :%s error: %+v", *input.CompanySfid, input.AuthorityEmail.String(), removeErr)
			}
		}
		return nil, err
	}

	// Update the company ACL
	log.WithFields(f).Debugf("Adding user with LFID: %s to company access list...", lfUsername)
	companyACLError := s.companyService.AddUserToCompanyAccessList(ctx, comp.CompanyID, lfUsername)
	if companyACLError != nil {
		log.WithFields(f).WithError(companyACLError).Warnf("Unable to add user with LFID: %s to company ACL, companyID: %s", lfUsername, *input.CompanySfid)
	}

	return &models.CorporateSignatureOutput{
		SignURL:     signature.SignatureSignURL,
		SignatureID: signature.SignatureID,
	}, nil
}

func (s *service) getCorporateSignatureCallbackUrl(companyId, projectId string) string {
	return fmt.Sprintf("%s/v4/signed/corporate/%s/%s", s.ClaV4ApiURL, companyId, projectId)
}

func (s *service) SignedIndividualCallbackGithub(ctx context.Context, payload []byte, installationID, changeRequestID, repositoryID string) error {
	f := logrus.Fields{
		"functionName":    "sign.SignedIndividualCallbackGithub",
		utils.XREQUESTID:  ctx.Value(utils.XREQUESTID),
		"installationID":  installationID,
		"changeRequestID": changeRequestID,
		"repositoryID":    repositoryID,
	}

	log.WithFields(f).Debug("processing signed individual callback...")

	var dataModel Payload

	err := json.Unmarshal(payload, &dataModel)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("unable to unmarshall payload")
		return err
	}

	log.WithFields(f).Debugf("webhook payload: %+v", dataModel)

	return nil

}

func (s *service) SignedIndividualCallbackGitlab(ctx context.Context, payload []byte, userID, organizationID, mergeRequestID, repositoryID string) error {
	return nil
}

func (s *service) SignedIndividualCallbackGerrit(ctx context.Context, payload []byte, userID string) error {
	return nil
}

func (s *service) SignedCorporateCallback(ctx context.Context, payload []byte, companyID, projectID string) error {
	return nil
}

func (s *service) RequestIndividualSignature(ctx context.Context, input *models.IndividualSignatureInput, preferredEmail string) (*models.IndividualSignatureOutput, error) {
	f := logrus.Fields{
		"functionName":   "sign.RequestIndividualSignature",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"projectID":      *input.ProjectID,
		"returnURL":      input.ReturnURL,
		"returnURLType":  input.ReturnURLType,
		"userID":         *input.UserID,
	}

	/**
	1. Ensure this is a valid user
	2. Ensure this is a valid project
	3. Check for active signature object with this project. If the user has signed the most recent version they should not be able to sign again.
	4. Generate signature callback url
	5. Get signature return URL
	6. Get latest document
	7. if the CCLA/ICLA template is missing we wont have a document and return an error
	8. Create new signature object
	9. Set signature ACL
	10. Populate sign url
	11. Save signature
	**/

	// 1. Ensure this is a valid user
	log.WithFields(f).Debugf("looking up user by ID: %s", *input.UserID)
	user, err := s.userService.GetUser(*input.UserID)
	if err != nil || user == nil {
		log.WithFields(f).WithError(err).Warnf("unable to lookup user by ID: %s", *input.UserID)
		return nil, err
	}

	// 2. Ensure this is a valid project
	log.WithFields(f).Debugf("looking up project by ID: %s", *input.ProjectID)
	claGroup, err := s.claGroupService.GetCLAGroup(ctx, *input.ProjectID)
	if err != nil || claGroup == nil {
		log.WithFields(f).WithError(err).Warnf("unable to lookup project by ID: %s", *input.ProjectID)
		return nil, err
	}

	// 3. Check for active signature object with this project. If the user has signed the most recent version they should not be able to sign again.
	log.WithFields(f).Debugf("checking for active signature object with this project...")
	approved := true
	signed := true

	userSignatures, err := s.signatureService.GetIndividualSignatures(ctx, *input.ProjectID, *input.UserID, &approved, &signed)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to lookup user signatures by user ID: %s", *input.UserID)
		return nil, err
	}
	latestSignature := getLatestSignature(userSignatures)

	// loading latest document
	log.WithFields(f).Debugf("loading latest individual document for project: %s", *input.ProjectID)
	latestDocument, err := common.GetCurrentDocument(ctx, claGroup.ProjectIndividualDocuments)

	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to lookup latest individual document for project: %s", *input.ProjectID)
		return nil, err
	}

	if common.AreClaGroupDocumentsEqual(latestDocument, v1Models.ClaGroupDocument{}) {
		log.WithFields(f).WithError(err).Warnf("unable to lookup latest individual document for project: %s", *input.ProjectID)
		return nil, errors.New("unable to lookup latest individual document for project")
	}

	// creating individual default values
	log.WithFields(f).Debugf("creating individual default values...")
	defaultValues := s.createDefaultIndividualValues(user, preferredEmail)

	// 4. Generate signature callback url
	log.WithFields(f).Debugf("generating signature callback url...")
	activeSignatureMetadata, err := s.storeRepository.GetActiveSignatureMetaData(ctx, *input.UserID)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to get active signature meta data for user: %s", *input.UserID)
		return nil, err
	}

	log.WithFields(f).Debugf("active signature metadata: %+v", activeSignatureMetadata)

	log.WithFields(f).Debugf("generating signature callback url...")
	var callBackURL string

	if strings.ToLower(input.ReturnURLType) == utils.GitHubType {
		callBackURL, err = s.getIndividualSignatureCallbackURL(ctx, *input.UserID, activeSignatureMetadata)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("unable to get signature callback url for user: %s", *input.UserID)
			return nil, err
		}
	} else if strings.ToLower(input.ReturnURLType) == utils.GitLabLower {
		callBackURL, err = s.getIndividualSignatureCallbackURLGitlab(ctx, *input.UserID, activeSignatureMetadata)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("unable to get signature callback url for user: %s", *input.UserID)
			return nil, err
		}
	}

	log.WithFields(f).Debugf("signature callback url: %s", callBackURL)

	if latestSignature != nil {
		if latestDocument.DocumentMajorVersion == latestSignature.SignatureDocumentMajorVersion {

			log.WithFields(f).Warnf("user: already has a signature with this project: %s", *input.ProjectID)

			// Regenerate and set the signing URL - This will update the signature record
			log.WithFields(f).Debugf("regenerating signing URL for user: %s", *input.UserID)
			_, currentTime := utils.CurrentTime()
			itemSignature := signatures.ItemSignature{
				SignatureID:  latestSignature.SignatureID,
				DateModified: currentTime,
			}
			signURL, signErr := s.populateSignURL(ctx, &itemSignature, callBackURL, "", "", false, "", "", defaultValues, preferredEmail)
			if signErr != nil {
				log.WithFields(f).WithError(err).Warnf("unable to populate sign url for user: %s", *input.UserID)
				return nil, signErr
			}

			return &models.IndividualSignatureOutput{
				SignURL:     signURL,
				SignatureID: latestSignature.SignatureID,
				UserID:      latestSignature.SignatureReferenceID,
				ProjectID:   *input.ProjectID,
			}, nil
		}
	}

	// 5. Get signature return URL
	log.WithFields(f).Debugf("getting signature return url...")
	var returnURL string
	if input.ReturnURL.String() == "" {
		log.WithFields(f).Warnf("signature return url is empty")
		returnURL, err = getActiveSignatureReturnURL(*input.UserID, activeSignatureMetadata)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("unable to get active signature return url for user: %s", *input.UserID)
			return nil, err
		}
		if returnURL == "" {
			log.WithFields(f).Warnf("signature return url is empty")
			return &models.IndividualSignatureOutput{
				UserID:    *input.UserID,
				ProjectID: *input.ProjectID,
			}, nil
		}

	}

	// 6. Get latest document
	log.WithFields(f).Debugf("getting latest document...")
	document, err := common.GetCurrentDocument(ctx, claGroup.ProjectIndividualDocuments)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to get latest document for project: %s", *input.ProjectID)
		return nil, err
	}

	// 7. if the CCLA/ICLA template is missing we wont have a document and return an error
	if common.AreClaGroupDocumentsEqual(document, v1Models.ClaGroupDocument{}) {
		log.WithFields(f).WithError(err).Warnf("unable to get latest document for project: %s", *input.ProjectID)
		return nil, errors.New("unable to get latest document for project")
	}

	// 8. Create new signature object
	log.WithFields(f).Debugf("creating new signature object...")
	signatureID := uuid.Must(uuid.NewV4()).String()
	_, currentTime := utils.CurrentTime()
	var acl string
	if input.ReturnURLType == "github" {
		acl = fmt.Sprintf("%s:%s", strings.ToLower(input.ReturnURLType), user.GithubID)
	} else if input.ReturnURLType == "gitlab" {
		acl = fmt.Sprintf("%s:%s", strings.ToLower(input.ReturnURLType), user.GitlabID)
	}

	majorVersion, err := strconv.Atoi(document.DocumentMajorVersion)

	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to convert document major version to int: %s", document.DocumentMajorVersion)
		return nil, err
	}

	minorVersion, err := strconv.Atoi(document.DocumentMinorVersion)

	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to convert document minor version to int: %s", document.DocumentMinorVersion)
		return nil, err
	}

	itemSignature := signatures.ItemSignature{
		SignatureID:                   signatureID,
		DateCreated:                   currentTime,
		DateModified:                  currentTime,
		SignatureSigned:               false,
		SignatureApproved:             true,
		SignatureDocumentMajorVersion: majorVersion,
		SignatureDocumentMinorVersion: minorVersion,
		SignatureReferenceID:          *input.UserID,
		SignatureReferenceName:        getUserName(user),
		SignatureType:                 utils.SignatureTypeCLA,
		SignatureReturnURLType:        input.ReturnURLType,
		SignatureProjectID:            *input.ProjectID,
		SignatureReturnURL:            input.ReturnURL.String(),
		SignatureCallbackURL:          callBackURL,
		SignatureReferenceType:        "user",
		SignatureACL:                  []string{acl},
		SigtypeSignedApprovedID:       fmt.Sprintf("%s#%v#%v#%s", utils.ClaTypeICLA, signed, approved, signatureID),
		SignatureUserCompanyID:        user.CompanyID,
		SignatureReferenceNameLower:   strings.ToLower(getUserName(user)),
	}

	// 10. Populate sign url
	log.WithFields(f).Debugf("populating sign url...")
	_, err = s.populateSignURL(ctx, &itemSignature, callBackURL, "", "", false, "", "", defaultValues, preferredEmail)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to populate sign url for user: %s", *input.UserID)
		return nil, err
	}

	log.WithFields(f).Debugf("Updated signature: %+v", itemSignature)

	return &models.IndividualSignatureOutput{
		UserID:      itemSignature.SignatureReferenceID,
		ProjectID:   itemSignature.SignatureProjectID,
		SignatureID: itemSignature.SignatureID,
		SignURL:     itemSignature.SignatureSignURL,
	}, nil
}

func getUserName(user *v1Models.User) string {

	if user.Username != "" {
		return user.Username
	}
	if user.LfUsername != "" {
		return user.LfUsername
	}

	if user.GithubUsername != "" {
		return user.GithubUsername
	}
	if user.GitlabUsername != "" {
		return user.GitlabUsername
	}
	return ""
}

func (s *service) getIndividualSignatureCallbackURLGitlab(ctx context.Context, userID string, metadata map[string]interface{}) (string, error) {
	f := logrus.Fields{
		"functionName": "sign.getIndividualSignatureCallbackURLGitlab",
		"userID":       userID,
	}

	log.WithFields(f).Debugf("generating signature callback url...")
	var err error
	var repositoryID string
	var mergeRequestID string

	if metadata == nil {
		metadata, err = s.storeRepository.GetActiveSignatureMetaData(ctx, userID)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("unable to get active signature meta data for user: %s", userID)
			return "", err
		}
	}

	if found, ok := metadata["repository_id"].(string); ok {
		repositoryID = found
	} else {
		log.WithFields(f).WithError(err).Warnf("unable to get repository ID for user: %s", userID)
		return "", err
	}

	if found, ok := metadata["merge_request_id"].(string); ok {
		mergeRequestID = found
	} else {
		log.WithFields(f).WithError(err).Warnf("unable to get pull request ID for user: %s", userID)
		return "", err
	}

	gitlabOrg, err := s.gitlabOrgService.GetGitLabOrganization(ctx, repositoryID)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to get organization ID for repository ID: %s", repositoryID)
		return "", err
	}

	if gitlabOrg.OrganizationID == "" {
		log.WithFields(f).WithError(err).Warnf("unable to get organization ID for repository ID: %s", repositoryID)
		return "", err
	}

	return fmt.Sprintf("%s/v4/signed/gitlab/individual/%s/%s/%s/%s", s.ClaV4ApiURL, userID, gitlabOrg.OrganizationID, repositoryID, mergeRequestID), nil

}

func (s *service) getIndividualSignatureCallbackURL(ctx context.Context, userID string, metadata map[string]interface{}) (string, error) {
	f := logrus.Fields{
		"functionName": "sign.getIndividualSignatureCallbackURL",
		"userID":       userID,
	}

	log.WithFields(f).Debugf("generating signature callback url...")
	var err error
	var installationId int64
	var repositoryID string
	var pullRequestID string

	if metadata == nil {
		metadata, err = s.storeRepository.GetActiveSignatureMetaData(ctx, userID)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("unable to get active signature meta data for user: %s", userID)
			return "", err
		}
	}

	if found, ok := metadata["repository_id"].(string); ok {
		repositoryID = found
	} else {
		log.WithFields(f).WithError(err).Warnf("unable to get repository ID for user: %s", userID)
		return "", err
	}

	if found, ok := metadata["pull_request_id"].(string); ok {
		pullRequestID = found
	} else {
		log.WithFields(f).WithError(err).Warnf("unable to get pull request ID for user: %s", userID)
		return "", err
	}

	// Get installation ID through a helper function
	log.WithFields(f).Debugf("getting repository...")
	githubRepository, err := s.repositoryService.GetRepositoryByExternalID(ctx, repositoryID)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to get installation ID for repository ID: %s", repositoryID)
		return "", err
	}
	// Get github organization
	log.WithFields(f).Debugf("getting github organization...")
	githubOrg, err := s.githubOrgService.GetGitHubOrganizationByName(ctx, githubRepository.RepositoryOrganizationName)

	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to get github organization for repository ID: %s", repositoryID)
		return "", err
	}

	installationId = githubOrg.OrganizationInstallationID
	if installationId == 0 {
		log.WithFields(f).WithError(err).Warnf("unable to get installation ID for repository ID: %s", repositoryID)
		return "", err
	}

	callbackURL := fmt.Sprintf("%s/v4/signed/individual/%d/%s/%s", s.ClaV4ApiURL, installationId, repositoryID, pullRequestID)

	log.WithFields(f).Debugf("return url: %s", callbackURL)

	return callbackURL, nil

}

//nolint:gocyclo
func (s *service) populateSignURL(ctx context.Context,
	latestSignature *signatures.ItemSignature, callbackURL string,
	authorityOrSignatoryName, authorityOrSignatoryEmail string,
	sendAsEmail bool,
	claManagerName, claManagerEmail string,
	defaultValues map[string]interface{}, preferredEmail string) (string, error) {

	f := logrus.Fields{
		"functionName":              "sign.populateSignURL",
		"authorityOrSignatoryName":  authorityOrSignatoryName,
		"authorityOrSignatoryEmail": authorityOrSignatoryEmail,
		"preferredEmail":            preferredEmail,
	}
	log.WithFields(f).Debugf("populating sign url...")
	signatureReferenceType := latestSignature.SignatureReferenceType

	log.WithFields(f).Debugf("signatureReferenceType: %s", signatureReferenceType)
	log.WithFields(f).Debugf("processing signing request...")

	var userSignatureName string
	var userSignatureEmail string
	var document v1Models.ClaGroupDocument
	var project *v1Models.ClaGroup
	var companyModel *v1Models.Company
	var err error
	var signer DocuSignRecipient
	var emailBody string
	var emailSubject string

	// populate user details
	userDetails, err := s.populateUserDetails(ctx, signatureReferenceType, latestSignature, claManagerName, claManagerEmail, sendAsEmail, preferredEmail)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to populate user details for signatureReferenceType: %s", signatureReferenceType)
		return "", err
	}

	userSignatureName = userDetails.userSignatureName
	userSignatureEmail = userDetails.userSignatureEmail

	log.WithFields(f).Debugf("userSignatureName: %s, userSignatureEmail: %s", userSignatureName, userSignatureEmail)

	// Get the document template to sign
	log.WithFields(f).Debugf("getting document template to sign...")
	project, err = s.projectRepo.GetCLAGroupByID(ctx, latestSignature.SignatureProjectID, DontLoadRepoDetails)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to lookup project by ID: %s", latestSignature.SignatureProjectID)
		return "", err
	}

	if project == nil {
		log.WithFields(f).WithError(err).Warnf("unable to lookup project by ID: %s", latestSignature.SignatureProjectID)
		return "", errors.New("no project lookup error")
	}

	if signatureReferenceType == utils.SignatureReferenceTypeCompany {
		log.WithFields(f).Debugf("loading project corporate document...")
		document, err = common.GetCurrentDocument(ctx, project.ProjectCorporateDocuments)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("unable to lookup project corporate document for project: %s", latestSignature.SignatureProjectID)
			return "", err
		}
	} else {
		log.WithFields(f).Debugf("loading project individual document...")
		document, err = common.GetCurrentDocument(ctx, project.ProjectIndividualDocuments)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("unable to lookup project individual document for project: %s", latestSignature.SignatureProjectID)
			return "", err
		}
	}

	// Void the existing envelope to prevent multiple envelopes pending for a signer
	envelopeID := latestSignature.SignatureEnvelopeID
	if envelopeID != "" {
		message := fmt.Sprintf("You are getting this message because your DocuSign Session for project %s expired. A new session will be in place for your signing process.", project.ProjectName)
		log.WithFields(f).Debug(message)
		err = s.VoidEnvelope(ctx, envelopeID, message)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("DocuSign error while voiding the envelope - regardless, continuing on..., error: %s", err)
		}
	}

	// create a new source and rand object
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src) //nolint:gosec

	randomInteger := r.Intn(1000000) //nolint:gosec
	documentID := strconv.Itoa(randomInteger)

	tab := getTabsFromDocument(&document, documentID, defaultValues)

	// # Create the envelope request object

	if sendAsEmail {
		log.WithFields(f).Warnf("assigning signatory name/email: %s/%s", authorityOrSignatoryName, authorityOrSignatoryEmail)
		signatoryEmail := authorityOrSignatoryEmail
		signatoryName := authorityOrSignatoryName

		var projectName string
		var companyName string

		if project != nil {
			projectName = project.ProjectName
		}

		if companyModel != nil {
			companyName = companyModel.CompanyName
		}

		pcgs, pcgErr := s.projectClaGroupsRepo.GetProjectsIdsForClaGroup(ctx, project.ProjectID)
		if pcgErr != nil {
			log.WithFields(f).Debugf("problem fetching project cla groups by id :%s, err: %+v", project.ProjectID, pcgErr)
			return "", pcgErr
		}

		if len(pcgs) == 0 {
			log.WithFields(f).Debugf("no project cla groups found for project id :%s", project.ProjectID)
			return "", errors.New("no project cla groups found for project id")
		}

		var projectNames []string
		for _, pcg := range pcgs {
			projectNames = append(projectNames, pcg.ProjectName)
		}

		if len(projectNames) == 0 {
			projectNames = []string{projectName}
		}

		claSignatoryParams := &ClaSignatoryEmailParams{
			ClaGroupName:    project.ProjectName,
			SignatoryName:   signatoryName,
			CompanyName:     companyName,
			ProjectNames:    projectNames,
			ProjectVersion:  project.Version,
			ClaManagerName:  claManagerName,
			ClaManagerEmail: claManagerEmail,
		}

		log.WithFields(f).Debugf("claSignatoryParams: %+v", claSignatoryParams)
		emailSubject, emailBody = claSignatoryEmailContent(*claSignatoryParams)
		log.WithFields(f).Debugf("subject: %s, body: %s", emailSubject, emailBody)

		signer = DocuSignRecipient{
			Email:       signatoryEmail,
			Name:        signatoryName,
			Tabs:        tab,
			RecipientId: "1",
			RoleName:    "signer",
		}

	} else {
		// This will be the Initial CLA Manager
		signatoryName := userSignatureName
		signatoryEmail := userSignatureEmail

		// Assigning a clientUserId does not send an email.
		// It assumes that the user handles the communication with the client.
		// In this case, the user opened the docusign document to manually sign it.
		// Thus the email does not need to be sent.

		log.WithFields(f).Debugf("signatoryName: %s, signatoryEmail: %s", signatoryName, signatoryEmail)

		// # Max length for emailSubject is 100 characters - guard/truncate if necessary
		emailSubject = fmt.Sprintf("EasyCLA: CLA Signature Request for %s", project.ProjectName)
		if len(emailSubject) > 100 {
			emailSubject = emailSubject[:97] + "..."
		}

		// # Update Signed for label according to signature_type (company or name)
		var userIdentifier string
		if signatureReferenceType == utils.SignatureReferenceTypeCompany && companyModel != nil {
			userIdentifier = companyModel.CompanyName
		} else {
			if signatoryName == "Unknown" || signatoryName == "" {
				userIdentifier = signatoryEmail
			} else {
				userIdentifier = signatoryName
			}
		}

		log.WithFields(f).Debugf("userIdentifier: %s", userIdentifier)

		emailBody = fmt.Sprintf("CLA Sign Request for %s", userIdentifier)

		signer = DocuSignRecipient{
			Email:        signatoryEmail,
			Name:         signatoryName,
			Tabs:         tab,
			RecipientId:  "1",
			ClientUserId: latestSignature.SignatureID,
			RoleName:     "signer",
		}
	}

	contentType := document.DocumentContentType
	var pdf []byte

	if document.DocumentS3URL != "" {
		log.WithFields(f).Debugf("getting document resource from s3: %s...", document.DocumentS3URL)
		pdf, err = s.getDocumentResource(document.DocumentS3URL)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("unable to get document resource from s3 for document: %s", document.DocumentS3URL)
			return "", err
		}
	} else if strings.HasPrefix(contentType, "url+") {
		log.WithFields(f).Debugf("getting document resource from url...")
		pdfURL := document.DocumentContent
		pdf, err = s.getDocumentResource(pdfURL)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("unable to get document resource from url: %s", pdfURL)
			return "", err
		}
	} else {
		log.WithFields(f).Debugf("getting document resource from content...")
		content := document.DocumentContent
		pdf = []byte(content)
	}

	documentName := document.DocumentName
	log.WithFields(f).Debugf("documentName: %s", documentName)
	log.WithFields(f).Debugf("contentType: %s", contentType)

	docusignDocument := DocuSignDocument{
		Name:           documentName,
		DocumentId:     documentID,
		FileExtension:  "pdf",
		FileFormatHint: "pdf",
		Order:          "1",
		DocumentBase64: base64.StdEncoding.EncodeToString(pdf),
	}

	var envelopeRequest DocuSignEnvelopeRequest

	if callbackURL != "" {
		// Webhook properties for callbacks after the user signs the document.
		// Ensure that a webhook is returned on the status "Completed" where
		// all signers on a document finish signing the document.
		log.WithFields(f).Debugf("setting up webhook properties with callback url: %s", callbackURL)
		recipientEvents := []DocuSignRecipientEvent{
			{
				EnvelopeEventStatusCode: "Completed",
			},
		}

		eventNotification := DocuSignEventNotification{
			URL:            callbackURL,
			LoggingEnabled: true,
			EnvelopeEvents: recipientEvents,
			UseSoapInterface: "true",
			IncludeDocuments: "true",
		}

		envelopeRequest = DocuSignEnvelopeRequest{
			Documents: []DocuSignDocument{
				docusignDocument,
			},
			EmailSubject:      emailSubject,
			EmailBlurb:        emailBody,
			EventNotification: eventNotification,
			Status:            "sent",
			Recipients: DocuSignRecipientType{
				Signers: []DocuSignRecipient{
					signer,
				},
			},
		}

	} else {

		envelopeRequest = DocuSignEnvelopeRequest{
			Documents: []DocuSignDocument{
				docusignDocument,
			},
			EmailSubject: emailSubject,
			EmailBlurb:   emailBody,
			Status:       "sent",
			Recipients: DocuSignRecipientType{
				Signers: []DocuSignRecipient{
					signer,
				},
			},
		}

	}

	envelopeResponse, err := s.PrepareSignRequest(ctx, &envelopeRequest)

	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to create envelope for user: %s", latestSignature.SignatureReferenceID)
		return "", err
	}

	log.WithFields(f).Debugf("envelopeID: %s", envelopeResponse.EnvelopeId)
	var signatureSignURL *string

	if !sendAsEmail {
		// The URL the user will be redirected to after signing.
		// This route will be in charge of extracting the signature's return_url and redirecting.
		recipients, recipientErr := s.getEnvelopeRecipients(ctx, envelopeResponse.EnvelopeId)
		if recipientErr != nil {
			log.WithFields(f).Debugf("unable to fetch recipients for envelope: %s", envelopeResponse.EnvelopeId)
			return "", recipientErr
		}

		if len(recipients) == 0 {
			log.WithFields(f).Debugf("no envelope recipients found : %s", envelopeResponse.EnvelopeId)
			return "", errors.New("no envelope recipients found")
		}
		recipient := recipients[0]
		returnURL := fmt.Sprintf("%s/v2/return-url/%s", s.ClaV1ApiURL, recipient.ClientUserId)

		log.WithFields(f).Debugf("generating signature sign_url, using return-url as: %s", returnURL)
		signURL, signErr := s.GetSignURL(signer.Email, signer.RecipientId, signer.Name, signer.ClientUserId, envelopeResponse.EnvelopeId, returnURL)

		if signErr != nil {
			log.WithFields(f).WithError(err).Warnf("unable to get sign url for user: %s", latestSignature.SignatureReferenceID)
			return "", signErr
		}

		log.WithFields(f).Debugf("setting signature sign_url as: %s", signURL)
		signatureSignURL = &signURL
	}

	// Save Envelope ID in signature.
	log.WithFields(f).Debugf("saving signature to database...")
	latestSignature.SignatureEnvelopeID = envelopeResponse.EnvelopeId
	latestSignature.SignatureSignURL = *signatureSignURL

	log.WithFields(f).Debugf("signature: %+v", latestSignature)

	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to save signature to database for user: %s", latestSignature.SignatureID)
		return "", err
	}

	err = s.signatureService.CreateSignature(ctx, latestSignature)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to save signature to database for user: %s", latestSignature.SignatureID)
		return "", err
	}

	log.WithFields(f).Debug("signature saved to database")

	log.WithFields(f).Debugf("populate_sign_url - complete: %s", *signatureSignURL)

	return *signatureSignURL, nil
}

type UserSignDetails struct {
	userSignatureName  string
	userSignatureEmail string
}

func (s *service) populateUserDetails(ctx context.Context, signatureReferenceType string, latestSignature *signatures.ItemSignature, claManagerName, claManagerEmail string, sendAsEmail bool, preferredEmail string) (*UserSignDetails, error) {
	f := logrus.Fields{
		"functionName": "sign.populateUserDetails",
	}
	log.WithFields(f).Debugf("populating user details...")
	userSignDetails := &UserSignDetails{
		userSignatureName:  Unknown,
		userSignatureEmail: Unknown,
	}

	if signatureReferenceType == utils.SignatureReferenceTypeCompany {
		companyModel, err := s.companyRepo.GetCompany(ctx, latestSignature.SignatureReferenceID)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("unable to lookup company by ID: %s", latestSignature.SignatureReferenceID)
			return nil, err
		}
		if companyModel == nil {
			log.WithFields(f).WithError(err).Warnf("unable to lookup company by ID: %s", latestSignature.SignatureReferenceID)
			return nil, errors.New("no CLA manager lookup error")
		}
		userSignDetails.userSignatureEmail = claManagerEmail
		userSignDetails.userSignatureName = claManagerName

	} else if signatureReferenceType == utils.SignatureReferenceTypeUser {
		if !sendAsEmail {
			userModel, userErr := s.userService.GetUser(latestSignature.SignatureReferenceID)
			if userErr != nil {
				log.WithFields(f).WithError(userErr).Warnf("unable to lookup user by ID: %s", latestSignature.SignatureReferenceID)
				return nil, userErr
			}
			log.WithFields(f).Debugf("loaded user : %+v", userModel)

			if userModel == nil {
				log.WithFields(f).WithError(userErr).Warnf("unable to lookup user by ID: %s", latestSignature.SignatureReferenceID)
				msg := fmt.Sprintf("No user lookup error for user ID: %s", latestSignature.SignatureReferenceID)
				return nil, errors.New(msg)
			}

			if userModel.Username != "" {
				userSignDetails.userSignatureName = userModel.Username
			}
			if getUserEmail(userModel, preferredEmail) != "" {
				userSignDetails.userSignatureEmail = getUserEmail(userModel, preferredEmail)
			}
		}
	} else {
		log.WithFields(f).Warnf("unknown signature reference type: %s", signatureReferenceType)
		return nil, errors.New("unknown signature reference type")
	}
	return userSignDetails, nil
}

func (s *service) getDocumentResource(urlString string) ([]byte, error) {

	// validate the URL
	u, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warnf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get document resource from url: %s, status code: %d", urlString, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// Helper function to extract the docusign tabs from the document
func getTabsFromDocument(document *v1Models.ClaGroupDocument, documentID string, defaultValues map[string]interface{}) DocuSignTab {
	var docTab DocuSignTab
	f := logrus.Fields{
		"functionName": "sign.getTabsFromDocument",
		"documentID":   documentID,
	}
	log.WithFields(f).Debugf("getting tabs from document...")
	for _, tab := range document.DocumentTabs {
		var args DocuSignTabDetails
		args.DocumentId = documentID
		args.PageNumber = strconv.FormatInt(tab.DocumentTabPage, 10)
		args.XPosition = strconv.FormatInt(tab.DocumentTabPositionx, 10)
		args.YPosition = strconv.FormatInt(tab.DocumentTabPositiony, 10)
		args.Width = strconv.FormatInt(tab.DocumentTabWidth, 10)
		args.Height = strconv.FormatInt(tab.DocumentTabHeight, 10)
		args.CustomTabId = tab.DocumentTabID
		args.TabLabel = tab.DocumentTabID
		args.Name = tab.DocumentTabName

		if tab.DocumentTabAnchorString != "" {
			args.AnchorString = tab.DocumentTabAnchorString
			args.AnchorIgnoreIfNotPresent = strconv.FormatBool(tab.DocumentTabAnchorIgnoreIfNotPresent)
			args.AnchorXOffset = strconv.FormatInt(tab.DocumentTabAnchorxOffset, 10)
			args.AnchorYOffset = strconv.FormatInt(tab.DocumentTabAnchoryOffset, 10)
		}

		if defaultValues != nil {
			if value, ok := defaultValues[tab.DocumentTabID].(string); ok {
				args.Value = value
			}
		}

		switch tab.DocumentTabType {
		case "text":
			docTab.TextTabs = append(docTab.TextTabs, args)
		case "text_unlocked":
			args.Locked = DocSignFalse
			docTab.TextTabs = append(docTab.TextTabs, args)
		case "text_optional":
			args.Required = DocSignFalse
			docTab.TextOptionalTabs = append(docTab.TextOptionalTabs, args)
		case "number":
			docTab.NumberTabs = append(docTab.NumberTabs, args)
		case "sign":
			docTab.SignHereTabs = append(docTab.SignHereTabs, args)
		case "sign_optional":
			args.Optional = "true"
			docTab.SignHereOptionalTabs = append(docTab.SignHereOptionalTabs, args)
		case "date":
			docTab.DateSignedTabs = append(docTab.DateSignedTabs, args)
		default:
			log.WithFields(f).Warnf("unknown document tab type: %s", tab.DocumentTabType)
			continue
		}
	}

	return docTab
}

// helper function to get user email
func getUserEmail(user *v1Models.User, preferredEmail string) string {
	if preferredEmail != "" {
		if utils.StringInSlice(preferredEmail, user.Emails) || user.LfEmail == strfmt.Email(preferredEmail) {
			return preferredEmail
		}
	}
	if user.LfEmail != "" {
		return string(user.LfEmail)
	}
	if len(user.Emails) > 0 {
		return user.Emails[0]
	}
	return ""
}

func getActiveSignatureReturnURL(userID string, metadata map[string]interface{}) (string, error) {

	f := logrus.Fields{
		"functionName": "sign.getActiveSignatureReturnURL",
	}

	var returnURL string
	var err error
	var pullRequestID int
	var installationID int64
	var repositoryID int64

	if found, ok := metadata["pull_request_id"].(int); ok {
		pullRequestID = found
	} else {
		log.WithFields(f).WithError(err).Warnf("unable to get pull request ID for user: %s", userID)
		return "", err
	}

	if found, ok := metadata["installation_id"].(int64); ok {
		installationID = found
	} else {
		log.WithFields(f).WithError(err).Warnf("unable to get installation ID for user: %s", userID)
		return "", err
	}

	if found, ok := metadata["repository_id"].(int64); ok {
		repositoryID = found
	} else {
		log.WithFields(f).WithError(err).Warnf("unable to get repository ID for user: %s", userID)
		return "", err
	}

	returnURL, err = github.GetReturnURL(context.Background(), installationID, repositoryID, pullRequestID)

	if err != nil {
		return "", err
	}

	return returnURL, nil

}

func (s *service) createDefaultIndividualValues(user *v1Models.User, preferredEmail string) map[string]interface{} {
	f := logrus.Fields{
		"functionName": "sign.createDefaultIndiviualValues",
	}
	log.WithFields(f).Debugf("creating individual default values...")

	defaultValues := make(map[string]interface{})

	if user != nil {
		if user.Username != "" {
			defaultValues["user_name"] = user.Username
			defaultValues["public_name"] = user.Username
		}
	}

	if preferredEmail != "" {
		if utils.StringInSlice(preferredEmail, user.Emails) || user.LfEmail == strfmt.Email(preferredEmail) {
			defaultValues["user_email"] = preferredEmail
		}
	}

	return defaultValues
}

func (s *service) createDefaultCorporateValues(company *v1Models.Company, signatoryName string, signatoryEmail string, managerName string, managerEmail string) map[string]interface{} {
	f := logrus.Fields{
		"functionName": "sign.createDefaultCorporateValues",
	}
	log.WithFields(f).Debugf("creating corporate default values...")

	defaultValues := make(map[string]interface{})

	if company != nil {
		if company.CompanyName != "" {
			defaultValues["corporation"] = company.CompanyName
		}
		if company.SigningEntityName != "" {
			defaultValues["corporation_name"] = company.SigningEntityName
		} else {
			defaultValues["corporation_name"] = company.CompanyName
		}
	}
	if signatoryName != "" {
		defaultValues["signatory_name"] = signatoryName
	}
	if signatoryEmail != "" {
		defaultValues["signatory_email"] = signatoryEmail
	}

	if managerName != "" {
		defaultValues["point_of_contact"] = managerName
		defaultValues["cla_manager_name"] = managerName
	}

	if managerEmail != "" {
		defaultValues["email"] = managerEmail
		defaultValues["cla_manager_email"] = managerEmail
	}

	if signatoryName != "" && signatoryEmail != "" {
		defaultValues["scheduleA"] = fmt.Sprintf("CLA Manager: %s, %s", signatoryName, signatoryEmail)
	}

	return defaultValues
}

func getLatestSignature(signatures []*v1Models.Signature) *v1Models.Signature {
	var latestSignature *v1Models.Signature
	for _, signature := range signatures {
		if latestSignature == nil {
			latestSignature = signature
		} else {
			if signature.SignatureMajorVersion > latestSignature.SignatureMajorVersion {
				latestSignature = signature
			} else if signature.SignatureMajorVersion == latestSignature.SignatureMajorVersion {
				if signature.SignatureMinorVersion > latestSignature.SignatureMinorVersion {
					latestSignature = signature
				}
			}
		}
	}
	return latestSignature
}

func (s *service) RequestIndividualSignatureGerrit(ctx context.Context, input *models.IndividualSignatureInput) (*models.IndividualSignatureOutput, error) {
	return nil, nil
}

func (s *service) requestCorporateSignature(ctx context.Context, apiURL string, input *requestCorporateSignatureInput, comp *v1Models.Company, proj *v1Models.ClaGroup, lfUsername string, currentUserEmail string) (*v1Models.Signature, error) {
	f := logrus.Fields{
		"functionName":      "requestCorporateSignature",
		"apiURL":            apiURL,
		"CompanyID":         input.CompanyID,
		"ProjectID":         input.ProjectID,
		"SigningEntityName": input.SigningEntityName,
		"AuthorityName":     input.AuthorityName,
		"AuthorityEmail":    input.AuthorityEmail,
		"ReturnURL":         input.ReturnURL,
		"SendAsEmail":       input.SendAsEmail,
	}
	/**
		1. Ensure User exists in easycla db, if not then create one by getting user by user service
	   	2. Create individual default values
		3. Load latest document
		4. Check for active corporate signature record for this project/company combination
		5. if signature doesn't exists then Create new signature object
		6. Set signature ACL
		7. Populate sign url
		8. Save signature
	**/
	// 1. Ensure User exists in easycla db, if not then create one by getting user by user service
	usc := userService.GetClient()
	log.WithFields(f).Debugf("Get UserProfile from easycla: %s...", lfUsername)
	claUser, err := s.userService.GetUserByUserName(lfUsername, true)
	if err != nil {
		return nil, err
	}
	if claUser == nil {
		log.WithFields(f).Debugf("Loading user by username from username: %s...", lfUsername)
		userModel, userErr := usc.GetUserByUsername(lfUsername)
		if userErr != nil {
			return nil, userErr
		}
		var lfEmail string
		var emailList []string
		emails := userModel.Emails
		if len(emails) > 0 {
			for _, email := range emails {
				if *email.IsPrimary {
					lfEmail = *email.EmailAddress
				}
				emailList = append(emailList, *email.EmailAddress)
			}
		}

		claUser, err = s.userService.CreateUser(&v1Models.User{
			Username:       userModel.Name,
			UserExternalID: userModel.ID,
			LfUsername:     lfUsername,
			Admin:          false,
			LfEmail:        strfmt.Email(lfEmail),
			Emails:         emailList,
		}, nil)
		if err != nil {
			return nil, err
		}
	}
	signatoryName := input.AuthorityName
	signatoryEmail := input.AuthorityEmail

	if input.AuthorityName == "" || input.AuthorityEmail == "" {
		signatoryName = claUser.Username
		signatoryEmail = currentUserEmail
	}

	// 2. Create individual default values
	log.WithFields(f).Debugf("creating individual default values...")
	defaultValues := s.createDefaultCorporateValues(comp, signatoryName, signatoryEmail, claUser.Username, currentUserEmail)

	// 3. Load latest document
	log.WithFields(f).Debugf("loading latest individual document for project: %s", input.ProjectID)
	latestDocument, err := common.GetCurrentDocument(ctx, proj.ProjectCorporateDocuments)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to lookup latest corporate document for project: %s", input.ProjectID)
		return nil, err
	}

	if common.AreClaGroupDocumentsEqual(latestDocument, v1Models.ClaGroupDocument{}) {
		log.WithFields(f).WithError(err).Warnf("unable to lookup latest corporate document for project: %s", input.ProjectID)
		return nil, errors.New("unable to lookup latest corporate document for project")
	}

	// 4. Check for active corporate signature record for this project/company combination
	approved := true
	log.WithFields(f).Debug("Forwarding request to v1 API for requestCorporateSignature...")
	companySignatures, err := s.signatureService.GetCorporateSignatures(ctx, input.ProjectID, input.CompanyID, &approved, nil)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to lookup user signatures by Company ID: %s, Project ID: %s", input.CompanyID, input.ProjectID)
		return nil, err
	}

	haveSigned := false
	for _, s := range companySignatures {
		if s.SignatureSigned {
			haveSigned = true
			break
		}
	}
	if haveSigned {
		haveSignedErr := fmt.Errorf("one or more corporate valid signature exists for Company ID: %s, Project ID: %s", input.CompanyID, input.ProjectID)
		log.WithFields(f).WithError(err).Warnf(haveSignedErr.Error())
		return nil, haveSignedErr
	}
	callbackURL := s.getCorporateSignatureCallbackUrl(input.ProjectID, input.CompanyID)
	var companySignature *v1Models.Signature
	var itemSignature *signatures.ItemSignature
	var signed bool
	if len(companySignatures) > 0 {
		companySignature = companySignatures[0]
		itemSignature = &signatures.ItemSignature{
			SignatureID:  companySignature.SignatureID,
			DateModified: companySignature.Modified,
		}
		signed = companySignature.SignatureSigned
		approved = companySignature.SignatureApproved
	} else {
		// 5. if signature doesn't exists then Create new signature object
		log.WithFields(f).Debugf("creating new signature object...")
		signatureID := uuid.Must(uuid.NewV4()).String()
		_, currentTime := utils.CurrentTime()
		signed = false
		approved = true
		majorVersion, majorErr := strconv.Atoi(latestDocument.DocumentMajorVersion)
		if majorErr != nil {
			log.WithFields(f).WithError(err).Warnf("unable to convert document major version to int: %s", latestDocument.DocumentMajorVersion)
			return nil, majorErr
		}
		minorVersion, minorErr := strconv.Atoi(latestDocument.DocumentMinorVersion)
		if minorErr != nil {
			log.WithFields(f).WithError(err).Warnf("unable to convert document minor version to int: %s", latestDocument.DocumentMinorVersion)
			return nil, minorErr
		}
		itemSignature = &signatures.ItemSignature{
			SignatureID:                   signatureID,
			SignatureDocumentMajorVersion: majorVersion,
			SignatureDocumentMinorVersion: minorVersion,
			SignatureReferenceID:          comp.CompanyID,
			SignatureReferenceType:        "company",
			SignatureReferenceName:        comp.CompanyName,
			SignatureProjectID:            input.ProjectID,
			DateCreated:                   currentTime,
			DateModified:                  currentTime,
			SignatureType:                 utils.SignatureTypeCCLA,
			SignatoryName:                 signatoryName,
			SignatureSigned:               false,
			SignatureApproved:             true,
			SigtypeSignedApprovedID:       fmt.Sprintf("%s#%v#%v#%s", utils.SignatureTypeCCLA, signed, approved, signatureID),
			SignatureReferenceNameLower:   strings.ToLower(comp.CompanyName),
		}

	}
	companySignature.SignatureCallbackURL = callbackURL

	if !input.SendAsEmail {
		companySignature.SignatureReturnURL = input.ReturnURL
	}

	// 6. Set signature ACL
	log.WithFields(f).Debugf("setting signature ACL...")
	companySignature.SignatureACL = []v1Models.User{
		*claUser,
	}

	// 7. Populate sign url
	log.WithFields(f).Debugf("populating sign url...")
	_, err = s.populateSignURL(ctx, itemSignature, callbackURL, input.AuthorityName, input.AuthorityEmail, input.SendAsEmail, claUser.Username, currentUserEmail, defaultValues, currentUserEmail)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to populate sign url for company: %s", input.CompanyID)
		return nil, err
	}

	companySignature, err = s.signatureService.GetCorporateSignature(ctx, input.ProjectID, input.CompanyID, &approved, &signed)

	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to lookup user signatures by Company ID: %s, Project ID: %s", input.CompanyID, input.ProjectID)
		return nil, err
	}

	return companySignature, nil
}

func removeSignatoryRole(ctx context.Context, userEmail string, companySFID string, projectSFID string) error {
	f := logrus.Fields{"functionName": "removeSignatoryRole", "user_email": userEmail, "company_sfid": companySFID, "project_sfid": projectSFID}
	log.WithFields(f).Debug("removing role for user")

	usc := userService.GetClient()
	// search user
	log.WithFields(f).Debug("searching user by email")
	user, err := usc.SearchUserByEmail(userEmail)
	if err != nil {
		log.WithFields(f).Debug("Failed to get user")
		return err
	}

	log.WithFields(f).Debug("Getting role id")
	acsClient := acsService.GetClient()
	roleID, roleErr := acsClient.GetRoleID("cla-signatory")
	if roleErr != nil {
		log.WithFields(f).Debug("Failed to get role id for cla-signatory")
		return roleErr
	}
	// Get scope id
	log.WithFields(f).Debug("getting scope id")
	orgClient := organizationService.GetClient()
	scopeID, scopeErr := orgClient.GetScopeID(ctx, companySFID, projectSFID, "cla-signatory", "project|organization", user.Username)

	if scopeErr != nil {
		log.WithFields(f).Debug("Failed to get scope id for cla-signatory role")
		return scopeErr
	}

	//Unassign role
	log.WithFields(f).Debug("Unassigning role")
	deleteErr := orgClient.DeleteOrgUserRoleOrgScopeProjectOrg(ctx, companySFID, roleID, scopeID, &user.Username, &userEmail)

	if deleteErr != nil {
		log.WithFields(f).Debug("Failed to remove cla-signatory role")
		return deleteErr
	}

	return nil

}

func prepareUserForSigning(ctx context.Context, userEmail string, companySFID, projectSFID, signedEntityName string) error {
	f := logrus.Fields{
		"functionName":     "sign.prepareUserForSigning",
		utils.XREQUESTID:   ctx.Value(utils.XREQUESTID),
		"user_email":       userEmail,
		"company_sfid":     companySFID,
		"project_sfid":     projectSFID,
		"signedEntityName": signedEntityName,
	}

	role := utils.CLASignatoryRole
	log.WithFields(f).Debug("called")
	usc := userService.GetClient()
	// search user
	log.WithFields(f).Debug("searching user by email")
	user, err := usc.SearchUserByEmail(userEmail)
	if err != nil {
		log.WithFields(f).WithError(err).Debugf("User with email: %s does not have an LF login", userEmail)
		return nil
	}

	ac := acsService.GetClient()
	log.WithFields(f).Debugf("getting role_id for %s", role)
	roleID, err := ac.GetRoleID(role)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("getting role_id for %s failed: %v", role, err.Error())
		return err
	}
	log.WithFields(f).Debugf("fetched role %s, role_id %s", role, roleID)
	// assign user role of cla signatory for this project
	osc := organizationService.GetClient()

	// Attempt to assign the cla-signatory role
	log.WithFields(f).Debugf("assigning user role of %s...", role)
	err = osc.CreateOrgUserRoleOrgScopeProjectOrg(ctx, userEmail, projectSFID, companySFID, roleID)
	if err != nil {
		// Log the error - but assigning the cla-signatory role is not a requirement as most users do not have a LF Login - do not throw an error
		if strings.Contains(err.Error(), "associated with some organization") {
			msg := fmt.Sprintf("user: %s already associated with some organization", user.Username)
			log.WithFields(f).WithError(err).Warn(msg)
			// return errors.New(msg)
		} else if _, ok := err.(*organizations.CreateOrgUsrRoleScopesConflict); !ok {
			log.WithFields(f).WithError(err).Warnf("assigning user role of %s failed - user already assigned the role: %v", role, err)
			// return err
		} else {
			log.WithFields(f).WithError(err).Warnf("assigning user role of %s failed: %v", role, err)
		}
	}

	return nil
}

func claSignatoryEmailContent(params ClaSignatoryEmailParams) (string, string) {
	projectNamesList := strings.Join(params.ProjectNames, ", ")

	emailSubject := fmt.Sprintf("EasyCLA: CLA Signature Request for %s", params.ClaGroupName)
	emailBody := fmt.Sprintf("<p>Hello %s,<p>", params.SignatoryName)
	emailBody += fmt.Sprintf("<p>This is a notification email from EasyCLA regarding the project(s) %s associated with the CLA Group %s. %s has designated you as an authorized signatory for the organization %s. In order for employees of your company to contribute to any of the above project(s), they must do so under a Contributor License Agreement signed by someone with authority on behalf of your company.</p>", projectNamesList, params.ClaGroupName, params.ClaManagerName, params.CompanyName)
	emailBody += fmt.Sprintf("<p>After you sign, %s (as the initial CLA Manager for your company) will be able to maintain the list of specific employees authorized to contribute to the project(s) under this signed CLA.</p>", params.ClaManagerName)
	emailBody += fmt.Sprintf("<p>If you are authorized to sign on your company’s behalf, and if you approve %s as your initial CLA Manager, please review the document and sign the CLA. If you have questions, or if you are not an authorized signatory of this company, please contact the requester at %s.</p>", params.ClaManagerName, params.ClaManagerEmail)
	// You would need to implement the appendEmailHelpSignOffContent function in Go separately

	return emailSubject, emailBody
}
