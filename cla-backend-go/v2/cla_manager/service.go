// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package cla_manager

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/LF-Engineering/lfx-kit/auth"
	"github.com/communitybridge/easycla/cla-backend-go/events"
	"github.com/communitybridge/easycla/cla-backend-go/utils"

	"github.com/communitybridge/easycla/cla-backend-go/company"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/restapi/operations/cla_manager"
	"github.com/communitybridge/easycla/cla-backend-go/project"
	"github.com/communitybridge/easycla/cla-backend-go/projects_cla_groups"
	"github.com/communitybridge/easycla/cla-backend-go/repositories"
	"github.com/communitybridge/easycla/cla-backend-go/v2/organization-service/client/organizations"

	v1ClaManager "github.com/communitybridge/easycla/cla-backend-go/cla_manager"
	v1Models "github.com/communitybridge/easycla/cla-backend-go/gen/models"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	v1User "github.com/communitybridge/easycla/cla-backend-go/user"
	easyCLAUser "github.com/communitybridge/easycla/cla-backend-go/users"
	v2AcsService "github.com/communitybridge/easycla/cla-backend-go/v2/acs-service"
	v2Company "github.com/communitybridge/easycla/cla-backend-go/v2/company"
	v2OrgService "github.com/communitybridge/easycla/cla-backend-go/v2/organization-service"
	v2ProjectService "github.com/communitybridge/easycla/cla-backend-go/v2/project-service"
	v2UserService "github.com/communitybridge/easycla/cla-backend-go/v2/user-service"
)

// Lead representing type of user
const Lead = "lead"

var (
	//ErrSalesForceProjectNotFound returned error if salesForce Project not found
	ErrSalesForceProjectNotFound = errors.New("salesforce Project not found")
	//ErrCLACompanyNotFound returned if EasyCLA company not found
	ErrCLACompanyNotFound = errors.New("company not found")
	//ErrGitHubRepoNotFound returned if GH Repos is not found
	ErrGitHubRepoNotFound = errors.New("gH Repo not found")
	//ErrCLAUserNotFound returned if EasyCLA User is not found
	ErrCLAUserNotFound = errors.New("cLA User not found")
	//ErrCLAManagersNotFound when cla managers arent found for given  project and company
	ErrCLAManagersNotFound = errors.New("cla Managers not found")
	//ErrLFXUserNotFound when user-service fails to find user
	ErrLFXUserNotFound = errors.New("lfx user not found")
	//ErrNoLFID thrown when users dont have an LFID
	ErrNoLFID = errors.New("user has no LFID")
	//ErrNotInOrg when user is not in organization
	ErrNotInOrg = errors.New("user not in organization")
	//ErrNoOrgAdmins when No admins found for organization
	ErrNoOrgAdmins = errors.New("no Admins in company ")
	//ErrRoleScopeConflict thrown if user already has role scope
	ErrRoleScopeConflict = errors.New("user is already cla-manager-designee")
	//ErrCLAManagerDesigneeConflict when user is already assigned cla-manager-designee role
	ErrCLAManagerDesigneeConflict = errors.New("user already assigned cla-manager-designee")
	//ErrScopeNotFound returns error when getting scopeID
	ErrScopeNotFound = errors.New("scope not found")
	//ErrProjectSigned returns error if project already signed
	ErrProjectSigned = errors.New("project already signed")
)

type service struct {
	companyService      company.IService
	projectService      project.Service
	repositoriesService repositories.Service
	managerService      v1ClaManager.IService
	easyCLAUserService  easyCLAUser.Service
	v2CompanyService    v2Company.Service
	eventService        events.Service
	projectCGRepo       projects_cla_groups.Repository
}

// Service interface
type Service interface {
	CreateCLAManager(claGroupID string, params cla_manager.CreateCLAManagerParams, authUsername string) (*models.CompanyClaManager, *models.ErrorResponse)
	DeleteCLAManager(claGroupID string, params cla_manager.DeleteCLAManagerParams) *models.ErrorResponse
	InviteCompanyAdmin(contactAdmin bool, companyID string, projectID string, userEmail string, contributor *v1User.User, lFxPortalURL string) (*models.ClaManagerDesignee, *models.ErrorResponse)
	CreateCLAManagerDesignee(companyID string, projectID string, userEmail string) (*models.ClaManagerDesignee, error)
	CreateCLAManagerRequest(contactAdmin bool, companyID string, projectID string, userEmail string, fullName string, authUser *auth.User, requestEmail, LfxPortalURL string) (*models.ClaManagerDesignee, error)
	NotifyCLAManagers(notifyCLAManagers *models.NotifyClaManagerList) error
}

