// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package cla_manager

import (
	"fmt"

	"github.com/go-openapi/strfmt"

	"github.com/sirupsen/logrus"

	"github.com/communitybridge/easycla/cla-backend-go/projects_cla_groups"

	"github.com/communitybridge/easycla/cla-backend-go/utils"

	"github.com/LF-Engineering/lfx-kit/auth"

	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/restapi/operations"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/restapi/operations/cla_manager"
	v1User "github.com/communitybridge/easycla/cla-backend-go/user"

	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/go-openapi/runtime/middleware"
)

const (
	//BadRequest error Response code
	BadRequest = "400"
	//Conflict error Response code
	Conflict = "409"
)

// Configure is the API handler routine for CLA Manager routes
func Configure(api *operations.EasyclaAPI, service Service, LfxPortalURL string, projectClaGroupRepo projects_cla_groups.Repository, easyCLAUserRepo v1User.RepositoryService) {
	api.ClaManagerCreateCLAManagerHandler = cla_manager.CreateCLAManagerHandlerFunc(func(params cla_manager.CreateCLAManagerParams, authUser *auth.User) middleware.Responder {
		utils.SetAuthUserProperties(authUser, params.XUSERNAME, params.XEMAIL)
		if !utils.IsUserAuthorizedForProjectOrganization(authUser, params.ProjectSFID, params.CompanySFID) {
			return cla_manager.NewCreateCLAManagerForbidden().WithPayload(&models.ErrorResponse{
				Code: "403",
				Message: fmt.Sprintf("EasyCLA - 403 Forbidden - user %s does not have access to CreateCLAManager with Project|Organization scope of %s | %s",
					authUser.UserName, params.ProjectSFID, params.CompanySFID),
			})
		}
		cginfo, err := projectClaGroupRepo.GetClaGroupIDForProject(params.ProjectSFID)
		if err != nil {
			if err == projects_cla_groups.ErrProjectNotAssociatedWithClaGroup {
				return cla_manager.NewCreateCLAManagerInternalServerError().WithPayload(&models.ErrorResponse{
					Code:    "400",
					Message: fmt.Sprintf("EasyCLA - Bad Request. error = %s", "No cla group is associated with this project"),
				})
			}
			return cla_manager.NewCreateCLAManagerInternalServerError().WithPayload(&models.ErrorResponse{
				Code:    "500",
				Message: fmt.Sprintf("EasyCLA - 500 Internal server error. error = %s", err.Error()),
			})
		}
		compCLAManager, errorResponse := service.CreateCLAManager(cginfo.ClaGroupID, params, authUser.UserName)
		if errorResponse != nil {
			if errorResponse.Code == BadRequest {
				return cla_manager.NewCreateCLAManagerBadRequest().WithPayload(errorResponse)
			} else if errorResponse.Code == Conflict {
				return cla_manager.NewCreateCLAManagerConflict().WithPayload(errorResponse)
			}
		}

		return cla_manager.NewCreateCLAManagerOK().WithPayload(compCLAManager)
	})

	api.ClaManagerDeleteCLAManagerHandler = cla_manager.DeleteCLAManagerHandlerFunc(func(params cla_manager.DeleteCLAManagerParams, authUser *auth.User) middleware.Responder {
		utils.SetAuthUserProperties(authUser, params.XUSERNAME, params.XEMAIL)
		if !utils.IsUserAuthorizedForProjectOrganization(authUser, params.ProjectSFID, params.CompanySFID) {
			return cla_manager.NewDeleteCLAManagerForbidden().WithPayload(&models.ErrorResponse{
				Code: "403",
				Message: fmt.Sprintf("EasyCLA - 403 Forbidden - user %s does not have access to DeleteCLAManager with Project|Organization scope of %s | %s",
					authUser.UserName, params.ProjectSFID, params.CompanySFID),
			})
		}
		cginfo, err := projectClaGroupRepo.GetClaGroupIDForProject(params.ProjectSFID)
		if err != nil {
			return cla_manager.NewDeleteCLAManagerBadRequest().WithPayload(&models.ErrorResponse{
				Code:    "400",
				Message: fmt.Sprintf("EasyCLA - Bad Request. No Cla Group associated with ProjectSFID: %s ", params.ProjectSFID),
			})
		}

		errResponse := service.DeleteCLAManager(cginfo.ClaGroupID, params)
		if errResponse != nil {
			return cla_manager.NewDeleteCLAManagerBadRequest().WithPayload(errResponse)
		}

		return cla_manager.NewDeleteCLAManagerNoContent()
	})

	api.ClaManagerCreateCLAManagerDesigneeHandler = cla_manager.CreateCLAManagerDesigneeHandlerFunc(func(params cla_manager.CreateCLAManagerDesigneeParams, authUser *auth.User) middleware.Responder {
		f := logrus.Fields{"functionName": "ClaManagerCreateCLAManagerDesigneeHandler", "CompanySFID": params.CompanySFID, "ProjectSFID": params.ProjectSFID, "authUser": *params.XUSERNAME}
		log.WithFields(f).Debugf("processing CLA Manager Desginee request")
		utils.SetAuthUserProperties(authUser, params.XUSERNAME, params.XEMAIL)
		if !utils.IsUserAuthorizedForOrganization(authUser, params.CompanySFID) {
			return cla_manager.NewCreateCLAManagerDesigneeForbidden().WithPayload(&models.ErrorResponse{
				Code: "403",
				Message: fmt.Sprintf("EasyCLA - 403 Forbidden - user %s does not have access to CreateCLAManagerDesignee with Organization scope of %s",
					authUser.UserName, params.CompanySFID),
			})
		}

		claManagerDesignee, err := service.CreateCLAManagerDesignee(params.CompanySFID, params.ProjectSFID, strfmt.Email(params.Body.UserEmail).String())
		if err != nil {
			msg := fmt.Sprintf("user :%s, error: %+v ", authUser.Email, err)
			return cla_manager.NewCreateCLAManagerBadRequest().WithPayload(
				&models.ErrorResponse{
					Message: msg,
					Code:    "400",
				})
		}

		log.Debugf("CLA Manager designee created : %+v", claManagerDesignee)
		return cla_manager.NewCreateCLAManagerDesigneeOK().WithPayload(claManagerDesignee)
	})

	api.ClaManagerCreateCLAManagerDesigneeByGroupHandler = cla_manager.CreateCLAManagerDesigneeByGroupHandlerFunc(
		func(params cla_manager.CreateCLAManagerDesigneeByGroupParams, authUser *auth.User) middleware.Responder {
			f := logrus.Fields{
				"functionName": "ClaManagerCreateCLAManagerDesigneeByGroupHandler",
				"CompanySFID":  params.CompanySFID,
				"ClaGroupID":   params.ClaGroupID,
				"Email":        strfmt.Email(params.Body.UserEmail).String(),
				"authUser":     *params.XUSERNAME,
			}
			log.WithFields(f).Debugf("processing CLA Manager Desginee by group request")
			if !utils.IsUserAuthorizedForOrganization(authUser, params.CompanySFID) {
				msg := fmt.Sprintf("EasyCLA - 403 Forbidden - user %s does not have access to CreateCLAManagerDesignee with Organization scope of %s",
					authUser.UserName, params.CompanySFID)
				log.WithFields(f).Warn(msg)
				return cla_manager.NewCreateCLAManagerDesigneeForbidden().WithPayload(&models.ErrorResponse{
					Code:    "403",
					Message: msg,
				})
			}

			log.WithFields(f).Debugf("getting project IDs for CLA group")
			projectCLAGroups, getErr := projectClaGroupRepo.GetProjectsIdsForClaGroup(params.ClaGroupID)
			if getErr != nil {
				msg := fmt.Sprintf("Error getting SF projects for claGroup: %s ", params.ClaGroupID)
				log.WithFields(f).Warn(msg)
				return cla_manager.NewCreateCLAManagerDesigneeByGroupBadRequest().WithPayload(
					&models.ErrorResponse{
						Message: msg,
						Code:    "400",
					})
			}
			log.WithFields(f).Debugf("found %d project IDs for CLA group", len(projectCLAGroups))
			if len(projectCLAGroups) == 0 {
				msg := fmt.Sprintf("no projects associated with CLA Group: %s", params.ClaGroupID)
				log.WithFields(f).Warn(msg)
				return cla_manager.NewCreateCLAManagerDesigneeByGroupNotFound().WithPayload(
					&models.ErrorResponse{
						Message: msg,
						Code:    "400",
					})

			}

			var designeeScopes []*models.ClaManagerDesignee
			for _, pcg := range projectCLAGroups {
				log.WithFields(f).Debugf("creating CLA Manager Designee for Project SFID: %s", pcg.ProjectSFID)
				claManagerDesignee, err := service.CreateCLAManagerDesignee(params.CompanySFID, pcg.ProjectSFID, strfmt.Email(params.Body.UserEmail).String())
				if err != nil {
					msg := fmt.Sprintf("Creating cla manager designee failed for Project SFID: %s ", pcg.ProjectSFID)
					log.WithFields(f).Warn(msg)
					return cla_manager.NewCreateCLAManagerDesigneeByGroupBadRequest().WithPayload(
						&models.ErrorResponse{
							Message: msg,
							Code:    "400",
						})
				}
				designeeScopes = append(designeeScopes, claManagerDesignee)
			}

			return cla_manager.NewCreateCLAManagerDesigneeByGroupOK().WithPayload(&models.ClaManagerDesignees{
				List: designeeScopes,
			})
		})

	api.ClaManagerInviteCompanyAdminHandler = cla_manager.InviteCompanyAdminHandlerFunc(func(params cla_manager.InviteCompanyAdminParams) middleware.Responder {

		// Get Contributor details
		user, userErr := easyCLAUserRepo.GetUser(params.UserID)

		if userErr != nil {
			msg := fmt.Sprintf("Problem getting user by ID : %s, error: %+v ", params.UserID, userErr)
			return cla_manager.NewInviteCompanyAdminBadRequest().WithPayload(
				&models.ErrorResponse{
					Code:    "400",
					Message: msg,
				})
		}

		claManagerDesignees, err := service.InviteCompanyAdmin(params.Body.ContactAdmin, params.Body.CompanyID, params.Body.ClaGroupID, strfmt.Email(params.Body.UserEmail).String(), params.Body.Name, &user, LfxPortalURL)

		if err != nil {
			return cla_manager.NewInviteCompanyAdminBadRequest().WithPayload(err)
		}
		// Check if admins successfully sent email
		if len(claManagerDesignees) == 0 {
			return cla_manager.NewInviteCompanyAdminNoContent()
		}

		// successfully created cla manager designee and sent invite
		return cla_manager.NewInviteCompanyAdminOK().WithPayload(&models.ClaManagerDesignees{
			List: claManagerDesignees,
		})
	})

	api.ClaManagerCreateCLAManagerRequestHandler = cla_manager.CreateCLAManagerRequestHandlerFunc(func(params cla_manager.CreateCLAManagerRequestParams, authUser *auth.User) middleware.Responder {
		f := logrus.Fields{"functionName": "ClaManagerCreateCLAManagerRequestHandler", "CompanySFID": params.CompanySFID, "ProjectSFID": params.ProjectSFID, "authUser": *params.XUSERNAME}
		log.WithFields(f).Debugf("processing CLA Manager request")
		utils.SetAuthUserProperties(authUser, params.XUSERNAME, params.XEMAIL)
		if !utils.IsUserAuthorizedForOrganization(authUser, params.CompanySFID) {
			return cla_manager.NewCreateCLAManagerRequestForbidden().WithPayload(&models.ErrorResponse{
				Code: "403",
				Message: fmt.Sprintf("EasyCLA - 403 Forbidden - user %s does not have access to CreateCLAManagerRequest with Project|Organization scope of %s | %s",
					authUser.UserName, params.ProjectSFID, params.CompanySFID),
			})
		}

		claManagerDesignee, err := service.CreateCLAManagerRequest(params.Body.ContactAdmin, params.CompanySFID, params.ProjectSFID, strfmt.Email(params.Body.UserEmail).String(),
			params.Body.FullName, authUser, *params.XEMAIL, LfxPortalURL)

		if err != nil {
			if err == ErrNoOrgAdmins || err == ErrNoLFID {
				return cla_manager.NewCreateCLAManagerRequestNotFound().WithPayload(
					&models.ErrorResponse{
						Message: err.Error(),
						Code:    "404",
					})
			}
			// Check if user is already assigned scope/role
			if err == ErrRoleScopeConflict {
				msg := fmt.Sprintf("User %s already has role scope assigned ", params.Body.FullName)
				return cla_manager.NewCreateCLAManagerRequestConflict().WithPayload(
					&models.ErrorResponse{
						Message: msg,
						Code:    "409",
					})
			}
			// Return Bad Request
			return cla_manager.NewCreateCLAManagerRequestBadRequest().WithPayload(
				&models.ErrorResponse{
					Message: err.Error(),
					Code:    "400",
				})
		}
		if params.Body.ContactAdmin {
			return cla_manager.NewCreateCLAManagerRequestNoContent()
		}
		return cla_manager.NewCreateCLAManagerRequestOK().WithPayload(claManagerDesignee)
	})

	api.ClaManagerNotifyCLAManagersHandler = cla_manager.NotifyCLAManagersHandlerFunc(
		func(params cla_manager.NotifyCLAManagersParams) middleware.Responder {
			err := service.NotifyCLAManagers(params.Body)
			if err != nil {
				if err == ErrCLAUserNotFound {
					return cla_manager.NewNotifyCLAManagersNotFound()
				}
				return cla_manager.NewNotifyCLAManagersBadRequest()
			}
			return cla_manager.NewNotifyCLAManagersNoContent()
		})

}

// buildErrorMessageCreate helper function to build an error message
func buildErrorMessageCreate(params cla_manager.CreateCLAManagerParams, err error) string {
	return fmt.Sprintf("problem creating new CLA Manager using company SFID: %s, project SFID: %s, firstName: %s, lastName: %s, user email: %s, error: %+v",
		params.CompanySFID, params.ProjectSFID, *params.Body.FirstName, *params.Body.LastName, *params.Body.UserEmail, err)
}

// buildErrorMessage helper function to build an error message
func buildErrorMessageDelete(params cla_manager.DeleteCLAManagerParams, err error) string {
	return fmt.Sprintf("problem deleting new CLA Manager Request using company SFID: %s, project SFID: %s, user ID: %s, error: %+v",
		params.CompanySFID, params.ProjectSFID, params.UserLFID, err)
}
