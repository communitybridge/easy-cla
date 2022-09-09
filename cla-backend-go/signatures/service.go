// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package signatures

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-openapi/strfmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/sirupsen/logrus"

	"github.com/communitybridge/easycla/cla-backend-go/events"
	"github.com/communitybridge/easycla/cla-backend-go/users"

	"github.com/LF-Engineering/lfx-kit/auth"
	"github.com/communitybridge/easycla/cla-backend-go/company"
	"github.com/communitybridge/easycla/cla-backend-go/utils"

	"github.com/communitybridge/easycla/cla-backend-go/gen/v1/restapi/operations/signatures"

	log "github.com/communitybridge/easycla/cla-backend-go/logging"

	"github.com/communitybridge/easycla/cla-backend-go/gen/v1/models"
	githubpkg "github.com/google/go-github/v37/github"
	"golang.org/x/oauth2"
)

// SignatureService interface
type SignatureService interface {
	GetSignature(ctx context.Context, signatureID string) (*models.Signature, error)
	GetIndividualSignature(ctx context.Context, claGroupID, userID string, approved, signed *bool) (*models.Signature, error)
	GetCorporateSignature(ctx context.Context, claGroupID, companyID string, approved, signed *bool) (*models.Signature, error)
	GetProjectSignatures(ctx context.Context, params signatures.GetProjectSignaturesParams) (*models.Signatures, error)
	CreateProjectSummaryReport(ctx context.Context, params signatures.CreateProjectSummaryReportParams) (*models.SignatureReport, error)
	GetProjectCompanySignature(ctx context.Context, companyID, projectID string, approved, signed *bool, nextKey *string, pageSize *int64) (*models.Signature, error)
	GetProjectCompanySignatures(ctx context.Context, params signatures.GetProjectCompanySignaturesParams) (*models.Signatures, error)
	GetProjectCompanyEmployeeSignatures(ctx context.Context, params signatures.GetProjectCompanyEmployeeSignaturesParams, criteria *ApprovalCriteria) (*models.Signatures, error)
	GetCompanySignatures(ctx context.Context, params signatures.GetCompanySignaturesParams) (*models.Signatures, error)
	GetCompanyIDsWithSignedCorporateSignatures(ctx context.Context, claGroupID string) ([]SignatureCompanyID, error)
	GetUserSignatures(ctx context.Context, params signatures.GetUserSignaturesParams) (*models.Signatures, error)
	InvalidateProjectRecords(ctx context.Context, projectID, note string) (int, error)

	GetGithubOrganizationsFromApprovalList(ctx context.Context, signatureID string, githubAccessToken string) ([]models.GithubOrg, error)
	AddGithubOrganizationToApprovalList(ctx context.Context, signatureID string, approvalListParams models.GhOrgWhitelist, githubAccessToken string) ([]models.GithubOrg, error)
	DeleteGithubOrganizationFromApprovalList(ctx context.Context, signatureID string, approvalListParams models.GhOrgWhitelist, githubAccessToken string) ([]models.GithubOrg, error)
	UpdateApprovalList(ctx context.Context, authUser *auth.User, claGroupModel *models.ClaGroup, companyModel *models.Company, claGroupID string, params *models.ApprovalList) (*models.Signature, error)

	AddCLAManager(ctx context.Context, signatureID, claManagerID string) (*models.Signature, error)
	RemoveCLAManager(ctx context.Context, ignatureID, claManagerID string) (*models.Signature, error)

	GetClaGroupICLASignatures(ctx context.Context, claGroupID string, searchTerm *string, approved, signed *bool, pageSize int64, nextKey string) (*models.IclaSignatures, error)
	GetClaGroupCCLASignatures(ctx context.Context, claGroupID string, approved, signed *bool) (*models.Signatures, error)
	GetClaGroupCorporateContributors(ctx context.Context, claGroupID string, companyID *string, searchTerm *string) (*models.CorporateContributorList, error)
}

type service struct {
	repo                SignatureRepository
	companyService      company.IService
	usersService        users.Service
	eventsService       events.Service
	githubOrgValidation bool
}

// NewService creates a new signature service
func NewService(repo SignatureRepository, companyService company.IService, usersService users.Service, eventsService events.Service, githubOrgValidation bool) SignatureService {
	return service{
		repo,
		companyService,
		usersService,
		eventsService,
		githubOrgValidation,
	}
}

// GetSignature returns the signature associated with the specified signature ID
func (s service) GetSignature(ctx context.Context, signatureID string) (*models.Signature, error) {
	return s.repo.GetSignature(ctx, signatureID)
}

// GetIndividualSignature returns the signature associated with the specified CLA Group and User ID
func (s service) GetIndividualSignature(ctx context.Context, claGroupID, userID string, approved, signed *bool) (*models.Signature, error) {
	return s.repo.GetIndividualSignature(ctx, claGroupID, userID, approved, signed)
}

// GetCorporateSignature returns the signature associated with the specified CLA Group and Company ID
func (s service) GetCorporateSignature(ctx context.Context, claGroupID, companyID string, approved, signed *bool) (*models.Signature, error) {
	return s.repo.GetCorporateSignature(ctx, claGroupID, companyID, approved, signed)
}

// GetProjectSignatures returns the list of signatures associated with the specified project
func (s service) GetProjectSignatures(ctx context.Context, params signatures.GetProjectSignaturesParams) (*models.Signatures, error) {

	projectSignatures, err := s.repo.GetProjectSignatures(ctx, params)
	if err != nil {
		return nil, err
	}

	return projectSignatures, nil
}

// CreateProjectSummaryReport generates a project summary report based on the specified input
func (s service) CreateProjectSummaryReport(ctx context.Context, params signatures.CreateProjectSummaryReportParams) (*models.SignatureReport, error) {

	projectSignatures, err := s.repo.CreateProjectSummaryReport(ctx, params)
	if err != nil {
		return nil, err
	}

	return projectSignatures, nil
}