// NewService returns instance of CLA Manager service
func NewService(compService company.IService, projService project.Service, mgrService v1ClaManager.IService, claUserService easyCLAUser.Service,
	repoService repositories.Service, v2CompService v2Company.Service,
	evService events.Service, projectCGroupRepo projects_cla_groups.Repository) Service {
	return &service{
		companyService:      compService,
		projectService:      projService,
		repositoriesService: repoService,
		managerService:      mgrService,
		easyCLAUserService:  claUserService,
		v2CompanyService:    v2CompService,
		eventService:        evService,
		projectCGRepo:       projectCGroupRepo,
	}
}

// CreateCLAManager creates Cla Manager
func (s *service) CreateCLAManager(claGroupID string, params cla_manager.CreateCLAManagerParams, authUsername string) (*models.CompanyClaManager, *models.ErrorResponse) {

	re := regexp.MustCompile(`^\w{1,30}$`)
	if !re.MatchString(*params.Body.FirstName) || !re.MatchString(*params.Body.LastName) {
		msg := "Firstname and last Name values should not exceed 30 characters in length"
		log.Warn(msg)
		return nil, &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}
	if *params.Body.UserEmail == "" {
		msg := "UserEmail cannot be empty"
		log.Warn(msg)
		return nil, &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}

	// Search for salesForce Company aka external Company
	log.Debugf("Getting company by external ID : %s", params.CompanySFID)
	companyModel, companyErr := s.companyService.GetCompanyByExternalID(params.CompanySFID)
	if companyErr != nil || companyModel == nil {
		msg := buildErrorMessage("company lookup error", claGroupID, params, companyErr)
		log.Warn(msg)
		return nil, &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}

	claGroup, err := s.projectService.GetCLAGroupByID(claGroupID)
	if err != nil || claGroup == nil {
		msg := buildErrorMessage("cla group search by ID failure", claGroupID, params, err)
		log.Warn(msg)
		return nil, &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}
	// Get user by email
	userServiceClient := v2UserService.GetClient()
	// Get Manager lf account by username. Used for email content
	managerUser, mgrErr := userServiceClient.GetUserByUsername(authUsername)
	if mgrErr != nil {
		msg := fmt.Sprintf("Failed to get Lfx User with username : %s ", authUsername)
		log.Warn(msg)
	}
	// GetSF Org
	orgClient := v2OrgService.GetClient()
	organizationSF, orgErr := orgClient.GetOrganization(params.CompanySFID)
	if orgErr != nil {
		msg := buildErrorMessage("organization service lookup error", claGroupID, params, orgErr)
		log.Warn(msg)
		return nil, &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}
	acsClient := v2AcsService.GetClient()
	user, userErr := userServiceClient.SearchUserByEmail(*params.Body.UserEmail)

	if userErr != nil {
		designeeName := fmt.Sprintf("%s %s", *params.Body.FirstName, *params.Body.LastName)
		designeeEmail := *params.Body.UserEmail
		msg := fmt.Sprintf("User does not have an LFID account and has been sent an email invite: %s.", *params.Body.UserEmail)
		log.Warn(msg)
		sendEmailErr := sendEmailToUserWithNoLFID(claGroup.ProjectName, authUsername, *managerUser.Emails[0].EmailAddress, designeeName, designeeEmail, organizationSF.ID)
		if sendEmailErr != nil {
			emailMessage := fmt.Sprintf("Failed to send email to user : %s ", designeeEmail)
			return nil, &models.ErrorResponse{
				Message: emailMessage,
				Code:    "400",
			}
		}
		return nil, &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}

	// Check if user exists in easyCLA DB, if not add User
	log.Debugf("Checking user: %s in easyCLA records", user.Username)
	claUser, claUserErr := s.easyCLAUserService.GetUserByLFUserName(user.Username)
	if claUserErr != nil {
		msg := fmt.Sprintf("Problem getting claUser by :%s, error: %+v ", user.Username, claUserErr)
		log.Warn(msg)
		return nil, &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}

	if claUser == nil {
		msg := fmt.Sprintf("User not found when searching by LFID: %s and shall be created", user.Username)
		log.Debug(msg)
		userName := fmt.Sprintf("%s %s", *params.Body.FirstName, *params.Body.LastName)
		_, currentTimeString := utils.CurrentTime()
		claUserModel := &v1Models.User{
			UserExternalID: params.CompanySFID,
			LfEmail:        *user.Emails[0].EmailAddress,
			Admin:          true,
			LfUsername:     user.Username,
			DateCreated:    currentTimeString,
			DateModified:   currentTimeString,
			Username:       userName,
			Version:        "v1",
		}
		newUserModel, userModelErr := s.easyCLAUserService.CreateUser(claUserModel, nil)
		if userModelErr != nil {
			msg := fmt.Sprintf("Failed to create user : %+v", claUserModel)
			log.Warn(msg)
			return nil, &models.ErrorResponse{
				Message: msg,
				Code:    "400",
			}
		}
		log.Debugf("Created easyCLAUser %+v ", newUserModel)
	}

	// Check if user is part of org
	log.Debugf("Check user: %s's organization ", user.Username)
	if user.Account.ID != strings.TrimSpace(params.CompanySFID) {
		msg := fmt.Sprintf("User : %s not in organization : %s ", user.Username, organizationSF.Name)
		log.Warn(msg)
		return nil, &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}

	// GetSFProject
	ps := v2ProjectService.GetClient()
	projectSF, projectErr := ps.GetProject(params.ProjectSFID)
	if projectErr != nil {
		msg := buildErrorMessage("project service lookup error", claGroupID, params, projectErr)
		log.Warn(msg)
		return nil, &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}

	// Add CLA Manager to Database
	signature, addErr := s.managerService.AddClaManager(companyModel.CompanyID, claGroupID, user.Username)
	if addErr != nil {
		msg := buildErrorMessageCreate(params, addErr)
		log.Warn(msg)
		return nil, &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}
	if signature == nil {
		sigMsg := fmt.Sprintf("Signature not found for project: %s and company: %s ", claGroupID, companyModel.CompanyID)
		log.Warn(sigMsg)
		return nil, &models.ErrorResponse{
			Message: sigMsg,
			Code:    "400",
		}
	}

	log.Warn("Getting role")
	// Get RoleID for cla-manager

	roleID, roleErr := acsClient.GetRoleID("cla-manager")
	if roleErr != nil {
		msg := buildErrorMessageCreate(params, roleErr)
		log.Warn(msg)
		return nil, &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}
	log.Debugf("Role ID for cla-manager-role : %s", roleID)
	log.Debugf("Creating user role Scope for user : %s ", *params.Body.UserEmail)

	hasScope, err := orgClient.IsUserHaveRoleScope("cla-manager", user.ID, params.CompanySFID, params.ProjectSFID)
	if err != nil {
		msg := buildErrorMessageCreate(params, err)
		log.Warn(msg)
		return nil, &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}
	if hasScope {
		msg := fmt.Sprintf("User %s is already cla-manager for Company: %s and Project: %s", user.Username, params.CompanySFID, params.ProjectSFID)
		log.Warn(msg)
		return nil, &models.ErrorResponse{
			Message: msg,
			Code:    "409",
		}
	}

	projectCLAGroups, getErr := s.projectCGRepo.GetProjectsIdsForClaGroup(claGroupID)
	log.Debugf("Getting associated SF projects for claGroup: %s ", claGroupID)

	if getErr != nil {
		msg := buildErrorMessageCreate(params, getErr)
		log.Warn(msg)
		return nil, &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}

	for _, projectCG := range projectCLAGroups {

		scopeErr := orgClient.CreateOrgUserRoleOrgScopeProjectOrg(*params.Body.UserEmail, projectCG.ProjectSFID, params.CompanySFID, roleID)
		if scopeErr != nil {
			msg := buildErrorMessageCreate(params, scopeErr)
			log.Warn(msg)
			return nil, &models.ErrorResponse{
				Message: msg,
				Code:    "400",
			}
		}
	}

	if user.Type == Lead {
		// convert user to contact
		log.Debug("converting lead to contact")
		err := userServiceClient.ConvertToContact(user.ID)
		if err != nil {
			msg := fmt.Sprintf("converting lead to contact failed: %v", err)
			log.Warn(msg)
			return nil, &models.ErrorResponse{
				Message: msg,
				Code:    "400",
			}
		}
	}

	claCompanyManager := &models.CompanyClaManager{
		LfUsername:       user.Username,
		Email:            *params.Body.UserEmail,
		UserSfid:         user.ID,
		ApprovedOn:       time.Now().String(),
		ProjectSfid:      params.ProjectSFID,
		ClaGroupName:     claGroup.ProjectName,
		ProjectID:        claGroupID,
		ProjectName:      projectSF.Name,
		OrganizationName: companyModel.CompanyName,
		OrganizationSfid: params.CompanySFID,
		Name:             fmt.Sprintf("%s %s", user.FirstName, user.LastName),
	}
	return claCompanyManager, nil
}

