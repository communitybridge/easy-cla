// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package cla_manager

import (
	"context"
	"fmt"

	"github.com/communitybridge/easycla/cla-backend-go/emails"

	"github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations/cla_manager"

	"github.com/communitybridge/easycla/cla-backend-go/utils"

	"github.com/aws/aws-sdk-go/aws"
	sigAPI "github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations/signatures"
	"github.com/communitybridge/easycla/cla-backend-go/signatures"

	"github.com/communitybridge/easycla/cla-backend-go/company"
	"github.com/communitybridge/easycla/cla-backend-go/events"
	"github.com/communitybridge/easycla/cla-backend-go/gen/models"
	"github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/communitybridge/easycla/cla-backend-go/project"
	"github.com/communitybridge/easycla/cla-backend-go/user"
	"github.com/communitybridge/easycla/cla-backend-go/users"
	"github.com/go-openapi/runtime/middleware"
)

// isValidUser is a helper function to determine if the user object is valid
func isValidUser(claUser *user.CLAUser) bool {
	return claUser != nil && claUser.UserID != "" && claUser.LFUsername != "" && claUser.LFEmail != ""
}

// Configure is the API handler routine for the CLA manager routes
func Configure(api *operations.ClaAPI, service IService, companyService company.IService, projectService project.Service, usersService users.Service, sigService signatures.SignatureService, eventsService events.Service, corporateConsoleURL string) { // nolint
	api.ClaManagerCreateCLAManagerRequestHandler = cla_manager.CreateCLAManagerRequestHandlerFunc(func(params cla_manager.CreateCLAManagerRequestParams, claUser *user.CLAUser) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		if !isValidUser(claUser) {
			return cla_manager.NewCreateCLAManagerRequestUnauthorized().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: "unauthorized",
				Code:    "401",
			})
		}

		companyModel, companyErr := companyService.GetCompany(ctx, params.CompanyID)
		if companyErr != nil || companyModel == nil {
			msg := buildErrorMessage("company lookup error", params, companyErr)
			log.Warn(msg)
			return cla_manager.NewCreateCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		claGroupModel, projectErr := projectService.GetCLAGroupByID(ctx, params.ProjectID)
		if projectErr != nil || claGroupModel == nil {
			msg := buildErrorMessage("project lookup error", params, projectErr)
			log.Warn(msg)
			return cla_manager.NewCreateCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		userModel, userErr := usersService.GetUserByLFUserName(params.Body.UserLFID)
		if userErr != nil || userModel == nil {
			msg := buildErrorMessage("user lookup error", params, userErr)
			log.Warn(msg)
			return cla_manager.NewCreateCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		existingRequests, getErr := service.GetRequestsByUserID(params.CompanyID, params.ProjectID, userModel.UserID)
		if getErr != nil {
			msg := buildErrorMessage("get existing requests", params, getErr)
			log.Warn(msg)
			return cla_manager.NewCreateCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		// Look up signature ACL to ensure the user can approve the request
		sigModels, sigErr := sigService.GetProjectCompanySignatures(ctx, sigAPI.GetProjectCompanySignaturesParams{
			HTTPRequest: nil,
			CompanyID:   params.CompanyID,
			ProjectID:   params.ProjectID,
			NextKey:     nil,
			PageSize:    aws.Int64(5),
		})
		if sigErr != nil || sigModels == nil {
			msg := buildErrorMessage("signature lookup error", params, sigErr)
			log.Warn(msg)
			return cla_manager.NewCreateCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: "CLA Manager Create Request - error reading CCLA Signatures - " + msg,
				Code:    "400",
			})
		}
		if len(sigModels.Signatures) > 1 {
			log.Warnf("returned multiple CCLA signature models for company ID: %s, project ID: %s",
				params.CompanyID, params.ProjectID)
		}

		sigModel := sigModels.Signatures[0]
		claManagers := sigModel.SignatureACL

		var request *models.ClaManagerRequest
		// If no previous requests...
		if existingRequests == nil || existingRequests.Requests == nil {
			var createErr error
			request, createErr = service.CreateRequest(&CLAManagerRequest{
				CompanyID:         params.CompanyID,
				CompanyExternalID: companyModel.CompanyExternalID,
				CompanyName:       companyModel.CompanyName,
				ProjectID:         params.ProjectID,
				ProjectExternalID: claGroupModel.ProjectExternalID,
				ProjectName:       claGroupModel.ProjectName,
				UserID:            params.Body.UserLFID,
				UserExternalID:    userModel.UserExternalID,
				UserName:          params.Body.UserName,
				UserEmail:         params.Body.UserEmail,
				Status:            "pending",
			})
			if createErr != nil {
				msg := buildErrorMessage("create request error", params, createErr)
				log.Warn(msg)
				return cla_manager.NewCreateCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
					Message: msg,
					Code:    "400",
				})
			}

		} else {
			// Ok - we have an existing request with some state...

			// Check to see if we have an existing request in a pending or approved state - if so, don't allow
			for _, existingRequest := range existingRequests.Requests {
				if existingRequest.Status == "pending" || existingRequest.Status == "approved" {
					return cla_manager.NewCreateCLAManagerRequestConflict().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
						Message: "an existing pending request exists for this user for this company and project",
						Code:    "409",
					})
				}
			}

			// Ok - existing state which is either denied or approved - allow them to create another request
			var updateErr error
			request, updateErr = service.PendingRequest(params.CompanyID, params.ProjectID, existingRequests.Requests[0].RequestID)
			if updateErr != nil {
				msg := buildErrorMessage("pending request error", params, updateErr)
				log.Warn(msg)
				return cla_manager.NewCreateCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
					Message: msg,
					Code:    "400",
				})
			}
		}

		// Send an event
		eventsService.LogEvent(&events.LogEventArgs{
			EventType:         events.ClaManagerAccessRequestCreated,
			ProjectID:         params.ProjectID,
			ClaGroupModel:     claGroupModel,
			CompanyID:         params.CompanyID,
			CompanyModel:      companyModel,
			LfUsername:        params.Body.UserLFID,
			UserID:            params.Body.UserLFID,
			UserModel:         userModel,
			ExternalProjectID: claGroupModel.ProjectExternalID,
			EventData: &events.CLAManagerRequestCreatedEventData{
				RequestID:   request.RequestID,
				CompanyName: companyModel.CompanyName,
				ProjectName: claGroupModel.ProjectName,
				UserName:    params.Body.UserName,
				UserEmail:   params.Body.UserEmail,
				UserLFID:    params.Body.UserLFID,
			},
		})

		// Send email to each manager
		for _, manager := range claManagers {
			sendRequestAccessEmailToCLAManagers(companyModel, claGroupModel,
				params.Body.UserName, params.Body.UserEmail,
				manager.Username, manager.LfEmail)
		}

		return cla_manager.NewCreateCLAManagerRequestOK().WithXRequestID(reqID).WithPayload(request)
	})

	// Get Requests
	api.ClaManagerGetCLAManagerRequestsHandler = cla_manager.GetCLAManagerRequestsHandlerFunc(func(params cla_manager.GetCLAManagerRequestsParams, claUser *user.CLAUser) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		//ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		if !isValidUser(claUser) {
			return cla_manager.NewCreateCLAManagerRequestUnauthorized().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: "unauthorized",
				Code:    "401",
			})
		}

		request, err := service.GetRequests(params.CompanyID, params.ProjectID)
		if err != nil {
			msg := buildErrorMessageForGetRequests(params, err)
			log.Warn(msg)
			return cla_manager.NewGetCLAManagerRequestsBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		return cla_manager.NewGetCLAManagerRequestsOK().WithXRequestID(reqID).WithPayload(request)
	})

	// Get Request
	api.ClaManagerGetCLAManagerRequestHandler = cla_manager.GetCLAManagerRequestHandlerFunc(func(params cla_manager.GetCLAManagerRequestParams, claUser *user.CLAUser) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		//ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		if !isValidUser(claUser) {
			return cla_manager.NewCreateCLAManagerRequestUnauthorized().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: "unauthorized",
				Code:    "401",
			})
		}

		request, err := service.GetRequest(params.RequestID)
		if err != nil {
			msg := buildErrorMessageForGetRequest(params, err)
			log.Warn(msg)
			return cla_manager.NewGetCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		// If we didn't find it
		if request == nil {
			return cla_manager.NewGetCLAManagerRequestNotFound().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: fmt.Sprintf("request not found for Company ID: %s, Project ID: %s, Request ID: %s",
					params.CompanyID, params.ProjectID, params.RequestID),
				Code: "404",
			})
		}

		return cla_manager.NewGetCLAManagerRequestOK().WithXRequestID(reqID).WithPayload(request)
	})

	// Approve Request
	api.ClaManagerApproveCLAManagerRequestHandler = cla_manager.ApproveCLAManagerRequestHandlerFunc(func(params cla_manager.ApproveCLAManagerRequestParams, claUser *user.CLAUser) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint

		companyModel, companyErr := companyService.GetCompany(ctx, params.CompanyID)
		if companyErr != nil || companyModel == nil {
			msg := buildErrorMessageForApprove(params, companyErr)
			log.Warn(msg)
			return cla_manager.NewCreateCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		claGroupModel, projectErr := projectService.GetCLAGroupByID(ctx, params.ProjectID)
		if projectErr != nil || claGroupModel == nil {
			msg := buildErrorMessageForApprove(params, projectErr)
			log.Warn(msg)
			return cla_manager.NewCreateCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		// Look up signature ACL to ensure the user can approve the request
		sigModels, sigErr := sigService.GetProjectCompanySignatures(ctx, sigAPI.GetProjectCompanySignaturesParams{
			HTTPRequest: nil,
			CompanyID:   params.CompanyID,
			ProjectID:   params.ProjectID,
			NextKey:     nil,
			PageSize:    aws.Int64(5),
		})
		if sigErr != nil || sigModels == nil {
			msg := buildErrorMessageForApprove(params, sigErr)
			log.Warn(msg)
			return cla_manager.NewApproveCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: "CLA Manager Approve Request - error reading CCLA Signatures - " + msg,
				Code:    "400",
			})
		}
		if len(sigModels.Signatures) > 1 {
			log.Warnf("returned multiple CCLA signature models for company ID: %s, project ID: %s for request ID: %s",
				params.CompanyID, params.ProjectID, params.RequestID)
		}

		sigModel := sigModels.Signatures[0]
		claManagers := sigModel.SignatureACL
		if !currentUserInACL(claUser, claManagers) {
			return cla_manager.NewApproveCLAManagerRequestUnauthorized().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: fmt.Sprintf("CLA Manager %s / %s / %s not authorized to approve request for company ID: %s, project ID: %s",
					claUser.UserID, claUser.Name, claUser.LFEmail, params.CompanyID, params.ProjectID),
				Code: "401",
			})
		}

		// Approve the request
		request, err := service.ApproveRequest(params.CompanyID, params.ProjectID, params.RequestID)
		if err != nil {
			msg := buildErrorMessageForApprove(params, err)
			log.Warn(msg)
			return cla_manager.NewApproveCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		// Update the signature ACL
		_, aclErr := sigService.AddCLAManager(ctx, sigModel.SignatureID, request.UserID)
		if aclErr != nil {
			msg := buildErrorMessageForApprove(params, aclErr)
			log.Warn(msg)
			return cla_manager.NewApproveCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		// Send an event
		eventsService.LogEvent(&events.LogEventArgs{
			EventType: events.ClaManagerAccessRequestApproved,
			ProjectID: params.ProjectID,
			CompanyID: params.CompanyID,
			UserID:    claUser.UserID,
			EventData: &events.CLAManagerRequestApprovedEventData{
				RequestID:    request.RequestID,
				CompanyName:  companyModel.CompanyName,
				ProjectName:  claGroupModel.ProjectName,
				UserName:     request.UserName,
				UserEmail:    request.UserEmail,
				ManagerName:  claUser.Name,    // from the request
				ManagerEmail: claUser.LFEmail, // from the request
			},
		})

		// Notify CLA Managers - send email to each manager
		for _, manager := range claManagers {
			sendRequestApprovedEmailToCLAManagers(companyModel, claGroupModel, request.UserName, request.UserEmail,
				manager.Username, manager.LfEmail)
		}

		// Notify the requester
		sendRequestApprovedEmailToRequester(companyModel, claGroupModel, request.UserName, request.UserEmail)

		return cla_manager.NewCreateCLAManagerRequestOK().WithXRequestID(reqID).WithPayload(request)
	})

	// Deny Request
	api.ClaManagerDenyCLAManagerRequestHandler = cla_manager.DenyCLAManagerRequestHandlerFunc(func(params cla_manager.DenyCLAManagerRequestParams, claUser *user.CLAUser) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint

		companyModel, companyErr := companyService.GetCompany(ctx, params.CompanyID)
		if companyErr != nil || companyModel == nil {
			msg := buildErrorMessageForDeny(params, companyErr)
			log.Warn(msg)
			return cla_manager.NewDenyCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		claGroupModel, projectErr := projectService.GetCLAGroupByID(ctx, params.ProjectID)
		if projectErr != nil || claGroupModel == nil {
			msg := buildErrorMessageForDeny(params, projectErr)
			log.Warn(msg)
			return cla_manager.NewDenyCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		// Look up signature ACL to ensure the user can deny the request
		sigModels, sigErr := sigService.GetProjectCompanySignatures(ctx, sigAPI.GetProjectCompanySignaturesParams{
			HTTPRequest: nil,
			CompanyID:   params.CompanyID,
			ProjectID:   params.ProjectID,
			NextKey:     nil,
			PageSize:    aws.Int64(5),
		})
		if sigErr != nil || sigModels == nil {
			msg := buildErrorMessageForDeny(params, sigErr)
			log.Warn(msg)
			return cla_manager.NewApproveCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: "CLA Manager Deny Request - error reading CCLA Signatures - " + msg,
				Code:    "400",
			})
		}
		if len(sigModels.Signatures) > 1 {
			log.Warnf("returned multiple CCLA signature models for company ID: %s, project ID: %s for request ID: %s",
				params.CompanyID, params.ProjectID, params.RequestID)
		}

		sigModel := sigModels.Signatures[0]
		claManagers := sigModel.SignatureACL
		if !currentUserInACL(claUser, claManagers) {
			return cla_manager.NewApproveCLAManagerRequestUnauthorized().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: fmt.Sprintf("CLA Manager %s / %s / %s not authorized to approve request for company ID: %s, project ID: %s",
					claUser.UserID, claUser.Name, claUser.LFEmail, params.CompanyID, params.ProjectID),
				Code: "401",
			})
		}

		request, err := service.DenyRequest(params.CompanyID, params.ProjectID, params.RequestID)
		if err != nil {
			msg := buildErrorMessageForDeny(params, err)
			log.Warn(msg)
			return cla_manager.NewDenyCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		// Send an event
		eventsService.LogEvent(&events.LogEventArgs{
			EventType: events.ClaManagerAccessRequestDenied,
			ProjectID: params.ProjectID,
			CompanyID: params.CompanyID,
			UserID:    claUser.UserID,
			EventData: &events.CLAManagerRequestDeniedEventData{
				RequestID:    request.RequestID,
				CompanyName:  companyModel.CompanyName,
				ProjectName:  claGroupModel.ProjectName,
				UserName:     request.UserName,
				UserEmail:    request.UserEmail,
				ManagerName:  claUser.Name,    // from the request
				ManagerEmail: claUser.LFEmail, // from the request
			},
		})

		// Notify CLA Managers - send email to each manager
		for _, manager := range claManagers {
			sendRequestDeniedEmailToCLAManagers(companyModel, claGroupModel, request.UserName, request.UserEmail,
				manager.Username, manager.LfEmail)
		}

		// Notify the requester
		sendRequestDeniedEmailToRequester(companyModel, claGroupModel, request.UserName, request.UserEmail)

		return cla_manager.NewCreateCLAManagerRequestOK().WithPayload(request)
	})

	// Delete Request
	api.ClaManagerDeleteCLAManagerRequestHandler = cla_manager.DeleteCLAManagerRequestHandlerFunc(func(params cla_manager.DeleteCLAManagerRequestParams, claUser *user.CLAUser) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint

		// Make sure the company id exists...
		companyModel, companyErr := companyService.GetCompany(ctx, params.CompanyID)
		if companyErr != nil || companyModel == nil {
			msg := buildErrorMessageForDelete(params, companyErr)
			return cla_manager.NewDeleteCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		// Make sure the project id exists...
		claGroupModel, projectErr := projectService.GetCLAGroupByID(ctx, params.ProjectID)
		if projectErr != nil || claGroupModel == nil {
			msg := buildErrorMessageForDelete(params, projectErr)
			log.Warn(msg)
			return cla_manager.NewDenyCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		// Make sure the request exists...
		request, err := service.GetRequest(params.RequestID)
		if err != nil {
			msg := buildErrorMessageForDelete(params, err)
			log.Warn(msg)
			return cla_manager.NewDeleteCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		if request == nil {
			msg := buildErrorMessageForDelete(params, err)
			log.Warn(msg)
			return cla_manager.NewDeleteCLAManagerRequestNotFound().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "404",
			})
		}

		// Look up signature ACL to ensure the user can delete the request
		sigModels, sigErr := sigService.GetProjectCompanySignatures(ctx, sigAPI.GetProjectCompanySignaturesParams{
			HTTPRequest: nil,
			CompanyID:   params.CompanyID,
			ProjectID:   params.ProjectID,
			NextKey:     nil,
			PageSize:    aws.Int64(5),
		})
		if sigErr != nil || sigModels == nil {
			msg := buildErrorMessageForDelete(params, sigErr)
			log.Warn(msg)
			return cla_manager.NewDeleteCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: "EasyCLA - 400 Bad Request - CLA Manager Delete Request - error reading CCLA Signatures - " + msg,
				Code:    "400",
			})
		}
		if len(sigModels.Signatures) > 1 {
			log.Warnf("returned multiple CCLA signature models for company ID: %s, project ID: %s for request ID: %s",
				params.CompanyID, params.ProjectID, params.RequestID)
		}

		sigModel := sigModels.Signatures[0]
		claManagers := sigModel.SignatureACL
		if !currentUserInACL(claUser, claManagers) {
			msg := fmt.Sprintf("EasyCLA - 401 Unauthorized - CLA Manager %s / %s / %s not authorized to delete requests for company ID: %s, project ID: %s",
				claUser.UserID, claUser.Name, claUser.LFEmail, params.CompanyID, params.ProjectID)
			log.Debug(msg)
			return cla_manager.NewDeleteCLAManagerRequestUnauthorized().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "401",
			})
		}

		// Delete the request
		deleteErr := service.DeleteRequest(params.RequestID)
		if deleteErr != nil {
			msg := buildErrorMessageForDelete(params, deleteErr)
			log.Warn(msg)
			return cla_manager.NewDeleteCLAManagerRequestBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		// Send an event
		eventsService.LogEvent(&events.LogEventArgs{
			EventType: events.ClaManagerAccessRequestDeleted,
			ProjectID: params.ProjectID,
			CompanyID: params.CompanyID,
			UserID:    claUser.UserID,
			EventData: &events.CLAManagerRequestDeniedEventData{
				RequestID:    params.RequestID,
				CompanyName:  companyModel.CompanyName,
				ProjectName:  claGroupModel.ProjectName,
				UserName:     request.UserName,
				UserEmail:    request.UserEmail,
				ManagerName:  claUser.Name,    // from the request
				ManagerEmail: claUser.LFEmail, // from the request
			},
		})

		log.Debug("CLA Manager Delete - Returning Success")
		return cla_manager.NewDeleteCLAManagerRequestNoContent().WithXRequestID(reqID)
	})

	api.ClaManagerAddCLAManagerHandler = cla_manager.AddCLAManagerHandlerFunc(func(params cla_manager.AddCLAManagerParams, claUser *user.CLAUser) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint

		userModel, userErr := usersService.GetUserByLFUserName(params.Body.UserLFID)
		if userErr != nil || userModel == nil {
			msg := fmt.Sprintf("User lookup for user by LFID: %s failed ", params.Body.UserLFID)
			log.Warn(msg)
			return cla_manager.NewAddCLAManagerBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: "EasyCLA - 400 Bad Request - Add CLA Manager - error getting user - " + msg,
				Code:    "400",
			})
		}
		companyModel, companyErr := companyService.GetCompany(ctx, params.CompanyID)
		if companyErr != nil || companyModel == nil {
			msg := fmt.Sprintf("User lookup for company by ID: %s failed ", params.CompanyID)
			log.Warn(msg)
			return cla_manager.NewAddCLAManagerBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: "EasyCLA - 400 Bad Request - Add CLA Manager - error getting company - " + msg,
				Code:    "400",
			})
		}

		claGroupModel, projectErr := projectService.GetCLAGroupByID(ctx, params.ProjectID)
		if projectErr != nil || claGroupModel == nil {
			msg := fmt.Sprintf("User lookup for project by ID: %s failed ", params.ProjectID)
			log.Warn(msg)
			return cla_manager.NewAddCLAManagerBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: "EasyCLA - 400 Bad Request - Add CLA Manager - error getting project - " + msg,
				Code:    "400",
			})
		}

		// Look up signature ACL to ensure the user is allowed to add CLA managers
		sigModels, sigErr := sigService.GetProjectCompanySignatures(ctx, sigAPI.GetProjectCompanySignaturesParams{
			HTTPRequest: nil,
			CompanyID:   params.CompanyID,
			ProjectID:   params.ProjectID,
			NextKey:     nil,
			PageSize:    aws.Int64(5),
		})
		if sigErr != nil || sigModels == nil {
			msg := buildErrorMessageAddManager("Add CLA Manager - signature lookup error", params, sigErr)
			log.Warn(msg)
			return cla_manager.NewAddCLAManagerBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: "EasyCLA - 400 Bad Request - Add CLA Manager - error reading CCLA Signatures - " + msg,
				Code:    "400",
			})
		}
		if len(sigModels.Signatures) > 1 {
			log.Warnf("returned multiple CCLA signature models for company ID: %s, project ID: %s",
				params.CompanyID, params.ProjectID)
		}

		sigModel := sigModels.Signatures[0]
		claManagers := sigModel.SignatureACL
		if !currentUserInACL(claUser, claManagers) {
			msg := fmt.Sprintf("EasyCLA - 401 Unauthorized - User %s / %s / %s is not authorized to add a CLA Manager for company ID: %s, project ID: %s",
				claUser.UserID, claUser.Name, claUser.LFEmail, params.CompanyID, params.ProjectID)
			log.Debug(msg)
			return cla_manager.NewAddCLAManagerUnauthorized().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "401",
			})
		}

		// Audit Event sent from service upon success
		signature, addErr := service.AddClaManager(ctx, params.CompanyID, params.ProjectID, params.Body.UserLFID)
		if addErr != nil {
			msg := buildErrorMessageAddManager("Add CLA Manager - Service Error", params, addErr)
			log.Warn(msg)
			return cla_manager.NewAddCLAManagerBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		return cla_manager.NewAddCLAManagerOK().WithXRequestID(reqID).WithPayload(signature)
	})

	// Delete CLA Manager
	api.ClaManagerDeleteCLAManagerHandler = cla_manager.DeleteCLAManagerHandlerFunc(func(params cla_manager.DeleteCLAManagerParams, claUser *user.CLAUser) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint

		userModel, userErr := usersService.GetUserByLFUserName(params.UserLFID)
		if userErr != nil || userModel == nil {
			msg := fmt.Sprintf("User lookup for user by LFID: %s failed ", params.UserLFID)
			log.Warn(msg)
			return cla_manager.NewDeleteCLAManagerBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: "EasyCLA - 400 Bad Request - Delete CLA Manager - error getting user - " + msg,
				Code:    "400",
			})
		}
		companyModel, companyErr := companyService.GetCompany(ctx, params.CompanyID)
		if companyErr != nil || companyModel == nil {
			msg := fmt.Sprintf("User lookup for company by ID: %s failed ", params.CompanyID)
			log.Warn(msg)
			return cla_manager.NewDeleteCLAManagerBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: "EasyCLA - 400 Bad Request - Delete CLA Manager - error getting company - " + msg,
				Code:    "400",
			})
		}

		claGroupModel, projectErr := projectService.GetCLAGroupByID(ctx, params.ProjectID)
		if projectErr != nil || claGroupModel == nil {
			msg := fmt.Sprintf("User lookup for project by ID: %s failed ", params.ProjectID)
			log.Warn(msg)
			return cla_manager.NewDeleteCLAManagerBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: "EasyCLA - 400 Bad Request - Delete CLA Manager - error getting project - " + msg,
				Code:    "400",
			})
		}
		// Look up signature ACL to ensure the user is allowed to remove CLA managers
		sigModels, sigErr := sigService.GetProjectCompanySignatures(ctx, sigAPI.GetProjectCompanySignaturesParams{
			HTTPRequest: nil,
			CompanyID:   params.CompanyID,
			ProjectID:   params.ProjectID,
			NextKey:     nil,
			PageSize:    aws.Int64(5),
		})
		if sigErr != nil || sigModels == nil {
			msg := buildErrorMessageDeleteManager("Delete CLA Manager - Signature Lookup Error", params, sigErr)
			log.Warn(msg)
			return cla_manager.NewDeleteCLAManagerBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: "EasyCLA - 400 Bad Request - Delete CLA Manager - error reading CCLA Signatures - " + msg,
				Code:    "400",
			})
		}
		if len(sigModels.Signatures) > 1 {
			log.Warnf("returned multiple CCLA signature models for company ID: %s, project ID: %s",
				params.CompanyID, params.ProjectID)
		}

		sigModel := sigModels.Signatures[0]
		claManagers := sigModel.SignatureACL
		if !currentUserInACL(claUser, claManagers) {
			msg := fmt.Sprintf("User %s / %s / %s is not authorized to remove a CLA Manager for company ID: %s, project ID: %s",
				claUser.UserID, claUser.Name, claUser.LFEmail, params.CompanyID, params.ProjectID)
			log.Debug(msg)
			return cla_manager.NewDeleteCLAManagerUnauthorized().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "401",
			})
		}

		// Audit Event sent from service upon success
		signature, deleteErr := service.RemoveClaManager(ctx, params.CompanyID, params.ProjectID, params.UserLFID)

		if deleteErr != nil {
			msg := buildErrorMessageDeleteManager("EasyCLA - 400 Bad Request - Delete CLA Manager - Service Error", params, deleteErr)
			log.Warn(msg)
			return cla_manager.NewDeleteCLAManagerBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		if signature == nil {
			msg := buildErrorMessageDeleteManager("EasyCLA - 400 Bad Request - Delete CLA Manager - Response Signature Missing", params, deleteErr)
			log.Warn(msg)
			return cla_manager.NewDeleteCLAManagerBadRequest().WithXRequestID(reqID).WithPayload(&models.ErrorResponse{
				Message: msg,
				Code:    "400",
			})
		}

		return cla_manager.NewDeleteCLAManagerOK().WithXRequestID(reqID).WithPayload(signature)
	})
}

