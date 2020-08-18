// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package cla_groups

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/LF-Engineering/lfx-kit/auth"
	v1ClaManager "github.com/communitybridge/easycla/cla-backend-go/cla_manager"
	"github.com/communitybridge/easycla/cla-backend-go/events"
	"github.com/communitybridge/easycla/cla-backend-go/gerrits"
	"github.com/communitybridge/easycla/cla-backend-go/repositories"
	signatureService "github.com/communitybridge/easycla/cla-backend-go/signatures"
	organization_service "github.com/communitybridge/easycla/cla-backend-go/v2/organization-service"

	"github.com/communitybridge/easycla/cla-backend-go/v2/metrics"

	"github.com/communitybridge/easycla/cla-backend-go/utils"

	"github.com/jinzhu/copier"

	v1Models "github.com/communitybridge/easycla/cla-backend-go/gen/models"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	v1Project "github.com/communitybridge/easycla/cla-backend-go/project"
	"github.com/communitybridge/easycla/cla-backend-go/projects_cla_groups"
	v1Template "github.com/communitybridge/easycla/cla-backend-go/template"
	v2ProjectService "github.com/communitybridge/easycla/cla-backend-go/v2/project-service"
	psproject "github.com/communitybridge/easycla/cla-backend-go/v2/project-service/client/project"
	"github.com/sirupsen/logrus"
)

// constants
const (
	DontLoadDetails = false
	LoadDetails     = true
)

type service struct {
	v1ProjectService      v1Project.Service
	v1TemplateService     v1Template.Service
	projectsClaGroupsRepo projects_cla_groups.Repository
	claManagerRequests    v1ClaManager.IService
	signatureService      signatureService.SignatureService
	metricsRepo           metrics.Repository
	gerritService         gerrits.Service
	repositoriesService   repositories.Service
	eventsService         events.Service
}

// Service interface
type Service interface {
	CreateCLAGroup(input *models.CreateClaGroupInput, projectManagerLFID string) (*models.ClaGroup, error)
	EnrollProjectsInClaGroup(claGroupID string, foundationSFID string, projectSFIDList []string) error
	DeleteCLAGroup(claGroupModel *v1Models.Project, authUser *auth.User) error
	ListClaGroupsForFoundationOrProject(foundationSFID string) (*models.ClaGroupList, error)
	ValidateCLAGroup(input *models.ClaGroupValidationRequest) (bool, []string)
	ListAllFoundationClaGroups(foundationID *string) (*models.FoundationMappingList, error)
}

// NewService returns instance of CLA group service
func NewService(projectService v1Project.Service, templateService v1Template.Service, projectsClaGroupsRepo projects_cla_groups.Repository, claMangerRequests v1ClaManager.IService, signatureService signatureService.SignatureService, metricsRepo metrics.Repository, gerritService gerrits.Service, repositoriesService repositories.Service, eventsService events.Service) Service {
	return &service{
		v1ProjectService:      projectService, // aka cla_group service of v1
		v1TemplateService:     templateService,
		projectsClaGroupsRepo: projectsClaGroupsRepo,
		claManagerRequests:    claMangerRequests,
		signatureService:      signatureService,
		metricsRepo:           metricsRepo,
		gerritService:         gerritService,
		repositoriesService:   repositoriesService,
		eventsService:         eventsService,
	}
}