func (s *service) DeleteCLAManager(claGroupID string, params cla_manager.DeleteCLAManagerParams) *models.ErrorResponse {
	// Get user by firstname,lastname and email parameters
	userServiceClient := v2UserService.GetClient()
	user, userErr := userServiceClient.GetUserByUsername(params.UserLFID)

	if userErr != nil {
		msg := fmt.Sprintf("Failed to get user when searching by username: %s , error: %v ", params.UserLFID, userErr)
		return &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}

	// Search for salesForce Company aka external Company
	companyModel, companyErr := s.companyService.GetCompanyByExternalID(params.CompanySFID)
	if companyErr != nil || companyModel == nil {
		msg := buildErrorMessageDelete(params, companyErr)
		log.Warn(msg)
		return &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}

	acsClient := v2AcsService.GetClient()

	roleID, roleErr := acsClient.GetRoleID("cla-manager")
	if roleErr != nil {
		msg := buildErrorMessageDelete(params, roleErr)
		log.Warn(msg)
		return &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}
	log.Debugf("Role ID for cla-manager-role : %s", roleID)

	projectCLAGroups, getErr := s.projectCGRepo.GetProjectsIdsForClaGroup(claGroupID)

	if getErr != nil {
		msg := buildErrorMessageDelete(params, getErr)
		log.Warn(msg)
		return &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}

	orgClient := v2OrgService.GetClient()

	for _, projectCG := range projectCLAGroups {
		scopeID, scopeErr := orgClient.GetScopeID(params.CompanySFID, projectCG.ProjectSFID, "cla-manager", "project|organization", params.UserLFID)
		if scopeErr != nil {
			msg := buildErrorMessageDelete(params, scopeErr)
			log.Warn(msg)
			return &models.ErrorResponse{
				Message: msg,
				Code:    "400",
			}
		}
		if scopeID == "" {
			msg := buildErrorMessageDelete(params, ErrScopeNotFound)
			log.Warn(msg)
			return &models.ErrorResponse{
				Message: msg,
				Code:    "400",
			}
		}
		email := *user.Emails[0].EmailAddress
		deleteErr := orgClient.DeleteOrgUserRoleOrgScopeProjectOrg(params.CompanySFID, roleID, scopeID, &user.Username, &email)
		if deleteErr != nil {
			msg := buildErrorMessageDelete(params, deleteErr)
			log.Warn(msg)
			return &models.ErrorResponse{
				Message: msg,
				Code:    "400",
			}
		}
	}

	signature, deleteErr := s.managerService.RemoveClaManager(companyModel.CompanyID, claGroupID, params.UserLFID)

	if deleteErr != nil {
		msg := buildErrorMessageDelete(params, deleteErr)
		log.Warn(msg)
		return &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}
	if signature == nil {
		msg := fmt.Sprintf("Not found signature for project: %s and company: %s ", claGroupID, companyModel.CompanyID)
		log.Warn(msg)
		return &models.ErrorResponse{
			Message: msg,
			Code:    "400",
		}
	}

	return nil
}

