package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/cors"

	"github.com/lucastomic/msBaseProj/internal/controller"
	"github.com/lucastomic/msBaseProj/internal/controller/apitypes"
	"github.com/lucastomic/msBaseProj/internal/errs"
	"github.com/lucastomic/msBaseProj/internal/logging"
	"github.com/lucastomic/msBaseProj/internal/middleware"
	"github.com/lucastomic/msBaseProj/internal/translator"
)

// Server represents the core structure of an HTTP server. It encapsulates all necessary components
// for server operation, including routing, logging, and middleware management.
type Server struct {
	listenAddr     string                  // The address on which the server listens for incoming requests.
	controller     []controller.Controller // Controller manages routing of requests to their respective handlers.
	logger         logging.Logger          // logicLogger is specialized for logging business logic related events.
	middlewares    []middleware.Middleware // middlewares is a slice of Middleware interfaces to be applied to all requests.
	authMiddleware middleware.Middleware   // authMiddleware is the middleware for those routes who requires authentication
	allowOrigins   []string                // allowOrigins is a list of origins that are allowed to make requests to the server
}

// New creates a new instance of the Server struct, initializing it with the provided parameters
// such as listen address, controller, API and logic loggers, and middlewares.
func New(
	listenAddr string,
	controller []controller.Controller,
	logger logging.Logger,
	middlewares []middleware.Middleware,
	authMiddleware middleware.Middleware,
	allowOrigins []string,
) Server {
	return Server{
		listenAddr,
		controller,
		logger,
		middlewares,
		authMiddleware,
		allowOrigins,
	}
}

// Run initializes the server's routes based on the controller's router, applies middlewares,
// starts listening on the specified address, and logs the server's start or any errors encountered.
func (s *Server) Run() {
	r := http.NewServeMux()
	for _, controller := range s.controller {
		for _, route := range controller.Router() {
			middlewares := s.middlewares
			if route.RequireAuth {
				middlewares = append(middlewares, s.authMiddleware)
			}
			handlerWithMiddlewares := middleware.ChainMiddleware(
				s.makeHTTPHandlerFunc(route.Handler),
				s.handleError,
				middlewares...,
			)
			r.Handle(fmt.Sprintf("%s /api%s", route.Method, route.Path), handlerWithMiddlewares)
		}
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   s.allowOrigins,
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Credentials"},
	})
	handler := c.Handler(r)

	s.logger.Info(context.Background(), "Service running in %s", s.listenAddr)
	if err := http.ListenAndServe(s.listenAddr, handler); err != nil {
		s.logger.Error(context.Background(), "Failed to start server: %v", err)
	}
}

// makeHTTPHandlerFunc wraps the API function into an http.HandlerFunc, facilitating the handling
// of HTTP requests and responses within the server's routing mechanism.
func (s *Server) makeHTTPHandlerFunc(apiFn apitypes.APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := apiFn(w, r)
		s.writeResponse(r, w, res)
	}
}

// handleError handles errors by writing an error message as a JSON response, utilizing the
// writeJSON method to ensure the response format is consistent.
func (s *Server) handleError(
	req *http.Request,
	w http.ResponseWriter,
	err error,
	statusCode int,
) {
	i18n := &errs.I18nError{}
	var message string
	if errors.As(err, i18n) {
		message = translator.TranslateGivenCtx(req.Context(), i18n.Code)
	} else {
		message = err.Error()
	}
	errMsg := map[string]string{"error": message}
	s.writeResponse(req, w, apitypes.Response{Status: statusCode, Content: errMsg})
}

// writeResponse prepares and sends an HTTP response based on the provided apitypes.Response struct.
// It sets custom headers, writes the status code, and sends the response content, which can vary in type.
func (s *Server) writeResponse(req *http.Request, w http.ResponseWriter, res apitypes.Response) {
	setCustomHeaders(w, res.Headers)
	w.WriteHeader(res.Status)
	if err := writeContent(w, res.Content); err != nil {
		s.logger.Error(req.Context(), "Failed to write response: %v", err)
	}
}

// writeContent writes the content as a JSON to the response writer.
// Returns an error if it encounters an issue during the write operation.
func writeContent(w http.ResponseWriter, content interface{}) error {
	return json.NewEncoder(w).Encode(content)
}

// setCustomHeaders sets the headers provided in the response struct.
func setCustomHeaders(w http.ResponseWriter, headers map[string]string) {
	for key, value := range headers {
		w.Header().Set(key, value)
	}
}
