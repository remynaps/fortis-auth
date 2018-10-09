package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"gitlab.com/gilden/fortis/authorization"
	"gitlab.com/gilden/fortis/correlationID"
	"gitlab.com/gilden/fortis/logging"
	"gitlab.com/gilden/fortis/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
)

type Env struct {
	db     models.UserStore
	auth   authorization.AuthorizationService
	logger *logrus.Logger
}

// Basic user info
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type Response struct {
	Data string `json:"data"`
}

type Token struct {
	Token string `json:"token"`
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

const (
	keyPath = "./config/jwt"
)

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type ErrorData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ErrorResponseData struct {
	APIVersion string    `json:"apiVersion"`
	Context    string    `json:"context,omitempty"`
	RequestID  string    `json:"id"`
	Method     string    `json:"method"`
	Error      ErrorData `json:"error,omitempty"`
}

func main() {

	// Get the console flags
	profile := flag.String("profile", "prod", "Environment profile")
	flag.Parse()

	var logger = logrus.New()

	logger.Info("Profile selected: " + *profile)

	// Set the log format
	if *profile == "dev" {
		logger.Info("Using text log formatter..")
		logger.Info("Setting log format profile..")
		logger.Formatter = (&logrus.TextFormatter{
			TimestampFormat: "01-02-2018T15:04:05.000",
			FullTimestamp:   true,
		})
	} else {
		logger.Info("Using JSON log formatter..")
		logger.Formatter = (&logrus.JSONFormatter{})
	}

	// Get the rsa keys from the file system.

	// Init services
	auth, err := authorization.InitService(keyPath, logger)

	if err != nil {
		logger.Error("Failed to initialize the auth service." + err.Error())
	}

	// Init the http multiplexer
	r := mux.NewRouter()

	logger.Info("starting server..")

	// Set up Cors
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080"},
		AllowedHeaders:   []string{"Origin", "X-Requested-With", "Content-Type", "Authorization"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	})

	db, err := models.InitDB("postgres://gilden:koekjeszijnlekker!@192.168.1.120/fortis?sslmode=disable")
	if err != nil {
		// db connection failed. start the retry logic
		logger.Error("Failed to connect to the database." + err.Error())
	}

	// Initialize the env object and
	env := &Env{db, auth, logger}

	// ----- oauth ------
	r.Handle("/login/google", c.Handler(http.HandlerFunc(env.GoogleLoginHandler)))
	r.Handle("/login/microsoft", c.Handler(http.HandlerFunc(env.MicrosoftLoginHandler)))

	// ----- protected handlers ------
	r.Handle("/status", c.Handler(ValidateTokenMiddleware(http.HandlerFunc(StatusHandler))))
	r.Handle("/refresh-token", ValidateTokenMiddleware(http.HandlerFunc(StatusHandler)))
	r.Handle("/validate-token", ValidateTokenMiddleware(http.HandlerFunc(StatusHandler)))
	r.Handle("/logout", ValidateTokenMiddleware(http.HandlerFunc(StatusHandler)))

	logger.Info("Init complete")

	//use the default servemux(nil)
	http.ListenAndServe(":8081", RequestLogMiddleWare(logger, r))
	fatal(err)
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// RequestLogMiddleWare - logs each incoming request
var RequestLogMiddleWare = func(
	logger *logrus.Logger, next http.Handler) http.HandlerFunc {
	// one time scope setup area for middleware
	return func(w http.ResponseWriter, r *http.Request) {

		lrw := newLoggingResponseWriter(w)

		start := time.Now()

		next.ServeHTTP(lrw, r)

		end := time.Now()
		latency := end.Sub(start)

		requestID, typeCheck := correlationID.FromContext(r.Context())
		if !typeCheck {
			logger.Error("Request id of wrong type")
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

// Json wrapper function
func JsonResponse(response interface{}, w http.ResponseWriter) {

	// Create a json object from the given interface type
	json, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Json created, write success
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func Error(w http.ResponseWriter, err error, requestId string, code int, logger *logrus.Logger) {
	// Hide error from client if it is internal.
	if code == http.StatusInternalServerError {
		err = errors.New("An internal server error occured")
	}

	result := &ErrorResponseData{
		APIVersion: "beta", // change this
		Method:     "GET",  // and this
		RequestID:  requestId,
		Error: ErrorData{
			Code:    code,
			Message: err.Error(),
		},
	}

	// Write generic error response.
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(result)
}

// Simple status handler to call to validate api
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("API is up and running"))
}

// Middleware handler for methods that are protected by login
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