//CreateCLAManagerDesignee creates designee for cla manager prospect
func (s *service) CreateCLAManagerDesignee(companyID string, projectID string, userEmail string) (*models.ClaManagerDesignee, error) {
	// integrate user,acs,org and project services
	userClient := v2UserService.GetClient()
	acServiceClient := v2AcsService.GetClient()
	orgClient := v2OrgService.GetClient()
	projectClient := v2ProjectService.GetClient()

	user, userErr := userClient.SearchUserByEmail(userEmail)
	if userErr != nil {
		log.Debugf("Failed to get user by email: %s , error: %+v", userEmail, userErr)
		return nil, ErrLFXUserNotFound
	}

	// Check if user is part of organization
	if user.Account.ID != strings.TrimSpace(companyID) {
		msg := fmt.Sprintf("User :%s does not belong to organization", userEmail)
		log.Warn(msg)
		return nil, ErrNotInOrg
	}

	projectSF, projectErr := projectClient.GetProject(projectID)
	if projectErr != nil {
		msg := fmt.Sprintf("Problem getting project :%s ", projectID)
		log.Debug(msg)
		return nil, projectErr
	}

	roleID, designeeErr := acServiceClient.GetRoleID("cla-manager-designee")
	if designeeErr != nil {
		msg := "Problem getting role ID for cla-manager-designee"
		log.Warn(msg)
		return nil, designeeErr
	}

	scopeErr := orgClient.CreateOrgUserRoleOrgScopeProjectOrg(userEmail, projectID, companyID, roleID)
	if scopeErr != nil {
		msg := fmt.Sprintf("Problem creating projectOrg scope for email: %s , projectID: %s, companyID: %s", userEmail, projectID, companyID)
		log.Warn(msg)
		if _, ok := scopeErr.(*organizations.CreateOrgUsrRoleScopesConflict); ok {
			return nil, ErrRoleScopeConflict
		}
		return nil, scopeErr
	}

	// Log Event
	s.eventService.LogEvent(
		&events.LogEventArgs{
			EventType:         events.AssignUserRoleScopeType,
			LfUsername:        user.Username,
			ExternalProjectID: projectID,
			EventData: &events.AssignRoleScopeData{
				Role:  "cla-manager-designee",
				Scope: fmt.Sprintf("%s|%s", projectID, companyID),
			},
		})

	if user.Type == Lead {
		log.Debugf("Converting user: %s from lead to contact ", userEmail)
		contactErr := userClient.ConvertToContact(user.ID)
		if contactErr != nil {
			log.Debugf("failed to convert user: %s to contact ", userEmail)
			return nil, contactErr
		}
		// Log user conversion event
		s.eventService.LogEvent(&events.LogEventArgs{
			EventType:         events.ConvertUserToContactType,
			LfUsername:        user.Username,
			ExternalProjectID: projectID,
			EventData:         &events.UserConvertToContactData{},
		})
	}

	claManagerDesignee := &models.ClaManagerDesignee{
		LfUsername:  user.Username,
		UserSfid:    user.ID,
		Type:        user.Type,
		AssignedOn:  time.Now().String(),
		Email:       userEmail,
		ProjectSfid: projectID,
		CompanySfid: companyID,
		ProjectName: projectSF.Name,
	}
	return claManagerDesignee, nil
}