// currentUserInACL is a helper function to determine if the current logged in user is in the CLA Manager list
func currentUserInACL(currentUser *user.CLAUser, managers []models.User) bool {
	//log.Debugf("checking if user: %+v is in the Signature ACL: %+v", currentUser, managers)
	var inACL = false
	for _, manager := range managers {
		if manager.UserID == currentUser.UserID {
			inACL = true
			break
		}
	}

	return inACL
}

// buildErrorMessage helper function to build an error message
func buildErrorMessage(errPrefix string, params cla_manager.CreateCLAManagerRequestParams, err error) string {
	return fmt.Sprintf("%s - problem creating new CLA Manager Request using company ID: %s, project ID: %s, user ID: %s, user name: %s, user email: %s, error: %+v",
		errPrefix, params.CompanyID, params.ProjectID, params.Body.UserLFID, params.Body.UserName, params.Body.UserEmail, err)
}

// buildErrorMessageForApprove is a helper function to build an error message
func buildErrorMessageForApprove(params cla_manager.ApproveCLAManagerRequestParams, err error) string {
	return fmt.Sprintf("problem approving the CLA Manager Request using company ID: %s, project ID: %s, request ID: %s, error: %+v",
		params.CompanyID, params.ProjectID, params.RequestID, err)
}

// buildErrorMessageForDeny is a helper function to build an error message
func buildErrorMessageForDeny(params cla_manager.DenyCLAManagerRequestParams, err error) string {
	return fmt.Sprintf("problem denying the CLA Manager Request using company ID: %s, project ID: %s, request ID: %s, error: %+v",
		params.CompanyID, params.ProjectID, params.RequestID, err)
}