// GetProjectCompanySignature returns the signature associated with the specified project and company
func (s service) GetProjectCompanySignature(ctx context.Context, companyID, projectID string, approved, signed *bool, nextKey *string, pageSize *int64) (*models.Signature, error) {
	return s.repo.GetProjectCompanySignature(ctx, companyID, projectID, approved, signed, nextKey, pageSize)
}

// GetProjectCompanySignatures returns the list of signatures associated with the specified project
func (s service) GetProjectCompanySignatures(ctx context.Context, params signatures.GetProjectCompanySignaturesParams) (*models.Signatures, error) {

	const defaultPageSize int64 = 10
	var pageSize = defaultPageSize
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	signed := true
	approved := true

	projectSignatures, err := s.repo.GetProjectCompanySignatures(
		ctx, params.CompanyID, params.ProjectID, &signed, &approved, params.NextKey, params.SortOrder, &pageSize)
	if err != nil {
		return nil, err
	}

	return projectSignatures, nil
}

// GetProjectCompanyEmployeeSignatures returns the list of employee signatures associated with the specified project
func (s service) GetProjectCompanyEmployeeSignatures(ctx context.Context, params signatures.GetProjectCompanyEmployeeSignaturesParams, criteria *ApprovalCriteria) (*models.Signatures, error) {

	if params.PageSize == nil {
		params.PageSize = utils.Int64(10)
	}

	projectSignatures, err := s.repo.GetProjectCompanyEmployeeSignatures(ctx, params, criteria)
	if err != nil {
		return nil, err
	}

	return projectSignatures, nil
}

// GetCompanySignatures returns the list of signatures associated with the specified company
func (s service) GetCompanySignatures(ctx context.Context, params signatures.GetCompanySignaturesParams) (*models.Signatures, error) {

	const defaultPageSize int64 = 50
	var pageSize = defaultPageSize
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	companySignatures, err := s.repo.GetCompanySignatures(ctx, params, pageSize, LoadACLDetails)
	if err != nil {
		return nil, err
	}

	return companySignatures, nil
}

// GetCompanyIDsWithSignedCorporateSignatures returns a list of company IDs that have signed a CLA agreement
func (s service) GetCompanyIDsWithSignedCorporateSignatures(ctx context.Context, claGroupID string) ([]SignatureCompanyID, error) {
	return s.repo.GetCompanyIDsWithSignedCorporateSignatures(ctx, claGroupID)
}

// GetUserSignatures returns the list of user signatures associated with the specified user
func (s service) GetUserSignatures(ctx context.Context, params signatures.GetUserSignaturesParams) (*models.Signatures, error) {

	const defaultPageSize int64 = 10
	var pageSize = defaultPageSize
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	userSignatures, err := s.repo.GetUserSignatures(ctx, params, pageSize)
	if err != nil {
		return nil, err
	}

	return userSignatures, nil
}

// GetGithubOrganizationsFromApprovalList retrieves the organization from the approval list
func (s service) GetGithubOrganizationsFromApprovalList(ctx context.Context, signatureID string, githubAccessToken string) ([]models.GithubOrg, error) {

	if signatureID == "" {
		msg := "unable to get GitHub organizations approval list - signature ID is nil"
		log.Warn(msg)
		return nil, errors.New(msg)
	}

	orgIds, err := s.repo.GetGithubOrganizationsFromApprovalList(ctx, signatureID)
	if err != nil {
		log.Warnf("error loading github organization from approval list using signatureID: %s, error: %v",
			signatureID, err)
		return nil, err
	}

	if githubAccessToken != "" {
		log.Debugf("already authenticated with github - scanning for user's orgs...")

		selectedOrgs := make(map[string]struct{}, len(orgIds))
		for _, selectedOrg := range orgIds {
			selectedOrgs[*selectedOrg.ID] = struct{}{}
		}

		// Since we're logged into github, lets get the list of organization we can add.
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: githubAccessToken},
		)
		tc := oauth2.NewClient(utils.NewContext(), ts)
		client := githubpkg.NewClient(tc)

		opt := &githubpkg.ListOptions{
			PerPage: 100,
		}

		orgs, _, err := client.Organizations.List(utils.NewContext(), "", opt)
		if err != nil {
			return nil, err
		}

		for _, org := range orgs {
			_, ok := selectedOrgs[*org.Login]
			if ok {
				continue
			}

			orgIds = append(orgIds, models.GithubOrg{ID: org.Login})
		}
	}

	return orgIds, nil
}

// AddGithubOrganizationToApprovalList adds the GH organization to the approval list
func (s service) AddGithubOrganizationToApprovalList(ctx context.Context, signatureID string, approvalListParams models.GhOrgWhitelist, githubAccessToken string) ([]models.GithubOrg, error) {
	organizationID := approvalListParams.OrganizationID

	if signatureID == "" {
		msg := "unable to add GitHub organization from approval list - signature ID is nil"
		log.Warn(msg)
		return nil, errors.New(msg)
	}

	if organizationID == nil {
		msg := "unable to add GitHub organization from approval list - organization ID is nil"
		log.Warn(msg)
		return nil, errors.New(msg)
	}

	// GH_ORG_VALIDATION environment - set to false to test locally which will by-pass the GH auth checks and
	// allow functional tests (e.g. with curl or postmon) - default is enabled

	if s.githubOrgValidation {
		// Verify the authenticated github user has access to the github organization being added.
		if githubAccessToken == "" {
			msg := fmt.Sprintf("unable to add github organization, not logged in using "+
				"signatureID: %s, github organization id: %s",
				signatureID, *organizationID)
			log.Warn(msg)
			return nil, errors.New(msg)
		}

		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: githubAccessToken},
		)
		tc := oauth2.NewClient(utils.NewContext(), ts)
		client := githubpkg.NewClient(tc)

		opt := &githubpkg.ListOptions{
			PerPage: 100,
		}

		log.Debugf("querying for user's github organizations...")
		orgs, _, err := client.Organizations.List(utils.NewContext(), "", opt)
		if err != nil {
			return nil, err
		}

		found := false
		for _, org := range orgs {
			if *org.Login == *organizationID {
				found = true
				break
			}
		}

		if !found {
			msg := fmt.Sprintf("user is not authorized for github organization id: %s", *organizationID)
			log.Warnf(msg)
			return nil, errors.New(msg)
		}
	}

	gitHubOrgApprovalList, err := s.repo.AddGithubOrganizationToApprovalList(ctx, signatureID, *organizationID)
	if err != nil {
		log.Warnf("issue adding github organization to approval list using signatureID: %s, gh org id: %s, error: %v",
			signatureID, *organizationID, err)
		return nil, err
	}

	return gitHubOrgApprovalList, nil
}

