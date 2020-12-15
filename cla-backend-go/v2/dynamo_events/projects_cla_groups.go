// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package dynamo_events

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations/signatures"
	organization_service "github.com/communitybridge/easycla/cla-backend-go/v2/organization-service"
	user_service "github.com/communitybridge/easycla/cla-backend-go/v2/user-service"

	acs_service "github.com/communitybridge/easycla/cla-backend-go/v2/acs-service"

	"github.com/aws/aws-lambda-go/events"
	claEvents "github.com/communitybridge/easycla/cla-backend-go/events"
	"github.com/communitybridge/easycla/cla-backend-go/gen/models"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/communitybridge/easycla/cla-backend-go/utils"
	v2ProjectService "github.com/communitybridge/easycla/cla-backend-go/v2/project-service"
	"github.com/sirupsen/logrus"
)

// ProjectClaGroup is database model for projects_cla_group table
type ProjectClaGroup struct {
	ProjectSFID       string `json:"project_sfid"`
	ClaGroupID        string `json:"cla_group_id"`
	FoundationSFID    string `json:"foundation_sfid"`
	RepositoriesCount int64  `json:"repositories_count"`
}

// ProjectServiceEnableCLAServiceHandler handles enabling the CLA Service attribute from the project service
func (s *service) ProjectServiceEnableCLAServiceHandler(event events.DynamoDBEventRecord) error {
	f := logrus.Fields{
		"functionName": "ProjectServiceEnableCLAServiceHandler",
		"eventID":      event.EventID,
		"eventName":    event.EventName,
		"eventSource":  event.EventSource,
	}

	log.WithFields(f).Debug("processing request")
	var newProject ProjectClaGroup
	err := unmarshalStreamImage(event.Change.NewImage, &newProject)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("project decoding add event")
		return err
	}

	f["projectSFID"] = newProject.ProjectSFID
	f["claGroupID"] = newProject.ClaGroupID
	f["foundationSFID"] = newProject.FoundationSFID

	psc := v2ProjectService.GetClient()
	log.WithFields(f).Debug("enabling CLA service...")
	start, _ := utils.CurrentTime()
	err = psc.EnableCLA(newProject.ProjectSFID)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("enabling CLA service failed")
		return err
	}
	finish, _ := utils.CurrentTime()
	log.WithFields(f).Debugf("enabling CLA service completed - took: %s", finish.Sub(start).String())

	// Log the event
	eventErr := s.eventsRepo.CreateEvent(&models.Event{
		ContainsPII:            false,
		EventData:              fmt.Sprintf("enabled CLA service for project: %s", newProject.ProjectSFID),
		EventSummary:           fmt.Sprintf("enabled CLA service for project: %s", newProject.ProjectSFID),
		EventFoundationSFID:    newProject.FoundationSFID,
		EventProjectExternalID: newProject.ProjectSFID,
		EventProjectID:         newProject.ClaGroupID,
		EventProjectSFID:       newProject.ProjectSFID,
		EventType:              claEvents.ProjectServiceCLAEnabled,
		LfUsername:             "easycla system",
		UserID:                 "easycla system",
		UserName:               "easycla system",
		// EventProjectName:       "",
		// EventProjectSFName:     "",
	})
	if eventErr != nil {
		log.WithFields(f).WithError(eventErr).Warn("problem logging event for enabling CLA service")
		// Ok - don't fail for now
	}

	return nil
}

// ProjectServiceDisableCLAServiceHandler handles disabling/removing the CLA Service attribute from the project service
func (s *service) ProjectServiceDisableCLAServiceHandler(event events.DynamoDBEventRecord) error {
	f := logrus.Fields{
		"functionName": "ProjectServiceDisableCLAServiceHandler",
		"eventID":      event.EventID,
		"eventName":    event.EventName,
		"eventSource":  event.EventSource,
	}

	log.WithFields(f).Debug("processing request")
	var oldProject ProjectClaGroup
	err := unmarshalStreamImage(event.Change.OldImage, &oldProject)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("problem unmarshalling stream image")
		return err
	}

	// Add more fields for the logger
	f["ProjectSFID"] = oldProject.ProjectSFID
	f["ClaGroupID"] = oldProject.ClaGroupID
	f["FoundationSFID"] = oldProject.FoundationSFID

	psc := v2ProjectService.GetClient()
	// Gathering metrics - grab the time before the API call
	before, _ := utils.CurrentTime()
	log.WithFields(f).Debug("disabling CLA service")
	err = psc.DisableCLA(oldProject.ProjectSFID)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("disabling CLA service failed")
		return err
	}
	log.WithFields(f).Debugf("disabling CLA service took %s", time.Since(before).String())

	// Log the event
	eventErr := s.eventsRepo.CreateEvent(&models.Event{
		ContainsPII:            false,
		EventData:              fmt.Sprintf("disabled CLA service for project: %s", oldProject.ProjectSFID),
		EventSummary:           fmt.Sprintf("disabled CLA service for project: %s", oldProject.ProjectSFID),
		EventFoundationSFID:    oldProject.FoundationSFID,
		EventProjectExternalID: oldProject.ProjectSFID,
		EventProjectID:         oldProject.ClaGroupID,
		EventProjectSFID:       oldProject.ProjectSFID,
		EventType:              claEvents.ProjectServiceCLADisabled,
		LfUsername:             "easycla system",
		UserID:                 "easycla system",
		UserName:               "easycla system",
		// EventProjectName:       "",
		// EventProjectSFName:     "",
	})
	if eventErr != nil {
		log.WithFields(f).WithError(eventErr).Warn("problem logging event for disabling CLA service")
		// Ok - don't fail for now
	}

	return nil
}