// buildErrorMessageForDelete is a helper function to build an error message
func buildErrorMessageForDelete(params cla_manager.DeleteCLAManagerRequestParams, err error) string {
	return fmt.Sprintf("problem deleting the CLA Manager Request using company ID: %s, project ID: %s, request ID: %s, error: %+v",
		params.CompanyID, params.ProjectID, params.RequestID, err)
}

// buildErrorMessageForGetRequests is a helper function to build an error message
func buildErrorMessageForGetRequests(params cla_manager.GetCLAManagerRequestsParams, err error) string {
	return fmt.Sprintf("problem fetching the CLA Manager Requests using company ID: %s, project ID: %s, error: %+v",
		params.CompanyID, params.ProjectID, err)
}

// buildErrorMessageForGetRequest is a helper function to build an error message
func buildErrorMessageForGetRequest(params cla_manager.GetCLAManagerRequestParams, err error) string {
	return fmt.Sprintf("problem fetching the CLA Manager Requests using company ID: %s, project ID: %s, request ID: %s, error: %+v",
		params.CompanyID, params.ProjectID, params.RequestID, err)
}

// buildErrorMessageAdd helper function to build an error message
func buildErrorMessageAddManager(errPrefix string, params cla_manager.AddCLAManagerParams, err error) string {
	return fmt.Sprintf("%s - problem adding CLA Manager to company ID: %s, project ID: %s, user ID: %s, user name: %s, user email: %s, error: %+v",
		errPrefix, params.CompanyID, params.ProjectID, params.Body.UserLFID, params.Body.UserName, params.Body.UserEmail, err)
}

