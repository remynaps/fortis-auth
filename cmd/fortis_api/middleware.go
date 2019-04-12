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
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, "Token is not valid")
			}
		} else {
			log.Println(err)
			// Client isnt authorized
			w.WriteHeader(http.StatusUnauthorized)
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

		if statusCode == http.StatusOK {
			contextLogger.Info("request handled")
		} else {
			contextLogger.Error("request failed")
		}
	}
}