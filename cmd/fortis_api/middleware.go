package main

// Middleware handler for methods that are protected by login
import (
	"fmt"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"gitlab.com/gilden/fortis/authorization"
	"gitlab.com/gilden/fortis/correlationID"
	"gitlab.com/gilden/fortis/logging"
)

// CorrelationIDMiddleware - generates a correlationID for the request if it was not found
var CorrelationIDMiddleware = func(f http.HandlerFunc) http.HandlerFunc {
	// one time scope setup area for middleware
	return func(w http.ResponseWriter, r *http.Request) {

		corrID := uuid.NewV4()

		ctx := correlationID.NewContext(r.Context(), corrID.String())

		f(w, r.WithContext(ctx))
	}
}

func (server *Server) ValidateClientMiddleWare(next http.Handler) Handler {
	// The top level handler
	return Handler(func(w http.ResponseWriter, r *http.Request) *RequestError {

		// Search for the client id and redirect url in the session and url
		// Needs to be saved in the session to be used later on
		clientID := r.URL.Query().Get("client_id")
		redirect := r.URL.Query().Get("redirect_url")

		session, err := server.session.Get(r, server.config.GetString("session.name"))
		if err != nil {
			logging.Warning("couldn't find existing encrypted secure cookie with name %s: %s (probably fine)", server.config.GetString("session.name"), err)
		}

		if err != nil {
			logging.Error(err)
		}

		// Try to find the values in the session
		if clientID == "" {
			sessionClientID := session.Values["client_id"]
			if sessionClientID != nil {
				clientID = sessionClientID.(string)
			}
		}
		if redirect == "" {
			sessionRedirect := session.Values["redirect"]
			if sessionRedirect != nil {
				redirect = sessionRedirect.(string)
			}
		}

		if clientID == "" {
			return &RequestError{err, 405, "No clientId supplied"}
		}
		if redirect == "" {
			return &RequestError{err, 405, "No redirect url supplied"}
		}

		// Client validation logic
		if server.store.ClientExists(clientID) {

			client, err := server.store.GetClientByID(clientID)

			// Redirect to the error page if the client does not exist
			if err != nil {
				return &RequestError{err, 405, "The client does not exist"}
			}

			if !isValueInList(redirect, client.RedirectUris) {
				return &RequestError{err, 405, "The redirect uri is not registred for this client"}
			}

			// Set the values
			session.Values["redirect"] = redirect
			session.Values["client_id"] = clientID

			// Store the session in the cookie
			if err := server.session.Save(r, w, session); err != nil {
				return &RequestError{err, 500, "Failed to save session"}

			}

			// Everything went OK! allow the request
			next.ServeHTTP(w, r)

		} else {
			return &RequestError{err, 405, "The client does not exist"}

		}

		return nil
	})
}

func ValidateTokenMiddleware(next http.Handler) http.Handler {
	// The top level handler
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to parse the token
		token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor,
			func(token *jwt.Token) (interface{}, error) {
				return authorization.VerifyKey, nil
			})

		// There should be no error if the token is parsed
		if err == nil {
			if token.Valid {

				// Token is valid. Execute the wrapped handler
				next.ServeHTTP(w, r)
			} else {

				// Notify the client about the invalid token
				fmt.Fprint(w, "Token is not valid")
			}
		} else {
			log.Println(err)
			// Client isnt authorized
			fmt.Fprint(w, "Unauthorized access to this resource")
		}
	})
}

// RequestLogMiddleWare - logs each incoming request
var RequestLogMiddleWare = func(next http.Handler) http.HandlerFunc {
	// one time scope setup area for middleware
	return func(w http.ResponseWriter, r *http.Request) {

		lrw := newLoggingResponseWriter(w)

		start := time.Now()

		next.ServeHTTP(lrw, r)

		end := time.Now()
		latency := end.Sub(start)

		requestID, typeCheck := correlationID.FromContext(r.Context())
		if !typeCheck {
			logging.Error("Request id of wrong type")
		}

		contextLogger := logging.Logger.WithFields(logrus.Fields{
			"request-id":  requestID,
			"url":         r.URL.Path,
			"method":      r.Method,
			"status-code": lrw.statusCode,
			"latency":     latency,
			"user-agent":  r.UserAgent(),
		})

		statusCode := lrw.statusCode

		if statusCode == http.StatusOK || statusCode == http.StatusTemporaryRedirect || statusCode == http.StatusFound {
			contextLogger.Info("request handled")
		} else {
			contextLogger.Error("request failed")
		}
	}
}

// Error is the expected return of a dae.Handler, or nil otherwise.
type RequestError struct {
	Error   error
	Code    int
	Message string
}

// NewError is a helper for creating an Error pointer.
func NewError(err error, code int, msg string) *RequestError {
	return &RequestError{err, code, msg}
}

// Handler is used to cast functions to its type to implement ServeHTTP.
// Code that panics is automatically recovered and delivers a server 500 error.
type Handler func(http.ResponseWriter, *http.Request) *RequestError

// ServeHTTP implements the http.Handler interface. If an appHandler returns an
// error, the error is inspected and an appropriate response is written out.
func (fn Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	requestID, typeCheck := correlationID.FromContext(r.Context())
	if !typeCheck {
		logging.Error("Request id of wrong type")
	}

	defer func() {
		if r := recover(); r != nil {

			// Construct a logger to show the request id.
			// TODO: use single logger for all situations
			criticalLogger := logging.Logger.WithFields(logrus.Fields{
				"request-id": requestID,
			})
			criticalLogger.Error(r)
			renderError(w, "A serious error has occured.", 500)
			// if Debug {
			// 	panic(r.(error))
			// }
		}
	}()

	if e := fn(w, r); e != nil {

		contextLogger := logging.Logger.WithFields(logrus.Fields{
			"request-id":  requestID,
			"url":         r.URL.Path,
			"method":      r.Method,
			"status-code": e.Code,
			"user-agent":  r.UserAgent(),
		})

		contextLogger.Error(fmt.Sprintf("Code: %v, Message: \"%s\", Error: %v", e.Code, e.Message, e.Error))
		switch e.Code {
		case 500:
			renderError(w, e.Message, e.Code)
		case 404:
			renderError(w, e.Message, e.Code)
		case 405:
			renderError(w, e.Message, e.Code)
		case 200:
			fmt.Fprint(w, e.Message)
		}
	}
}