// buildErrorMessage helper function to build an error message
func buildErrorMessageDeleteManager(errPrefix string, params cla_manager.DeleteCLAManagerParams, err error) string {
	return fmt.Sprintf("%s - problem deleting CLA Manager for company ID: %s, project ID: %s, user ID: %s, error: %+v",
		errPrefix, params.CompanyID, params.ProjectID, params.UserLFID, err)
}

// sendRequestAccessEmailToCLAManagers sends the request access email to the specified CLA Managers
func sendRequestAccessEmailToCLAManagers(companyModel *models.Company, claGroupModel *models.ClaGroup, requesterName, requesterEmail, recipientName, recipientAddress string) {
	companyName := companyModel.CompanyName
	projectName := claGroupModel.ProjectName

	// subject string, body string, recipients []string
	subject := fmt.Sprintf("EasyCLA: New CLA Manager Access Request for %s on %s", companyName, projectName)
	recipients := []string{recipientAddress}
	body, err := emails.RenderTemplate(claGroupModel.Version, emails.RequestAccessToCLAManagersTemplateName,
		emails.RequestAccessToCLAManagersTemplate, emails.RequestAccessToCLAManagersTemplateParams{
			CLAManagerTemplateParams: emails.CLAManagerTemplateParams{
				RecipientName: recipientName,
				ProjectName:   projectName,
				CompanyName:   companyName,
			},
			RequesterName:  requesterName,
			RequesterEmail: requesterEmail,
			CorporateURL:   utils.GetCorporateURL(claGroupModel.Version == utils.V2),
		})
	if err != nil {
		log.Warnf("rendering email template : %s failed : %v", emails.RequestAccessToCLAManagersTemplateName, err)
		return
	}

	err = utils.SendEmail(subject, body, recipients)
	if err != nil {
		log.Warnf("problem sending email with subject: %s to recipients: %+v, error: %+v", subject, recipients, err)
	} else {
		log.Debugf("sent email with subject: %s to recipients: %+v", subject, recipients)
	}
}