func (s *service) ProjectUnenrolledDisableRepositoryHandler(event events.DynamoDBEventRecord) error {
	ctx := utils.NewContext()
	f := logrus.Fields{
		"functionName":   "ProjectUnenrolledDisableRepositoryHandler",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"eventID":        event.EventID,
		"eventName":      event.EventName,
		"eventSource":    event.EventSource,
	}

	log.WithFields(f).Debug("processing request")
	var oldProject ProjectClaGroup
	err := unmarshalStreamImage(event.Change.OldImage, &oldProject)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("problem unmarshalling stream image")
		return err
	}

	// Add more fields for the logger
	f["ProjectSFID"] = oldProject.ProjectSFID
	f["ClaGroupID"] = oldProject.ClaGroupID
	f["FoundationSFID"] = oldProject.FoundationSFID

	// Disable GitHub repos associated with this project
	enabled := true // only care about enabled repos
	gitHubRepos, githubRepoErr := s.repositoryService.GetRepositoryByProjectSFID(ctx, oldProject.ProjectSFID, &enabled)
	if githubRepoErr != nil {
		log.WithFields(f).WithError(githubRepoErr).Warn("problem listing github repositories by project sfid")
		return githubRepoErr
	}
	if gitHubRepos != nil && len(gitHubRepos.List) > 0 {
		log.WithFields(f).Debugf("discovered %d github repositories for project with sfid: %s - disabling repositories...",
			len(gitHubRepos.List), oldProject.ProjectSFID)

		// For each GitHub repository...
		for _, gitHubRepo := range gitHubRepos.List {
			log.WithFields(f).Debugf("disabling github repository: %s with id: %s for project with sfid: %s",
				gitHubRepo.RepositoryName, gitHubRepo.RepositoryID, gitHubRepo.ProjectSFID)
			disableErr := s.repositoryService.DisableRepository(ctx, gitHubRepo.RepositoryID)
			if disableErr != nil {
				log.WithFields(f).WithError(disableErr).Warnf("problem disabling github repository: %s with id: %s", gitHubRepo.RepositoryName, gitHubRepo.RepositoryID)
				return disableErr
			}
		}
	} else {
		log.WithFields(f).Debugf("no github repositories for project with sfid: %s - nothing to disable",
			oldProject.ProjectSFID)
	}

	gerrits, gerritRepoErr := s.gerritService.GetGerritsByProjectSFID(ctx, oldProject.ProjectSFID)
	if gerritRepoErr != nil {
		log.WithFields(f).WithError(gerritRepoErr).Warn("problem listing gerrit repositories by project sfid")
		return gerritRepoErr
	}
	if gerrits != nil && len(gerrits.List) > 0 {
		log.WithFields(f).Debugf("discovered %d gerrit repositories for project with sfid: %s - deleting gerrit instances...",
			len(gerrits.List), oldProject.ProjectSFID)
		for _, gerritRepo := range gerrits.List {
			log.WithFields(f).Debugf("deleting gerrit instance: %s with id: %s for project with sfid: %s",
				gerritRepo.GerritName, gerritRepo.GerritID.String(), gerritRepo.ProjectSFID)
			gerritDeleteErr := s.gerritService.DeleteGerrit(ctx, gerritRepo.GerritID.String())
			if gerritDeleteErr != nil {
				log.WithFields(f).WithError(gerritDeleteErr).Warnf("problem deleting gerrit instance: %s with id: %s",
					gerritRepo.GerritName, gerritRepo.GerritID.String())
				return gerritDeleteErr
			}
		}
	} else {
		log.WithFields(f).Debugf("no gerrit instances for project with sfid: %s - nothing to delete",
			oldProject.ProjectSFID)
	}

	return nil
}

