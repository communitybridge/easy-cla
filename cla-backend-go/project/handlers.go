// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package project

import (
	"context"
	"fmt"

	"github.com/communitybridge/easycla/cla-backend-go/utils"

	"github.com/sirupsen/logrus"

	"github.com/communitybridge/easycla/cla-backend-go/events"
	"github.com/communitybridge/easycla/cla-backend-go/gerrits"
	"github.com/communitybridge/easycla/cla-backend-go/repositories"
	"github.com/communitybridge/easycla/cla-backend-go/signatures"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/communitybridge/easycla/cla-backend-go/gen/models"
	"github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations"
	"github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations/project"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/communitybridge/easycla/cla-backend-go/user"

	"github.com/go-openapi/runtime/middleware"
)

const defaultPageSize int64 = 50

// isValidUser is a helper function to determine if the user object is valid
func isValidUser(claUser *user.CLAUser) bool {
	return claUser != nil && claUser.UserID != "" && claUser.LFUsername != "" && claUser.LFEmail != ""
}

// Configure establishes the middleware handlers for the project service
func Configure(api *operations.ClaAPI, service Service, eventsService events.Service, gerritService gerrits.Service, repositoryService repositories.Service, signatureService signatures.SignatureService) {
	// Create CLA Group/Project Handler
	api.ProjectCreateProjectHandler = project.CreateProjectHandlerFunc(func(params project.CreateProjectParams, claUser *user.CLAUser) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		if params.Body.ProjectName == "" || params.Body.ProjectACL == nil {
			msg := "Missing Project Name or Project ACL parameter."
			log.Warnf("Create Project Failed - %s", msg)
			return project.NewCreateProjectBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Code:    "400",
				Message: msg,
			})
		}

		exitingModel, getErr := service.GetCLAGroupByName(ctx, params.Body.ProjectName)
		if getErr != nil {
			msg := fmt.Sprintf("Error querying the project by name, error: %+v", getErr)
			log.Warnf("Create Project Failed - %s", msg)
			return project.NewCreateProjectBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Code:    "400",
				Message: msg,
			})
		}

		// If the project with the same name exists...
		if exitingModel != nil {
			msg := fmt.Sprintf("Project with same name exists: %s", params.Body.ProjectName)
			log.Warnf("Create Project Failed - %s", msg)
			return project.NewCreateProjectConflict().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Code:    "409",
				Message: msg,
			})
		}

		// Ok, safe to create now
		claGroupModel, err := service.CreateCLAGroup(ctx, &params.Body)
		if err != nil {
			log.Warnf("Create Project Failed - %+v", err)
			return project.NewCreateProjectBadRequest().WithPayload(errorResponse(err))
		}

		eventsService.LogEvent(&events.LogEventArgs{
			EventType:     events.CLAGroupCreated,
			ClaGroupModel: claGroupModel,
			UserID:        claUser.UserID,
			EventData:     &events.CLAGroupCreatedEventData{},
		})

		log.Infof("Create Project Succeeded, project name: %s, project external ID: %s",
			params.Body.ProjectName, params.Body.ProjectExternalID)
		return project.NewCreateProjectOK().WithXRequestID(reqID).WithPayload(claGroupModel)
	})

	// Get Projects
	api.ProjectGetProjectsHandler = project.GetProjectsHandlerFunc(func(params project.GetProjectsParams, claUser *user.CLAUser) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		if !isValidUser(claUser) {
			return project.NewGetProjectsUnauthorized().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: "unauthorized",
				Code:    "401",
			})
		}
		projects, err := service.GetCLAGroups(ctx, &params)
		if err != nil {
			return project.NewGetProjectsBadRequest().WithXRequestID(reqID).WithPayload(errorResponse(err))
		}

		return project.NewGetProjectsOK().WithXRequestID(reqID).WithPayload(projects)
	})

	// Get Project By ID
	api.ProjectGetProjectByIDHandler = project.GetProjectByIDHandlerFunc(func(params project.GetProjectByIDParams, claUser *user.CLAUser) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint

		claGroupModel, err := service.GetCLAGroupByID(ctx, params.ProjectID)
		if err != nil {
			return project.NewGetProjectByIDBadRequest().WithXRequestID(reqID).WithPayload(errorResponse(err))
		}
		if claGroupModel == nil {
			return project.NewGetProjectByIDNotFound().WithXRequestID(reqID)
		}

		return project.NewGetProjectByIDOK().WithXRequestID(reqID).WithPayload(claGroupModel)
	})

	// Get Project By External ID Handler
	api.ProjectGetProjectsByExternalIDHandler = project.GetProjectsByExternalIDHandlerFunc(func(params project.GetProjectsByExternalIDParams, claUser *user.CLAUser) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint

		log.Debugf("Project Handler - GetProjectsByExternalID")
		if params.ProjectSFID == "" {
			return project.NewGetProjectsByExternalIDBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Code:    "400",
				Message: "External ID is empty",
			})
		}

		// Set the default page size
		if params.PageSize == nil {
			params.PageSize = aws.Int64(defaultPageSize)
		}

		log.Debugf("Project Handler - GetProjectsByExternalID - invoking service")
		projectsModel, err := service.GetCLAGroupsByExternalID(ctx, &params)
		if err != nil {
			return project.NewGetProjectsByExternalIDBadRequest().WithXRequestID(reqID).WithPayload(errorResponse(err))
		}
		if projectsModel == nil {
			return project.NewGetProjectsByExternalIDNotFound().WithXRequestID(reqID)
		}

		return project.NewGetProjectsByExternalIDOK().WithXRequestID(reqID).WithPayload(projectsModel)
	})

	// Get Project By Name
	api.ProjectGetProjectByNameHandler = project.GetProjectByNameHandlerFunc(func(params project.GetProjectByNameParams, claUser *user.CLAUser) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint

		claGroupModel, err := service.GetCLAGroupByName(ctx, params.ProjectName)
		if err != nil {
			return project.NewGetProjectByNameBadRequest().WithXRequestID(reqID).WithPayload(errorResponse(err))
		}
		if claGroupModel == nil {
			return project.NewGetProjectByNameNotFound().WithXRequestID(reqID)
		}

		return project.NewGetProjectByNameOK().WithXRequestID(reqID).WithPayload(claGroupModel)
	})

	// Delete Project By ID
	api.ProjectDeleteProjectByIDHandler = project.DeleteProjectByIDHandlerFunc(func(params project.DeleteProjectByIDParams, claUser *user.CLAUser) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		f := logrus.Fields{
			"functionName":                "ProjectDeleteProjectByIDHandler",
			utils.XREQUESTID:              ctx.Value(utils.XREQUESTID),
			"claGroupID":                  params.ProjectID,
			"authenticatedUserLFUsername": claUser.LFUsername,
			"authenticatedUserLFEmail":    claUser.LFEmail,
			"authenticatedUserUserID":     claUser.UserID,
			"authenticatedUserName":       claUser.Name,
		}
		log.WithFields(f).Debug("Processing delete request")
		claGroupModel, err := service.GetCLAGroupByID(ctx, params.ProjectID)
		if err != nil {
			if err == ErrProjectDoesNotExist {
				return project.NewDeleteProjectByIDNotFound().WithXRequestID(reqID)
			}
			return project.NewDeleteProjectByIDBadRequest().WithXRequestID(reqID).WithPayload(errorResponse(err))
		}

		// Delete gerrit repositories
		log.WithFields(f).Debug("Processing gerrit delete")
		howMany, err := gerritService.DeleteClaGroupGerrits(ctx, params.ProjectID)
		if err != nil {
			return project.NewDeleteProjectByIDBadRequest().WithXRequestID(reqID).WithPayload(errorResponse(err))
		}
		// Log gerrit event
		if howMany > 0 {
			log.WithFields(f).Debugf("Deleted %d gerrit groups", howMany)
			eventsService.LogEvent(&events.LogEventArgs{
				EventType:     events.GerritRepositoryDeleted,
				ClaGroupModel: claGroupModel,
				UserID:        claUser.UserID,
				EventData: &events.GerritProjectDeletedEventData{
					DeletedCount: howMany,
				},
			})
		}

		// Delete github repositories
		log.WithFields(f).Debug("Processing github repository disable/delete")
		howMany, err = repositoryService.DisableRepositoriesByProjectID(ctx, params.ProjectID)
		if err != nil {
			return project.NewDeleteProjectByIDBadRequest().WithXRequestID(reqID).WithPayload(errorResponse(err))
		}
		if howMany > 0 {
			log.WithFields(f).Debugf("Deleted %d github repositories", howMany)

			// Log github delete event
			eventsService.LogEvent(&events.LogEventArgs{
				EventType:     events.RepositoryDisabled,
				ClaGroupModel: claGroupModel,
				UserID:        claUser.UserID,
				EventData: &events.GithubProjectDeletedEventData{
					DeletedCount: howMany,
				},
			})
		}

		// Invalidate project signatures
		log.WithFields(f).Debug("Invalidating signatures")
		howMany, err = signatureService.InvalidateProjectRecords(ctx, params.ProjectID, claGroupModel.ProjectName)
		if err != nil {
			return project.NewDeleteProjectByIDBadRequest().WithXRequestID(reqID).WithPayload(errorResponse(err))
		}
		if howMany > 0 {
			log.WithFields(f).Debugf("Invalidated %d signatures", howMany)
			// Log invalidate signatures
			eventsService.LogEvent(&events.LogEventArgs{
				EventType:     events.InvalidatedSignature,
				ClaGroupModel: claGroupModel,
				UserID:        claUser.UserID,
				EventData: &events.SignatureProjectInvalidatedEventData{
					InvalidatedCount: howMany,
				},
			})
		}

		err = service.DeleteCLAGroup(ctx, params.ProjectID)
		if err != nil {
			if err == ErrProjectDoesNotExist {
				return project.NewDeleteProjectByIDNotFound()
			}
			return project.NewDeleteProjectByIDBadRequest().WithXRequestID(reqID).WithPayload(errorResponse(err))
		}
		eventsService.LogEvent(&events.LogEventArgs{
			EventType:     events.CLAGroupDeleted,
			ClaGroupModel: claGroupModel,
			UserID:        claUser.UserID,
			EventData:     &events.CLAGroupDeletedEventData{},
		})

		return project.NewDeleteProjectByIDNoContent().WithXRequestID(reqID)
	})

	// Update Project By Name
	api.ProjectUpdateProjectHandler = project.UpdateProjectHandlerFunc(func(projectParams project.UpdateProjectParams, claUser *user.CLAUser) middleware.Responder {
		reqID := utils.GetRequestID(projectParams.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint

		exitingModel, getErr := service.GetCLAGroupByID(ctx, projectParams.Body.ProjectID)
		if getErr != nil {
			msg := fmt.Sprintf("Error querying the project by ID, error: %+v", getErr)
			log.Warnf("Update Project Failed - %s", msg)
			return project.NewCreateProjectBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Code:    "500",
				Message: msg,
			})
		}

		// If the project with the same name exists...
		if exitingModel == nil {
			msg := fmt.Sprintf("unable to locate project with ID: %s", projectParams.Body.ProjectID)
			log.Warn(msg)
			return project.NewUpdateProjectNotFound().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Code:    "404",
				Message: msg,
			})
		}

		claGroupModel, err := service.UpdateCLAGroup(ctx, &projectParams.Body)
		if err != nil {
			if err == ErrProjectDoesNotExist {
				return project.NewUpdateProjectNotFound()
			}
			return project.NewUpdateProjectBadRequest().WithXRequestID(reqID).WithPayload(errorResponse(err))
		}

		// Log an event
		eventsService.LogEvent(&events.LogEventArgs{
			EventType:     events.CLAGroupUpdated,
			ClaGroupModel: claGroupModel,
			UserID:        claUser.UserID,
			EventData: &events.CLAGroupUpdatedEventData{
				ClaGroupName:        projectParams.Body.ProjectName,
				ClaGroupDescription: projectParams.Body.ProjectDescription,
			},
		})

		return project.NewUpdateProjectOK().WithXRequestID(reqID).WithPayload(claGroupModel)
	})
}