func sendRequestApprovedEmailToCLAManagers(companyModel *models.Company, claGroupModel *models.ClaGroup, requesterName, requesterEmail, recipientName, recipientAddress string) {
	companyName := companyModel.CompanyName
	projectName := claGroupModel.ProjectName

	// subject string, body string, recipients []string
	subject := fmt.Sprintf("EasyCLA: CLA Manager Access Approval Notice for %s", projectName)
	recipients := []string{recipientAddress}
	body, err := emails.RenderTemplate(
		claGroupModel.Version,
		emails.RequestApprovedToCLAManagersTemplateName,
		emails.RequestApprovedToCLAManagersTemplate,
		emails.RequestApprovedToCLAManagersTemplateParams{
			CLAManagerTemplateParams: emails.CLAManagerTemplateParams{
				RecipientName: recipientName,
				ProjectName:   projectName,
				CompanyName:   companyName,
			},
			RequesterName:  requesterName,
			RequesterEmail: requesterEmail,
		})

	if err != nil {
		log.Warnf("rendering email template : %s failed : %v", emails.RequestApprovedToCLAManagersTemplateName, err)
		return
	}
	err = utils.SendEmail(subject, body, recipients)
	if err != nil {
		log.Warnf("problem sending email with subject: %s to recipients: %+v, error: %+v", subject, recipients, err)
	} else {
		log.Debugf("sent email with subject: %s to recipients: %+v", subject, recipients)
	}
}