// DeleteGithubOrganizationFromApprovalList deletes the specified GH organization from the approval list
func (s service) DeleteGithubOrganizationFromApprovalList(ctx context.Context, signatureID string, approvalListParams models.GhOrgWhitelist, githubAccessToken string) ([]models.GithubOrg, error) {

	// Extract the payload values
	organizationID := approvalListParams.OrganizationID

	if signatureID == "" {
		msg := "unable to delete GitHub organization from approval list - signature ID is nil"
		log.Warn(msg)
		return nil, errors.New(msg)
	}

	if organizationID == nil {
		msg := "unable to delete GitHub organization from approval list - organization ID is nil"
		log.Warn(msg)
		return nil, errors.New(msg)
	}

	// GH_ORG_VALIDATION environment - set to false to test locally which will by-pass the GH auth checks and
	// allow functional tests (e.g. with curl or postmon) - default is enabled

	if s.githubOrgValidation {
		// Verify the authenticated github user has access to the github organization being added.
		if githubAccessToken == "" {
			msg := fmt.Sprintf("unable to delete github organization, not logged in using "+
				"signatureID: %s, github organization id: %s",
				signatureID, *organizationID)
			log.Warn(msg)
			return nil, errors.New(msg)
		}

		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: githubAccessToken},
		)
		tc := oauth2.NewClient(context.Background(), ts)
		client := githubpkg.NewClient(tc)

		opt := &githubpkg.ListOptions{
			PerPage: 100,
		}

		log.Debugf("querying for user's github organizations...")
		orgs, _, err := client.Organizations.List(context.Background(), "", opt)
		if err != nil {
			return nil, err
		}

		found := false
		for _, org := range orgs {
			if *org.Login == *organizationID {
				found = true
				break
			}
		}

		if !found {
			msg := fmt.Sprintf("user is not authorized for github organization id: %s", *organizationID)
			log.Warnf(msg)
			return nil, errors.New(msg)
		}
	}

	gitHubOrgApprovalList, err := s.repo.DeleteGithubOrganizationFromApprovalList(ctx, signatureID, *organizationID)
	if err != nil {
		return nil, err
	}

	return gitHubOrgApprovalList, nil
}

