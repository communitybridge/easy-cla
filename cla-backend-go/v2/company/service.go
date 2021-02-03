// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package company

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-openapi/strfmt"

	"github.com/sirupsen/logrus"

	"github.com/communitybridge/easycla/cla-backend-go/events"
	"github.com/communitybridge/easycla/cla-backend-go/project"
	"github.com/communitybridge/easycla/cla-backend-go/projects_cla_groups"

	"github.com/jinzhu/copier"

	"github.com/LF-Engineering/lfx-kit/auth"
	"github.com/communitybridge/easycla/cla-backend-go/company"

	"github.com/aws/aws-sdk-go/aws"
	v1Company "github.com/communitybridge/easycla/cla-backend-go/company"
	v1Models "github.com/communitybridge/easycla/cla-backend-go/gen/models"
	v1ProjectParams "github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations/project"
	v1SignatureParams "github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations/signatures"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	v2Models "github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/communitybridge/easycla/cla-backend-go/signatures"
	"github.com/communitybridge/easycla/cla-backend-go/users"
	"github.com/communitybridge/easycla/cla-backend-go/utils"
	acs_service "github.com/communitybridge/easycla/cla-backend-go/v2/acs-service"

	orgService "github.com/communitybridge/easycla/cla-backend-go/v2/organization-service"
	"github.com/communitybridge/easycla/cla-backend-go/v2/organization-service/client/organizations"
	v2ProjectService "github.com/communitybridge/easycla/cla-backend-go/v2/project-service"
	v2ProjectServiceModels "github.com/communitybridge/easycla/cla-backend-go/v2/project-service/models"
	v2UserService "github.com/communitybridge/easycla/cla-backend-go/v2/user-service"
	v2UserServiceModels "github.com/communitybridge/easycla/cla-backend-go/v2/user-service/models"
)

// errors
var (
	ErrProjectNotFound = errors.New("project not found")
	ErrCLAUserNotFound = errors.New("claUser not found")
	ErrNoLfUsername    = errors.New("user has no LF username")
	// ErrNoValidEmail    = errors.New("user with no valid email")

	//ErrProjectSigned returns error if project already signed
	ErrProjectSigned = errors.New("project already signed")
	//ErrLFXUserNotFound when user-service fails to find user
	ErrLFXUserNotFound = errors.New("lfx user not found")
	//ErrContributorConflict when user is already assigned contributor role
	ErrContributorConflict = errors.New("user already assigned contributor")
	//ErrRoleScopeConflict thrown if user already has role scope
	ErrRoleScopeConflict = errors.New("user is already contributor")
	//ErrClaGroupNotFound returns error if cla group not found
	ErrClaGroupNotFound = errors.New("cla group not found")
	//ErrClaGroupBadRequest returns error if cla group bad request
	ErrClaGroupBadRequest = errors.New("cla group bad request")
)

// constants
const (
	// used when we want to query all data from dependent service.
	HugePageSize = int64(10000)
	// LoadRepoDetails     = true
	DontLoadRepoDetails = false
	// FoundationType the SF foundation type string - previously was "Foundation", now "Project Group"
	FoundationType = "Project Group"
	// Lead representing type of user
	Lead = "lead"
	//NoAccount
	NoAccount = "Individual - No Account"
	//OrgAssociated stating whether user has user association with another org
	OrgAssociated = "are already associated with other organization"
)

// Service functions for company
type Service interface {
	GetCompanyProjectCLAManagers(ctx context.Context, v1CompanyModel *models.Company, projectSFID string) (*models.CompanyClaManagers, error)
	GetCompanyProjectActiveCLAs(ctx context.Context, companyID string, projectSFID string) (*models.ActiveClaList, error)
	GetCompanyProjectContributors(ctx context.Context, projectSFID string, companySFID string, searchTerm string) (*models.CorporateContributorList, error)
	GetCompanyProjectCLA(ctx context.Context, authUser *auth.User, companySFID, projectSFID string, companyID *string) (*models.CompanyProjectClaList, error)
	CreateCompany(ctx context.Context, companyName, signingEntityName, companyWebsite, userEmail, userID string) (*models.CompanyOutput, error)
	GetCompanyByName(ctx context.Context, companyName string) (*models.Company, error)
	GetCompanyBySigningEntityName(ctx context.Context, signingEntityName string) (*models.Company, error)
	GetCompanyByID(ctx context.Context, companyID string) (*models.Company, error)
	GetCompanyBySFID(ctx context.Context, companySFID string) (*models.Company, error)
	DeleteCompanyByID(ctx context.Context, companyID string) error
	DeleteCompanyBySFID(ctx context.Context, companySFID string) error
	GetCompanyCLAGroupManagers(ctx context.Context, companyID, claGroupID string) (*models.CompanyClaManagers, error)
	AssociateContributor(ctx context.Context, companySFID, userEmail string) (*models.Contributor, error)
	AssociateContributorByGroup(ctx context.Context, companySFID, userEmail string, projectCLAGroups []*projects_cla_groups.ProjectClaGroup, ClaGroupID string) ([]*models.Contributor, string, error)
	GetCompanyAdmins(ctx context.Context, companyID string) (*models.CompanyAdminList, error)
	RequestCompanyAdmin(ctx context.Context, userID string, claManagerEmail string, claManagerName string, contributorName string, contributorEmail string, projectName string, companyName string, lFxPortalURL string) error

	// org service lookup
	GetCompanyLookup(ctx context.Context, companyName string, websiteName string) (*models.Lookup, error)
}

// ProjectRepo contains project repo methods
type ProjectRepo interface {
	GetCLAGroupByID(ctx context.Context, projectID string, loadRepoDetails bool) (*v1Models.ClaGroup, error)
	GetCLAGroupsByExternalID(ctx context.Context, params *v1ProjectParams.GetProjectsByExternalIDParams, loadRepoDetails bool) (*v1Models.ClaGroups, error)
	GetCLAGroupByName(ctx context.Context, projectName string) (*v1Models.ClaGroup, error)
}

// NewService returns instance of company service
func NewService(v1CompanyService v1Company.IService, sigRepo signatures.SignatureRepository, projectRepo ProjectRepo, usersRepo users.UserRepository, companyRepo company.IRepository, pcgRepo projects_cla_groups.Repository, evService events.Service) Service {
	return &service{
		v1CompanyService:     v1CompanyService,
		signatureRepo:        sigRepo,
		projectRepo:          projectRepo,
		userRepo:             usersRepo,
		companyRepo:          companyRepo,
		projectClaGroupsRepo: pcgRepo,
		eventService:         evService,
	}
}

