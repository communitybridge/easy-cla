// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package project

import (
	"github.com/communitybridge/easycla/cla-backend-go/gen/models"
	"github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations"
	"github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations/project"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/communitybridge/easycla/cla-backend-go/user"

	"github.com/go-openapi/runtime/middleware"
)

// Configure establishes the middleware handlers for the project service
func Configure(api *operations.ClaAPI, service Service) {
	api.ProjectCreateProjectHandler = project.CreateProjectHandlerFunc(func(params project.CreateProjectParams, claUser *user.CLAUser) middleware.Responder {
		projectModel, err := service.CreateProject(&params.Body)
		if err != nil {
			return project.NewCreateProjectBadRequest().WithPayload(errorResponse(err))
		}

		return project.NewCreateProjectOK().WithPayload(projectModel)
	})

	// Get Projects
	api.ProjectGetProjectsHandler = project.GetProjectsHandlerFunc(func(params project.GetProjectsParams, claUser *user.CLAUser) middleware.Responder {

		projects, err := service.GetProjects(&params)
		if err != nil {
			return project.NewGetProjectsBadRequest().WithPayload(errorResponse(err))
		}

		return project.NewGetProjectsOK().WithPayload(projects)
	})

	// Get Project By ID
	api.ProjectGetProjectByIDHandler = project.GetProjectByIDHandlerFunc(func(projectParams project.GetProjectByIDParams, claUser *user.CLAUser) middleware.Responder {

		projectModel, err := service.GetProjectByID(projectParams.ProjectSfdcID)
		if err != nil {
			return project.NewGetProjectByIDBadRequest().WithPayload(errorResponse(err))
		}
		if projectModel == nil {
			return project.NewGetProjectByIDNotFound()
		}

		return project.NewGetProjectByIDOK().WithPayload(projectModel)
	})

	// Delete Project By ID
	api.ProjectDeleteProjectByIDHandler = project.DeleteProjectByIDHandlerFunc(func(projectParams project.DeleteProjectByIDParams, claUser *user.CLAUser) middleware.Responder {
		log.Debugf("Processing delete request with project id: %s", projectParams.ProjectSfdcID)
		err := service.DeleteProject(projectParams.ProjectSfdcID)
		if err != nil {
			if err == ErrProjectDoesNotExist {
				return project.NewDeleteProjectByIDNotFound()
			}
			return project.NewDeleteProjectByIDBadRequest().WithPayload(errorResponse(err))
		}

		return project.NewDeleteProjectByIDNoContent()
	})

	// Update Project By ID
	api.ProjectUpdateProjectHandler = project.UpdateProjectHandlerFunc(func(projectParams project.UpdateProjectParams, claUser *user.CLAUser) middleware.Responder {
		projectModel, err := service.UpdateProject(&projectParams.Body)
		if err != nil {
			if err == ErrProjectDoesNotExist {
				return project.NewUpdateProjectNotFound()
			}
			return project.NewUpdateProjectBadRequest().WithPayload(errorResponse(err))
		}

		return project.NewUpdateProjectOK().WithPayload(projectModel)
	})
}

// codedResponse interface
type codedResponse interface {
	Code() string
}

// errorResponse is a helper to wrap the specified error into an error response model
func errorResponse(err error) *models.ErrorResponse {
	code := ""
	if e, ok := err.(codedResponse); ok {
		code = e.Code()
	}

	e := models.ErrorResponse{
		Code:    code,
		Message: err.Error(),
	}

	return &e
}