// UpdateApprovalList service method which handles updating the various approval lists
func (s service) UpdateApprovalList(ctx context.Context, authUser *auth.User, claGroupModel *models.ClaGroup, companyModel *models.Company, claGroupID string, params *models.ApprovalList) (*models.Signature, error) {
	f := logrus.Fields{
		"functionName":      "v1.signatures.service.UpdateApprovalList",
		utils.XREQUESTID:    ctx.Value(utils.XREQUESTID),
		"authUser.UserName": authUser.UserName,
		"authUser.Email":    authUser.Email,
		"claGroupID":        claGroupID,
		"claGroupName":      claGroupModel.ProjectName,
		"companyName":       companyModel.CompanyName,
		"companyID":         companyModel.CompanyID,
	}

	// Lookup the project corporate signature - should have one
	pageSize := int64(1)
	signed, approved := true, true
	corporateSigModel, sigErr := s.GetProjectCompanySignature(ctx, companyModel.CompanyID, claGroupID, &signed, &approved, nil, &pageSize)
	if sigErr != nil {
		msg := fmt.Sprintf("unable to locate project company signature by Company ID: %s, Project ID: %s, CLA Group ID: %s, error: %+v",
			companyModel.CompanyID, claGroupModel.ProjectID, claGroupID, sigErr)
		log.WithFields(f).WithError(sigErr).Warn(msg)
		return nil, NewBadRequestError(msg)
	}
	// If not found, return error
	if corporateSigModel == nil {
		msg := fmt.Sprintf("unable to locate signature for company ID: %s CLA Group ID: %s, type: ccla, signed: %t, approved: %t",
			companyModel.CompanyID, claGroupID, signed, approved)
		log.WithFields(f).Warn(msg)
		return nil, NewBadRequestError(msg)
	}

	// Ensure current user is in the Signature ACL
	claManagers := corporateSigModel.SignatureACL
	if !utils.CurrentUserInACL(authUser, claManagers) {
		msg := fmt.Sprintf("EasyCLA - 403 Forbidden - CLA Manager %s / %s is not authorized to approve request for company ID: %s / %s / %s, project ID: %s / %s / %s",
			authUser.UserName, authUser.Email,
			companyModel.CompanyName, companyModel.CompanyExternalID, companyModel.CompanyID,
			claGroupModel.ProjectName, claGroupModel.ProjectExternalID, claGroupModel.ProjectID)
		return nil, NewForbiddenError(msg)
	}

	// Lookup the user making the request - should be the CLA Manager
	userModel, userErr := s.usersService.GetUserByUserName(authUser.UserName, true)
	if userErr != nil {
		log.WithFields(f).WithError(userErr).Warnf("unable to lookup user by user name: %s", authUser.UserName)
		return nil, userErr
	}

	eventArgs := &events.LogEventArgs{
		EventType:     events.InvalidatedSignature,
		ProjectID:     claGroupModel.ProjectExternalID,
		ClaGroupModel: claGroupModel,
		CompanyID:     companyModel.CompanyID,
		CompanyModel:  companyModel,
		LfUsername:    userModel.LfUsername,
		UserID:        userModel.UserID,
		UserModel:     userModel,
		ProjectSFID:   claGroupModel.ProjectExternalID,
	}

	updatedSig, err := s.repo.UpdateApprovalList(ctx, userModel, claGroupModel, companyModel.CompanyID, params, eventArgs)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("problem updating approval list for company ID: %s, project ID: %s, cla group ID: %s", companyModel.CompanyID, claGroupModel.ProjectID, claGroupID)
		return updatedSig, err
	}

	// Log Events that the CLA manager updated the approval lists
	s.createEventLogEntries(ctx, companyModel, claGroupModel, userModel, params)

	// Send an email to each of the CLA Managers
	for _, claManager := range claManagers {
		claManagerEmail := getBestEmail(&claManager) // nolint
		s.sendApprovalListUpdateEmailToCLAManagers(companyModel, claGroupModel, claManager.Username, claManagerEmail, params)
	}

	// TODO: DAD - update email template to indicate that if auto crate ECLA is enabled, that users should be good-to-go
	// Send emails to contributors if email or GitHub/GitLab username was added or removed
	s.sendRequestAccessEmailToContributors(authUser, companyModel, claGroupModel, params)

	// If auto create ECLA is enabled for this Corporate Agreement, then create an ECLA for each employee that was added to the approval list
	if corporateSigModel.AutoCreateECLA {
		for _, email := range params.AddEmailApprovalList {
			// Lookup the user making the request
			employeeUserModel, userLookupErr := s.usersService.GetUserByEmail(email)
			if userLookupErr != nil {
				log.WithFields(f).WithError(userLookupErr).Warnf("unable to lookup user by email: %s - unable to auto-create employee acknowledgement", email)
				continue
			}
			if employeeUserModel == nil {
				// TODO: Need to create a new user record based on the email
				employeeUserModel, userCreateErr := s.usersService.CreateUser(&models.User{
					Admin:          false,
					CompanyID:      "",
					DateCreated:    "",
					DateModified:   "",
					Emails:         []string{email},
					GithubID:       "",
					GithubUsername: "",
					GitlabID:       "",
					GitlabUsername: "",
					LfEmail:        strfmt.Email(email),
					LfUsername:     "",
					Note:           "",
					UserExternalID: "",
					UserID:         "",
					Username:       "",
					Version:        "v1",
				}, &user.ClaUser{})
			}

			createErr := s.repo.CreateProjectCompanyEmployeeSignature(ctx, companyModel, claGroupModel, employeeUserModel)
			if createErr != nil {
				log.WithFields(f).WithError(createErr).Warnf("unable to create project company employee signature record for: %+v", employeeUserModel)
				continue
			}
		}
		for _, gitHubUserName := range params.AddGithubUsernameApprovalList {
			// Lookup the user making the request
			employeeUserModel, userLookupErr := s.usersService.GetUserByGitHubUsername(gitHubUserName)
			if userLookupErr != nil {
				log.WithFields(f).WithError(userLookupErr).Warnf("unable to lookup user by GitHub username: %s - unable to auto-create employee acknowledgement", gitHubUserName)
				continue
			}
			if employeeUserModel == nil {
				// TODO: Need to create a new user record based on the GitHub username
			}

			createErr := s.repo.CreateProjectCompanyEmployeeSignature(ctx, companyModel, claGroupModel, employeeUserModel)
			if createErr != nil {
				log.WithFields(f).WithError(createErr).Warnf("unable to create project company employee signature record for: %+v", employeeUserModel)
				continue
			}
		}
		for _, gitLabUserName := range params.AddGitlabUsernameApprovalList {
			// Lookup the user making the request
			employeeUserModel, userLookupErr := s.usersService.GetUserByGitlabUsername(gitLabUserName)
			if userLookupErr != nil {
				log.WithFields(f).WithError(userLookupErr).Warnf("unable to lookup user by GitLab username: %s - unable to auto-create employee acknowledgement", gitLabUserName)
				continue
			}
			if employeeUserModel == nil {
				// TODO: Need to create a new user record based on the GitLab username
			}

			createErr := s.repo.CreateProjectCompanyEmployeeSignature(ctx, companyModel, claGroupModel, employeeUserModel)
			if createErr != nil {
				log.WithFields(f).WithError(createErr).Warnf("unable to create project company employee signature record for: %+v", employeeUserModel)
				continue
			}
		}
	}

	return updatedSig, nil
}

// InvalidateProjectRecords disassociates project signatures
func (s service) InvalidateProjectRecords(ctx context.Context, projectID, note string) (int, error) {
	f := logrus.Fields{
		"functionName": "v1.signatures.service.InvalidateProjectRecords",
		"projectID":    projectID,
	}

	result, err := s.repo.ProjectSignatures(ctx, projectID)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf(fmt.Sprintf("Unable to get signatures for project: %s", projectID))
		return 0, err
	}

	if len(result.Signatures) > 0 {
		var wg sync.WaitGroup
		wg.Add(len(result.Signatures))
		log.WithFields(f).Debugf(fmt.Sprintf("Invalidating %d signatures for project: %s ",
			len(result.Signatures), projectID))
		for _, signature := range result.Signatures {
			// Do this in parallel, as we could have a lot to invalidate
			go func(sigID, projectID string) {
				defer wg.Done()
				updateErr := s.repo.InvalidateProjectRecord(ctx, sigID, note)
				if updateErr != nil {
					log.WithFields(f).WithError(updateErr).Warnf("Unable to update signature: %s with project ID: %s, error: %v", sigID, projectID, updateErr)
				}
			}(signature.SignatureID, projectID)
		}

		// Wait until all the workers are done
		wg.Wait()
	}

	return len(result.Signatures), nil
}