func (s *service) GetCompanyProjectCLAManagers(ctx context.Context, v1CompanyModel *models.Company, projectSFID string) (*models.CompanyClaManagers, error) {
	f := logrus.Fields{
		"functionName":      "GetCompanyProjectCLAManagers",
		utils.XREQUESTID:    ctx.Value(utils.XREQUESTID),
		"projectSFID":       projectSFID,
		"companyID":         v1CompanyModel.CompanyID,
		"companySFID":       v1CompanyModel.CompanyExternalID,
		"companyName":       v1CompanyModel.CompanyName,
		"signingEntityName": v1CompanyModel.SigningEntityName,
	}

	log.WithFields(f).Debugf("locating CLA Group(s) under project or foundation...")
	var err error
	claGroups, err := s.getCLAGroupsUnderProjectOrFoundation(ctx, projectSFID)
	if err != nil {
		log.WithFields(f).Warnf("problem fetching CLA Groups under project or foundation, error: %+v", err)
		return nil, err
	}

	// get the org client for org info filling
	orgClient := orgService.GetClient()
	orgModel, err := orgClient.GetOrganization(ctx, v1CompanyModel.CompanyExternalID)
	if err != nil {
		return nil, fmt.Errorf("fetching org model failed for companySFID : %s : %w", v1CompanyModel.CompanyExternalID, err)
	}

	signed, approved := true, true
	maxLoad := int64(10)
	var sigs []*v1Models.Signature
	for _, claGroup := range claGroups {
		var sigErr error
		// Should only have 1 per CLA Group/Company pair
		if claGroup.ClaGroupID == "" {
			log.WithFields(f).Debugf("claGroupID missing for project : %s ", claGroup.ProjectSFID)
			continue
		}
		sig, sigErr := s.signatureRepo.GetProjectCompanySignature(ctx, v1CompanyModel.CompanyID, claGroup.ClaGroupID, &signed, &approved, nil, &maxLoad)
		if sigErr != nil {
			log.WithFields(f).Warnf("problem fetching CLA signatures, error: %+v", sigErr)
			return nil, sigErr
		}
		if sig != nil {
			sigs = append(sigs, sig)
		}
	}

	claManagers := make([]*models.CompanyClaManager, 0)
	lfUsernames := utils.NewStringSet()

	// Get CLA managers
	for _, sig := range sigs {
		if _, ok := claGroups[sig.ProjectID]; !ok {
			continue
		}
		for _, user := range sig.SignatureACL {
			claManagers = append(claManagers, &models.CompanyClaManager{
				// DB doesn't have approved_on value
				ApprovedOn:        sig.SignatureCreated,
				LfUsername:        user.LfUsername,
				ProjectID:         sig.ProjectID,
				OrganizationSfid:  v1CompanyModel.CompanyExternalID,
				OrganizationID:    v1CompanyModel.CompanyID,
				OrganizationName:  orgModel.Name,
				SigningEntityName: v1CompanyModel.SigningEntityName,
			})
			lfUsernames.Add(user.LfUsername)
		}
	}

	if len(claManagers) == 0 {
		return &models.CompanyClaManagers{List: claManagers}, nil
	}

	// get userinfo and project info
	var usermap map[string]*v2UserServiceModels.User
	usermap, err = getUsersInfo(lfUsernames.List())
	if err != nil {
		log.WithFields(f).Warnf("problem fetching users information, error: %+v", err)
		return nil, err
	}

	// fill user info
	fillUsersInfo(claManagers, usermap)
	// fill project info
	fillProjectInfo(claManagers, claGroups)
	// fetch the cla_manager.added events so can fill the addedOn field
	claManagerAddedEvents, err := s.eventService.GetCompanyEvents(v1CompanyModel.CompanyID, events.ClaManagerCreated, nil, aws.Int64(100), true)
	if err != nil {
		log.WithFields(f).Warnf("fetching events for companyID failed : %s : %v", v1CompanyModel.CompanyID, err)
		return nil, err
	}
	// fill events info
	fillEventsInfo(claManagers, claManagerAddedEvents)

	log.WithFields(f).Debug("sorting response for client")
	// sort result by cla manager name
	sort.Slice(claManagers, func(i, j int) bool {
		return claManagers[i].Name < claManagers[j].Name
	})

	return &models.CompanyClaManagers{List: claManagers}, nil
}

func fillEventsInfo(claManagers []*v2Models.CompanyClaManager, addedEvents *v1Models.EventList) {
	eventMap := map[string]*v1Models.Event{}
	if addedEvents == nil || len(addedEvents.Events) == 0 {
		return
	}
	for _, e := range addedEvents.Events {
		lfUserName := e.LfUsername
		eventMap[lfUserName] = e
	}

	for _, claManager := range claManagers {
		if eventMap[claManager.LfUsername] == nil {
			continue
		}
		claManager.AddedOn = eventMap[claManager.LfUsername].EventTime
	}
}

func (s *service) GetCompanyAdmins(ctx context.Context, companySFID string) (*models.CompanyAdminList, error) {
	f := logrus.Fields{
		"functionName":   "GetCompanyAdmins",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"companySFID":    companySFID,
	}
	orgClient := orgService.GetClient()

	log.WithFields(f).Info("Getting user admins for company")
	admins, adminErr := orgClient.ListOrgUserAdminScopes(ctx, companySFID, nil)
	adminList := make([]*models.AdminSf, 0)
	if adminErr != nil {
		if _, ok := adminErr.(*organizations.ListOrgUsrAdminScopesNotFound); ok {
			log.WithFields(f).Info(" No admins found ")
			return &models.CompanyAdminList{
				List: adminList,
			}, nil
		}
		return nil, adminErr
	}

	// if 404 and no error parse the userroles list
	for _, userRole := range admins.Userroles {
		adminList = append(adminList, &models.AdminSf{
			Email:    userRole.Contact.EmailAddress,
			Username: userRole.Contact.Username,
			ID:       userRole.Contact.ID,
		})
	}

	return &models.CompanyAdminList{
		List: adminList,
	}, nil
}

func (s *service) GetCompanyProjectActiveCLAs(ctx context.Context, companyID string, projectSFID string) (*models.ActiveClaList, error) {
	f := logrus.Fields{
		"functionName":   "GetCompanyProjectActiveCLAs",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"projectSFID":    projectSFID,
		"companyID":      companyID,
	}
	var err error
	claGroups, err := s.getCLAGroupsUnderProjectOrFoundation(ctx, projectSFID)
	if err != nil {
		log.WithFields(f).Warnf("problem fetching CLA Groups under project or foundation, error: %+v", err)
		return nil, err
	}
	var out models.ActiveClaList
	sigs, err := s.getAllCCLASignatures(ctx, companyID)
	if err != nil {
		log.WithFields(f).Warnf("problem fetching CCLA signatures, error: %+v", err)
		return nil, err
	}
	out.List = make([]*models.ActiveCla, 0, len(sigs))
	if len(sigs) == 0 {
		return &out, nil
	}
	var wg sync.WaitGroup
	wg.Add(len(sigs))
	for _, sig := range sigs {
		if _, ok := claGroups[sig.ProjectID]; !ok {
			// skip the cla_group which are not under current foundation/project
			wg.Done()
			continue
		}
		activeCla := &models.ActiveCla{}
		out.List = append(out.List, activeCla)
		go func(swg *sync.WaitGroup, signature *v1Models.Signature, acla *models.ActiveCla) {
			s.fillActiveCLA(ctx, swg, signature, acla, claGroups, companyID)
		}(&wg, sig, activeCla)
	}
	wg.Wait()
	return &out, nil
}

func (s *service) GetCompanyProjectContributors(ctx context.Context, projectSFID string, companyID string, searchTerm string) (*models.CorporateContributorList, error) {
	f := logrus.Fields{
		"functionName":   "GetCompanyProjectContributors",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"projectSFID":    projectSFID,
		"companyID":      companyID,
		"searchTerm":     searchTerm,
	}
	list := make([]*models.CorporateContributor, 0)
	sigs, err := s.getAllCompanyProjectEmployeeSignatures(ctx, companyID, projectSFID)
	if err != nil {
		log.WithFields(f).Warnf("problem fetching all company project employee signatures, error: %+v", err)
		return nil, err
	}
	if len(sigs) == 0 {
		return &models.CorporateContributorList{
			List: list,
		}, nil
	}
	var wg sync.WaitGroup
	result := make(chan *models.CorporateContributor)
	wg.Add(len(sigs))
	go func() {
		wg.Wait()
		close(result)
	}()

	for _, sig := range sigs {
		go fillCorporateContributorModel(&wg, s.userRepo, sig, result, searchTerm)
	}

	for corpContributor := range result {
		list = append(list, corpContributor)
	}

	return &models.CorporateContributorList{
		List: list,
	}, nil
}