func (s *service) CreateCLAManagerRequest(contactAdmin bool, companyID string, projectID string, userEmail string, fullName string, authUser *auth.User, requestEmail, LfxPortalURL string) (*models.ClaManagerDesignee, error) {
	orgService := v2OrgService.GetClient()

	isSigned, signedErr := s.isSigned(projectID)
	if signedErr != nil {
		msg := fmt.Sprintf("EasyCLA - 400 Bad Request- %s", signedErr)
		log.Warn(msg)
		return nil, signedErr
	}

	if isSigned {
		msg := fmt.Sprintf("EasyCLA - 400 Bad Request - Project :%s is already signed ", projectID)
		log.Warn(msg)
		return nil, ErrProjectSigned
	}

	// GetSFProject
	ps := v2ProjectService.GetClient()
	projectSF, projectErr := ps.GetProject(projectID)
	if projectErr != nil {
		msg := fmt.Sprintf("EasyCLA - 400 Bad Request - Project service lookup error for SFID: %s, error : %+v",
			projectID, projectErr)
		log.Warn(msg)
		return nil, projectErr
	}

	// Search for salesForce Company aka external Company
	companyModel, companyErr := orgService.GetOrganization(companyID)
	if companyErr != nil || companyModel == nil {
		msg := fmt.Sprintf("EasyCLA - 400 Bad Request - Problem getting company by SFID: %s, error: %+v",
			companyID, companyErr)
		log.Warn(msg)
		return nil, companyErr
	}

	// Check if sending cla manager request to company admin
	if contactAdmin {
		log.Debugf("Sending email to company Admin")
		scopes, listScopeErr := orgService.ListOrgUserAdminScopes(companyID)
		if listScopeErr != nil {
			msg := fmt.Sprintf("EasyCLA - 400 Bad Request - Admin lookup error for organisation SFID: %s, error: %+v ",
				companyID, listScopeErr)
			log.Warn(msg)
			return nil, listScopeErr
		}

		if len(scopes.Userroles) == 0 {
			msg := fmt.Sprintf("EasyCLA - 404 NotFound - No admins for organization SFID: %s",
				companyID)
			log.Warn(msg)
			return nil, ErrNoOrgAdmins
		}

		for _, admin := range scopes.Userroles {
			sendEmailToOrgAdmin(admin.Contact.EmailAddress, admin.Contact.Name, companyModel.Name, projectSF.Name, authUser.Email, authUser.UserName, LfxPortalURL)
			// Make a note in the event log
			s.eventService.LogEvent(&events.LogEventArgs{
				EventType:         events.ContributorNotifyCompanyAdminType,
				LfUsername:        authUser.UserName,
				ExternalProjectID: projectID,
				CompanyID:         companyModel.ID,
				EventData: &events.ContributorNotifyCompanyAdminData{
					AdminName:  admin.Contact.Name,
					AdminEmail: admin.Contact.EmailAddress,
				},
			})
		}

		return nil, nil
	}

	userService := v2UserService.GetClient()
	lfxUser, userErr := userService.SearchUserByEmail(userEmail)
	if userErr != nil {
		msg := fmt.Sprintf("EasyCLA - 404 Not Found - User: %s does not have an LFID ", userEmail)
		log.Warn(msg)
		// Send email
		sendEmailErr := sendEmailToUserWithNoLFID(projectSF.Name, authUser.UserName, requestEmail, fullName, userEmail, companyModel.ID)
		if sendEmailErr != nil {
			return nil, sendEmailErr
		}
		return nil, ErrNoLFID
	}

	claManagerDesignee, err := s.CreateCLAManagerDesignee(companyID, projectID, userEmail)
	if err != nil {
		// Check conflict for role scope
		if err == err.(*organizations.CreateOrgUsrRoleScopesConflict) {
			return nil, ErrRoleScopeConflict
		}
		return nil, err
	}

	// Make a note in the event log
	s.eventService.LogEvent(&events.LogEventArgs{
		EventType:         events.ContributorAssignCLADesigneeType,
		LfUsername:        authUser.UserName,
		ExternalProjectID: projectID,
		CompanyID:         companyModel.ID,
		EventData: &events.ContributorAssignCLADesignee{
			DesigneeName:  claManagerDesignee.LfUsername,
			DesigneeEmail: claManagerDesignee.Email,
		},
	})

	log.Debugf("Sending Email to CLA Manager Designee email: %s ", userEmail)
	designeeName := fmt.Sprintf("%s %s", lfxUser.FirstName, lfxUser.LastName)
	sendEmailToCLAManagerDesignee(LfxPortalURL, companyModel.Name, projectSF.Name, userEmail, designeeName, authUser.Email, authUser.UserName)
	// Make a note in the event log
	s.eventService.LogEvent(&events.LogEventArgs{
		EventType:         events.ContributorNotifyCLADesigneeType,
		LfUsername:        authUser.UserName,
		ExternalProjectID: projectID,
		CompanyID:         companyModel.ID,
		EventData: &events.ContributorNotifyCLADesignee{
			DesigneeName:  claManagerDesignee.LfUsername,
			DesigneeEmail: claManagerDesignee.Email,
		},
	})

	log.Debugf("CLA Manager designee created : %+v", claManagerDesignee)
	return claManagerDesignee, nil
}

