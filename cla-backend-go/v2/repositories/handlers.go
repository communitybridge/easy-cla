// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	log "github.com/communitybridge/easycla/cla-backend-go/logging"

	"github.com/communitybridge/easycla/cla-backend-go/github"

	"github.com/LF-Engineering/lfx-kit/auth"
	"github.com/communitybridge/easycla/cla-backend-go/events"
	v1Models "github.com/communitybridge/easycla/cla-backend-go/gen/models"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/restapi/operations"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/restapi/operations/github_repositories"
	"github.com/communitybridge/easycla/cla-backend-go/repositories"
	"github.com/communitybridge/easycla/cla-backend-go/utils"
	"github.com/go-openapi/runtime/middleware"
	"github.com/jinzhu/copier"
)

// Configure establishes the middleware handlers for the repository service
func Configure(api *operations.EasyclaAPI, service Service, eventService events.Service) {
	api.GithubRepositoriesGetProjectGithubRepositoriesHandler = github_repositories.GetProjectGithubRepositoriesHandlerFunc(
		func(params github_repositories.GetProjectGithubRepositoriesParams, authUser *auth.User) middleware.Responder {
			reqID := utils.GetRequestID(params.XREQUESTID)
			utils.SetAuthUserProperties(authUser, params.XUSERNAME, params.XEMAIL)
			ctx := context.WithValue(params.HTTPRequest.Context(), utils.XREQUESTID, reqID) // nolint
			f := logrus.Fields{
				"functionName":   "GithubRepositoriesGetProjectGithubRepositoriesHandler",
				utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
				"authUser":       authUser.UserName,
				"authEmail":      authUser.Email,
				"projectSFID":    params.ProjectSFID,
			}

			if !utils.IsUserAuthorizedForProjectTree(authUser, params.ProjectSFID, utils.ALLOW_ADMIN_SCOPE) {
				msg := fmt.Sprintf("user %s does not have access to Get GitHub Repositories with Project scope of %s",
					authUser.UserName, params.ProjectSFID)
				log.WithFields(f).Debug(msg)
				return github_repositories.NewGetProjectGithubRepositoriesForbidden().WithPayload(
					utils.ErrorResponseForbidden(reqID, msg))
			}

			result, err := service.ListProjectRepositories(ctx, params.ProjectSFID)
			if err != nil {
				if strings.ContainsAny(err.Error(), "getProjectNotFound") {
					msg := fmt.Sprintf("repository not found for projectSFID: %s", params.ProjectSFID)
					log.WithFields(f).WithError(err).Warn(msg)
					return github_repositories.NewGetProjectGithubRepositoriesNotFound().WithPayload(
						utils.ErrorResponseNotFound(reqID, msg))
				}

				msg := fmt.Sprintf("problem looking up repositories for projectSFID: %s", params.ProjectSFID)
				log.WithFields(f).WithError(err).Warn(msg)
				return github_repositories.NewGetProjectGithubRepositoriesBadRequest().WithPayload(
					utils.ErrorResponseBadRequestWithError(reqID, msg, err))
			}

			response := &models.ListGithubRepositories{}
			err = copier.Copy(response, result)
			if err != nil {
				msg := fmt.Sprintf("problem converting response for projectSFID: %s", params.ProjectSFID)
				log.WithFields(f).WithError(err).Warn(msg)
				return github_repositories.NewGetProjectGithubRepositoriesInternalServerError().WithPayload(
					utils.ErrorResponseInternalServerErrorWithError(reqID, msg, err))
			}

			return github_repositories.NewGetProjectGithubRepositoriesOK().WithPayload(response)
		})

	api.GithubRepositoriesAddProjectGithubRepositoryHandler = github_repositories.AddProjectGithubRepositoryHandlerFunc(
		func(params github_repositories.AddProjectGithubRepositoryParams, authUser *auth.User) middleware.Responder {
			reqID := utils.GetRequestID(params.XREQUESTID)
			utils.SetAuthUserProperties(authUser, params.XUSERNAME, params.XEMAIL)
			ctx := context.WithValue(params.HTTPRequest.Context(), utils.XREQUESTID, reqID) // nolint
			f := logrus.Fields{
				"functionName":   "GithubRepositoriesAddProjectGithubRepositoryHandler",
				utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
				"authUser":       authUser.UserName,
				"authEmail":      authUser.Email,
				"projectSFID":    params.ProjectSFID,
			}

			if !utils.IsUserAuthorizedForProjectTree(authUser, params.ProjectSFID, utils.ALLOW_ADMIN_SCOPE) {
				msg := fmt.Sprintf("user %s does not have access to Add GitHub Repositories with Project scope of %s",
					authUser.UserName, params.ProjectSFID)
				log.WithFields(f).Debug(msg)
				return github_repositories.NewAddProjectGithubRepositoryForbidden().WithPayload(
					utils.ErrorResponseForbidden(reqID, msg))
			}

			result, err := service.AddGithubRepository(ctx, params.ProjectSFID, params.GithubRepositoryInput)
			if err != nil {
				msg := fmt.Sprintf("problem adding github repositories for projectSFID: %s", params.ProjectSFID)
				log.WithFields(f).WithError(err).Warn(msg)
				return github_repositories.NewAddProjectGithubRepositoryBadRequest().WithPayload(
					utils.ErrorResponseBadRequestWithError(reqID, msg, err))
			}

			// Log the event
			eventService.LogEvent(&events.LogEventArgs{
				EventType:         events.RepositoryAdded,
				ProjectID:         utils.StringValue(params.GithubRepositoryInput.ClaGroupID),
				ExternalProjectID: params.ProjectSFID,
				LfUsername:        authUser.UserName,
				ClaGroupModel: &v1Models.ClaGroup{
					ProjectExternalID: params.ProjectSFID,
					ProjectID:         utils.StringValue(params.GithubRepositoryInput.ClaGroupID),
				},
				EventData: &events.RepositoryAddedEventData{
					RepositoryName: result.RepositoryName,
				},
			})

			response := &models.GithubRepository{}
			err = copier.Copy(response, result)
			if err != nil {
				msg := fmt.Sprintf("problem converting response for projectSFID: %s", params.ProjectSFID)
				log.WithFields(f).WithError(err).Warn(msg)
				return github_repositories.NewAddProjectGithubRepositoryInternalServerError().WithPayload(
					utils.ErrorResponseInternalServerErrorWithError(reqID, msg, err))
			}

			return github_repositories.NewAddProjectGithubRepositoryOK().WithPayload(response)
		})

	api.GithubRepositoriesDeleteProjectGithubRepositoryHandler = github_repositories.DeleteProjectGithubRepositoryHandlerFunc(
		func(params github_repositories.DeleteProjectGithubRepositoryParams, authUser *auth.User) middleware.Responder {
			reqID := utils.GetRequestID(params.XREQUESTID)
			utils.SetAuthUserProperties(authUser, params.XUSERNAME, params.XEMAIL)
			ctx := context.WithValue(params.HTTPRequest.Context(), utils.XREQUESTID, reqID) // nolint
			f := logrus.Fields{
				"functionName":   "GithubRepositoriesDeleteProjectGithubRepositoryHandler",
				utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
				"authUser":       authUser.UserName,
				"authEmail":      authUser.Email,
				"projectSFID":    params.ProjectSFID,
				"repositoryID":   params.RepositoryID,
			}

			if !utils.IsUserAuthorizedForProjectTree(authUser, params.ProjectSFID, utils.ALLOW_ADMIN_SCOPE) {
				msg := fmt.Sprintf("user %s does not have access to Delete GitHub Repositories with Project scope of %s",
					authUser.UserName, params.ProjectSFID)
				log.WithFields(f).Debug(msg)
				return github_repositories.NewDeleteProjectGithubRepositoryForbidden().WithPayload(
					utils.ErrorResponseForbidden(reqID, msg))
			}

			ghRepo, err := service.GetRepository(ctx, params.RepositoryID)
			if err != nil {
				if err == repositories.ErrGithubRepositoryNotFound {
					msg := fmt.Sprintf("repository not found for projectSFID: %s", params.ProjectSFID)
					log.WithFields(f).WithError(err).Warn(msg)
					return github_repositories.NewDeleteProjectGithubRepositoryNotFound().WithPayload(
						utils.ErrorResponseNotFound(reqID, msg))
				}

				msg := fmt.Sprintf("problem looking up repository for projectSFID: %s", params.ProjectSFID)
				log.WithFields(f).WithError(err).Warn(msg)
				return github_repositories.NewDeleteProjectGithubRepositoryBadRequest().WithPayload(
					utils.ErrorResponseBadRequestWithError(reqID, msg, err))
			}

			err = service.DisableRepository(ctx, params.RepositoryID)
			if err != nil {
				msg := fmt.Sprintf("problem disabling repository for projectSFID: %s, error: %+v", params.ProjectSFID, err)
				log.WithFields(f).WithError(err).Warn(msg)
				return github_repositories.NewDeleteProjectGithubRepositoryBadRequest().WithPayload(
					utils.ErrorResponseBadRequestWithError(reqID, msg, err))
			}

			eventService.LogEvent(&events.LogEventArgs{
				EventType:         events.RepositoryDisabled,
				ExternalProjectID: params.ProjectSFID,
				ProjectID:         ghRepo.RepositoryProjectID,
				LfUsername:        authUser.UserName,
				EventData: &events.RepositoryDisabledEventData{
					RepositoryName: ghRepo.RepositoryName,
				},
			})

			return github_repositories.NewDeleteProjectGithubRepositoryNoContent()
		})

	api.GithubRepositoriesGetProjectGithubRepositoryBranchProtectionHandler = github_repositories.GetProjectGithubRepositoryBranchProtectionHandlerFunc(
		func(params github_repositories.GetProjectGithubRepositoryBranchProtectionParams, authUser *auth.User) middleware.Responder {
			reqID := utils.GetRequestID(params.XREQUESTID)
			utils.SetAuthUserProperties(authUser, params.XUSERNAME, params.XEMAIL)
			ctx := context.WithValue(params.HTTPRequest.Context(), utils.XREQUESTID, reqID) // nolint
			f := logrus.Fields{
				"functionName":   "GithubRepositoriesGetProjectGithubRepositoryBranchProtectionHandler",
				utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
				"authUser":       authUser.UserName,
				"authEmail":      authUser.Email,
				"projectSFID":    params.ProjectSFID,
				"repositoryID":   params.RepositoryID,
			}

			if !utils.IsUserAuthorizedForProjectTree(authUser, params.ProjectSFID, utils.ALLOW_ADMIN_SCOPE) {
				msg := fmt.Sprintf("user %s does not have access to Query Protected Branch GitHub Repositories with Project scope of %s",
					authUser.UserName, params.ProjectSFID)
				log.WithFields(f).Debug(msg)
				return github_repositories.NewGetProjectGithubRepositoryBranchProtectionForbidden().WithPayload(
					utils.ErrorResponseForbidden(reqID, msg))
			}

			protectedBranch, err := service.GetProtectedBranch(ctx, params.ProjectSFID, params.RepositoryID)
			if err != nil {
				if err == repositories.ErrGithubRepositoryNotFound {
					msg := fmt.Sprintf("unable to locatate branch protection projectSFID: %s, repository: %s", params.ProjectSFID, params.RepositoryID)
					log.WithFields(f).WithError(err).Warn(msg)
					return github_repositories.NewGetProjectGithubRepositoryBranchProtectionNotFound().WithPayload(
						utils.ErrorResponseNotFound(reqID, msg))
				}
				if errors.Is(err, github.ErrAccessDenied) {
					msg := fmt.Sprintf("access denied for branch protection for projectSFID: %s, repository: %s", params.ProjectSFID, params.RepositoryID)
					log.WithFields(f).WithError(err).Warn(msg)
					return github_repositories.NewGetProjectGithubRepositoryBranchProtectionForbidden().WithPayload(
						utils.ErrorResponseNotFound(reqID, msg))
				}

				// shall we return the actual code for rate liming ?
				if errors.Is(err, github.ErrRateLimited) {
					msg := fmt.Sprintf("problem loading branch protection for projectSFID: %s, repository: %s", params.ProjectSFID, params.RepositoryID)
					log.WithFields(f).WithError(err).Warn(msg)
					return github_repositories.NewGetProjectGithubRepositoryBranchProtectionInternalServerError().WithPayload(
						utils.ErrorResponseInternalServerErrorWithError(reqID, msg, err))
				}

				msg := fmt.Sprintf("problem loading branch protection for projectSFID: %s, repository: %s, error: %+v", params.ProjectSFID, params.RepositoryID, err)
				log.WithFields(f).WithError(err).Warn(msg)
				return github_repositories.NewGetProjectGithubRepositoryBranchProtectionBadRequest().WithPayload(
					utils.ErrorResponseBadRequestWithError(reqID, msg, err))
			}

			return github_repositories.NewGetProjectGithubRepositoryBranchProtectionOK().WithPayload(protectedBranch)
		})

	api.GithubRepositoriesUpdateProjectGithubRepositoryBranchProtectionHandler = github_repositories.UpdateProjectGithubRepositoryBranchProtectionHandlerFunc(
		func(params github_repositories.UpdateProjectGithubRepositoryBranchProtectionParams, authUser *auth.User) middleware.Responder {
			reqID := utils.GetRequestID(params.XREQUESTID)
			utils.SetAuthUserProperties(authUser, params.XUSERNAME, params.XEMAIL)
			ctx := context.WithValue(params.HTTPRequest.Context(), utils.XREQUESTID, reqID) // nolint
			f := logrus.Fields{
				"functionName":   "GithubRepositoriesUpdateProjectGithubRepositoryBranchProtectionHandler",
				utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
				"authUser":       authUser.UserName,
				"authEmail":      authUser.Email,
				"projectSFID":    params.ProjectSFID,
				"repositoryID":   params.RepositoryID,
			}

			if !utils.IsUserAuthorizedForProjectTree(authUser, params.ProjectSFID, utils.ALLOW_ADMIN_SCOPE) {
				msg := fmt.Sprintf("user %s does not have access to Update Protected Branch GitHub Repositories with Project scope of %s",
					authUser.UserName, params.ProjectSFID)
				log.WithFields(f).Debug(msg)
				return github_repositories.NewUpdateProjectGithubRepositoryBranchProtectionForbidden().WithPayload(
					utils.ErrorResponseForbidden(reqID, msg))
			}

			protectedBranch, err := service.UpdateProtectedBranch(ctx, params.RepositoryID, params.ProjectSFID, params.GithubRepositoryBranchProtectionInput)
			if err != nil {
				log.Warnf("UpdateProjectGithubRepositoryBranchProtectionHandler : failed for repo %s : %v", params.RepositoryID, err)
				if err == repositories.ErrGithubRepositoryNotFound {
					msg := fmt.Sprintf("unable to update branch protection projectSFID: %s, repository: %s", params.ProjectSFID, params.RepositoryID)
					log.WithFields(f).WithError(err).Warn(msg)
					return github_repositories.NewGetProjectGithubRepositoryBranchProtectionNotFound().WithPayload(
						utils.ErrorResponseNotFound(reqID, msg))
				}
				if errors.Is(err, github.ErrAccessDenied) {
					msg := fmt.Sprintf("access denied for branch protection for projectSFID: %s, repository: %s", params.ProjectSFID, params.RepositoryID)
					log.WithFields(f).WithError(err).Warn(msg)
					return github_repositories.NewGetProjectGithubRepositoryBranchProtectionForbidden().WithPayload(
						utils.ErrorResponseNotFound(reqID, msg))
				}

				// shall we return the actual code for rate liming ?
				if errors.Is(err, github.ErrRateLimited) {
					msg := fmt.Sprintf("problem updating branch protection for projectSFID: %s, repository: %s", params.ProjectSFID, params.RepositoryID)
					log.WithFields(f).WithError(err).Warn(msg)
					return github_repositories.NewGetProjectGithubRepositoryBranchProtectionInternalServerError().WithPayload(
						utils.ErrorResponseInternalServerErrorWithError(reqID, msg, err))
				}

				if errors.Is(err, ErrInvalidBranchProtectionName) {
					msg := fmt.Sprintf("problem updating branch protection for projectSFID: %s, repository: %s, error: %+v", params.ProjectSFID, params.RepositoryID, err)
					log.WithFields(f).WithError(err).Warn(msg)
					return github_repositories.NewGetProjectGithubRepositoryBranchProtectionBadRequest().WithPayload(
						utils.ErrorResponseInternalServerErrorWithError(reqID, msg, err))
				}

				msg := fmt.Sprintf("problem updating branch protection for projectSFID: %s, repository: %s, error: %+v", params.ProjectSFID, params.RepositoryID, err)
				log.WithFields(f).WithError(err).Warn(msg)
				return github_repositories.NewGetProjectGithubRepositoryBranchProtectionBadRequest().WithPayload(
					utils.ErrorResponseInternalServerErrorWithError(reqID, msg, err))
			}

			return github_repositories.NewGetProjectGithubRepositoryBranchProtectionOK().WithPayload(protectedBranch)
		})
}