func sendRequestApprovedEmailToRequester(companyModel *models.Company, claGroupModel *models.ClaGroup, requesterName, requesterEmail string) {
	companyName := companyModel.CompanyName
	projectName := claGroupModel.ProjectName

	// subject string, body string, recipients []string
	subject := fmt.Sprintf("EasyCLA: New CLA Manager Access Approved for %s", projectName)
	recipients := []string{requesterEmail}
	body, err := emails.RenderTemplate(claGroupModel.Version, emails.RequestApprovedToRequesterTemplateName,
		emails.RequestApprovedToRequesterTemplate, emails.RequestApprovedToRequesterTemplateParams{
			CLAManagerTemplateParams: emails.CLAManagerTemplateParams{
				RecipientName: requesterName,
				ProjectName:   projectName,
				CompanyName:   companyName,
			},
			CorporateURL: utils.GetCorporateURL(claGroupModel.Version == utils.V2),
		})
	if err != nil {
		log.Warnf("email template : %s failed rendering : %s", emails.RequestApprovedToRequesterTemplateName, err)
		return
	}
	err = utils.SendEmail(subject, body, recipients)
	if err != nil {
		log.Warnf("problem sending email with subject: %s to recipients: %+v, error: %+v", subject, recipients, err)
	} else {
		log.Debugf("sent email with subject: %s to recipients: %+v", subject, recipients)
	}
}