func (s *service) InviteCompanyAdmin(contactAdmin bool, companyID string, projectID string, userEmail string, contributor *v1User.User, LfxPortalURL string) (*models.ClaManagerDesignee, *models.ErrorResponse) {
	orgService := v2OrgService.GetClient()
	projectService := v2ProjectService.GetClient()
	userService := v2UserService.GetClient()

	// Get repo instance (assist in getting salesforce project)
	log.Debugf("Get salesforce project by claGroupID: %s ", projectID)
	ghRepoModel, ghRepoErr := s.repositoriesService.GetGithubRepositoryByCLAGroup(projectID)
	if ghRepoErr != nil || ghRepoModel.RepositorySfdcID == "" {
		msg := fmt.Sprintf("Problem getting salesforce project by claGroupID : %s ", projectID)
		log.Warn(msg)
		return nil, &models.ErrorResponse{
			Code:    "404",
			Message: msg,
		}
	}

	// Get company
	log.Debugf("Get company for companyID: %s ", companyID)
	companyModel, companyErr := s.companyService.GetCompany(companyID)
	if companyErr != nil || companyModel.CompanyExternalID == "" {
		msg := fmt.Sprintf("Problem getting company for companyID: %s ", companyID)
		log.Warn(msg)
		return nil, &models.ErrorResponse{
			Code:    "404",
			Message: msg,
		}
	}

	project, projectErr := projectService.GetProject(ghRepoModel.RepositorySfdcID)
	if projectErr != nil {
		msg := fmt.Sprintf("Problem getting project by ID: %s ", projectID)
		log.Warn(msg)
		return nil, &models.ErrorResponse{
			Code:    "400",
			Message: msg,
		}
	}

	organization, orgErr := orgService.GetOrganization(companyModel.CompanyExternalID)
	if orgErr != nil {
		msg := fmt.Sprintf("Problem getting company by ID: %s ", companyID)
		log.Warn(msg)
		return nil, &models.ErrorResponse{
			Code:    "400",
			Message: msg,
		}
	}

	// Get suggested CLA Manager user details
	user, userErr := userService.SearchUserByEmail(userEmail)
	if userErr != nil {
		msg := fmt.Sprintf("UserEmail: %s has no LFID and has been sent an invite email to create an account , error: %+v", userEmail, userErr)
		log.Warn(msg)
		// Send Email
		sendErr := sendEmailToUserWithNoLFID(project.Name, contributor.UserName, contributor.UserEmails[0], userEmail, userEmail, organization.ID)
		if sendErr != nil {
			return nil, &models.ErrorResponse{
				Code:    "400",
				Message: sendErr.Error(),
			}
		}
		return nil, &models.ErrorResponse{
			Code:    "400",
			Message: msg,
		}
	}

	// Check if sending cla manager request to company admin
	if contactAdmin {
		log.Debugf("Sending email to company Admin")
		scopes, listScopeErr := orgService.ListOrgUserAdminScopes(companyModel.CompanyExternalID)
		if listScopeErr != nil {
			msg := fmt.Sprintf("Admin lookup error for organisation SFID: %s ", companyModel.CompanyExternalID)
			return nil, &models.ErrorResponse{
				Code:    "400",
				Message: msg,
			}
		}
		for _, admin := range scopes.Userroles {
			// Check if is Gerrit User or GH User
			if contributor.LFUsername != "" && contributor.LFEmail != "" {
				sendEmailToOrgAdmin(admin.Contact.EmailAddress, admin.Contact.Name, organization.Name, project.Name, contributor.LFEmail, contributor.LFUsername, LfxPortalURL)
			} else {
				sendEmailToOrgAdmin(admin.Contact.EmailAddress, admin.Contact.Name, organization.Name, project.Name, contributor.UserGithubID, contributor.UserGithubUsername, LfxPortalURL)
			}

		}
		return nil, nil
	}

	claManagerDesignee, err := s.CreateCLAManagerDesignee(organization.ID, project.ID, userEmail)

	if err != nil {
		msg := fmt.Sprintf("Problem creating cla Manager Designee for user :%s, error: %+v ", userEmail, err)
		return nil, &models.ErrorResponse{
			Code:    "400",
			Message: msg,
		}
	}

	log.Debugf("Sending Email to CLA Manager Designee email: %s ", userEmail)

	if contributor.LFUsername != "" && contributor.LFEmail != "" {
		sendEmailToCLAManagerDesignee(LfxPortalURL, organization.Name, project.Name, userEmail, user.Name, contributor.LFEmail, contributor.LFUsername)
	} else {
		sendEmailToCLAManagerDesignee(LfxPortalURL, organization.Name, project.Name, userEmail, user.Name, contributor.UserGithubID, contributor.UserGithubUsername)
	}

	log.Debugf("CLA Manager designee created : %+v", claManagerDesignee)
	return claManagerDesignee, nil

}

