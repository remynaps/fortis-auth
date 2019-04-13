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

func (server *Server) ValidateClientMiddleWare(next http.Handler) http.Handler {
	// The top level handler
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
			renderError(w, "No ClientID supplied")
			return
		}
		if redirect == "" {
			renderError(w, "No redirect url supplied")
			return
		}

		// Client validation logic
		if server.store.ClientExists(clientID) {

			client, err := server.store.GetClientByID(clientID)

			// Redirect to the error page if the client does not exist
			if err != nil {
				logging.Error("The client does not exist")
				renderError(w, "The client does not exist")
				return
			}

			if !isValueInList(redirect, client.RedirectUris) {
				logging.Error("The redirect uri is not registred for this client")
				renderError(w, "The redirect uri is not registred for this client")
				return
			}

			// Set the values
			session.Values["redirect"] = redirect
			session.Values["client_id"] = clientID

			// Store the session in the cookie
			if err := server.session.Save(r, w, session); err != nil {
				renderError(w, "The redirect uri is not registred for this client")
				Error(w, err, "", 500, server.logger)
				return
			}

			// Everything went OK! allow the request
			next.ServeHTTP(w, r)

		} else {
			renderError(w, "The client does not exist")
			return
		}
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