func (s *service) CreateCompany(ctx context.Context, companyName, signingEntityName, companyWebsite, userEmail, userID string) (*models.CompanyOutput, error) {
	f := logrus.Fields{
		"functionName":      "CreateCompany",
		utils.XREQUESTID:    ctx.Value(utils.XREQUESTID),
		"companyName":       companyName,
		"signingEntityName": signingEntityName,
		"companyWebsite":    companyWebsite,
		"userEmail":         userEmail,
		"userID":            userID,
	}
	var lfUser *v2UserServiceModels.User

	// Create SalesForce company
	orgClient := orgService.GetClient()
	log.WithFields(f).Debugf("Creating Organization: %s, Signing Entity Name: %s, Website: %s", companyName, signingEntityName, companyWebsite)
	org, err := orgClient.CreateOrg(ctx, companyName, signingEntityName, companyWebsite)
	if err != nil {
		log.WithFields(f).Warnf("unable to create platform organization service, error: %+v", err)
		return nil, err
	}

	acsClient := acs_service.GetClient()
	userClient := v2UserService.GetClient()

	lfUser, lfErr := userClient.SearchUserByEmail(userEmail)
	if lfErr != nil {
		msg := fmt.Sprintf("User : %s has no LFID", userEmail)
		log.WithFields(f).Warn(msg)
	}
	if lfUser != nil && lfUser.Username == "" {
		msg := fmt.Sprintf("User: %s has no LF username", userEmail)
		log.WithFields(f).Warn(msg)
	}
	if lfUser != nil && lfUser.Username != "" {
		log.WithFields(f).Debugf("User :%s has been assigned the %s role to organization: %s ",
			userEmail, utils.CompanyAdminRole, org.Name)
		// Assign company-admin to user
		roleID, adminErr := acsClient.GetRoleID(utils.CompanyAdminRole)
		if adminErr != nil {
			msg := "Problem getting companyAdmin role ID for contributor"
			log.WithFields(f).Warn(msg)
			return nil, adminErr
		}

		scopeErr := orgClient.CreateOrgUserRoleOrgScope(ctx, userEmail, org.ID, roleID)
		if scopeErr != nil {
			msg := fmt.Sprintf("Problem creating Org scope for email: %s , companyID: %s", userEmail, org.ID)
			log.WithFields(f).Warn(msg)
			// Ignore conflict - role has already been assigned - otherwise, return error
			if !strings.Contains(scopeErr.Error(), OrgAssociated) {
				return nil, scopeErr
			}
		}
		// Associate User (Not associated) with Newly created org
		if lfUser.Account.ID == NoAccount {
			lfUser.Account.ID = org.ID
		}
	}

	// Create Easy CLA Company
	log.WithFields(f).Debugf("Creating EasyCLA company: %s ", companyName)

	if signingEntityName == "" {
		log.WithFields(f).Debugf("Setting signing entity with company name value :%s ", companyName)
		signingEntityName = companyName
	}
	// OrgID used as externalID for the easyCLA Company
	// Create a new company model for the create function
	createCompanyModel := &v1Models.Company{
		CompanyACL:        nil,
		CompanyExternalID: org.ID,
		CompanyManagerID:  userID,
		CompanyName:       companyName,
		SigningEntityName: signingEntityName,
	}

	_, createErr := s.companyRepo.CreateCompany(ctx, createCompanyModel)
	//easyCLAErr := s.repo.CreateCompany(companyName, org.ID, userID)
	if createErr != nil {
		log.WithFields(f).Warnf("Failed to create EasyCLA company for company: %s, error: %+v",
			companyName, createErr)
		return nil, createErr
	}

	return &models.CompanyOutput{
		CompanyName:       org.Name,
		SigningEntityName: signingEntityName,
		CompanyWebsite:    companyWebsite,
		LogoURL:           org.LogoURL,
		CompanyID:         org.ID,
	}, nil
}

// GetCompanyByName deletes the company by name
func (s *service) GetCompanyByName(ctx context.Context, companyName string) (*models.Company, error) {
	f := logrus.Fields{
		"functionName":   "GetCompanyByName",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"companyName":    companyName,
	}
	companyModel, err := s.companyRepo.GetCompanyByName(ctx, companyName)
	if err != nil {
		return nil, err
	}

	if companyModel == nil {
		log.WithFields(f).Debugf("search by company name: %s didn't locate the record", companyName)
		return nil, nil
	}

	// Convert from v1 to v2 model - use helper: Copy(toValue interface{}, fromValue interface{})
	var v2CompanyModel v2Models.Company
	copyErr := copier.Copy(&v2CompanyModel, &companyModel)
	if copyErr != nil {
		log.WithFields(f).Warnf("problem converting v1 company model to a v2 company model, error: %+v", copyErr)
		return nil, copyErr
	}

	return &v2CompanyModel, nil
}

// GetCompanyBySigningEntityName retrieves the company by signing entity name
func (s *service) GetCompanyBySigningEntityName(ctx context.Context, signingEntityName string) (*models.Company, error) {
	f := logrus.Fields{
		"functionName":      "company.service.GetCompanyBySigningEntityName",
		utils.XREQUESTID:    ctx.Value(utils.XREQUESTID),
		"signingEntityName": signingEntityName,
	}

	log.WithFields(f).Warn("looking up company record by signing entity name...")
	companyModel, err := s.companyRepo.GetCompanyBySigningEntityName(ctx, signingEntityName)
	if err != nil {
		if _, ok := err.(*utils.CompanyNotFound); ok { // nolint
			// As a backup, in case the signing entity name was not set on the old records, lookup the company by it's normal name
			log.WithFields(f).Debugf("signing entity name not found. as a backup, searching company by name using signing entity name value: %s", signingEntityName)
			companyModel, err = s.companyRepo.GetCompanyByName(ctx, signingEntityName)
			if err != nil {
				log.WithFields(f).WithError(err).Warn("unable to lookup company name by attempting to use the signing entity name")
				return nil, err
			}
		} else {
			log.WithFields(f).WithError(err).Warn("unable to lookup company by signing entity name")
			return nil, err
		}
	}

	if companyModel == nil {
		log.WithFields(f).Debugf("search by company signing entity name: %s didn't locate the record", signingEntityName)
		// As a backup, in case the signing entity name was not set on the old records, lookup the company by it's normal name
		log.WithFields(f).Debugf("as a backup, searching company by name using signing entity name value: %s", signingEntityName)
		companyModel, err = s.companyRepo.GetCompanyByName(ctx, signingEntityName)
		if err != nil {
			log.WithFields(f).WithError(err).Warn("unable to lookup company name by attempting to use the signing entity name")
			return nil, err
		}

		if companyModel == nil {
			log.WithFields(f).Debugf("search by company name: %s didn't locate the record", signingEntityName)
			return nil, nil
		}
	}

	// Convert from v1 to v2 model - use helper: Copy(toValue interface{}, fromValue interface{})
	var v2CompanyModel v2Models.Company
	copyErr := copier.Copy(&v2CompanyModel, &companyModel)
	if copyErr != nil {
		log.WithFields(f).Warnf("problem converting v1 company model to a v2 company model, error: %+v", copyErr)
		return nil, copyErr
	}

	return &v2CompanyModel, nil
}

// GetCompanyByID retrieves the company by internal ID
func (s *service) GetCompanyByID(ctx context.Context, companyID string) (*models.Company, error) {
	f := logrus.Fields{
		"functionName":   "company.service.GetCompanyByID",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"companyID":      companyID,
	}
	companyModel, err := s.companyRepo.GetCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	if companyModel == nil {
		log.WithFields(f).Debugf("search by company ID: %s didn't locate the record", companyID)
		return nil, nil
	}

	// Convert from v1 to v2 model - use helper: Copy(toValue interface{}, fromValue interface{})
	var v2CompanyModel v2Models.Company
	copyErr := copier.Copy(&v2CompanyModel, &companyModel)
	if copyErr != nil {
		log.WithFields(f).Warnf("problem converting v1 company model to a v2 company model, error: %+v", copyErr)
		return nil, copyErr
	}

	return &v2CompanyModel, nil
}

func (s *service) AssociateContributor(ctx context.Context, companySFID string, userEmail string) (*models.Contributor, error) {
	f := logrus.Fields{
		"functionName":   "company.service.AssociateContributor",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"companySFID":    companySFID,
		"userEmail":      userEmail,
	}

	orgClient := orgService.GetClient()

	userService := v2UserService.GetClient()
	log.WithFields(f).Info("searching for LFX User")
	lfxUser, userErr := userService.SearchUserByEmail(userEmail)
	if userErr != nil {
		log.WithFields(f).Warnf("unable to get user")
		return nil, userErr
	}

	acsServiceClient := acs_service.GetClient()

	log.WithFields(f).Info("Getting roleID for the contributor role")
	roleID, roleErr := acsServiceClient.GetRoleID("contributor")
	if roleErr != nil {
		log.WithFields(f).Warn("Problem getting roleID for contributor role ")
		return nil, roleErr
	}

	log.WithFields(f).Info("creating contributor role scope")
	scopeErr := orgClient.CreateOrgUserRoleOrgScope(ctx, userEmail, companySFID, roleID)
	if scopeErr != nil {
		log.WithFields(f).Warnf("Problem creating role scope")
		return nil, scopeErr
	}

	contributor := &models.Contributor{
		LfUsername:  lfxUser.Username,
		UserSfid:    lfxUser.ID,
		Email:       strfmt.Email(userEmail),
		AssignedOn:  time.Now().String(),
		CompanySfid: companySFID,
		Role:        *aws.String("contributor"),
	}

	return contributor, nil
}