func sendRequestDeniedEmailToCLAManagers(companyModel *models.Company, claGroupModel *models.ClaGroup, requesterName, requesterEmail, recipientName, recipientAddress string) {
	companyName := companyModel.CompanyName
	projectName := claGroupModel.ProjectName

	// subject string, body string, recipients []string
	subject := fmt.Sprintf("EasyCLA: CLA Manager Access Denied Notice for %s", projectName)
	recipients := []string{recipientAddress}
	body, err := emails.RenderTemplate(
		claGroupModel.Version,
		emails.RequestDeniedToCLAManagersTemplateName,
		emails.RequestDeniedToCLAManagersTemplate,
		emails.RequestDeniedToCLAManagersTemplateParams{
			CLAManagerTemplateParams: emails.CLAManagerTemplateParams{
				RecipientName: recipientName,
				ProjectName:   projectName,
				CompanyName:   companyName,
			},
			RequesterName:  requesterName,
			RequesterEmail: requesterEmail,
		},
	)

	if err != nil {
		log.Warnf("email template render : %s failed : %v", emails.RequestDeniedToCLAManagersTemplateName, err)
		return
	}

	err = utils.SendEmail(subject, body, recipients)
	if err != nil {
		log.Warnf("problem sending email with subject: %s to recipients: %+v, error: %+v", subject, recipients, err)
	} else {
		log.Debugf("sent email with subject: %s to recipients: %+v", subject, recipients)
	}
}