// AddCLAManager adds the specified manager to the signature ACL list
func (s service) AddCLAManager(ctx context.Context, signatureID, claManagerID string) (*models.Signature, error) {
	return s.repo.AddCLAManager(ctx, signatureID, claManagerID)
}

// RemoveCLAManager removes the specified manager from the signature ACL list
func (s service) RemoveCLAManager(ctx context.Context, signatureID, claManagerID string) (*models.Signature, error) {
	return s.repo.RemoveCLAManager(ctx, signatureID, claManagerID)
}

// appendList is a helper function to generate the email content of the Approval List changes
func appendList(approvalList []string, message string) string {
	approvalListSummary := ""

	if len(approvalList) > 0 {
		for _, value := range approvalList {
			approvalListSummary += fmt.Sprintf("<li>%s %s</li>", message, value)
		}
	}

	return approvalListSummary
}

// buildApprovalListSummary is a helper function to generate the email content of the Approval List changes
func buildApprovalListSummary(approvalListChanges *models.ApprovalList) string {
	approvalListSummary := "<ul>"
	approvalListSummary += appendList(approvalListChanges.AddEmailApprovalList, "Added Email:")
	approvalListSummary += appendList(approvalListChanges.RemoveEmailApprovalList, "Removed Email:")
	approvalListSummary += appendList(approvalListChanges.AddDomainApprovalList, "Added Domain:")
	approvalListSummary += appendList(approvalListChanges.RemoveDomainApprovalList, "Removed Domain:")
	approvalListSummary += appendList(approvalListChanges.AddGithubUsernameApprovalList, "Added GitHub User:")
	approvalListSummary += appendList(approvalListChanges.RemoveGithubUsernameApprovalList, "Removed GitHub User:")
	approvalListSummary += appendList(approvalListChanges.AddGithubOrgApprovalList, "Added GitHub Organization:")
	approvalListSummary += appendList(approvalListChanges.RemoveGithubOrgApprovalList, "Removed GitHub Organization:")
	approvalListSummary += appendList(approvalListChanges.AddGitlabUsernameApprovalList, "Added Gitlab User:")
	approvalListSummary += appendList(approvalListChanges.RemoveGitlabUsernameApprovalList, "Removed Gitlab User:")
	approvalListSummary += appendList(approvalListChanges.AddGitlabOrgApprovalList, "Added Gitlab Organization:")
	approvalListSummary += appendList(approvalListChanges.RemoveGitlabOrgApprovalList, "Removed Gitlab Organization:")
	approvalListSummary += "</ul>"
	return approvalListSummary
}

// sendRequestAccessEmailToCLAManagers sends the request access email to the specified CLA Managers
func (s service) sendApprovalListUpdateEmailToCLAManagers(companyModel *models.Company, claGroupModel *models.ClaGroup, recipientName, recipientAddress string, approvalListChanges *models.ApprovalList) {
	f := logrus.Fields{
		"functionName":      "sendApprovalListUpdateEmailToCLAManagers",
		"projectName":       claGroupModel.ProjectName,
		"projectExternalID": claGroupModel.ProjectExternalID,
		"foundationSFID":    claGroupModel.FoundationSFID,
		"companyName":       companyModel.CompanyName,
		"companyExternalID": companyModel.CompanyExternalID,
		"recipientName":     recipientName,
		"recipientAddress":  recipientAddress}

	companyName := companyModel.CompanyName
	projectName := claGroupModel.ProjectName

	// subject string, body string, recipients []string
	subject := fmt.Sprintf("EasyCLA: Approval List Update for %s on %s", companyName, projectName)
	recipients := []string{recipientAddress}
	body := fmt.Sprintf(`
<p>Hello %s,</p>
<p>This is a notification email from EasyCLA regarding the project %s.</p>
<p>The EasyCLA approval list for %s for project %s was modified.</p>
<p>The modification was as follows:</p>
%s
%s
%s`,
		recipientName, projectName, companyName, projectName, buildApprovalListSummary(approvalListChanges),
		utils.GetEmailHelpContent(claGroupModel.Version == utils.V2), utils.GetEmailSignOffContent())

	err := utils.SendEmail(subject, body, recipients)
	if err != nil {
		log.WithFields(f).Warnf("problem sending email with subject: %s to recipients: %+v, error: %+v", subject, recipients, err)
	} else {
		log.WithFields(f).Debugf("sent email with subject: %s to recipients: %+v", subject, recipients)
	}
}

// getAddEmailContributors is a helper function to lookup the contributors impacted by the Approval List update
func (s service) getAddEmailContributors(approvalList *models.ApprovalList) []*models.User {
	var userModelList []*models.User
	for _, value := range approvalList.AddEmailApprovalList {
		userModel, err := s.usersService.GetUserByEmail(value)
		if err != nil {
			log.Warnf("unable to lookup user by LF email: %s, error: %+v", value, err)
		} else {
			userModelList = append(userModelList, userModel)
		}
	}

	return userModelList
}

// getRemoveEmailContributors is a helper function to lookup the contributors impacted by the Approval List update
func (s service) getRemoveEmailContributors(approvalList *models.ApprovalList) []*models.User {
	var userModelList []*models.User
	for _, value := range approvalList.RemoveEmailApprovalList {
		userModel, err := s.usersService.GetUserByEmail(value)
		if err != nil {
			log.Warnf("unable to lookup user by LF email: %s, error: %+v", value, err)
		} else {
			userModelList = append(userModelList, userModel)
		}
	}

	return userModelList
}

// getAddGitHubContributors is a helper function to lookup the contributors impacted by the Approval List update
func (s service) getAddGitHubContributors(approvalList *models.ApprovalList) []*models.User {
	var userModelList []*models.User
	for _, value := range approvalList.AddGithubUsernameApprovalList {
		userModel, err := s.usersService.GetUserByGitHubUsername(value)
		if err != nil {
			log.Warnf("unable to lookup user by GitHub username: %s, error: %+v", value, err)
		} else {
			userModelList = append(userModelList, userModel)
		}
	}

	return userModelList
}