// ValidateCLAGroup is the service handler for validating a CLA Group
func (s *service) ValidateCLAGroup(input *models.ClaGroupValidationRequest) (bool, []string) {

	var valid = true
	var validationErrors []string

	// All parameters are optional - caller can specify which fields they want to validate based on what they provide
	// in the request payload.  If the value is there, we will attempt to validate it.  Note: some validation
	// happens at the Swagger API specification level (and rejected) before our API handler will be invoked.

	// Note: CLA Group Name Min/Max Character Length validated via Swagger Spec restrictions
	if input.ClaGroupName != nil {
		claGroupModel, err := s.v1ProjectService.GetCLAGroupByName(*input.ClaGroupName)
		if err != nil {
			valid = false
			validationErrors = append(validationErrors, fmt.Sprintf("unable to query project service - error: %+v", err))
		}
		if claGroupModel != nil {
			valid = false
			validationErrors = append(validationErrors, fmt.Sprintf("CLA Group with name %s already exist", *input.ClaGroupName))
		}
	}

	// Note: CLA Group Description Min/Max Character Length validated via Swagger Spec restrictions

	// Optional - we can expand this API logic to validate other fields if needed.

	return valid, validationErrors
}

// validateClaGroupInput validates the cla group input. It there is validation error then it returns the error
// if foundation_sfid is root project i.e project without parent and if it does not have subprojects then return boolean
// flag would be true
func (s *service) validateClaGroupInput(input *models.CreateClaGroupInput) (bool, error) {
	if *input.FoundationSfid == "" {
		return false, fmt.Errorf("bad request: foundation_sfid cannot be empty")
	}
	if !*input.IclaEnabled && !*input.CclaEnabled {
		return false, fmt.Errorf("bad request: can not create cla group with both icla and ccla disabled")
	}
	if *input.CclaRequiresIcla {
		if !(*input.IclaEnabled && *input.CclaEnabled) {
			return false, fmt.Errorf("bad request: ccla_requires_icla can not be enabled if one of icla/ccla is disabled")
		}
	}
	claGroupModel, err := s.v1ProjectService.GetCLAGroupByName(*input.ClaGroupName)
	if err != nil {
		return false, err
	}
	if claGroupModel != nil {
		return false, fmt.Errorf("bad request: cla_group with name %s already exist", *input.ClaGroupName)
	}

	psc := v2ProjectService.GetClient()
	rootProjectDetails, err := psc.GetProject(*input.FoundationSfid)
	if err != nil {
		if _, ok := err.(*psproject.GetProjectNotFound); ok {
			return false, errors.New("bad request: invalid foundation_sfid")
		}
		return false, err
	}

	if rootProjectDetails.Parent == "" && len(rootProjectDetails.Projects) == 0 {
		// this is standalone project
		if len(input.ProjectSfidList) != 0 {
			return false, fmt.Errorf("bad request: invalid project_sfid_list. This project does not have subprojects")
		}
		return true, nil
	}
	err = s.validateEnrollProjectsInput(*input.FoundationSfid, input.ProjectSfidList)
	if err != nil {
		return false, err
	}
	return false, nil
}

func (s *service) validateEnrollProjectsInput(foundationSFID string, projectSFIDList []string) error {
	psc := v2ProjectService.GetClient()

	if len(projectSFIDList) == 0 {
		return fmt.Errorf("bad request: there should be at least one subproject associated")
	}

	// fetch foundation and its sub projects
	rootProjectDetails, err := psc.GetProject(foundationSFID)
	if err != nil {
		return err
	}

	if rootProjectDetails.Parent != "" {
		return fmt.Errorf("bad request: invalid input foundation_sfid. It have parent project")
	}
	if len(rootProjectDetails.Projects) == 0 {
		return fmt.Errorf("bad request: invalid input to enroll projects. project does not have subprojects")
	}

	// check if all enrolled projects are part of foundation
	foundationProjectList := utils.NewStringSet()
	for _, pr := range rootProjectDetails.Projects {
		foundationProjectList.Add(pr.ID)
	}
	invalidProjectSFIDs := utils.NewStringSet()
	for _, projectSFID := range projectSFIDList {
		if !foundationProjectList.Include(projectSFID) {
			invalidProjectSFIDs.Add(projectSFID)
		}
	}
	if invalidProjectSFIDs.Length() != 0 {
		return fmt.Errorf("bad request: invalid project_sfid: %v. These project is not under foundation", invalidProjectSFIDs.List())
	}

	// check if projects are not already enabled
	enabledProjects, err := s.projectsClaGroupsRepo.GetProjectsIdsForFoundation(foundationSFID)
	if err != nil {
		return err
	}
	enabledProjectList := utils.NewStringSet()
	for _, pr := range enabledProjects {
		enabledProjectList.Add(pr.ProjectSFID)
	}
	invalidProjectSFIDs = utils.NewStringSet()
	for _, projectSFID := range projectSFIDList {
		if enabledProjectList.Include(projectSFID) {
			invalidProjectSFIDs.Add(projectSFID)
		}
	}
	if invalidProjectSFIDs.Length() != 0 {
		return fmt.Errorf("bad request: invalid project_sfid passed : %v. These project is already enrolled in one of the cla_group", invalidProjectSFIDs.List())
	}

	return nil
}

