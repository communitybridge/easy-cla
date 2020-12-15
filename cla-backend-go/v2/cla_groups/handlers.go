// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package cla_groups

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/communitybridge/easycla/cla-backend-go/projects_cla_groups"

	"github.com/sirupsen/logrus"

	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/restapi/operations/foundation"

	"github.com/communitybridge/easycla/cla-backend-go/events"

	"github.com/LF-Engineering/lfx-kit/auth"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/restapi/operations"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/restapi/operations/cla_group"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	v1Project "github.com/communitybridge/easycla/cla-backend-go/project"
	"github.com/communitybridge/easycla/cla-backend-go/utils"
	v2ProjectService "github.com/communitybridge/easycla/cla-backend-go/v2/project-service"
	v2ProjectServiceClient "github.com/communitybridge/easycla/cla-backend-go/v2/project-service/client/project"
	"github.com/go-openapi/runtime/middleware"
)

// Configure configures the cla group api
func Configure(api *operations.EasyclaAPI, service Service, v1ProjectService v1Project.Service, projectClaGroupsRepo projects_cla_groups.Repository, eventsService events.Service) { //nolint

	api.ClaGroupCreateClaGroupHandler = cla_group.CreateClaGroupHandlerFunc(func(params cla_group.CreateClaGroupParams, authUser *auth.User) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		utils.SetAuthUserProperties(authUser, params.XUSERNAME, params.XEMAIL)
		f := logrus.Fields{
			"functionName":        "ClaGroupCreateClaGroupHandler",
			utils.XREQUESTID:      ctx.Value(utils.XREQUESTID),
			"claGroupName":        aws.StringValue(params.ClaGroupInput.ClaGroupName),
			"claGroupDescription": params.ClaGroupInput.ClaGroupDescription,
			"projectSFIDList":     strings.Join(params.ClaGroupInput.ProjectSfidList, ","),
			"authUsername":        params.XUSERNAME,
			"authEmail":           params.XEMAIL,
		}

		// Check permissions
		if !isUserHaveAccessToCLAProject(ctx, authUser, aws.StringValue(params.ClaGroupInput.FoundationSfid), projectClaGroupsRepo) {
			msg := fmt.Sprintf("user %s does not have access to create a CLA Group with project scope of: %s", authUser.UserName, aws.StringValue(params.ClaGroupInput.FoundationSfid))
			log.WithFields(f).Warn(msg)
			return cla_group.NewCreateClaGroupForbidden().WithXRequestID(reqID).WithPayload(utils.ErrorResponseForbidden(reqID, msg))
		}

		claGroup, err := service.CreateCLAGroup(ctx, params.ClaGroupInput, utils.StringValue(params.XUSERNAME))
		if err != nil {
			log.WithFields(f).WithError(err).Warn("unable to create the CLA Group")
			return cla_group.NewCreateClaGroupBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Code:       "400",
				Message:    fmt.Sprintf("EasyCLA - 400 Bad Request - %s", err.Error()),
				XRequestID: reqID,
			})
		}

		// Log the event
		eventsService.LogEvent(&events.LogEventArgs{
			EventType:  events.CLAGroupCreated,
			ProjectID:  claGroup.ClaGroupID,
			LfUsername: authUser.UserName,
			EventData:  &events.CLAGroupCreatedEventData{},
		})

		return cla_group.NewCreateClaGroupOK().WithXRequestID(reqID).WithPayload(claGroup)
	})

	api.ClaGroupUpdateClaGroupHandler = cla_group.UpdateClaGroupHandlerFunc(func(params cla_group.UpdateClaGroupParams, authUser *auth.User) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		utils.SetAuthUserProperties(authUser, params.XUSERNAME, params.XEMAIL)
		f := logrus.Fields{
			"functionName":   "ClaGroupUpdateClaGroupHandler",
			utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
			"claGroupID":     params.ClaGroupID,
			"authUsername":   params.XUSERNAME,
			"authEmail":      params.XEMAIL,
		}

		// Make sure we have some parameters to process...
		if params.Body == nil || (params.Body.ClaGroupName == "" && params.Body.ClaGroupDescription == "") {
			log.WithFields(f).Warn("missing CLA Group update parameters - body missing required values")
			return cla_group.NewUpdateClaGroupBadRequest().WithXRequestID(reqID).WithPayload(
				utils.ErrorResponseBadRequest(reqID, "missing update parameters - body missing required values"))
		}

		// Attempt to load the existing CLA Group model
		log.WithFields(f).Debug("Loading existing CLA Group by ID...")
		claGroupModel, err := v1ProjectService.GetCLAGroupByID(ctx, params.ClaGroupID)
		if err != nil {
			log.WithFields(f).WithError(err).Warn("problem loading CLA group by ID")

			var e *utils.CLAGroupNotFound
			//if err, ok := err.(*utils.CLAGroupNotFound); ok {
			if errors.As(err, &e) {
				log.WithFields(f).WithError(err).Warn("problem loading CLA group by ID - cla group not found")
				return cla_group.NewUpdateClaGroupNotFound().WithXRequestID(reqID).WithPayload(
					utils.ErrorResponseNotFoundWithError(reqID, "CLA Group not found", err))
			}
			if errors.Is(err, v1Project.ErrProjectDoesNotExist) {
				return cla_group.NewUpdateClaGroupNotFound().WithXRequestID(reqID).WithPayload(
					utils.ErrorResponseNotFoundWithError(reqID, "CLA Group not found", err))
			}
			return cla_group.NewUpdateClaGroupBadRequest().WithXRequestID(reqID).WithPayload(
				utils.ErrorResponseBadRequestWithError(reqID, fmt.Sprintf("unable to lookup CLA Group by ID: %s", params.ClaGroupID), err))
		}

		// Check permissions
		if !isUserHaveAccessToCLAProject(ctx, authUser, claGroupModel.FoundationSFID, projectClaGroupsRepo) {
			msg := fmt.Sprintf("user %s does not have access to update an existing CLA Group with project scope of: %s", authUser.UserName, claGroupModel.FoundationSFID)
			log.WithFields(f).Warn(msg)
			return cla_group.NewUpdateClaGroupForbidden().WithXRequestID(reqID).WithPayload(utils.ErrorResponseForbidden(reqID, msg))
		}

		// Only update if either the CLA Group Name or Description is changed - if both are the same, abort.
		if claGroupModel.ProjectName == params.Body.ClaGroupName && claGroupModel.ProjectDescription == params.Body.ClaGroupDescription {
			log.WithFields(f).Warn("unable to update the CLA Group Name or Description - provided values are the same as the existing record")
			return cla_group.NewUpdateClaGroupBadRequest().WithXRequestID(reqID).WithPayload(
				utils.ErrorResponseBadRequest(reqID, fmt.Sprintf("unable to update the CLA Group Name or Description - values are the same for CLA Group ID: %s", params.ClaGroupID)))
		}

		claGroup, err := service.UpdateCLAGroup(ctx, claGroupModel, params.Body, utils.StringValue(params.XUSERNAME))
		if err != nil {
			log.WithFields(f).WithError(err).Warn("unable to update the CLA Group Name and/or Description - update failed")
			return cla_group.NewUpdateClaGroupBadRequest().WithXRequestID(reqID).WithPayload(
				utils.ErrorResponseBadRequestWithError(reqID, fmt.Sprintf("unable to update CLA Group by ID: %s", params.ClaGroupID), err))
		}

		// Log the event
		eventsService.LogEvent(&events.LogEventArgs{
			EventType:  events.CLAGroupUpdated,
			ProjectID:  claGroup.ClaGroupID,
			LfUsername: authUser.UserName,
			EventData: &events.CLAGroupUpdatedEventData{
				ClaGroupName:        params.Body.ClaGroupName,
				ClaGroupDescription: params.Body.ClaGroupDescription,
			},
		})

		return cla_group.NewUpdateClaGroupOK().WithXRequestID(reqID).WithPayload(claGroup)
	})

	api.ClaGroupDeleteClaGroupHandler = cla_group.DeleteClaGroupHandlerFunc(func(params cla_group.DeleteClaGroupParams, authUser *auth.User) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		utils.SetAuthUserProperties(authUser, params.XUSERNAME, params.XEMAIL)
		f := logrus.Fields{
			"functionName":   "ClaGroupDeleteClaGroupHandler",
			utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
			"claGroupID":     params.ClaGroupID,
			"authUsername":   params.XUSERNAME,
			"authEmail":      params.XEMAIL,
		}

		claGroupModel, err := v1ProjectService.GetCLAGroupByID(ctx, params.ClaGroupID)
		if err != nil {
			log.WithFields(f).Warn(err)
			if err, ok := err.(*utils.CLAGroupNotFound); ok {
				return cla_group.NewDeleteClaGroupNotFound().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
					Code:       "404",
					Message:    fmt.Sprintf("EasyCLA - 404 Not Found - %s", err.Error()),
					XRequestID: reqID,
				})
			}
			if err == v1Project.ErrProjectDoesNotExist {
				return cla_group.NewDeleteClaGroupNotFound().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
					Code: "404",
					Message: fmt.Sprintf("EasyCLA - 404 Not Found - cla_group %s not found",
						params.ClaGroupID),
					XRequestID: reqID,
				})
			}
			return cla_group.NewDeleteClaGroupInternalServerError().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Code: "500",
				Message: fmt.Sprintf("EasyCLA - 500 Internal server error - unable to lookup CLA Group by ID: %s, error: %+v",
					params.ClaGroupID, err),
				XRequestID: reqID,
			})
		}

		// Check permissions
		if !isUserHaveAccessToCLAProject(ctx, authUser, claGroupModel.FoundationSFID, projectClaGroupsRepo) {
			msg := fmt.Sprintf("user %s does not have access to delete the CLA Group with project scope of: %s", authUser.UserName, claGroupModel.FoundationSFID)
			log.WithFields(f).Warn(msg)
			return cla_group.NewDeleteClaGroupForbidden().WithXRequestID(reqID).WithPayload(utils.ErrorResponseForbidden(reqID, msg))
		}

		err = service.DeleteCLAGroup(ctx, claGroupModel, authUser)
		if err != nil {
			log.WithFields(f).Warn(err)
			return cla_group.NewDeleteClaGroupInternalServerError().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Code: "500",
				Message: fmt.Sprintf("EasyCLA - 500 Internal server error - error deleting CLA Group %s, error: %+v",
					params.ClaGroupID, err),
				XRequestID: reqID,
			})
		}

		eventsService.LogEvent(&events.LogEventArgs{
			EventType:     events.CLAGroupDeleted,
			ClaGroupModel: claGroupModel,
			LfUsername:    authUser.UserName,
			EventData:     &events.CLAGroupDeletedEventData{},
		})

		return cla_group.NewDeleteClaGroupNoContent().WithXRequestID(reqID)
	})

	api.ClaGroupEnrollProjectsHandler = cla_group.EnrollProjectsHandlerFunc(func(params cla_group.EnrollProjectsParams, authUser *auth.User) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		utils.SetAuthUserProperties(authUser, params.XUSERNAME, params.XEMAIL)
		f := logrus.Fields{
			"functionName":    "ClaGroupEnrollProjectsHandler",
			utils.XREQUESTID:  ctx.Value(utils.XREQUESTID),
			"ClaGroupID":      params.ClaGroupID,
			"authUsername":    params.XUSERNAME,
			"authEmail":       params.XEMAIL,
			"projectSFIDList": strings.Join(params.ProjectSFIDList, ","),
		}

		cg, err := v1ProjectService.GetCLAGroupByID(ctx, params.ClaGroupID)
		if err != nil {
			if err, ok := err.(*utils.CLAGroupNotFound); ok {
				return cla_group.NewEnrollProjectsNotFound().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
					Code:       "404",
					Message:    fmt.Sprintf("EasyCLA - 404 Not Found - %s", err.Error()),
					XRequestID: reqID,
				})
			}
			if err == v1Project.ErrProjectDoesNotExist {
				return cla_group.NewEnrollProjectsNotFound().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
					Code: "404",
					Message: fmt.Sprintf("EasyCLA - 404 Not Found - cla_group %s not found",
						params.ClaGroupID),
					XRequestID: reqID,
				})
			}
			return cla_group.NewEnrollProjectsInternalServerError().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Code:       "400",
				Message:    fmt.Sprintf("EasyCLA - 500 Internal server error - error = %s", err.Error()),
				XRequestID: reqID,
			})
		}

		// Check permissions
		if !isUserHaveAccessToCLAProject(ctx, authUser, cg.FoundationSFID, projectClaGroupsRepo) {
			msg := fmt.Sprintf("user %s does not have access to enroll projects with project scope of: %s", authUser.UserName, cg.FoundationSFID)
			log.WithFields(f).Warn(msg)
			return cla_group.NewEnrollProjectsForbidden().WithXRequestID(reqID).WithPayload(utils.ErrorResponseForbidden(reqID, msg))
		}

		if !cg.FoundationLevelCLA {
			log.WithFields(f).Debug("locating project by sfid...")
			psc := v2ProjectService.GetClient()
			for _, projectSFID := range params.ProjectSFIDList {
				project, projectErr := psc.GetProject(projectSFID)
				if projectErr != nil || project == nil {
					msg := fmt.Sprintf("Failed to get salesforce project: %s", projectSFID)
					log.WithFields(f).Warn(msg)
					if _, ok := projectErr.(*v2ProjectServiceClient.GetProjectNotFound); ok {
						return cla_group.NewEnrollProjectsNotFound().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
							Code:    "404",
							Message: fmt.Sprintf("project not found with given ID. [%s]", projectSFID),
						})
					}
					return cla_group.NewEnrollProjectsBadRequest().WithXRequestID(reqID).WithPayload(utils.ErrorResponseNotFoundWithError(reqID, msg, projectErr))
				}
				if project.ProjectType == utils.ProjectTypeProjectGroup {
					msg := fmt.Sprintf("Unable to enroll salesforce foundation project: %s in project level cla-group.", projectSFID)
					return cla_group.NewEnrollProjectsBadRequest().WithXRequestID(reqID).WithPayload(utils.ErrorResponseBadRequest(reqID, msg))
				}
			}

		}

		err = service.EnrollProjectsInClaGroup(ctx, params.ClaGroupID, cg.FoundationSFID, params.ProjectSFIDList)
		if err != nil {
			if strings.Contains(err.Error(), "bad request") {
				return cla_group.NewEnrollProjectsBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
					Code:       "400",
					Message:    fmt.Sprintf("EasyCLA - 400 Bad Request - %s", err.Error()),
					XRequestID: reqID,
				})
			}
			return cla_group.NewEnrollProjectsInternalServerError().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Code:       "500",
				Message:    fmt.Sprintf("EasyCLA - 500 Internal server error - error = %s", err.Error()),
				XRequestID: reqID,
			})
		}

		eventsService.LogEvent(&events.LogEventArgs{
			EventType:     events.CLAGroupUpdated,
			ClaGroupModel: cg,
			LfUsername:    authUser.UserName,
			EventData: &events.CLAGroupUpdatedEventData{
				ClaGroupName:        cg.ProjectName,
				ClaGroupDescription: cg.ProjectDescription,
			},
		})

		return cla_group.NewEnrollProjectsOK().WithXRequestID(reqID)
	})

	api.ClaGroupUnenrollProjectsHandler = cla_group.UnenrollProjectsHandlerFunc(func(params cla_group.UnenrollProjectsParams, authUser *auth.User) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		utils.SetAuthUserProperties(authUser, params.XUSERNAME, params.XEMAIL)
		f := logrus.Fields{
			"functionName":    "ClaGroupUnenrollProjectsHandler",
			utils.XREQUESTID:  ctx.Value(utils.XREQUESTID),
			"ClaGroupID":      params.ClaGroupID,
			"authUsername":    params.XUSERNAME,
			"authEmail":       params.XEMAIL,
			"projectSFIDList": strings.Join(params.ProjectSFIDList, ","),
		}

		cg, err := v1ProjectService.GetCLAGroupByID(ctx, params.ClaGroupID)
		if err != nil {
			if err, ok := err.(*utils.CLAGroupNotFound); ok {
				return cla_group.NewUnenrollProjectsNotFound().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
					Code:       "404",
					Message:    fmt.Sprintf("EasyCLA - 404 Not Found - %s", err.Error()),
					XRequestID: reqID,
				})
			}
			if err == v1Project.ErrProjectDoesNotExist {
				return cla_group.NewUnenrollProjectsNotFound().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
					Code:       "404",
					Message:    fmt.Sprintf("EasyCLA - 404 Not Found - cla_group %s not found", params.ClaGroupID),
					XRequestID: reqID,
				})
			}
			return cla_group.NewUnenrollProjectsInternalServerError().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Code:       "400",
				Message:    fmt.Sprintf("EasyCLA - 500 Internal server error - error = %s", err.Error()),
				XRequestID: reqID,
			})
		}

		// Check permissions
		if !isUserHaveAccessToCLAProject(ctx, authUser, cg.FoundationSFID, projectClaGroupsRepo) {
			msg := fmt.Sprintf("user %s does not have access to unenroll projects with project scope of: %s", authUser.UserName, cg.FoundationSFID)
			log.WithFields(f).Warn(msg)
			return cla_group.NewUnenrollProjectsForbidden().WithXRequestID(reqID).WithPayload(utils.ErrorResponseForbidden(reqID, msg))
		}

		err = service.UnenrollProjectsInClaGroup(ctx, params.ClaGroupID, cg.FoundationSFID, params.ProjectSFIDList)
		if err != nil {
			if strings.Contains(err.Error(), "bad request") {
				return cla_group.NewUnenrollProjectsBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
					Code:       "400",
					Message:    fmt.Sprintf("EasyCLA - 400 Bad Request - %s", err.Error()),
					XRequestID: reqID,
				})
			}
			return cla_group.NewUnenrollProjectsInternalServerError().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Code:       "500",
				Message:    fmt.Sprintf("EasyCLA - 500 Internal server error - error = %s", err.Error()),
				XRequestID: reqID,
			})
		}

		eventsService.LogEvent(&events.LogEventArgs{
			EventType:     events.CLAGroupUpdated,
			ClaGroupModel: cg,
			LfUsername:    authUser.UserName,
			EventData: &events.CLAGroupUpdatedEventData{
				ClaGroupName:        cg.ProjectName,
				ClaGroupDescription: cg.ProjectDescription,
			},
		})

		return cla_group.NewUnenrollProjectsOK().WithXRequestID(reqID)
	})

	api.ClaGroupListClaGroupsUnderFoundationHandler = cla_group.ListClaGroupsUnderFoundationHandlerFunc(func(params cla_group.ListClaGroupsUnderFoundationParams, authUser *auth.User) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		utils.SetAuthUserProperties(authUser, params.XUSERNAME, params.XEMAIL)
		f := logrus.Fields{
			"functionName":   "ClaGroupListClaGroupsUnderFoundationHandler",
			utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
			"projectSFID":    params.ProjectSFID,
			"authUsername":   params.XUSERNAME,
			"authEmail":      params.XEMAIL,
		}

		log.WithFields(f).Debug("locating project by sfid...")
		psc := v2ProjectService.GetClient()
		project, projectErr := psc.GetProject(params.ProjectSFID)
		if projectErr != nil || project == nil {
			msg := fmt.Sprintf("Failed to get salesforce project: %s", params.ProjectSFID)
			log.WithFields(f).Warn(msg)
			if _, ok := projectErr.(*v2ProjectServiceClient.GetProjectNotFound); ok {
				return cla_group.NewListClaGroupsUnderFoundationNotFound().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
					Code:    "404",
					Message: fmt.Sprintf("project not found with given ID. [%s]", params.ProjectSFID),
				})
			}
			return cla_group.NewListClaGroupsUnderFoundationBadRequest().WithXRequestID(reqID).WithPayload(utils.ErrorResponseNotFoundWithError(reqID, msg, projectErr))
		}

		log.WithFields(f).Debug("found project - evaluating parent...")
		var projectSFIDs []string
		// Add the foundation ID, if available
		if project.Foundation != nil && project.Foundation.ID != "" {
			log.WithFields(f).Debugf("parent project - found %s - adding to list of project IDs...", project.Foundation.ID)
			projectSFIDs = append(projectSFIDs, project.Foundation.ID)
		}
		log.WithFields(f).Debug("project - adding to list of project IDs...")
		projectSFIDs = append(projectSFIDs, project.ID)

		// Check permissions
		log.WithFields(f).Debugf("checking permissions for %s", strings.Join(projectSFIDs, ","))
		if !utils.IsUserAuthorizedForAnyProjects(authUser, projectSFIDs) {
			msg := fmt.Sprintf("user %s does not have access to list projects with project scope of: %s", authUser.UserName, params.ProjectSFID)
			log.WithFields(f).Warn(msg)
			return cla_group.NewListClaGroupsUnderFoundationForbidden().WithXRequestID(reqID).WithPayload(utils.ErrorResponseForbidden(reqID, msg))
		}

		log.WithFields(f).Debug("locating CLA groups for foundation or project...")
		result, err := service.ListClaGroupsForFoundationOrProject(ctx, params.ProjectSFID)
		if err != nil {
			msg := fmt.Sprintf("problem loading CLA Group for foundation or project: %s", params.ProjectSFID)
			log.WithFields(f).WithError(err).Warn(msg)

			if err, ok := err.(*utils.SFProjectNotFound); ok {
				msg := fmt.Sprintf("salesforce project not found: %s", params.ProjectSFID)
				log.WithFields(f).WithError(err).Warn(msg)
			}
			if _, ok := err.(*utils.ProjectCLAGroupMappingNotFound); ok {
				msg := fmt.Sprintf("project cla grouping not found for project: %s", params.ProjectSFID)
				log.WithFields(f).WithError(err).Warn(msg)
			}
			if err, ok := err.(*utils.CLAGroupNotFound); ok {
				msg := fmt.Sprintf("project cla group not found for project: %s", params.ProjectSFID)
				log.WithFields(f).WithError(err).Warn(msg)
			}

			// UI wants 200 empty list response
			return cla_group.NewListClaGroupsUnderFoundationOK().WithXRequestID(reqID).WithPayload(&models.ClaGroupListSummary{
				List: []*models.ClaGroupSummary{},
			})
		}

		// No results - empty OK response
		if result == nil {
			log.WithFields(f).Debug("no results found - returning empty list")
			return cla_group.NewListClaGroupsUnderFoundationOK().WithXRequestID(reqID).WithPayload(&models.ClaGroupListSummary{
				List: []*models.ClaGroupSummary{},
			})
		}

		return cla_group.NewListClaGroupsUnderFoundationOK().WithXRequestID(reqID).WithPayload(result)
	})

	api.ClaGroupValidateClaGroupHandler = cla_group.ValidateClaGroupHandlerFunc(func(params cla_group.ValidateClaGroupParams, authUser *auth.User) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		utils.SetAuthUserProperties(authUser, params.XUSERNAME, params.XEMAIL)

		// No API user validation - anyone can confirm or use the validate API endpoint

		valid, validationErrors := service.ValidateCLAGroup(ctx, params.ValidationInputRequest)
		return cla_group.NewValidateClaGroupOK().WithXRequestID(reqID).WithPayload(&models.ClaGroupValidationResponse{
			Valid:            valid,
			ValidationErrors: validationErrors,
		})
	})

	api.FoundationListFoundationClaGroupsHandler = foundation.ListFoundationClaGroupsHandlerFunc(func(params foundation.ListFoundationClaGroupsParams, authUser *auth.User) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		utils.SetAuthUserProperties(authUser, params.XUSERNAME, params.XEMAIL)
		result, err := service.ListAllFoundationClaGroups(ctx, params.FoundationSFID)
		if err != nil {
			return foundation.NewListFoundationClaGroupsInternalServerError().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Code:    "500",
				Message: err.Error(),
			})
		}

		return foundation.NewListFoundationClaGroupsOK().WithXRequestID(reqID).WithPayload(result)
	})
}

