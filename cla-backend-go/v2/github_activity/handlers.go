// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package github_activity

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/v32/github"

	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/restapi/operations"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/restapi/operations/github_activity"
	"github.com/communitybridge/easycla/cla-backend-go/utils"
	"github.com/go-openapi/runtime/middleware"
	"github.com/gofrs/uuid"
)

// signatureCheckMiddleware is used to get access to raw http request so can do the
// signature validation properly
func signatureCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := github.ValidatePayload(r, nil)
		if err != nil {
			http.Error(w, "signature check failure", 401)
			return
		}
		defer r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
		// call the next middleware
		next.ServeHTTP(w, r)
	})
}

// Configure setups handlers on api with service
func Configure(api *operations.EasyclaAPI, service Service) {
	api.AddMiddlewareFor("POST", "/github/activity", signatureCheckMiddleware)
	api.GithubActivityGithubActivityHandler = github_activity.GithubActivityHandlerFunc(
		func(params github_activity.GithubActivityParams) middleware.Responder {
			githubEvent := utils.GetGithubEvent(params.XGITHUBEVENT)
			if githubEvent == "" {
				return github_activity.NewGithubActivityBadRequest()
			}

			// we need the raw payload so we can use the github utilities
			payload, err := params.GithubActivityInput.MarshalJSON()
			if err != nil {
				return github_activity.NewGithubActivityBadRequest().WithPayload(&models.ErrorResponse{
					Code:    "400",
					Message: "json marshall",
				})
			}

			event, err := github.ParseWebHook(githubEvent, payload)
			if err != nil {
				return github_activity.NewGithubActivityBadRequest().WithPayload(&models.ErrorResponse{
					Code:    "400",
					Message: fmt.Sprintf("parsing event failed : %v", err),
				})
			}

			var processError error
			switch event := event.(type) {
			case *github.InstallationRepositoriesEvent:
				processError = service.ProcessInstallationRepositoriesEvent(event)
			case *github.RepositoryEvent:
				processError = service.ProcessRepositoryEvent(event)
			default:
				// TODO: this will be removed when switched to real github activity
				return github_activity.NewGithubActivityInternalServerError().WithPayload(&models.ErrorResponse{
					Code:    "500",
					Message: fmt.Sprintf("unsupported event : %s", githubEvent),
				})
			}

			if processError != nil {
				return github_activity.NewGithubActivityInternalServerError().WithPayload(&models.ErrorResponse{
					Code:    "500",
					Message: processError.Error(),
				})
			}

			return github_activity.NewGithubActivityOK()
		})
}

type codedResponse interface {
	Code() string
}

func errorResponse(reqID string, err error) *models.ErrorResponse {
	if reqID == "" {
		requestID, _ := uuid.NewV4()
		reqID = requestID.String()
	}
	code := ""
	if e, ok := err.(codedResponse); ok {
		code = e.Code()
	}

	e := models.ErrorResponse{
		Code:       code,
		Message:    err.Error(),
		XRequestID: reqID,
	}

	return &e
}
