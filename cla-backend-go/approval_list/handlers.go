// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package approval_list

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/communitybridge/easycla/cla-backend-go/events"
	"github.com/communitybridge/easycla/cla-backend-go/gen/models"
	"github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations"
	"github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations/company"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/communitybridge/easycla/cla-backend-go/signatures"
	"github.com/communitybridge/easycla/cla-backend-go/user"
	"github.com/communitybridge/easycla/cla-backend-go/utils"
	"github.com/go-openapi/runtime/middleware"
	"github.com/savaki/dynastore"
)

// Configure setups handlers on api with service
func Configure(api *operations.ClaAPI, service IService, sessionStore *dynastore.Store, signatureService signatures.SignatureService, eventsService events.Service) {

	api.CompanyAddCclaWhitelistRequestHandler = company.AddCclaWhitelistRequestHandlerFunc(
		func(params company.AddCclaWhitelistRequestParams) middleware.Responder {
			reqID := utils.GetRequestID(params.XREQUESTID)
			ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
			requestID, err := service.AddCclaWhitelistRequest(ctx, params.CompanyID, params.ProjectID, params.Body)
			if err != nil {
				return company.NewAddCclaWhitelistRequestBadRequest().WithXRequestID(reqID).WithPayload(errorResponse(err))
			}

			eventsService.LogEvent(&events.LogEventArgs{
				EventType: events.CCLAApprovalListRequestCreated,
				ProjectID: params.ProjectID,
				CompanyID: params.CompanyID,
				UserID:    params.Body.ContributorID,
				EventData: &events.CCLAApprovalListRequestCreatedEventData{RequestID: requestID},
			})

			return company.NewAddCclaWhitelistRequestOK().WithXRequestID(reqID)
		})

	api.CompanyApproveCclaWhitelistRequestHandler = company.ApproveCclaWhitelistRequestHandlerFunc(
		func(params company.ApproveCclaWhitelistRequestParams, claUser *user.CLAUser) middleware.Responder {
			reqID := utils.GetRequestID(params.XREQUESTID)
			ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
			err := service.ApproveCclaWhitelistRequest(ctx, claUser, params.CompanyID, params.ProjectID, params.RequestID)
			if err != nil {
				return company.NewApproveCclaWhitelistRequestBadRequest().WithXRequestID(reqID).WithPayload(errorResponse(err))
			}

			eventsService.LogEvent(&events.LogEventArgs{
				EventType: events.CCLAApprovalListRequestApproved,
				ProjectID: params.ProjectID,
				CompanyID: params.CompanyID,
				UserID:    claUser.UserID,
				EventData: &events.CCLAApprovalListRequestApprovedEventData{RequestID: params.RequestID},
			})

			return company.NewApproveCclaWhitelistRequestOK().WithXRequestID(reqID)
		})

	api.CompanyRejectCclaWhitelistRequestHandler = company.RejectCclaWhitelistRequestHandlerFunc(
		func(params company.RejectCclaWhitelistRequestParams, claUser *user.CLAUser) middleware.Responder {
			reqID := utils.GetRequestID(params.XREQUESTID)
			ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
			err := service.RejectCclaWhitelistRequest(ctx, params.CompanyID, params.ProjectID, params.RequestID)
			if err != nil {
				return company.NewRejectCclaWhitelistRequestBadRequest().WithXRequestID(reqID).WithPayload(errorResponse(err))
			}

			eventsService.LogEvent(&events.LogEventArgs{
				EventType: events.CCLAApprovalListRequestRejected,
				ProjectID: params.ProjectID,
				CompanyID: params.CompanyID,
				UserID:    claUser.UserID,
				EventData: &events.CCLAApprovalListRequestRejectedEventData{RequestID: params.RequestID},
			})

			return company.NewRejectCclaWhitelistRequestOK().WithXRequestID(reqID)
		})

	api.CompanyListCclaWhitelistRequestsHandler = company.ListCclaWhitelistRequestsHandlerFunc(
		func(params company.ListCclaWhitelistRequestsParams, claUser *user.CLAUser) middleware.Responder {
			reqID := utils.GetRequestID(params.XREQUESTID)
			ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
			f := logrus.Fields{
				"functionName":   "CompanyListCclaWhitelistRequestsHandler",
				utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
			}
			log.WithFields(f).Debugf("Invoking ListCclaWhitelistRequest with Company ID: %+v, Project ID: %+v, Status: %+v",
				params.CompanyID, params.ProjectID, params.Status)
			result, err := service.ListCclaWhitelistRequest(params.CompanyID, params.ProjectID, params.Status)
			if err != nil {
				return company.NewListCclaWhitelistRequestsBadRequest().WithXRequestID(reqID).WithPayload(errorResponse(err))
			}

			return company.NewListCclaWhitelistRequestsOK().WithXRequestID(reqID).WithPayload(result)
		})

	api.CompanyListCclaWhitelistRequestsByCompanyAndProjectHandler = company.ListCclaWhitelistRequestsByCompanyAndProjectHandlerFunc(
		func(params company.ListCclaWhitelistRequestsByCompanyAndProjectParams, claUser *user.CLAUser) middleware.Responder {
			log.Debugf("Invoking ListCclaWhitelistRequestByCompanyProjectUser with Company ID: %+v, Project ID: %+v, Status: %+v",
				params.CompanyID, params.ProjectID, params.Status)
			result, err := service.ListCclaWhitelistRequestByCompanyProjectUser(params.CompanyID, &params.ProjectID, params.Status, nil)
			if err != nil {
				return company.NewListCclaWhitelistRequestsByCompanyAndProjectBadRequest().WithPayload(errorResponse(err))
			}

			return company.NewListCclaWhitelistRequestsByCompanyAndProjectOK().WithPayload(result)
		})

	api.CompanyListCclaWhitelistRequestsByCompanyAndProjectAndUserHandler = company.ListCclaWhitelistRequestsByCompanyAndProjectAndUserHandlerFunc(
		func(params company.ListCclaWhitelistRequestsByCompanyAndProjectAndUserParams, claUser *user.CLAUser) middleware.Responder {
			log.Debugf("Invoking ListCclaWhitelistRequestByCompanyProjectUser with Company ID: %+v, Project ID: %+v, Status: %+v, User: %+v",
				params.CompanyID, params.ProjectID, params.Status, claUser.LFUsername)
			result, err := service.ListCclaWhitelistRequestByCompanyProjectUser(params.CompanyID, &params.ProjectID, params.Status, &claUser.LFUsername)
			if err != nil {
				return company.NewListCclaWhitelistRequestsByCompanyAndProjectAndUserBadRequest().WithPayload(errorResponse(err))
			}

			return company.NewListCclaWhitelistRequestsByCompanyAndProjectAndUserOK().WithPayload(result)
		})
}

type codedResponse interface {
	Code() string
}

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