// isUserHaveAccessToCLAProject is a helper function to determine if the user has access to the specified project
func isUserHaveAccessToCLAProject(ctx context.Context, authUser *auth.User, projectSFID string, projectClaGroupsRepo projects_cla_groups.Repository) bool { // nolint
	f := logrus.Fields{
		"functionName":   "isUserHaveAccessToCLAProject",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"projectSFID":    projectSFID,
		"userName":       authUser.UserName,
		"userEmail":      authUser.Email,
	}

	log.WithFields(f).Debug("testing if user has access to project SFID")
	if utils.IsUserAuthorizedForProject(authUser, projectSFID) {
		return true
	}

	log.WithFields(f).Debug("user doesn't have direct access to the projectSFID - loading CLA Group from project id...")
	projectCLAGroupModel, err := projectClaGroupsRepo.GetClaGroupIDForProject(projectSFID)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("problem loading project -> cla group mapping - returning false")
		return false
	}
	if projectCLAGroupModel == nil {
		log.WithFields(f).WithError(err).Warnf("problem loading project -> cla group mapping - no mapping found - returning false")
		return false
	}

	f["foundationSFID"] = projectCLAGroupModel.FoundationSFID
	log.WithFields(f).Debug("testing if user has access to parent foundation...")
	if utils.IsUserAuthorizedForProjectTree(authUser, projectCLAGroupModel.FoundationSFID) {
		log.WithFields(f).Debug("user has access to parent foundation tree...")
		return true
	}
	if utils.IsUserAuthorizedForProject(authUser, projectCLAGroupModel.FoundationSFID) {
		log.WithFields(f).Debug("user has access to parent foundation...")
		return true
	}
	log.WithFields(f).Debug("user does not have access to parent foundation...")

	// Lookup the other project IDs for the CLA Group
	log.WithFields(f).Debug("looking up other projects associated with the CLA Group...")
	projectCLAGroupModels, err := projectClaGroupsRepo.GetProjectsIdsForClaGroup(projectCLAGroupModel.ClaGroupID)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("problem loading project cla group mappings by CLA Group ID - returning false")
		return false
	}

	projectSFIDs := getProjectIDsFromModels(f, projectCLAGroupModel.FoundationSFID, projectCLAGroupModels)
	f["projectIDs"] = strings.Join(projectSFIDs, ",")
	log.WithFields(f).Debug("testing if user has access to any projects")
	if utils.IsUserAuthorizedForAnyProjects(authUser, projectSFIDs) {
		log.WithFields(f).Debug("user has access to at least of of the projects...")
		return true
	}

	log.WithFields(f).Debug("exhausted project checks - user does not have access to project")
	return false
}

// getProjectIDsFromModels is a helper function to extract the project SFIDs from the project CLA Group models
func getProjectIDsFromModels(f logrus.Fields, foundationSFID string, projectCLAGroupModels []*projects_cla_groups.ProjectClaGroup) []string {
	// Build a list of projects associated with this CLA Group
	log.WithFields(f).Debug("building list of project IDs associated with the CLA Group...")
	var projectSFIDs []string
	projectSFIDs = append(projectSFIDs, foundationSFID)
	for _, projectCLAGroupModel := range projectCLAGroupModels {
		projectSFIDs = append(projectSFIDs, projectCLAGroupModel.ProjectSFID)
	}
	log.WithFields(f).Debugf("%d projects associated with the CLA Group...", len(projectSFIDs))
	return projectSFIDs
}