func (s *service) enrollProjects(claGroupID string, foundationSFID string, projectSFIDList []string) error {
	f := logrus.Fields{"function": "enrollProjects"}
	for _, projectSFID := range projectSFIDList {
		log.WithFields(f).Debugf("associating cla_group with project : %s", projectSFID)
		err := s.projectsClaGroupsRepo.AssociateClaGroupWithProject(claGroupID, projectSFID, foundationSFID)
		if err != nil {
			log.WithFields(f).Errorf("associating cla_group with project : %s failed", projectSFID)
			log.WithFields(f).Debug("deleting stale entries from cla_group project association")
			deleteErr := s.projectsClaGroupsRepo.RemoveProjectAssociatedWithClaGroup(claGroupID, projectSFIDList, false)
			if deleteErr != nil {
				log.WithFields(f).Error("deleting stale entries from cla_group project association failed", deleteErr)
			}
			return err
		}
	}
	return nil
}

func (s *service) CreateCLAGroup(input *models.CreateClaGroupInput, projectManagerLFID string) (*models.ClaGroup, error) {
	f := logrus.Fields{"function": "CreateCLAGroup"}
	// Validate the input
	log.WithFields(f).WithField("input", input).Debugf("validating create cla group input")
	if input.IclaEnabled == nil ||
		input.CclaEnabled == nil ||
		input.CclaRequiresIcla == nil ||
		input.ClaGroupName == nil ||
		input.FoundationSfid == nil {
		return nil, fmt.Errorf("bad request: required parameters are not passed")
	}
	standaloneProject, err := s.validateClaGroupInput(input)
	if err != nil {
		log.WithFields(f).Warnf("validation of create cla group input failed")
		return nil, err
	}

	// Create cla group
	log.WithFields(f).WithField("input", input).Debugf("creating cla group")
	claGroup, err := s.v1ProjectService.CreateCLAGroup(&v1Models.Project{
		FoundationSFID:          *input.FoundationSfid,
		ProjectDescription:      input.ClaGroupDescription,
		ProjectCCLAEnabled:      *input.CclaEnabled,
		ProjectCCLARequiresICLA: *input.CclaRequiresIcla,
		ProjectExternalID:       *input.FoundationSfid,
		ProjectACL:              []string{projectManagerLFID},
		ProjectICLAEnabled:      *input.IclaEnabled,
		ProjectName:             *input.ClaGroupName,
		Version:                 "v2",
	})
	if err != nil {
		log.WithFields(f).Errorf("creating cla group failed. error = %s", err.Error())
		return nil, err
	}
	log.WithFields(f).WithField("cla_group", claGroup).Debugf("cla group created")
	f["cla_group_id"] = claGroup.ProjectID

	// Attach template with cla group
	var templateFields v1Models.CreateClaGroupTemplate
	err = copier.Copy(&templateFields, &input.TemplateFields)
	if err != nil {
		log.WithFields(f).Error("unable to create v1 create cla group template model", err)
		return nil, err
	}
	log.WithFields(f).Debug("attaching cla_group_template")
	if templateFields.TemplateID == "" {
		log.WithFields(f).Debug("using apache style template as template_id is not passed")
		templateFields.TemplateID = v1Template.ApacheStyleTemplateID
	}
	pdfUrls, err := s.v1TemplateService.CreateCLAGroupTemplate(context.Background(), claGroup.ProjectID, &templateFields)
	if err != nil {
		log.WithFields(f).Error("attaching cla_group_template failed", err)
		log.WithFields(f).Debug("deleting created cla group")
		deleteErr := s.v1ProjectService.DeleteCLAGroup(claGroup.ProjectID)
		if deleteErr != nil {
			log.WithFields(f).Error("deleting created cla group failed.", deleteErr)
		}
		return nil, err
	}
	log.WithFields(f).Debug("cla_group_template attached", pdfUrls)

	// Associate projects with cla group

	if standaloneProject {
		// for standalone project, root_project_sfid i.e foundation_sfid and project_sfid
		// would be same
		input.ProjectSfidList = append(input.ProjectSfidList, *input.FoundationSfid)
	}
	err = s.enrollProjects(claGroup.ProjectID, *input.FoundationSfid, input.ProjectSfidList)
	if err != nil {
		log.WithFields(f).Debug("deleting created cla group")
		deleteErr := s.v1ProjectService.DeleteCLAGroup(claGroup.ProjectID)
		if deleteErr != nil {
			log.WithFields(f).Error("deleting created cla group failed.", deleteErr)
		}
		return nil, err
	}

	subProjectList, err := s.projectsClaGroupsRepo.GetProjectsIdsForClaGroup(claGroup.ProjectID)
	if err != nil {
		return nil, err
	}
	var foundationName string
	projectList := make([]*models.ClaGroupProject, 0)
	for _, p := range subProjectList {
		foundationName = p.FoundationName
		if p.ProjectSFID == p.FoundationSFID {
			// For standalone project, we dont need to return same project as subproject
			continue
		}
		projectList = append(projectList, &models.ClaGroupProject{
			ProjectName:       p.ProjectName,
			ProjectSfid:       p.ProjectSFID,
			RepositoriesCount: p.RepositoriesCount,
		})
	}

	return &models.ClaGroup{
		CclaEnabled:         claGroup.ProjectCCLAEnabled,
		CclaPdfURL:          pdfUrls.CorporatePDFURL,
		CclaRequiresIcla:    claGroup.ProjectCCLARequiresICLA,
		ClaGroupDescription: claGroup.ProjectDescription,
		ClaGroupID:          claGroup.ProjectID,
		ClaGroupName:        claGroup.ProjectName,
		FoundationSfid:      claGroup.FoundationSFID,
		FoundationName:      foundationName,
		IclaEnabled:         claGroup.ProjectICLAEnabled,
		IclaPdfURL:          pdfUrls.IndividualPDFURL,
		ProjectList:         projectList,
	}, nil
}

