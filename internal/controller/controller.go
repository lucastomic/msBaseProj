package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/lucastomic/msBaseProj/internal/controller/apitypes"
	"github.com/lucastomic/msBaseProj/internal/errs"
	"github.com/lucastomic/msBaseProj/internal/logging"
	"github.com/lucastomic/msBaseProj/internal/translator"
)

// Controller defines the interface for an HTTP controller.
// It outlines the necessary functionality for routing HTTP requests to their respective handlers.
type Controller interface {
	// Router returns a Router which maps HTTP paths to API functions.
	// This API functions act like the handler for every route.
	Router() apitypes.Router
}

// CommonController provides shared functionalities for handling HTTP requests across different controllers.
type CommonController struct {
	logger logging.Logger
}

// ParseError takes an error and maps it to an HTTP response.
// It uses the mapDomainErrorToHTTP function to convert domain-specific errors to HTTP errors,
// encapsulating the process of translating backend errors into user-friendly HTTP responses.
// It's important to take into account that if an error is not already defined into the domain-specific errors,
// it will map the error as a internalServerError
func (c *CommonController) ParseError(
	ctx context.Context,
	r *http.Request,
	w http.ResponseWriter,
	err error,
) apitypes.Response {
	if errors.Is(err, errs.ErrinternalError) {
		c.logger.Error(ctx, "internal error: %v", err)
	}
	httpErr := mapDomainErrorToHTTP(r.Context(), err)
	return apitypes.Response{
		Status:  httpErr.Code,
		Content: map[string]any{"error": httpErr.Error()},
		Headers: map[string]string{"Content-Type": "application/json"},
	}
}

// ReadIDFromPath reads the "id" path variable. For example, for /boat/{id}/book
// and the concret url /boat/12/book ReadIDFromPath(r) would return 12.
// This metod was thought only to be used to retrieve uint ids.
// In case of no id var, it will throw an error. Also if the id isn't an uint.
func (b CommonController) ReadIDFromPath(r *http.Request) (uint, error) {
	idString := r.PathValue("id")
	if idString == "" {
		return 0, errors.New("no id provided at path")
	}
	id, err := strconv.ParseUint(idString, 10, 32)
	if err != nil {
		return 0, errors.New("id is not an unsigned integer")
	}
	return uint(id), nil
}

// mapDomainErrorToHTTP converts domain-specific errors to errs.HTTPError instances.
// It checks for specific known errors and maps them to appropriate HTTP status codes and messages.
// For unrecognized errors, it defaults to returning an "internal error" with a 500 status code.
func mapDomainErrorToHTTP(ctx context.Context, err error) errs.HTTPError {
	i18n := &errs.I18nError{}
	var message string
	if errors.As(err, i18n) {
		message = translator.TranslateGivenCtx(ctx, i18n.Code)
		err = i18n.Unwrap()
	} else {
		message = err.Error()
	}

	switch {
	case errors.Is(err, errs.ErrInvalidInput):
		return *errs.NewHTTPError(http.StatusBadRequest, message)
	case errors.Is(err, errs.ErrinternalError):
		return *errs.NewHTTPError(http.StatusInternalServerError, message)
	case errors.Is(err, errs.ErrNotFound):
		return *errs.NewHTTPError(http.StatusNotFound, message)
	case errors.Is(err, errs.ErrNotAuthorized):
		return *errs.NewHTTPError(http.StatusUnauthorized, message)
	case errors.Is(err, errs.ErrConflict):
		return *errs.NewHTTPError(http.StatusConflict, message)
	default:
		return *errs.NewHTTPError(http.StatusInternalServerError, translator.TranslateGivenCtx(ctx, "internalerror"))
	}
}