func sendRequestDeniedEmailToRequester(companyModel *models.Company, claGroupModel *models.ClaGroup, requesterName, requesterEmail string) {
	companyName := companyModel.CompanyName
	projectName := claGroupModel.ProjectName

	// subject string, body string, recipients []string
	subject := fmt.Sprintf("EasyCLA: New CLA Manager Access Denied for %s", projectName)
	recipients := []string{requesterEmail}
	body, err := emails.RenderTemplate(claGroupModel.Version, emails.RequestDeniedToRequesterTemplateName,
		emails.RequestDeniedToRequesterTemplate,
		emails.RequestDeniedToRequesterTemplateParams{
			CLAManagerTemplateParams: emails.CLAManagerTemplateParams{
				RecipientName: requesterName,
				ProjectName:   projectName,
				CompanyName:   companyName,
			},
		})

	if err != nil {
		log.Warnf("email template rendering %s failed : %v", emails.RequestDeniedToRequesterTemplateName, err)
		return
	}

	err = utils.SendEmail(subject, body, recipients)
	if err != nil {
		log.Warnf("problem sending email with subject: %s to recipients: %+v, error: %+v", subject, recipients, err)
	} else {
		log.Debugf("sent email with subject: %s to recipients: %+v", subject, recipients)
	}
}