//CreateContributor creates contributor for contributor prospect
func (s *service) CreateContributor(ctx context.Context, companyID string, projectID string, userEmail string, ClaGroupID string) (*models.Contributor, error) {
	f := logrus.Fields{
		"functionName":   "company.service.CreateContributor",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"companyID":      companyID,
		"projectID":      projectID,
		"ClaGroupID":     ClaGroupID,
		"userEmail":      userEmail,
	}
	// integrate user,acs,org and project services
	userClient := v2UserService.GetClient()
	acServiceClient := acs_service.GetClient()
	orgClient := orgService.GetClient()

	user, userErr := userClient.SearchUserByEmail(userEmail)
	if userErr != nil {
		log.WithFields(f).Debugf("Failed to get user by email: %s , error: %+v", userEmail, userErr)
		return nil, ErrLFXUserNotFound
	}

	// Check if user is already contributor of project|organization scope
	hasRoleScope, hasRoleScopeErr := orgClient.IsUserHaveRoleScope(ctx, "contributor", user.ID, companyID, projectID)
	if hasRoleScopeErr != nil {
		// Skip 404 for ListOrgUsrServiceScopes endpoint
		if _, ok := hasRoleScopeErr.(*organizations.ListOrgUsrServiceScopesNotFound); !ok {
			log.WithFields(f).Debugf("Failed to check roleScope: contributor  for user: %s", user.Username)
			return nil, hasRoleScopeErr
		}
	}
	if hasRoleScope {
		log.WithFields(f).Debugf("Conflict ")
		return nil, ErrContributorConflict
	}

	roleID, designeeErr := acServiceClient.GetRoleID("contributor")
	if designeeErr != nil {
		msg := "Problem getting role ID for contributor"
		log.Warn(msg)
		return nil, designeeErr
	}

	scopeErr := orgClient.CreateOrgUserRoleOrgScopeProjectOrg(ctx, userEmail, projectID, companyID, roleID)
	if scopeErr != nil {
		msg := fmt.Sprintf("Problem creating projectOrg scope for email: %s , projectID: %s, companyID: %s", userEmail, projectID, companyID)
		log.Warn(msg)
		if _, ok := scopeErr.(*organizations.CreateOrgUsrRoleScopesConflict); ok {
			return nil, ErrRoleScopeConflict
		}
		return nil, scopeErr
	}

	v1CompanyModel, companyErr := s.v1CompanyService.GetCompanyByExternalID(ctx, companyID)
	if companyErr != nil {
		log.Error("company not found", companyErr)
	}

	projectModel, projErr := s.projectRepo.GetCLAGroupByID(ctx, ClaGroupID, DontLoadRepoDetails)
	if projErr != nil {
		msg := fmt.Sprintf("unable to query CLA Group ID: %s, error: %+v", ClaGroupID, projErr)
		log.WithFields(f).Warnf(msg)
	}

	// Log Event
	s.eventService.LogEvent(
		&events.LogEventArgs{
			EventType:         events.AssignUserRoleScopeType,
			LfUsername:        user.Username,
			UserID:            user.ID,
			ExternalProjectID: projectID,
			CompanyModel:      v1CompanyModel,
			ClaGroupModel:     projectModel,
			UserModel:         &v1Models.User{LfUsername: user.Username, UserID: user.ID},
			EventData: &events.AssignRoleScopeData{
				Role:  "contributor",
				Scope: fmt.Sprintf("%s|%s", projectID, companyID),
			},
		})

	contributor := &models.Contributor{
		LfUsername:  user.Username,
		UserSfid:    user.ID,
		AssignedOn:  time.Now().String(),
		Email:       strfmt.Email(userEmail),
		CompanySfid: companyID,
		Role:        *aws.String("contributor"),
	}
	return contributor, nil
}

//AssociateContributorByGroup creates contributor by group for contributor prospect
func (s *service) AssociateContributorByGroup(ctx context.Context, companySFID, userEmail string, projectCLAGroups []*projects_cla_groups.ProjectClaGroup, ClaGroupID string) ([]*models.Contributor, string, error) {
	f := logrus.Fields{
		"functionName":   "company.service.AssociateContributorByGroup",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"companySFID":    companySFID,
		"ClaGroupID":     ClaGroupID,
		"userEmail":      userEmail,
	}
	var contributors []*models.Contributor
	foundationSFID := projectCLAGroups[0].FoundationSFID
	if foundationSFID != "" {
		contributor, err := s.CreateContributor(ctx, companySFID, foundationSFID, userEmail, ClaGroupID)
		if err != nil {
			if err == ErrContributorConflict {
				msg := fmt.Sprintf("Conflict assigning contributor role for Foundation SFID: %s ", foundationSFID)
				return nil, msg, err
			}
			msg := fmt.Sprintf("Creating contributor failed for Foundation SFID: %s ", foundationSFID)
			return nil, msg, err
		}
		contributors = append(contributors, contributor)
	}

	for _, pcg := range projectCLAGroups {
		log.WithFields(f).Debugf("creating contributor for Project SFID: %s", pcg.ProjectSFID)
		if foundationSFID != pcg.ProjectSFID {
			contributor, err := s.CreateContributor(ctx, companySFID, pcg.ProjectSFID, userEmail, ClaGroupID)
			if err != nil {
				if err == ErrContributorConflict {
					msg := fmt.Sprintf("Conflict assigning contributor role for Project SFID: %s ", pcg.ProjectSFID)
					return nil, msg, err
				}
				msg := fmt.Sprintf("Creating contributor failed for Project SFID: %s ", pcg.ProjectSFID)
				return nil, msg, err
			}
			contributors = append(contributors, contributor)
		}

	}
	return contributors, "", nil
}

// GetCompanyBySFID retrieves the company by external SFID
func (s *service) GetCompanyBySFID(ctx context.Context, companySFID string) (*models.Company, error) {
	f := logrus.Fields{
		"functionName":   "company.service.GetCompanyBySFID",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"companySFID":    companySFID,
	}
	companyModel, err := s.companyRepo.GetCompanyByExternalID(ctx, companySFID)
	if err != nil {
		// If we were unable to find the company/org in our local database, try to auto-create based
		// on the existing SF record
		if _, ok := err.(*utils.CompanyNotFound); ok {
			log.WithFields(f).Debug("company not found in EasyCLA database - attempting to auto-create from platform organization service record")
			newCompanyModel, createCompanyErr := s.autoCreateCompany(ctx, companySFID)
			if createCompanyErr != nil {
				log.WithFields(f).Warnf("problem creating company from platform organization SF record, error: %+v",
					createCompanyErr)
				return nil, createCompanyErr
			}
			if newCompanyModel == nil {
				log.WithFields(f).Warnf("problem creating company from SF records - created model is nil")
				return nil, &utils.CompanyNotFound{
					Message:     "unable to auto-create company",
					CompanySFID: companySFID,
				}
			}
			// Success, fall through and continue processing
			companyModel = newCompanyModel
		}
		return nil, err
	}

	if companyModel == nil {
		log.WithFields(f).Debugf("search by company SFID didn't locate the record")
		return nil, nil
	}

	// Convert from v1 to v2 model - use helper: Copy(toValue interface{}, fromValue interface{})
	var v2CompanyModel v2Models.Company
	copyErr := copier.Copy(&v2CompanyModel, &companyModel)
	if copyErr != nil {
		log.WithFields(f).Warnf("problem converting v1 company model to a v2 company model, error: %+v", copyErr)
		return nil, copyErr
	}

	return &v2CompanyModel, nil
}