func (s *service) EnrollProjectsInClaGroup(claGroupID string, foundationSFID string, projectSFIDList []string) error {
	f := logrus.Fields{"cla_group_id": claGroupID, "foundation_sfid": foundationSFID, "project_sfid_list": projectSFIDList}
	log.WithFields(f).Debug("validating enroll project input")
	err := s.validateEnrollProjectsInput(foundationSFID, projectSFIDList)
	if err != nil {
		log.WithFields(f).Errorf("validating enroll project input failed. error = %s", err)
		return err
	}
	log.WithFields(f).Debug("validating enroll project input passed")
	log.WithFields(f).Debug("enrolling projects in cla_group")
	err = s.enrollProjects(claGroupID, foundationSFID, projectSFIDList)
	if err != nil {
		log.WithFields(f).Errorf("enrolling projects in cla_group failed. error = %s", err)
		return err
	}
	log.WithFields(f).Debug("projects enrolled successfully in cla_group")
	return nil
}

// DeleteCLAGroup handles deleting and invalidating the CLA group, removing permissions, cleaning up pending requests, etc.
func (s *service) DeleteCLAGroup(claGroupModel *v1Models.Project, authUser *auth.User) error {
	f := logrus.Fields{
		"functionName":             "DeleteCLAGroup",
		"claGroupID":               claGroupModel.ProjectID,
		"claGroupExternalID":       claGroupModel.ProjectExternalID,
		"claGroupName":             claGroupModel.ProjectName,
		"claGroupFoundationSFID":   claGroupModel.FoundationSFID,
		"claGroupVersion":          claGroupModel.Version,
		"claGroupICLAEnabled":      claGroupModel.ProjectICLAEnabled,
		"claGroupCCLAEnabled":      claGroupModel.ProjectCCLAEnabled,
		"claGroupCCLARequiresICLA": claGroupModel.ProjectCCLARequiresICLA,
	}
	log.WithFields(f).Debug("start deleting CLA Group")

	// Delete gerrit repositories
	numDeleted, err := s.gerritService.DeleteClaGroupGerrits(claGroupModel.ProjectID)
	if err != nil {
		log.WithFields(f).Warn(err)
		return err
	}
	if numDeleted > 0 {
		log.WithFields(f).Debugf("deleted %d gerrit repositories", numDeleted)
		// Log gerrit event
		s.eventsService.LogEvent(&events.LogEventArgs{
			EventType:    events.GerritRepositoryDeleted,
			ProjectModel: claGroupModel,
			LfUsername:   authUser.UserName,
			EventData:    &events.GerritProjectDeletedEventData{},
		})
	} else {
		log.WithFields(f).Debug("no gerrit repositories found to delete")
	}

	// Delete github repositories
	numDeleted, delGHReposErr := s.repositoriesService.DeleteProject(claGroupModel.ProjectID)
	if delGHReposErr != nil {
		log.WithFields(f).Warn(delGHReposErr)
		return err
	}
	if numDeleted > 0 {
		log.WithFields(f).Debugf("deleted %d github repositories", numDeleted)
		// Log github delete event
		s.eventsService.LogEvent(&events.LogEventArgs{
			EventType:    events.GithubRepositoryDeleted,
			ProjectModel: claGroupModel,
			LfUsername:   authUser.UserName,
			EventData:    &events.GithubProjectDeletedEventData{},
		})
	} else {
		log.WithFields(f).Debug("no github repositories found to delete")
	}

	// Locate all the signed/approved corporate CLA signature records - need all the Organization IDs so we can
	// remove CLA Manager/CLA Manager Designee/CLA Signatory Permissions
	log.WithFields(f).Debug("locating signed corporate signatures")
	signatureCompanyIDModels, companyIDErr := s.signatureService.GetCompanyIDsWithSignedCorporateSignatures(claGroupModel.ProjectID)
	if companyIDErr != nil {
		log.WithFields(f).Warnf("unable to fetch list of company IDs, error: %+v", companyIDErr)
		return companyIDErr
	}
	log.WithFields(f).Debugf("discovered %d corporate signatures to investigate", len(signatureCompanyIDModels))

	// Invalidate project signatures
	log.WithFields(f).Debug("locating signatures to invalidate")
	numInvalidated, invalidateErr := s.signatureService.InvalidateProjectRecords(claGroupModel.ProjectID, claGroupModel.ProjectName)
	if invalidateErr != nil {
		log.WithFields(f).Warn(invalidateErr)
		return invalidateErr
	}
	if numInvalidated > 0 {
		log.WithFields(f).Debugf("invalidated %d signatures", numInvalidated)
		// Log invalidate signatures
		s.eventsService.LogEvent(&events.LogEventArgs{
			EventType:    events.InvalidatedSignature,
			ProjectModel: claGroupModel,
			LfUsername:   authUser.UserName,
			EventData:    &events.SignatureProjectInvalidatedEventData{},
		})
	} else {
		log.WithFields(f).Debug("no signatures found to invalidate")
	}

	// Search ACS for users with cla-manager role with scope of ProjectSFID|CompanySFID => remove cla-manage role
	oscClient := organization_service.GetClient()

	// Error channel to send back the results
	errChan := make(chan error)

	// Don't remove project-manager or contributor roles (association with org)
	// Basically, we want to clean up all who have: Project|Organization scope (corporate console stuff)
	// For each organization/company...
	for _, signatureCompanyIDModel := range signatureCompanyIDModels {
		go func(signatureCompanyIDModel signatureService.SignatureCompanyID, projectSFID string, authUser *auth.User) {
			// Additional fields for logging
			f["companySFID"] = signatureCompanyIDModel.CompanySFID
			f["companyID"] = signatureCompanyIDModel.CompanyID
			f["companyName"] = signatureCompanyIDModel.CompanyName

			log.WithFields(f).Debugf("locating CLA Manager requests for company: %s", signatureCompanyIDModel.CompanyName)
			// Fetch any pending CLA manager requests for this company/project
			requestList, requestErr := s.claManagerRequests.GetRequests(signatureCompanyIDModel.CompanyID, claGroupModel.ProjectID)
			if requestErr != nil {
				log.WithFields(f).Warn(requestErr)
				errChan <- requestErr
				return
			}

			// If we have any CLA manager requests - delete them
			if requestList != nil && len(requestList.Requests) > 0 {
				log.WithFields(f).Debugf("removing %d CLA Manager Requests found for company and project", len(requestList.Requests))
				for _, request := range requestList.Requests {
					reqDelErr := s.claManagerRequests.DeleteRequest(request.RequestID)
					log.WithFields(f).Warn(reqDelErr)
					errChan <- reqDelErr
					return
				}
			} else {
				log.WithFields(f).Debug("no CLA Manager Requests found for company and project")
			}

			log.WithFields(f).Debugf("removing permissions for CLA Managers, CLA Manager Designees, and CLA Signatories for company: %s", signatureCompanyIDModel.CompanyName)
			// CLA Managers
			claMgrErr := oscClient.DeleteRolePermissions(signatureCompanyIDModel.CompanySFID, projectSFID, "cla-manager", authUser)
			if claMgrErr != nil {
				log.WithFields(f).Warn(err)
				errChan <- claMgrErr
				return
			}

			// CLA Manager Designee
			claMgrDesigneeErr := oscClient.DeleteRolePermissions(signatureCompanyIDModel.CompanySFID, projectSFID, "cla-manager-designee", authUser)
			if claMgrDesigneeErr != nil {
				log.WithFields(f).Warn(err)
				errChan <- claMgrDesigneeErr
				return
			}

			// CLA Signatories
			claSignatoryErr := oscClient.DeleteRolePermissions(signatureCompanyIDModel.CompanySFID, projectSFID, "cla-signatory", authUser)
			if claSignatoryErr != nil {
				log.WithFields(f).Warn(err)
				errChan <- claSignatoryErr
				return
			}

			// No errors - nice...return nil
			errChan <- nil
		}(signatureCompanyIDModel, claGroupModel.ProjectID, authUser)
	}

	// Process the results
	for range signatureCompanyIDModels {
		errFromFunc := <-errChan
		if errFromFunc != nil {
			log.WithFields(f).Warnf("problem removing removing requests or removing permissions, error: %+v - continuing with CLA Group delete", errFromFunc)
		}
	}

	// clear fields for logging - don't need them anymore
	delete(f, "companySFID")
	delete(f, "companyID")
	delete(f, "companyName")

	// Is this done via trigger?
	log.WithFields(f).Debug("deleting cla_group project associations")
	err = s.projectsClaGroupsRepo.RemoveProjectAssociatedWithClaGroup(claGroupModel.ProjectID, []string{}, true)
	if err != nil {
		return nil
	}

	log.WithFields(f).Debug("deleting cla_group from dynamodb")
	err = s.v1ProjectService.DeleteCLAGroup(claGroupModel.ProjectID)
	if err != nil {
		log.WithFields(f).Errorf("deleting cla_group from dynamodb failed. error = %s", err.Error())
		return err
	}

	return nil
}

