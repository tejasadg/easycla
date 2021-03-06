// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package template

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/LF-Engineering/lfx-kit/auth"
	"github.com/communitybridge/easycla/cla-backend-go/events"
	v1Events "github.com/communitybridge/easycla/cla-backend-go/events"
	v1Models "github.com/communitybridge/easycla/cla-backend-go/gen/models"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/restapi/operations"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/restapi/operations/template"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	v1Template "github.com/communitybridge/easycla/cla-backend-go/template"
	"github.com/communitybridge/easycla/cla-backend-go/utils"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/jinzhu/copier"
)

// Configure API call
func Configure(api *operations.EasyclaAPI, service v1Template.Service, eventsService v1Events.Service) {
	// Retrieve a list of available templates
	api.TemplateGetTemplatesHandler = template.GetTemplatesHandlerFunc(func(params template.GetTemplatesParams, user *auth.User) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		utils.SetAuthUserProperties(user, params.XUSERNAME, params.XEMAIL)
		f := logrus.Fields{
			"functionName":   "TemplateGetTemplatesHandler",
			utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		}

		templates, err := service.GetTemplates(params.HTTPRequest.Context())
		if err != nil {
			log.WithFields(f).WithError(err).Warn("problem loading templates")
			return template.NewGetTemplatesBadRequest().WithPayload(errorResponse(reqID, err))
		}
		var response []models.Template
		err = copier.Copy(&response, templates)
		if err != nil {
			log.WithFields(f).WithError(err).Warn("problem converting templates")
			return template.NewGetTemplatesInternalServerError().WithPayload(errorResponse(reqID, err))
		}
		return template.NewGetTemplatesOK().WithPayload(response)
	})

	api.TemplateCreateCLAGroupTemplateHandler = template.CreateCLAGroupTemplateHandlerFunc(func(params template.CreateCLAGroupTemplateParams, user *auth.User) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		utils.SetAuthUserProperties(user, params.XUSERNAME, params.XEMAIL)
		f := logrus.Fields{
			"functionName":   "TemplateCreateCLAGroupTemplateHandler",
			utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
			"claGroupID":     params.ClaGroupID,
		}

		input := &v1Models.CreateClaGroupTemplate{}
		err := copier.Copy(input, &params.Body)
		if err != nil {
			log.WithFields(f).WithError(err).Warn("problem converting templates")
			return template.NewGetTemplatesInternalServerError().WithPayload(errorResponse(reqID, err))
		}
		pdfUrls, err := service.CreateCLAGroupTemplate(params.HTTPRequest.Context(), params.ClaGroupID, input)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("Error generating PDFs from provided templates, error: %v", err)
			return template.NewGetTemplatesBadRequest().WithPayload(errorResponse(reqID, err))
		}

		eventsService.LogEvent(&events.LogEventArgs{
			EventType:  events.CLATemplateCreated,
			ProjectID:  params.ClaGroupID,
			LfUsername: user.UserName,
			EventData:  &events.CLATemplateCreatedEventData{},
		})

		response := &models.TemplatePdfs{}
		err = copier.Copy(response, pdfUrls)
		if err != nil {
			log.WithFields(f).WithError(err).Warn("problem converting templates")
			return template.NewGetTemplatesInternalServerError().WithPayload(errorResponse(reqID, err))
		}

		return template.NewCreateCLAGroupTemplateOK().WithPayload(response)
	})

	api.TemplateTemplatePreviewHandler = template.TemplatePreviewHandlerFunc(func(params template.TemplatePreviewParams, user *auth.User) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		utils.SetAuthUserProperties(user, params.XUSERNAME, params.XEMAIL)
		f := logrus.Fields{
			"functionName":   "TemplateTemplatePreviewHandler",
			utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
			"templateFor":    params.TemplateFor,
		}

		var param v1Models.CreateClaGroupTemplate
		err := copier.Copy(&param, &params.TemplatePreviewInput)
		if err != nil {
			log.WithFields(f).WithError(err).Warn("problem converting templates")
			return writeResponse(http.StatusInternalServerError, runtime.JSONMime, runtime.JSONProducer(), reqID, errorResponse(reqID, err))
		}
		pdf, err := service.CreateTemplatePreview(&param, params.TemplateFor)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("Error generating PDFs from provided templates, error: %v", err)
			return writeResponse(http.StatusBadRequest, runtime.JSONMime, runtime.JSONProducer(), reqID, errorResponse(reqID, err))
		}
		return middleware.ResponderFunc(func(rw http.ResponseWriter, pr runtime.Producer) {
			rw.WriteHeader(http.StatusOK)
			_, err := rw.Write(pdf)
			if err != nil {
				log.Warnf("Error writing pdf, error: %v", err)
			}
		})
	})

	api.TemplateGetCLATemplatePreviewHandler = template.GetCLATemplatePreviewHandlerFunc(func(params template.GetCLATemplatePreviewParams) middleware.Responder {
		reqID := utils.GetRequestID(params.XREQUESTID)
		ctx := context.WithValue(context.Background(), utils.XREQUESTID, reqID) // nolint
		f := logrus.Fields{
			"functionName":   "TemplateGetCLATemplatePreviewHandler",
			utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		}
		pdf, err := service.GetCLATemplatePreview(params.HTTPRequest.Context(), params.ClaGroupID, params.ClaType, *params.Watermark)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("Error getting PDFs for provided cla group ID : %s, error: %v", params.ClaGroupID, err)
			return writeResponse(http.StatusBadRequest, runtime.JSONMime, runtime.JSONProducer(), reqID, errorResponse(reqID, err))
		}

		return middleware.ResponderFunc(func(rw http.ResponseWriter, pr runtime.Producer) {
			rw.WriteHeader(http.StatusOK)
			_, err := rw.Write(pdf)
			if err != nil {
				log.WithFields(f).WithError(err).Warnf("Error writing pdf, error: %v", err)
			}
		})
	})
}

type codedResponse interface {
	Code() string
}

func errorResponse(reqID string, err error) *models.ErrorResponse {
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

func writeResponse(httpStatus int, contentType string, contentProducer runtime.Producer, reqID string, data interface{}) middleware.Responder {
	return middleware.ResponderFunc(func(rw http.ResponseWriter, pr runtime.Producer) {
		rw.Header().Set(utils.XREQUESTID, reqID)
		rw.Header().Set(runtime.HeaderContentType, contentType)
		rw.WriteHeader(httpStatus)
		err := contentProducer.Produce(rw, data)
		if err != nil {
			log.Warnf("failed to write data. error = %v", err)
		}
	})
}