// getRemoveGitHubContributors is a helper function to lookup the contributors impacted by the Approval List update
func (s service) getRemoveGitHubContributors(approvalList *models.ApprovalList) []*models.User {
	var userModelList []*models.User
	for _, value := range approvalList.RemoveGithubUsernameApprovalList {
		userModel, err := s.usersService.GetUserByGitHubUsername(value)
		if err != nil {
			log.Warnf("unable to lookup user by GitHub username: %s, error: %+v", value, err)
		} else {
			userModelList = append(userModelList, userModel)
		}
	}

	return userModelList
}

// getAddGitlabContributors is a helper function to look up the Gitlab contributors impacted by the Approval List update
func (s service) getAddGitlabContributors(approvalList *models.ApprovalList) []*models.User {
	var userModelList []*models.User
	for _, value := range approvalList.AddGitlabUsernameApprovalList {
		userModel, err := s.usersService.GetUserByGitHubUsername(value)
		if err != nil {
			log.Warnf("unable to lookup user by Gitlab username: %s, error: %+v", value, err)
		} else {
			userModelList = append(userModelList, userModel)
		}
	}

	return userModelList
}

// getRemoveGitlabContributors is a helper function to look up the Gitlab contributors impacted by the Approval List update
func (s service) getRemoveGitlabContributors(approvalList *models.ApprovalList) []*models.User {
	var userModelList []*models.User
	for _, value := range approvalList.RemoveGitlabUsernameApprovalList {
		userModel, err := s.usersService.GetUserByGitHubUsername(value)
		if err != nil {
			log.Warnf("unable to lookup user by Gitlab username: %s, error: %+v", value, err)
		} else {
			userModelList = append(userModelList, userModel)
		}
	}

	return userModelList
}

func (s service) sendRequestAccessEmailToContributors(authUser *auth.User, companyModel *models.Company, claGroupModel *models.ClaGroup, approvalList *models.ApprovalList) {
	addEmailUsers := s.getAddEmailContributors(approvalList)
	for _, user := range addEmailUsers {
		sendRequestAccessEmailToContributorRecipient(authUser, companyModel, claGroupModel, user.Username, user.LfEmail.String(), "added", "to",
			fmt.Sprintf("you are authorized to contribute to %s on behalf of %s", claGroupModel.ProjectName, companyModel.CompanyName))
	}
	removeEmailUsers := s.getRemoveEmailContributors(approvalList)
	for _, user := range removeEmailUsers {
		sendRequestAccessEmailToContributorRecipient(authUser, companyModel, claGroupModel, user.Username, user.LfEmail.String(), "removed", "from",
			fmt.Sprintf("you are no longer authorized to contribute to %s on behalf of %s ", claGroupModel.ProjectName, companyModel.CompanyName))
	}
	addGitHubUsers := s.getAddGitHubContributors(approvalList)
	for _, user := range addGitHubUsers {
		sendRequestAccessEmailToContributorRecipient(authUser, companyModel, claGroupModel, user.Username, user.LfEmail.String(), "added", "to",
			fmt.Sprintf("you are authorized to contribute to %s on behalf of %s", claGroupModel.ProjectName, companyModel.CompanyName))
	}
	removeGitHubUsers := s.getRemoveGitHubContributors(approvalList)
	for _, user := range removeGitHubUsers {
		sendRequestAccessEmailToContributorRecipient(authUser, companyModel, claGroupModel, user.Username, user.LfEmail.String(), "removed", "from",
			fmt.Sprintf("you are no longer authorized to contribute to %s on behalf of %s ", claGroupModel.ProjectName, companyModel.CompanyName))
	}
	addGitlabUsers := s.getAddGitlabContributors(approvalList)
	for _, user := range addGitlabUsers {
		sendRequestAccessEmailToContributorRecipient(authUser, companyModel, claGroupModel, user.Username, user.LfEmail.String(), "added", "to",
			fmt.Sprintf("you are authorized to contribute to %s on behalf of %s", claGroupModel.ProjectName, companyModel.CompanyName))
	}
	removeGitlabUsers := s.getRemoveGitlabContributors(approvalList)
	for _, user := range removeGitlabUsers {
		sendRequestAccessEmailToContributorRecipient(authUser, companyModel, claGroupModel, user.Username, user.LfEmail.String(), "removed", "from",
			fmt.Sprintf("you are no longer authorized to contribute to %s on behalf of %s ", claGroupModel.ProjectName, companyModel.CompanyName))
	}
}

