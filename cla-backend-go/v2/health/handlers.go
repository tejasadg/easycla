// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package health

import (
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/restapi/operations"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/restapi/operations/health"
	v1Health "github.com/communitybridge/easycla/cla-backend-go/health"
	"github.com/go-openapi/runtime/middleware"
)

// Configure setups handlers on api with service
func Configure(api *operations.EasyclaAPI, service v1Health.Service) {

	api.HealthHealthCheckHandler = health.HealthCheckHandlerFunc(func(params health.HealthCheckParams) middleware.Responder {
		result, err := service.HealthCheck(params.HTTPRequest.Context())
		if err != nil {
			return health.NewHealthCheckBadRequest().WithPayload(errorResponse(err))
		}

		return health.NewHealthCheckOK().WithPayload(result)
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