// DeleteCompanyByID deletes the company by ID
func (s *service) DeleteCompanyByID(ctx context.Context, companyID string) error {
	return s.companyRepo.DeleteCompanyByID(ctx, companyID)
}

// DeleteCompanyBySFID deletes the company by SFID
func (s *service) DeleteCompanyBySFID(ctx context.Context, companyID string) error {
	return s.companyRepo.DeleteCompanyBySFID(ctx, companyID)
}

func (s *service) GetCompanyProjectCLA(ctx context.Context, authUser *auth.User, companySFID, projectSFID string, companyID *string) (*models.CompanyProjectClaList, error) {
	f := logrus.Fields{
		"functionName":   "company.service.GetCompanyProjectCLA",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"authUserName":   authUser.UserName,
		"authUserEmail":  authUser.Email,
		"companySFID":    companySFID,
		"projectSFID":    projectSFID,
	}
	var canSign bool
	resources := authUser.ResourceIDsByTypeAndRole(auth.ProjectOrganization, utils.CLADesigneeRole)
	projectOrg := fmt.Sprintf("%s|%s", projectSFID, companySFID)
	for _, r := range resources {
		if r == projectOrg {
			canSign = true
			break
		}
	}

	// Attempt to locate the company model in our database
	log.WithFields(f).Debug("locating company by SF ID")
	var companyModel *v1Models.Company
	var companies []*v1Models.Company
	var companyErr error
	companies, companyErr = s.companyRepo.GetCompaniesByExternalID(ctx, companySFID)
	if companyErr != nil {
		// If we were unable to find the company/org in our local database, try to auto-create based
		// on the existing SF record
		if _, ok := companyErr.(*utils.CompanyNotFound); ok { // nolint
			log.WithFields(f).WithError(companyErr).Debug("company not found in EasyCLA database - attempting to auto-create from platform organization service record")
			var createCompanyErr error
			companyModel, createCompanyErr = s.autoCreateCompany(ctx, companySFID)
			if createCompanyErr != nil {
				log.WithFields(f).WithError(createCompanyErr).Warn("problem creating company from platform organization SF record")
				return nil, createCompanyErr
			}
			if companyModel == nil {
				log.WithFields(f).Warnf("problem creating company from SF records - created model is nil")
				return nil, &utils.CompanyNotFound{
					Message:     "unable to auto-create company",
					CompanySFID: companySFID,
				}
			}
			// Success, fall through and continue processing
		} else {
			log.WithFields(f).WithError(companyErr).Warnf("problem fetching company by SFID")
			return nil, companyErr
		}
	}

	claGroups, err := s.getCLAGroupsUnderProjectOrFoundation(ctx, projectSFID)
	if err != nil {
		log.WithFields(f).Warnf("problem fetching CLA Groups under project or foundation, error: %+v", err)
		return nil, err
	}

	var companyProjectClaList = make([]*models.CompanyProjectCla, 0)
	if companyID != nil {
		log.WithFields(f).Debugf("Filtering company for ID: %s ", *companyID)
		index, found := findCompany(companies, *companyID)
		if found {
			log.WithFields(f).Debugf("Found company: %v ", companies[index])
			companies = []*v1Models.Company{companies[index]}
		} else {
			companies = []*v1Models.Company{}
		}
	}
	for _, company := range companies {
		activeCLAList, err := s.GetCompanyProjectActiveCLAs(ctx, company.CompanyID, projectSFID)
		if err != nil {
			log.WithFields(f).Warnf("problem fetching company project active CLAs, error: %+v", err)
			return nil, err
		}
		var companyProjectCLA = &models.CompanyProjectCla{
			SignedClaList:       activeCLAList.List,
			UnsignedProjectList: make([]*models.UnsignedProject, 0),
		}
		for _, activeCLA := range activeCLAList.List {
			// remove cla groups for which we have signed cla
			log.WithFields(f).Debugf("removing CLA Groups with active CLA, CLA Group: %+v, error: %+v", activeCLA, err)
			delete(claGroups, activeCLA.ProjectID)
		}
		// Get Company details
		company, compErr := s.GetCompanyByID(ctx, company.CompanyID)
		if compErr != nil {
			log.WithFields(f).WithError(compErr).Warnf("unable to fetch company by ID: %s ", company.CompanyID)
			return nil, compErr
		}
		// fill details for not signed cla
		for claGroupID, claGroup := range claGroups {
			unsignedProject := &models.UnsignedProject{
				CompanyName:       company.CompanyName,
				SigningEntityName: company.SigningEntityName,
				SigningEntityID:   company.CompanyID,
				CanSign:           canSign,
				ClaGroupID:        claGroupID,
				ClaGroupName:      claGroup.ClaGroupName,
				ProjectName:       claGroup.ProjectName,
				ProjectSfid:       claGroup.ProjectSFID,
				SubProjects:       claGroup.SubProjects,
				IclaEnabled:       claGroup.IclaEnabled,
				CclaEnabled:       claGroup.CclaEnabled,
			}
			log.WithFields(f).Debugf("adding unsigned CLA Group: %+v, error: %+v", unsignedProject, err)
			companyProjectCLA.UnsignedProjectList = append(companyProjectCLA.UnsignedProjectList, unsignedProject)
		}
		companyProjectClaList = append(companyProjectClaList, companyProjectCLA)

		// refresh clagroups for next company instance
		claGroups, err = s.getCLAGroupsUnderProjectOrFoundation(ctx, projectSFID)
		if err != nil {
			log.WithFields(f).Warnf("problem fetching CLA Groups under project or foundation, error: %+v", err)
			return nil, err
		}
	}

	return &models.CompanyProjectClaList{
		List: companyProjectClaList,
	}, nil
}

// GetCompanyCLAGroupManagers when provided the internal company ID and CLA Groups ID, this routine returns the list of
// corresponding CLA managers
func (s *service) GetCompanyCLAGroupManagers(ctx context.Context, companyID, claGroupID string) (*models.CompanyClaManagers, error) {
	f := logrus.Fields{
		"functionName":   "company.service.GetCompanyCLAGroupManagers",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"companyID":      companyID,
		"claGroupID":     claGroupID,
	}
	signed, approved := true, true
	pageSize := int64(10)
	sigModel, err := s.signatureRepo.GetProjectCompanySignature(ctx, companyID, claGroupID, &signed, &approved, nil, &pageSize)
	if err != nil {
		log.WithFields(f).Warnf("unable to query CCLA signature using Company ID: %s and CLA Group ID: %s, signed: true, approved: true, error: %+v",
			companyID, claGroupID, err)
		return nil, err
	}

	if sigModel == nil {
		log.WithFields(f).Warnf("unable to query CCLA signature using Company ID: %s and CLA Group ID: %s, signed: true, approved: true - no signature found",
			companyID, claGroupID)
		return &models.CompanyClaManagers{}, nil
	}

	projectModel, projErr := s.projectRepo.GetCLAGroupByID(ctx, claGroupID, DontLoadRepoDetails)
	if projErr != nil {
		log.WithFields(f).Warnf("unable to query CLA Group ID: %s, error: %+v", claGroupID, err)
		return nil, err
	}

	if projectModel == nil {
		log.WithFields(f).Warnf("unable to query CLA Group ID: %s - no CLA Group found", claGroupID)
		return nil, nil
	}

	companyModel, companyErr := s.companyRepo.GetCompany(ctx, companyID)
	if companyErr != nil {
		log.WithFields(f).Warnf("unable to query Company ID: %s, error: %+v", companyID, companyErr)
		return nil, err
	}

	if companyModel == nil {
		log.WithFields(f).Warnf("unable to query Company ID: %s - no company by ID found", companyID)
		return nil, nil
	}

	claManagers := make([]*models.CompanyClaManager, 0)
	for _, user := range sigModel.SignatureACL {
		claManagers = append(claManagers, &models.CompanyClaManager{
			// DB doesn't have approved_on value - just use sig created date/time
			ApprovedOn:       sigModel.SignatureCreated,
			LfUsername:       user.LfUsername,
			Email:            strfmt.Email(user.LfEmail),
			Name:             user.Username,
			UserSfid:         user.UserExternalID,
			ProjectID:        sigModel.ProjectID,
			ProjectName:      projectModel.ProjectName,
			ClaGroupName:     projectModel.ProjectName,
			OrganizationName: companyModel.CompanyName,
			OrganizationSfid: companyModel.CompanyExternalID,
		})
	}

	return &models.CompanyClaManagers{List: claManagers}, nil
}