func getS3Url(claGroupID string, docs []v1Models.ProjectDocument) string {
	if len(docs) == 0 {
		return ""
	}
	var version int64
	var url string
	for _, doc := range docs {
		maj, err := strconv.Atoi(doc.DocumentMajorVersion)
		if err != nil {
			log.WithField("cla_group_id", claGroupID).Error("invalid major number in cla_group")
			continue
		}
		min, err := strconv.Atoi(doc.DocumentMinorVersion)
		if err != nil {
			log.WithField("cla_group_id", claGroupID).Error("invalid minor number in cla_group")
			continue
		}
		docVersion := int64(maj)<<32 | int64(min)
		if docVersion > version {
			url = doc.DocumentS3URL
		}
	}
	return url
}

// ListClaGroupsForFoundationOrProject returns the CLA Group list for the specified foundation ID
func (s *service) ListClaGroupsForFoundationOrProject(foundationSFID string) (*models.ClaGroupList, error) {
	out := &models.ClaGroupList{List: make([]*models.ClaGroup, 0)}
	v1ClaGroups, err := s.v1ProjectService.GetClaGroupsByFoundationSFID(foundationSFID, DontLoadDetails)
	if err != nil {
		return nil, err
	}

	m := make(map[string]*models.ClaGroup)
	claGroupIDList := utils.NewStringSet()
	for _, v1ClaGroup := range v1ClaGroups.Projects {

		// Lookup the foundation name
		var foundationName = "Not Defined"
		projectServiceModel, projErr := v2ProjectService.GetClient().GetProject(v1ClaGroup.FoundationSFID)
		if projErr != nil {
			log.Warnf("unable to lookup foundation SFID: %s - error: %+v - using 'Not Defined' as the default value",
				v1ClaGroup.FoundationSFID, projErr)
		} else {
			foundationName = projectServiceModel.Name
		}

		cg := &models.ClaGroup{
			CclaEnabled:         v1ClaGroup.ProjectCCLAEnabled,
			CclaRequiresIcla:    v1ClaGroup.ProjectCCLARequiresICLA,
			ClaGroupDescription: v1ClaGroup.ProjectDescription,
			ClaGroupID:          v1ClaGroup.ProjectID,
			ClaGroupName:        v1ClaGroup.ProjectName,
			FoundationSfid:      v1ClaGroup.FoundationSFID,
			FoundationName:      foundationName,
			IclaEnabled:         v1ClaGroup.ProjectICLAEnabled,
			CclaPdfURL:          getS3Url(v1ClaGroup.ProjectID, v1ClaGroup.ProjectCorporateDocuments),
			IclaPdfURL:          getS3Url(v1ClaGroup.ProjectID, v1ClaGroup.ProjectIndividualDocuments),
			ProjectList:         make([]*models.ClaGroupProject, 0),
			// Add root_project_repositories_count to repositories_count initially
			RepositoriesCount:            v1ClaGroup.RootProjectRepositoriesCount,
			RootProjectRepositoriesCount: v1ClaGroup.RootProjectRepositoriesCount,
		}
		claGroupIDList.Add(cg.ClaGroupID)
		m[cg.ClaGroupID] = cg
	}

	// Fill projectSFID list in cla group
	cgprojects, err := s.projectsClaGroupsRepo.GetProjectsIdsForFoundation(foundationSFID)
	if err != nil {
		return nil, err
	}
	for _, cgproject := range cgprojects {
		if cgproject.ProjectSFID == cgproject.FoundationSFID {
			// dont include itself as subproject
			continue
		}
		cg, ok := m[cgproject.ClaGroupID]
		if !ok {
			log.Warnf("stale data present in cla-group-projects table. cla_group_id : %s", cgproject.ClaGroupID)
			continue
		}
		cg.ProjectList = append(cg.ProjectList, &models.ClaGroupProject{
			ProjectSfid:       cgproject.ProjectSFID,
			ProjectName:       cgproject.ProjectName,
			RepositoriesCount: cgproject.RepositoriesCount,
		})
		cg.RepositoriesCount += cgproject.RepositoriesCount
	}
	cgmetrics := s.getMetrics(claGroupIDList.List())

	// now build output array
	for _, cg := range m {
		pm, ok := cgmetrics[cg.ClaGroupID]
		if ok {
			cg.TotalSignatures = pm.CorporateContributorsCount + pm.IndividualContributorsCount
		}
		out.List = append(out.List, cg)
	}

	// Sort the response based on the Foundation and CLA group name
	sort.Slice(out.List, func(i, j int) bool {
		switch strings.Compare(out.List[i].FoundationName, out.List[j].FoundationName) {
		case -1:
			return true
		case 1:
			return false
		}
		return out.List[i].ClaGroupName < out.List[j].ClaGroupName
	})

	return out, nil
}