// AddCLAPermissions handles adding CLA permissions
func (s *service) AddCLAPermissions(event events.DynamoDBEventRecord) error {
	f := logrus.Fields{
		"functionName": "AddCLAPermissions",
		"eventID":      event.EventID,
		"eventName":    event.EventName,
		"eventSource":  event.EventSource,
	}

	log.WithFields(f).Debug("processing event")
	var newProject ProjectClaGroup
	err := unmarshalStreamImage(event.Change.NewImage, &newProject)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("problem unmarshalling stream image")
		return err
	}

	// Add more fields for the logger
	f["ProjectSFID"] = newProject.ProjectSFID
	f["ClaGroupID"] = newProject.ClaGroupID
	f["FoundationSFID"] = newProject.FoundationSFID

	// Add any relevant CLA related permissions for this CLA Group/Project SFID
	permErr := s.addCLAPermissions(newProject.ClaGroupID, newProject.ProjectSFID)
	if permErr != nil {
		log.WithFields(f).WithError(permErr).Warn("problem removing CLA permissions for projectSFID")
		// Ok - don't fail for now
	}

	return nil
}

// RemoveCLAPermissions handles removing existing CLA permissions
func (s *service) RemoveCLAPermissions(event events.DynamoDBEventRecord) error {
	f := logrus.Fields{
		"functionName": "RemoveCLAPermissions",
		"eventID":      event.EventID,
		"eventName":    event.EventName,
		"eventSource":  event.EventSource,
	}

	log.WithFields(f).Debug("processing event")
	var oldProject ProjectClaGroup
	err := unmarshalStreamImage(event.Change.OldImage, &oldProject)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("problem unmarshalling stream image")
		return err
	}

	// Add more fields for the logger
	f["ProjectSFID"] = oldProject.ProjectSFID
	f["ClaGroupID"] = oldProject.ClaGroupID
	f["FoundationSFID"] = oldProject.FoundationSFID

	// Remove any CLA related permissions
	permErr := s.removeCLAPermissions(oldProject.ProjectSFID)
	if permErr != nil {
		log.WithFields(f).WithError(permErr).Warn("problem removing CLA permissions for projectSFID")
		// Ok - don't fail for now
	}

	return nil
}