func v2ProjectToMap(projectDetails *v2ProjectServiceModels.ProjectOutputDetailed) (map[string]*v2ProjectServiceModels.ProjectOutput, error) {
	epmap := make(map[string]*v2ProjectServiceModels.ProjectOutput) // key project_sfid
	var pr v2ProjectServiceModels.ProjectOutput
	err := copier.Copy(&pr, projectDetails)
	if err != nil {
		return nil, err
	}
	epmap[projectDetails.ID] = &pr
	for _, p := range projectDetails.Projects {
		epmap[p.ID] = p
	}
	return epmap, nil
}

func (s *service) getCLAGroupsUnderProjectOrFoundation(ctx context.Context, id string) (map[string]*claGroupModel, error) {
	result := make(map[string]*claGroupModel)
	psc := v2ProjectService.GetClient()
	projectDetails, err := psc.GetProject(id)
	if err != nil {
		return nil, err
	}
	var allProjectMapping []*projects_cla_groups.ProjectClaGroup
	if projectDetails.ProjectType == FoundationType {
		// get all projects for all cla group under foundation
		allProjectMapping, err = s.projectClaGroupsRepo.GetProjectsIdsForFoundation(id)
		if err != nil {
			return nil, err
		}
	} else {
		// get cla group id from project
		projectMapping, perr := s.projectClaGroupsRepo.GetClaGroupIDForProject(id)
		if perr != nil {
			return nil, err
		}
		// get all projects for that cla group
		allProjectMapping, err = s.projectClaGroupsRepo.GetProjectsIdsForClaGroup(projectMapping.ClaGroupID)
		if err != nil {
			return nil, err
		}
		if len(allProjectMapping) > 1 {
			// reload data in projectDetails for all projects of foundation
			projectDetails, err = psc.GetProject(projectDetails.Foundation.ID)
			if err != nil {
				return nil, err
			}
		}
	}
	// v2ProjectMap will contains projectSFID -> salesforce details of that project
	v2ProjectMap, err := v2ProjectToMap(projectDetails)
	if err != nil {
		return nil, err
	}
	// for all cla-groups create claGroupModel
	for _, pm := range allProjectMapping {
		cg, ok := result[pm.ClaGroupID]
		if !ok {
			cg = &claGroupModel{
				FoundationSFID: pm.FoundationSFID,
				SubProjects:    make([]string, 0),
			}
			result[pm.ClaGroupID] = cg
		}
		cg.SubProjectIDs = append(cg.SubProjectIDs, pm.ProjectSFID)
	}
	// if no cla-group found, return empty result
	if len(result) == 0 {
		return result, nil
	}
	var wg sync.WaitGroup
	wg.Add(len(result))
	for id, cg := range result {
		go func(claGroupID string, claGroup *claGroupModel) {
			defer wg.Done()
			// get cla-group info
			cginfo, err := s.projectRepo.GetCLAGroupByID(ctx, claGroupID, DontLoadRepoDetails)
			if err != nil || cginfo == nil {
				log.Warnf("Unable to get details of cla_group: %s", claGroupID)
				return
			}
			claGroup.ClaGroupName = cginfo.ProjectName
			claGroup.ClaGroupID = cginfo.ProjectID
			claGroup.IclaEnabled = cginfo.ProjectICLAEnabled
			claGroup.CclaEnabled = cginfo.ProjectCCLAEnabled

			var pid string
			if len(claGroup.SubProjectIDs) == 1 {
				// use project info if cla-group have only one project
				pid = claGroup.SubProjectIDs[0]
			} else {
				// use foundation info if cla-group have multiple project
				pid = claGroup.FoundationSFID
				for _, spid := range claGroup.SubProjectIDs {
					subProject, ok := v2ProjectMap[spid]
					if !ok {
						log.Warnf("Unable to fill details for cla_group: %s with project details of %s", claGroupID, spid)
						return
					}
					claGroup.SubProjects = append(claGroup.SubProjects, subProject.Name)
				}
			}
			project, ok := v2ProjectMap[pid]
			if !ok {
				log.Warnf("Unable to fill details for cla_group: %s with project details of %s", claGroupID, claGroup.ProjectSFID)
				return
			}
			claGroup.ProjectLogo = project.ProjectLogo
			claGroup.ProjectName = project.Name
			claGroup.ProjectType = project.ProjectType
			claGroup.ProjectSFID = pid
		}(id, cg)
	}
	wg.Wait()
	return result, nil
}

func (s *service) getAllCCLASignatures(ctx context.Context, companyID string) ([]*v1Models.Signature, error) {
	f := logrus.Fields{
		"functionName":   "company.service.getAllCCLASignatures",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"companyID":      companyID,
	}
	log.WithFields(f).Debug("getAllCCLASignatures")
	var sigs []*v1Models.Signature
	var lastScannedKey *string
	for {
		sigModels, err := s.signatureRepo.GetCompanySignatures(ctx, v1SignatureParams.GetCompanySignaturesParams{
			CompanyID:     companyID,
			SignatureType: aws.String("ccla"),
			NextKey:       lastScannedKey,
		}, HugePageSize, signatures.DontLoadACLDetails)
		if err != nil {
			return nil, err
		}
		sigs = append(sigs, sigModels.Signatures...)
		if sigModels.LastKeyScanned == "" {
			break
		}
		lastScannedKey = aws.String(sigModels.LastKeyScanned)
	}
	return sigs, nil
}

func getUsersInfo(lfUsernames []string) (map[string]*v2UserServiceModels.User, error) {
	userMap := make(map[string]*v2UserServiceModels.User)
	if len(lfUsernames) == 0 {
		return userMap, nil
	}
	userServiceClient := v2UserService.GetClient()
	userModels, err := userServiceClient.GetUsersByUsernames(lfUsernames)
	if err != nil {
		return nil, err
	}
	for _, user := range userModels {
		userMap[user.Username] = user
	}
	return userMap, nil
}

func fillUsersInfo(claManagers []*models.CompanyClaManager, usermap map[string]*v2UserServiceModels.User) {
	f := logrus.Fields{
		"functionName": "company.service.fillUsersInfo",
	}
	log.WithFields(f).Debug("filling users info...")

	for _, cm := range claManagers {
		user, ok := usermap[cm.LfUsername]
		if !ok {
			log.WithFields(f).Warnf("Unable to get user with username %s", cm.LfUsername)
			continue
		}
		cm.Name = user.Name
		// cm.LogoURL = user.LogoURL
		cm.UserSfid = user.ID
		for _, email := range user.Emails {
			if email != nil && email.IsPrimary != nil && *email.IsPrimary && email.EmailAddress != nil {
				cm.Email = strfmt.Email(*email.EmailAddress)
				break
			}
		}
	}
}

func fillProjectInfo(claManagers []*models.CompanyClaManager, claGroups map[string]*claGroupModel) {
	f := logrus.Fields{
		"functionName": "company.service.fillProjectInfo",
	}
	log.WithFields(f).Debug("filling project info...")
	for _, claManager := range claManagers {
		cg, ok := claGroups[claManager.ProjectID]
		if !ok {
			continue
		}
		claManager.ClaGroupName = cg.ClaGroupName
		claManager.ProjectSfid = cg.ProjectSFID
		claManager.ProjectName = cg.ProjectName
	}
}