func (s *service) NotifyCLAManagers(notifyCLAManagers *models.NotifyClaManagerList) error {
	// Search for Easy CLA User
	log.Debugf("Getting user by ID: %s", notifyCLAManagers.UserID)
	userModel, userErr := s.easyCLAUserService.GetUser(notifyCLAManagers.UserID)
	if userErr != nil {
		msg := fmt.Sprintf("Problem getting user by ID: %s ", notifyCLAManagers.UserID)
		log.Warn(msg)
		return ErrCLAUserNotFound
	}

	log.Debugf("Sending notification emails to claManagers: %+v", notifyCLAManagers.List)
	for _, claManager := range notifyCLAManagers.List {
		sendEmailToCLAManager(claManager.Name, claManager.Email, userModel.GithubUsername, notifyCLAManagers.CompanyName, notifyCLAManagers.ProjectName)
	}

	return nil
}

func sendEmailToCLAManager(manager string, managerEmail string, contributorName string, company string, project string) {
	subject := fmt.Sprintf("EasyCLA: Approval Request for contributor: %s  ", contributorName)
	recipients := []string{managerEmail}
	body := fmt.Sprintf(`
	<p>Hello %s,</p>
	<p>This is a notification email from EasyCLA regarding the organization %s.</p>
	<p>The following contributor would like to submit a contribution to %s 
	   and is requesting to be approved as a contributor for your organization: </p>
	<p>%s</p>
	<p>Please notify the contributor once they are added so that they may complete the contribution process.</p>
	%s
    %s`,
		manager, company, project, contributorName,
		utils.GetEmailHelpContent(true), utils.GetEmailSignOffContent())
	err := utils.SendEmail(subject, body, recipients)
	if err != nil {
		log.Warnf("problem sending email with subject: %s to recipients: %+v, error: %+v", subject, recipients, err)
	} else {
		log.Debugf("sent email with subject: %s to recipients: %+v", subject, recipients)
	}
}

// Helper function to check if project/claGroup is signed
func (s *service) isSigned(projectID string) (bool, error) {
	isSigned := false
	// Get claGroup ID
	cgGroup, cgErr := s.projectCGRepo.GetClaGroupIDForProject(projectID)
	if cgErr != nil {
		msg := fmt.Sprintf("EasyCLA - 400 Bad Request - CLAGroup lookup fail for project : %s ", projectID)
		log.Warn(msg)
		return isSigned, cgErr
	}

	// Check if group is signed
	claGroup, claGroupErr := s.projectService.GetCLAGroupByID(cgGroup.ClaGroupID)
	if claGroupErr != nil {
		msg := fmt.Sprintf("EasyCLA - 400 Bad Request - CLAGroup lookup fail for project : %s", cgGroup.ClaGroupID)
		log.Warn(msg)
		return isSigned, claGroupErr
	}
	if claGroup.ProjectCCLAEnabled {
		msg := fmt.Sprintf("EasyCLA - 400 Bad Request - CLA Group signed for project : %s ", projectID)
		log.Warn(msg)
		isSigned = true
	}

	return isSigned, nil
}