func (s service) createEventLogEntries(ctx context.Context, companyModel *models.Company, claGroupModel *models.ClaGroup, userModel *models.User, approvalList *models.ApprovalList) {
	for _, value := range approvalList.AddEmailApprovalList {
		// Send an event
		s.eventsService.LogEventWithContext(ctx, &events.LogEventArgs{
			EventType:     events.ClaApprovalListUpdated,
			ProjectID:     claGroupModel.ProjectExternalID,
			ClaGroupModel: claGroupModel,
			CompanyID:     companyModel.CompanyID,
			CompanyModel:  companyModel,
			LfUsername:    userModel.LfUsername,
			UserID:        userModel.UserID,
			UserModel:     userModel,
			ProjectSFID:   claGroupModel.ProjectExternalID,
			EventData: &events.CLAApprovalListAddEmailData{
				ApprovalListEmail: value,
			},
		})
	}
	for _, value := range approvalList.RemoveEmailApprovalList {
		// Send an event
		s.eventsService.LogEventWithContext(ctx, &events.LogEventArgs{
			EventType:     events.ClaApprovalListUpdated,
			ProjectID:     claGroupModel.ProjectExternalID,
			ClaGroupModel: claGroupModel,
			CompanyID:     companyModel.CompanyID,
			CompanyModel:  companyModel,
			LfUsername:    userModel.LfUsername,
			UserID:        userModel.UserID,
			UserModel:     userModel,
			ProjectSFID:   claGroupModel.ProjectExternalID,
			EventData: &events.CLAApprovalListRemoveEmailData{
				ApprovalListEmail: value,
			},
		})
	}
	for _, value := range approvalList.AddDomainApprovalList {
		// Send an event
		s.eventsService.LogEventWithContext(ctx, &events.LogEventArgs{
			EventType:     events.ClaApprovalListUpdated,
			ProjectID:     claGroupModel.ProjectExternalID,
			ClaGroupModel: claGroupModel,
			CompanyID:     companyModel.CompanyID,
			CompanyModel:  companyModel,
			LfUsername:    userModel.LfUsername,
			UserID:        userModel.UserID,
			UserModel:     userModel,
			ProjectSFID:   claGroupModel.ProjectExternalID,
			EventData: &events.CLAApprovalListAddDomainData{
				ApprovalListDomain: value,
			},
		})
	}
	for _, value := range approvalList.RemoveDomainApprovalList {
		// Send an event
		s.eventsService.LogEventWithContext(ctx, &events.LogEventArgs{
			EventType:     events.ClaApprovalListUpdated,
			ProjectID:     claGroupModel.ProjectExternalID,
			ClaGroupModel: claGroupModel,
			CompanyID:     companyModel.CompanyID,
			CompanyModel:  companyModel,
			LfUsername:    userModel.LfUsername,
			UserID:        userModel.UserID,
			UserModel:     userModel,
			ProjectSFID:   claGroupModel.ProjectExternalID,
			EventData: &events.CLAApprovalListRemoveDomainData{
				ApprovalListDomain: value,
			},
		})
	}
	for _, value := range approvalList.AddGithubUsernameApprovalList {
		// Send an event
		s.eventsService.LogEventWithContext(ctx, &events.LogEventArgs{
			EventType:     events.ClaApprovalListUpdated,
			ProjectID:     claGroupModel.ProjectExternalID,
			ClaGroupModel: claGroupModel,
			CompanyID:     companyModel.CompanyID,
			CompanyModel:  companyModel,
			LfUsername:    userModel.LfUsername,
			UserID:        userModel.UserID,
			UserModel:     userModel,
			ProjectSFID:   claGroupModel.ProjectExternalID,
			EventData: &events.CLAApprovalListAddGitHubUsernameData{
				ApprovalListGitHubUsername: value,
			},
		})
	}
	for _, value := range approvalList.RemoveGithubUsernameApprovalList {
		// Send an event
		s.eventsService.LogEventWithContext(ctx, &events.LogEventArgs{
			EventType:     events.ClaApprovalListUpdated,
			ProjectID:     claGroupModel.ProjectExternalID,
			ClaGroupModel: claGroupModel,
			CompanyID:     companyModel.CompanyID,
			CompanyModel:  companyModel,
			LfUsername:    userModel.LfUsername,
			UserID:        userModel.UserID,
			UserModel:     userModel,
			ProjectSFID:   claGroupModel.ProjectExternalID,
			EventData: &events.CLAApprovalListRemoveGitHubUsernameData{
				ApprovalListGitHubUsername: value,
			},
		})
	}
	for _, value := range approvalList.AddGithubOrgApprovalList {
		// Send an event
		s.eventsService.LogEventWithContext(ctx, &events.LogEventArgs{
			EventType:     events.ClaApprovalListUpdated,
			ProjectID:     claGroupModel.ProjectExternalID,
			ClaGroupModel: claGroupModel,
			CompanyID:     companyModel.CompanyID,
			CompanyModel:  companyModel,
			LfUsername:    userModel.LfUsername,
			UserID:        userModel.UserID,
			UserModel:     userModel,
			ProjectSFID:   claGroupModel.ProjectExternalID,
			EventData: &events.CLAApprovalListAddGitHubOrgData{
				ApprovalListGitHubOrg: value,
			},
		})
	}
	for _, value := range approvalList.RemoveGithubOrgApprovalList {
		// Send an event
		s.eventsService.LogEventWithContext(ctx, &events.LogEventArgs{
			EventType:     events.ClaApprovalListUpdated,
			CLAGroupID:    claGroupModel.ProjectID,
			ProjectID:     claGroupModel.ProjectExternalID,
			ClaGroupModel: claGroupModel,
			CompanyID:     companyModel.CompanyID,
			CompanyModel:  companyModel,
			LfUsername:    userModel.LfUsername,
			UserID:        userModel.UserID,
			UserModel:     userModel,
			ProjectSFID:   claGroupModel.ProjectExternalID,
			EventData: &events.CLAApprovalListRemoveGitHubOrgData{
				ApprovalListGitHubOrg: value,
			},
		})
	}
	for _, value := range approvalList.AddGitlabUsernameApprovalList {
		// Send an event
		s.eventsService.LogEventWithContext(ctx, &events.LogEventArgs{
			EventType:     events.ClaApprovalListUpdated,
			ProjectID:     claGroupModel.ProjectExternalID,
			ClaGroupModel: claGroupModel,
			CompanyID:     companyModel.CompanyID,
			CompanyModel:  companyModel,
			LfUsername:    userModel.LfUsername,
			UserID:        userModel.UserID,
			UserModel:     userModel,
			ProjectSFID:   claGroupModel.ProjectExternalID,
			EventData: &events.CLAApprovalListAddGitLabUsernameData{
				ApprovalListGitLabUsername: value,
			},
		})
	}
	for _, value := range approvalList.RemoveGitlabUsernameApprovalList {
		// Send an event
		s.eventsService.LogEventWithContext(ctx, &events.LogEventArgs{
			EventType:     events.ClaApprovalListUpdated,
			ProjectID:     claGroupModel.ProjectExternalID,
			ClaGroupModel: claGroupModel,
			CompanyID:     companyModel.CompanyID,
			CompanyModel:  companyModel,
			LfUsername:    userModel.LfUsername,
			UserID:        userModel.UserID,
			UserModel:     userModel,
			ProjectSFID:   claGroupModel.ProjectExternalID,
			EventData: &events.CLAApprovalListRemoveGitLabUsernameData{
				ApprovalListGitLabUsername: value,
			},
		})
	}
	for _, value := range approvalList.AddGitlabOrgApprovalList {
		// Send an event
		s.eventsService.LogEventWithContext(ctx, &events.LogEventArgs{
			EventType:     events.ClaApprovalListUpdated,
			ProjectID:     claGroupModel.ProjectExternalID,
			ClaGroupModel: claGroupModel,
			CompanyID:     companyModel.CompanyID,
			CompanyModel:  companyModel,
			LfUsername:    userModel.LfUsername,
			UserID:        userModel.UserID,
			UserModel:     userModel,
			ProjectSFID:   claGroupModel.ProjectExternalID,
			EventData: &events.CLAApprovalListAddGitLabGroupData{
				ApprovalListGitLabGroup: value,
			},
		})
	}
	for _, value := range approvalList.RemoveGitlabOrgApprovalList {
		// Send an event
		s.eventsService.LogEventWithContext(ctx, &events.LogEventArgs{
			EventType:     events.ClaApprovalListUpdated,
			CLAGroupID:    claGroupModel.ProjectID,
			ProjectID:     claGroupModel.ProjectExternalID,
			ClaGroupModel: claGroupModel,
			CompanyID:     companyModel.CompanyID,
			CompanyModel:  companyModel,
			LfUsername:    userModel.LfUsername,
			UserID:        userModel.UserID,
			UserModel:     userModel,
			ProjectSFID:   claGroupModel.ProjectExternalID,
			EventData: &events.CLAApprovalListRemoveGitLabGroupData{
				ApprovalListGitLabGroup: value,
			},
		})
	}
}