// addCLAPermissions handles adding the CLA Group (projects table) permissions for the specified project group (foundation) and project
func (s *service) addCLAPermissions(claGroupID, projectSFID string) error {
	ctx := utils.NewContext()
	f := logrus.Fields{
		"functionName": "addCLAPermissions",
		"projectSFID":  projectSFID,
		"claGroupID":   claGroupID,
	}
	log.WithFields(f).Debug("adding CLA permissions...")

	sigModels, err := s.signatureRepo.GetProjectSignatures(ctx, signatures.GetProjectSignaturesParams{
		ClaType:   aws.String(utils.ClaTypeCCLA),
		PageSize:  aws.Int64(1000),
		ProjectID: claGroupID,
	}, 1000)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("problem querying CCLA signatures for CLA Group - skipping %s role review/assignment for this project", utils.CLAManagerRole)
		return err
	}
	if sigModels == nil || len(sigModels.Signatures) == 0 {
		log.WithFields(f).WithError(err).Warnf("no signatures found CLA Group - unable to determine existing CLA Managers - skipping %s role review/assignment for this project", utils.CLAManagerRole)
		return err
	}

	// ACS Client
	acsClient := acs_service.GetClient()
	log.WithFields(f).Debugf("locating role ID for role: %s", utils.CLAManagerRole)
	claManagerRoleID, roleErr := acsClient.GetRoleID(utils.CLAManagerRole)
	if roleErr != nil {
		log.WithFields(f).Warnf("problem looking up details for role: %s, error: %+v", utils.CLAManagerRole, roleErr)
		return roleErr
	}
	orgClient := organization_service.GetClient()
	userClient := user_service.GetClient()

	// For each signature...
	for _, sig := range sigModels.Signatures {

		// Make sure we can load the company and grab the SFID
		sig := sig
		companyInternalID := sig.SignatureReferenceID.String()
		log.WithFields(f).Debugf("locating company by internal ID: %s", companyInternalID)
		companyModel, err := s.companyRepo.GetCompany(ctx, companyInternalID)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("problem loading company by internal ID: %s - skipping %s role review/assignment for this project", companyInternalID, utils.CLAManagerRole)
			continue
		}
		if companyModel == nil || companyModel.CompanyExternalID == "" {
			log.WithFields(f).WithError(err).Warnf("problem loading company ID: %s or external SFID for company not set - skipping %s role review/assignment for this project", companyInternalID, utils.CLAManagerRole)
			continue
		}
		log.WithFields(f).Debugf("loaded company by internal ID: %s with name: %s", companyInternalID, companyModel.CompanyName)
		companySFID := companyModel.CompanyExternalID

		// Make sure we can load the CLA Manger list (ACL)
		if len(sig.SignatureACL) == 0 {
			log.WithFields(f).Warnf("no CLA Manager list (acl) established for signature %s - skipping %s role review/assigment for this project", sig.SignatureID, utils.CLAManagerRole)
			continue
		}
		existingCLAManagers := sig.SignatureACL

		var wg sync.WaitGroup
		wg.Add(len(existingCLAManagers))

		// For each CLA manager for this company...
		log.WithFields(f).Debugf("processing %d CLA managers for company ID: %s/%s with name: %s", len(existingCLAManagers), companyInternalID, companySFID, companyModel.CompanyName)
		for _, signatureUserModel := range existingCLAManagers {
			// handle unpredictability with addresses o0f different signatureUserModel
			signatureUserModel := signatureUserModel
			go func(signatureUserModel models.User) {
				defer wg.Done()

				log.WithFields(f).Debugf("looking up existing CLA manager by LF username: %s...", signatureUserModel.LfUsername)
				userModel, userLookupErr := userClient.GetUserByUsername(signatureUserModel.LfUsername)
				if userLookupErr != nil {
					log.WithFields(f).WithError(userLookupErr).Warnf("unable to lookup user %s - skipping %s role review/assigment for this project",
						signatureUserModel.LfUsername, utils.CLAManagerRole)
					return
				}
				if userModel == nil || userModel.ID == "" || userModel.Email == nil {
					log.WithFields(f).Warnf("unable to lookup user %s - user object is empty or missing either the ID or email - skipping %s role review/assigment for project: %s, company: %s",
						signatureUserModel.LfUsername, utils.CLAManagerRole, projectSFID, companySFID)
					return
				}

				// Determine if the user already has the cla-manager role scope for this Project and Company
				hasRole, roleLookupErr := orgClient.IsUserHaveRoleScope(utils.CLAManagerRole, userModel.ID, companySFID, projectSFID)
				if roleLookupErr != nil {
					log.WithFields(f).WithError(roleLookupErr).Warnf("unable to lookup role scope %s for user %s/%s - skipping %s role review/assigment for this project",
						utils.CLAManagerRole, signatureUserModel.LfUsername, userModel.ID, utils.CLAManagerRole)
					return
				}

				// Does the user already have the cla-manager role?
				if hasRole {
					log.WithFields(f).Debugf("user %s/%s already has role %s for the project %s and organization %s",
						signatureUserModel.LfUsername, userModel.ID, utils.CLAManagerRole, projectSFID, companySFID)
					// Nothing to do here - move along...
					return
				}

				// Finally....assign the role to this user
				roleErr := orgClient.CreateOrgUserRoleOrgScopeProjectOrg(aws.StringValue(userModel.Email), projectSFID, companySFID, claManagerRoleID)
				if roleErr != nil {
					log.WithFields(f).WithError(roleErr).Warnf("%s, role assignment for user user %s/%s/%s failed for this project: %s, company: %s",
						utils.CLAManagerRole, signatureUserModel.LfUsername, userModel.ID, *userModel.Email, projectSFID, companySFID)
					return
				}

			}(signatureUserModel)
		}

		// Wait for the go routines to finish
		log.WithFields(f).Debugf("waiting for role assignment to complete for %d project: %s", len(sigModels.Signatures), projectSFID)
		wg.Wait()
	}

	return nil
}

// removeCLAPermissions handles removing CLA Group (projects table) permissions for the specified project
func (s *service) removeCLAPermissions(projectSFID string) error {
	f := logrus.Fields{
		"functionName": "removeCLAPermissions",
		"projectSFID":  projectSFID,
	}
	log.WithFields(f).Debug("removing CLA permissions...")

	client := acs_service.GetClient()
	err := client.RemoveCLAUserRolesByProject(projectSFID, []string{utils.CLAManagerRole, utils.CLADesigneeRole, utils.CLASignatoryRole})
	if err != nil {
		log.WithFields(f).WithError(err).Warn("problem removing CLA user roles by projectSFID")
	}

	return err
}

// removeCLAPermissionsByProjectOrganizationRole handles removal of the specified role for the given SF Project and SF Organization
func (s *service) removeCLAPermissionsByProjectOrganizationRole(projectSFID, organizationSFID string, roleNames []string) error {
	f := logrus.Fields{
		"functionName":     "removeCLAPermissionsByProjectOrganizationRole",
		"projectSFID":      projectSFID,
		"organizationSFID": organizationSFID,
		"roleNames":        strings.Join(roleNames, ","),
	}

	log.WithFields(f).Debug("removing CLA permissions...")
	client := acs_service.GetClient()
	err := client.RemoveCLAUserRolesByProjectOrganization(projectSFID, organizationSFID, roleNames)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("problem removing CLA user roles by projectSFID and organizationSFID")
	}

	return err
}