func sendEmailToOrgAdmin(adminEmail string, admin string, company string, projectName string, contributorID string, contributorName string, corporateConsole string) {
	subject := fmt.Sprintf("EasyCLA:  Invitation to Sign the %s Corporate CLA and add to approved list %s ", company, contributorID)
	recipients := []string{adminEmail}
	body := fmt.Sprintf(`
<p>Hello %s,</p>
<p>This is a notification email from EasyCLA regarding the project %s.</p>
<p>The following contributor is requesting to sign CLA for organization: </p>
<p> %s %s </p>
<p>Before the user contribution can be accepted, your organization must sign a CLA.
<p>Kindly login to this portal %s and sign the CLA for this project %s. </p>
<p>Please notify the contributor once they are added so that they may complete the contribution process.</p>
%s
%s`,
		admin, projectName, contributorName, contributorID, corporateConsole, projectName,
		utils.GetEmailHelpContent(true), utils.GetEmailSignOffContent())

	err := utils.SendEmail(subject, body, recipients)
	if err != nil {
		log.Warnf("problem sending email with subject: %s to recipients: %+v, error: %+v", subject, recipients, err)
	} else {
		log.Debugf("sent email with subject: %s to recipients: %+v", subject, recipients)
	}
}

func sendEmailToCLAManagerDesignee(corporateConsole string, companyName string, projectName string, designeeEmail string, designeeName string, contributorID string, contributorName string) {
	subject := fmt.Sprintf("EasyCLA:  Invitation to Sign the %s Corporate CLA and add to approved list %s ",
		companyName, contributorID)
	recipients := []string{designeeEmail}
	body := fmt.Sprintf(`
<p>Hello %s,</p>
<p>This is a notification email from EasyCLA regarding the project %s.</p>
<p>The following contributor is requesting to sign CLA for organization: </p>
<p> %s %s </p>
<p>Before the user contribution can be accepted, your organization must sign a CLA.
<p>Kindly login to this portal %s and sign the CLA for this project %s. </p>
<p>Please notify the contributor once they are added so that they may complete the contribution process.</p>
%s
%s`,
		designeeName, projectName, contributorName, contributorID, corporateConsole, projectName,
		utils.GetEmailHelpContent(true), utils.GetEmailSignOffContent())

	err := utils.SendEmail(subject, body, recipients)
	if err != nil {
		log.Warnf("problem sending email with subject: %s to recipients: %+v, error: %+v", subject, recipients, err)
	} else {
		log.Debugf("sent email with subject: %s to recipients: %+v", subject, recipients)
	}
}

// sendEmailToUserWithNoLFID helper function to send email to a given user with no LFID
func sendEmailToUserWithNoLFID(projectName, requesterUsername, requesterEmail, userWithNoLFIDName, userWithNoLFIDEmail, organizationID string) error {
	// subject string, body string, recipients []string
	subject := "EasyCLA: Invitation to create LFID and complete process of becoming CLA Manager"
	body := fmt.Sprintf(`
<p>Hello %s,</p>
<p>This is a notification email from EasyCLA regarding the Project %s in the EasyCLA system.</p>
<p>User %s (%s) was trying to add you as a CLA Manager for Project %s but was unable to identify your account details in
the EasyCLA system. In order to become a CLA Manager for Project %s, you will need to accept invite below.
Once complete, notify the user %s and they will be able to add you as a CLA Manager.</p>
%s
%s`,
		userWithNoLFIDName, projectName,
		requesterUsername, requesterEmail, projectName, projectName,
		requesterUsername,
		utils.GetEmailHelpContent(true), utils.GetEmailSignOffContent())

	acsClient := v2AcsService.GetClient()
	acsErr := acsClient.SendUserInvite(&userWithNoLFIDEmail, "cla-manager", "organization", organizationID, "userinvite", &subject, &body)
	if acsErr != nil {
		return acsErr
	}
	return nil
}

// buildErrorMessage helper function to build an error message
func buildErrorMessage(errPrefix string, claGroupID string, params cla_manager.CreateCLAManagerParams, err error) string {
	return fmt.Sprintf("%s - problem creating new CLA Manager Request using company SFID: %s, project ID: %s, first name: %s, last name: %s, user email: %s, error: %+v",
		errPrefix, params.CompanySFID, claGroupID, *params.Body.FirstName, *params.Body.LastName, *params.Body.UserEmail, err)
}