func (s service) GetClaGroupICLASignatures(ctx context.Context, claGroupID string, searchTerm *string, approved, signed *bool, pageSize int64, nextKey string) (*models.IclaSignatures, error) {
	return s.repo.GetClaGroupICLASignatures(ctx, claGroupID, searchTerm, approved, signed, pageSize, nextKey)
}

func (s service) GetClaGroupCCLASignatures(ctx context.Context, claGroupID string, approved, signed *bool) (*models.Signatures, error) {
	pageSize := utils.Int64(1000)
	return s.repo.GetProjectSignatures(ctx, signatures.GetProjectSignaturesParams{
		ClaType:   aws.String(utils.ClaTypeCCLA),
		ProjectID: claGroupID,
		PageSize:  pageSize,
		Approved:  approved,
		Signed:    signed,
	})
}

func (s service) GetClaGroupCorporateContributors(ctx context.Context, claGroupID string, companyID *string, searchTerm *string) (*models.CorporateContributorList, error) {
	return s.repo.GetClaGroupCorporateContributors(ctx, claGroupID, companyID, searchTerm)
}

// sendRequestAccessEmailToContributors sends the request access email to the specified contributors
func sendRequestAccessEmailToContributorRecipient(authUser *auth.User, companyModel *models.Company, claGroupModel *models.ClaGroup, recipientName, recipientAddress, addRemove, toFrom, authorizedString string) {
	companyName := companyModel.CompanyName
	projectName := claGroupModel.ProjectName

	// subject string, body string, recipients []string
	subject := fmt.Sprintf("EasyCLA: Approval List Update for %s on %s", companyName, projectName)
	recipients := []string{recipientAddress}
	body := fmt.Sprintf(`
<p>Hello %s,</p>
<p>This is a notification email from EasyCLA regarding the project %s.</p>
<p>You have been %s %s the Approval List of %s for %s by CLA Manager %s. This means that %s.</p>
<b>
<p>If you are a GitHub user and If you had previously submitted a pull request to EasyCLA Test Group that had failed, you can now go back to it, re-click the “Not Covered” button in the EasyCLA message in your pull request, and then follow these steps</p>
<ol>
<li>Select “Corporate Contributor”.</li>
<li>Select your company from the organization drop down list</li>
<li>Click Proceed</li>
</ol>
<p>If you are a Gerrit user and if you had previously submitted a pull request to EasyCLA Test Group that had failed, then navigate to Agreements Settings page on Gerrit, click on "New Contributor Agreement" link under Agreements section, select the radio button corresponding to Corporate CLA, click on "Please review the agreement" link, and then follow these steps</p>
<ol>
<li>Select “Corporate Contributor”.</li>
<li>Select your company from the organization drop down list</li>
<li>Click Proceed</li>
</ol>
<p>These steps will confirm your organization association and you will only need to do these once. After completing these steps, the EasyCLA check will be complete and enabled for all future code contributions for this project.</p>
</b>
%s
%s`,
		recipientName, projectName, addRemove, toFrom,
		companyName, projectName, authUser.UserName, authorizedString,
		utils.GetEmailHelpContent(claGroupModel.Version == utils.V2), utils.GetEmailSignOffContent())

	err := utils.SendEmail(subject, body, recipients)
	if err != nil {
		log.Warnf("problem sending email with subject: %s to recipients: %+v, error: %+v", subject, recipients, err)
	} else {
		log.Debugf("sent email with subject: %s to recipients: %+v", subject, recipients)
	}
}

// getBestEmail is a helper function to return the best email address for the user model
func getBestEmail(userModel *models.User) string {
	if userModel.LfEmail != "" {
		return userModel.LfEmail.String()
	}

	for _, email := range userModel.Emails {
		if email != "" && !strings.Contains(email, "noreply.github.com") {
			return email
		}
	}

	return ""
}