func (s *service) fillActiveCLA(ctx context.Context, wg *sync.WaitGroup, sig *v1Models.Signature, activeCla *models.ActiveCla, claGroups map[string]*claGroupModel, companyID string) {
	f := logrus.Fields{
		"functionName":   "v1CompanyModel.service.fillActiveCLA",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"companyID":      companyID,
	}
	defer wg.Done()
	cg, ok := claGroups[sig.ProjectID]
	if !ok {
		log.Warn("unable to get project details")
		return
	}

	// Get Company details
	v1CompanyModel, compErr := s.GetCompanyByID(ctx, companyID)
	if compErr != nil {
		log.WithFields(f).WithError(compErr).Warnf("unable to fetch v1CompanyModel by ID: %s ", companyID)
		return
	}

	// Update acl
	var acl = make([]string, 0)
	if len(sig.SignatureACL) > 0 {
		log.WithFields(f).Debugf("updating signature acl: %+v  list for lfusernames...", sig.SignatureACL)
		for _, manager := range sig.SignatureACL {
			if manager.LfUsername != "" {
				acl = append(acl, manager.LfUsername)
			}
		}
	}

	// fill details from dynamodb
	activeCla.CompanyName = v1CompanyModel.CompanyName
	if v1CompanyModel.SigningEntityName == "" {
		activeCla.SigningEntityName = v1CompanyModel.CompanyName
	} else {
		activeCla.SigningEntityName = v1CompanyModel.SigningEntityName
	}
	activeCla.ProjectID = sig.ProjectID
	if sig.SignedOn == "" {
		activeCla.SignedOn = sig.SignatureCreated
	} else {
		activeCla.SignedOn = sig.SignedOn
	}
	activeCla.SigningEntityID = companyID
	activeCla.SignatureACL = &models.ActiveClaSignatureACL{
		UsernameList: acl,
	}
	activeCla.ClaGroupName = cg.ClaGroupName
	activeCla.CompanyID = companyID
	activeCla.CompanySfid = v1CompanyModel.CompanyExternalID
	activeCla.SignatureID = sig.SignatureID

	// fill details from project service
	activeCla.ProjectName = cg.ProjectName
	activeCla.ProjectSfid = cg.ProjectSFID
	activeCla.ProjectType = cg.ProjectType
	activeCla.ProjectLogo = cg.ProjectLogo
	activeCla.SubProjects = cg.SubProjects
	var signatoryName string
	var cwg sync.WaitGroup
	cwg.Add(2)

	var cclaURL string
	go func() {
		var err error
		defer cwg.Done()
		cclaURL, err = utils.GetDownloadLink(utils.SignedCLAFilename(sig.ProjectID, sig.SignatureType, sig.SignatureReferenceID, sig.SignatureID))
		if err != nil {
			log.Error("fillActiveCLA : unable to get ccla s3 link", err)
			return
		}
	}()

	go func() {
		defer cwg.Done()
		if sig.SignatoryName != "" {
			signatoryName = sig.SignatoryName
			return
		}
		usc := v2UserService.GetClient()
		if len(sig.SignatureACL) == 0 {
			log.WithFields(f).Warnf("signature : %s have empty signature_acl", sig.SignatureID)
			return
		}
		lfUsername := sig.SignatureACL[0].LfUsername
		user, err := usc.GetUserByUsername(lfUsername)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("unable to get user with lf username : %s", lfUsername)
			return
		}
		signatoryName = user.Name
	}()

	cwg.Wait()

	activeCla.SignatoryName = signatoryName
	activeCla.CclaURL = cclaURL
}

// return projects output for which cla_group is present in cla
func (s *service) filterClaProjects(ctx context.Context, projects []*v2ProjectServiceModels.ProjectOutput) []*v2ProjectServiceModels.ProjectOutput { //nolint
	results := make([]*v2ProjectServiceModels.ProjectOutput, 0)
	prChan := make(chan *v2ProjectServiceModels.ProjectOutput)
	for _, v := range projects {
		go func(projectOutput *v2ProjectServiceModels.ProjectOutput) {
			project, err := s.projectRepo.GetCLAGroupsByExternalID(ctx, &v1ProjectParams.GetProjectsByExternalIDParams{
				ProjectSFID: projectOutput.ID,
				PageSize:    aws.Int64(1),
			}, DontLoadRepoDetails)
			if err != nil {
				log.Warnf("Unable to fetch project details for project with external id %s. error = %s", projectOutput.ID, err)
				prChan <- nil
				return
			}
			if project.ResultCount == 0 {
				prChan <- nil
				return
			}
			prChan <- projectOutput
		}(v)
	}
	for range projects {
		project := <-prChan
		if project != nil {
			results = append(results, project)
		}
	}
	return results
}

func fillCorporateContributorModel(wg *sync.WaitGroup, usersRepo users.UserRepository, sig *v1Models.Signature, result chan *models.CorporateContributor, searchTerm string) {
	defer wg.Done()
	user, err := usersRepo.GetUser(sig.SignatureReferenceID)
	if err != nil {
		log.Error("fillCorporateContributorModel: unable to get user info", err)
		return
	}
	if searchTerm != "" {
		ls := strings.ToLower(searchTerm)
		if !(strings.Contains(strings.ToLower(user.Username), ls) || strings.Contains(strings.ToLower(user.LfUsername), ls)) {
			return
		}
	}
	var contributor models.CorporateContributor
	var sigSignedTime = sig.SignatureCreated
	contributor.GithubID = user.GithubID
	contributor.LinuxFoundationID = user.LfUsername
	contributor.Name = user.Username
	t, err := utils.ParseDateTime(sig.SignatureCreated)
	if err != nil {
		log.Error("fillCorporateContributorModel: unable to parse time", err)
	} else {
		sigSignedTime = utils.TimeToString(t)
	}
	contributor.Timestamp = sigSignedTime
	contributor.SignatureVersion = fmt.Sprintf("v%s.%s", sig.SignatureMajorVersion, sig.SignatureMinorVersion)

	// send contributor struct on result channel
	result <- &contributor
}

func (s *service) getAllCompanyProjectEmployeeSignatures(ctx context.Context, companyID string, projectSFID string) ([]*v1Models.Signature, error) {
	f := logrus.Fields{
		"functionName":   "company.service.getAllCompanyProjectEmployeeSignatures",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"companyID":      companyID,
		"projectSFID":    projectSFID,
	}
	log.WithFields(f).Debug("getAllCompanyProjectEmployeeSignatures")
	_, claGroup, err := s.getCompanyAndClaGroup(ctx, companyID, projectSFID)
	if err != nil {
		return nil, err
	}
	params := v1SignatureParams.GetProjectCompanyEmployeeSignaturesParams{
		HTTPRequest: nil,
		CompanyID:   companyID,
		ProjectID:   claGroup.ProjectID,
	}
	sigs, err := s.signatureRepo.GetProjectCompanyEmployeeSignatures(ctx, params, HugePageSize)
	if err != nil {
		return nil, err
	}
	return sigs.Signatures, nil
}

// get company and project in parallel
func (s *service) getCompanyAndClaGroup(ctx context.Context, companyID, projectSFID string) (*v1Models.Company, *v1Models.ClaGroup, error) {
	f := logrus.Fields{
		"functionName":   "company.service.getCompanyAndClaGroup",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"companyID":      companyID,
		"projectSFID":    projectSFID,
	}
	var comp *v1Models.Company
	var claGroup *v1Models.ClaGroup
	var companyErr, projectErr error
	// query projects and company
	var cp sync.WaitGroup
	cp.Add(2)
	go func() {
		defer cp.Done()
		comp, companyErr = s.companyRepo.GetCompany(ctx, companyID)
	}()
	go func() {
		defer cp.Done()
		t := time.Now()
		var pm *projects_cla_groups.ProjectClaGroup
		pm, projectErr = s.projectClaGroupsRepo.GetClaGroupIDForProject(projectSFID)
		if projectErr != nil {
			log.WithFields(f).Debugf("cla group mapping not found for projectSFID %s", projectSFID)
			return
		}
		claGroup, projectErr = s.projectRepo.GetCLAGroupByID(ctx, pm.ClaGroupID, DontLoadRepoDetails)
		if claGroup == nil {
			projectErr = ErrProjectNotFound
		}
		log.WithField("time_taken", time.Since(t).String()).Debugf("getting project by external id : %s completed", projectSFID)
	}()
	cp.Wait()
	if companyErr != nil {
		return nil, nil, companyErr
	}
	if projectErr != nil {
		return nil, nil, projectErr
	}
	return comp, claGroup, nil
}