func (s *service) getMetrics(claGroupIDList []string) map[string]*metrics.ProjectMetric {
	m := make(map[string]*metrics.ProjectMetric)
	type result struct {
		claGroupID string
		metric     *metrics.ProjectMetric
		err        error
	}
	rchan := make(chan *result)
	var wg sync.WaitGroup
	wg.Add(len(claGroupIDList))
	go func() {
		wg.Wait()
		close(rchan)
	}()
	for _, cgid := range claGroupIDList {
		go func(swg *sync.WaitGroup, claGroupID string, resultChan chan *result) {
			defer swg.Done()
			metric, err := s.metricsRepo.GetProjectMetric(claGroupID)
			resultChan <- &result{
				claGroupID: claGroupID,
				metric:     metric,
				err:        err,
			}
		}(&wg, cgid, rchan)
	}
	for r := range rchan {
		if r.err != nil {
			log.WithField("cla_group_id", r.claGroupID).Error("unable to get cla_group metrics")
			continue
		}
		m[r.claGroupID] = r.metric
	}
	return m
}

func (s *service) ListAllFoundationClaGroups(foundationID *string) (*models.FoundationMappingList, error) {
	var out []*projects_cla_groups.ProjectClaGroup
	var err error
	if foundationID != nil {
		out, err = s.projectsClaGroupsRepo.GetProjectsIdsForFoundation(*foundationID)
	} else {
		out, err = s.projectsClaGroupsRepo.GetProjectsIdsForAllFoundation()
	}
	if err != nil {
		return nil, err
	}
	return toFoundationMapping(out), nil
}

func toFoundationMapping(list []*projects_cla_groups.ProjectClaGroup) *models.FoundationMappingList {
	out := &models.FoundationMappingList{List: make([]*models.FoundationMapping, 0)}
	foundationMap := make(map[string]*models.FoundationMapping)
	claGroups := make(map[string]*models.ClaGroupProjects)
	for _, in := range list {
		cgp, ok := claGroups[in.ClaGroupID]
		if !ok {
			cgp = &models.ClaGroupProjects{
				ClaGroupID:      in.ClaGroupID,
				ProjectSfidList: []string{in.ProjectSFID},
			}
			claGroups[in.ClaGroupID] = cgp
			foundation, ok := foundationMap[in.FoundationSFID]
			if !ok {
				foundation = &models.FoundationMapping{
					ClaGroups:      []*models.ClaGroupProjects{cgp},
					FoundationSfid: in.FoundationSFID,
				}
				foundationMap[in.FoundationSFID] = foundation
				out.List = append(out.List, foundation)
			} else {
				foundation.ClaGroups = append(foundation.ClaGroups, cgp)
			}
		} else {
			cgp.ProjectSfidList = append(cgp.ProjectSfidList, in.ProjectSFID)
		}
	}
	return out
}
