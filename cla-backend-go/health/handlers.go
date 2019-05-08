package health

import (
	"github.com/LF-Engineering/cla-monorepo/cla-backend-go/gen/models"
	"github.com/LF-Engineering/cla-monorepo/cla-backend-go/gen/restapi/operations"
	"github.com/LF-Engineering/cla-monorepo/cla-backend-go/user"

	"github.com/go-openapi/runtime/middleware"
)

// Configure setups handlers on api with service
func Configure(api *operations.ClaAPI, service service) {

	api.HealthCheckHandler = operations.HealthCheckHandlerFunc(func(params operations.HealthCheckParams, user *user.CLAUser) middleware.Responder {
		result, err := service.HealthCheck(params.HTTPRequest.Context(), params)
		if err != nil {
			return operations.NewHealthCheckBadRequest().WithPayload(errorResponse(err))
		}

		return operations.NewHealthCheckOK().WithPayload(result)
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