// autoCreateCompany helper function to create a new company record based on the SF ID and underlying record in SF
func (s service) autoCreateCompany(ctx context.Context, companySFID string) (*v1Models.Company, error) {
	f := logrus.Fields{
		"functionName":   "company.service.autoCreateCompany",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"companySFID":    companySFID,
	}
	// Get a reference to the platform organization service client
	orgClient := orgService.GetClient()
	log.WithFields(f).Debug("locating Organization in SF")

	// Lookup organization by ID in the Org Service
	sfOrgModel, sfOrgErr := orgClient.GetOrganization(ctx, companySFID)
	if sfOrgErr != nil {
		log.WithFields(f).Warnf("unable to locate platform organization record by SF ID, error: %+v", sfOrgErr)
		return nil, sfOrgErr
	}

	// If we were unable to lookup the company record in SF - we tried our best - return not exist error
	if sfOrgModel == nil {
		msg := "unable to locate platform organization record by SF ID - record not found"
		log.WithFields(f).Warn(msg)
		return nil, &utils.CompanyNotFound{
			Message:     msg,
			CompanySFID: companySFID,
		}
	}

	log.WithFields(f).Debug("found platform organization record in SF")
	// Auto-create based on the SF record information
	companyModel, companyCreateErr := s.companyRepo.CreateCompany(ctx, &v1Models.Company{
		CompanyExternalID: companySFID,
		CompanyName:       sfOrgModel.Name,

		Note: "created on-demand by v4 service based on SF Organization Service record",
	})

	if companyCreateErr != nil || companyModel == nil {
		log.WithFields(f).Warnf("unable to create EasyCLA company from platform SF organization record, error: %+v",
			companyCreateErr)
		return nil, companyCreateErr
	}

	log.WithFields(f).Debugf("successfully created EasyCLA company record: %+v", companyModel)
	return companyModel, nil
}

func (s *service) GetCompanyLookup(ctx context.Context, orgName string, websiteName string) (*models.Lookup, error) {
	f := logrus.Fields{
		"functionName":   "company.service.GetCompanyLookup",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"orgName":        orgName,
		"websiteName":    websiteName,
	}
	orgClient := orgService.GetClient()
	log.WithFields(f).Debug("Looking up organization by name and website")
	org, err := orgClient.SearchOrgLookup(ctx, &orgName, &websiteName)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("unable to lookup organization by name or website")
		return nil, err
	}

	var result *models.Lookup
	if org != nil {
		var signingEntityName []string
		// Sadly, search org lookup doesn't return signing entity name values
		if org.Payload.ID != "" {
			// lookup org with the signing entity name values
			log.WithFields(f).Debugf("Looking up organization by ID: %s", org.Payload.ID)
			orgModel, orgQueryErr := orgClient.GetOrganization(ctx, org.Payload.ID)
			if orgQueryErr != nil {
				log.WithFields(f).WithError(orgQueryErr).Warnf("unable to lookup organization using SFID: %s", org.Payload.ID)
			} else {
				log.WithFields(f).Debugf("Found organization by ID: %s", org.Payload.ID)
				// If we have any values in the project...
				if orgModel != nil && len(orgModel.SigningEntityName) > 0 {
					log.WithFields(f).Debugf("Adding signing entity name values to response: %s", strings.Join(orgModel.SigningEntityName, ";"))
					signingEntityName = orgModel.SigningEntityName
				} else {
					log.WithFields(f).Debug("No signing entity name values for Organization.")
				}
			}
		}

		result = &models.Lookup{
			Employees:          org.Payload.Employees,
			ID:                 org.Payload.ID,
			Industry:           org.Payload.Industry,
			Link:               org.Payload.Link,
			Name:               org.Payload.Name,
			Sector:             org.Payload.Sector,
			Source:             org.Payload.Source,
			SigningEntityNames: signingEntityName,
		}
	}
	return result, nil
}

func (s *service) RequestCompanyAdmin(ctx context.Context, userID string, claManagerEmail string, claManagerName string, contributorName string, contributorEmail string, projectName string, companyName string, corporateLink string) error {
	orgServices := orgService.GetClient()
	f := logrus.Fields{
		"functionName":     "RequestCompanyAdmin",
		utils.XREQUESTID:   ctx.Value(utils.XREQUESTID),
		"companyName":      companyName,
		"claGroupName":     projectName,
		"claManagerEmail":  claManagerEmail,
		"claManagerName":   claManagerName,
		"contributorEmail": contributorEmail,
		"contributorName":  contributorName,
	}

	validateError := s.ValidateRequestCompanyAdminCheck(ctx, f, userID, claManagerEmail, claManagerName, contributorName, contributorEmail, projectName, companyName)
	if validateError != nil {
		return validateError
	}

	organizations, orgErr := orgServices.ListOrg(ctx, companyName)
	if orgErr != nil {
		msg := fmt.Sprintf("Problem getting company by ID: %s ", companyName)
		log.Warn(msg)
		return orgErr
	}
	if len(organizations.Data) > 0 {
		msg := fmt.Sprintf("Comapny already exist with the name: %s ", companyName)
		log.Warn(msg)
		return errors.New(msg)
	}

	subject := fmt.Sprintf("EasyCLA: Request to start CLA signature process for : %s", projectName)
	recipients := []string{claManagerEmail}
	body := fmt.Sprintf(`
	<p>Hello %s,</p>
	<p>This is a notification email from EasyCLA regarding the project %s.</p>
	<p> %s uses EasyCLA to ensure that before a contribution is accepted, the contributor is covered under a signed CLA. </p>
	<p> %s (%s) has designated you as the proposed initial CLA Manager for contributions 
	from %s to %s. This would mean that, after the CLA is signed, you would be able to maintain 
	the list of employees allowed to contribute to %s on behalf of your company,
	as well as the list of your company’s CLA Managers for %s. </p>
	<p> If you can be the initial CLA Manager from your company for %s , please log into the EasyCLA
	Corporate Console at %s to begin the CLA signature process. You might not be authorized to 
	sign the CLA yourself on behalf of your company; if not, the signature process will prompt you 
	to designate somebody else who is authorized to sign the CLA. </p>
	%s
    %s`,
		claManagerName, projectName, projectName, contributorName,
		contributorEmail, companyName, projectName, projectName,
		projectName, projectName, corporateLink,
		utils.GetEmailHelpContent(true), utils.GetEmailSignOffContent())
	err := utils.SendEmail(subject, body, recipients)
	if err != nil {
		log.Warnf("problem sending email with subject: %s to recipients: %+v, error: %+v", subject, recipients, err)
	} else {
		log.Debugf("sent email with subject: %s to recipients: %+v", subject, recipients)
	}

	return nil
}

func (s *service) ValidateRequestCompanyAdminCheck(ctx context.Context, f logrus.Fields, userID string, claManagerEmail string, claManagerName string, contributorName string, contributorEmail string, projectName string, companyName string) error {
	validateError := validateRequestCompanyAdmin(userID, claManagerName, contributorName, contributorEmail)
	if validateError != nil {
		return validateError
	}

	claGroupModel, projectErr := s.projectRepo.GetCLAGroupByName(ctx, strings.ToLower(projectName))
	if projectErr != nil || claGroupModel == nil {
		log.WithFields(f).WithError(projectErr).Warn("problem loading CLA group by ID")

		var e *utils.CLAGroupNotFound
		if errors.As(projectErr, &e) {
			log.WithFields(f).WithError(projectErr).Warn("problem loading CLA group by ID - cla group not found")
			return ErrClaGroupNotFound

		}
		if errors.Is(projectErr, project.ErrProjectDoesNotExist) {
			log.WithFields(f).WithError(projectErr).Warn("problem cla group not found")
			return ErrClaGroupNotFound
		}
		return ErrClaGroupBadRequest

	}
	return nil
}

func validateRequestCompanyAdmin(userID string, claManagerName string, contributorName string, contributorEmail string) error {
	if userID == "" {
		return ErrCLAUserNotFound
	}
	if claManagerName == "" {
		return errors.New("cla manager name is required")
	}

	if contributorName == "" {
		return errors.New("contributor name is required")
	}
	if contributorEmail == "" {
		return errors.New("contributor is required")
	}

	return nil
}

func findCompany(companies []*v1Models.Company, companyID string) (int, bool) {
	for index, company := range companies {
		if company.CompanyID == companyID {
			return index, true
		}
	}
	return -1, false
}
