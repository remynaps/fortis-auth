package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gitlab.com/gilden/fortis/models"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type Server struct {
	config  *viper.Viper
	logger  *logrus.Logger
	server  *http.Server
	session *sessions.CookieStore
	store   *models.DB
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

type ErrorData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

const (
	keyPath = "/etc/keys"
	Timeout = 5 * time.Second
)

type ErrorResponseData struct {
	APIVersion string    `json:"apiVersion"`
	Context    string    `json:"context,omitempty"`
	RequestID  string    `json:"id"`
	Method     string    `json:"method"`
	Error      ErrorData `json:"error,omitempty"`
}

// NewServer returns a new instance of a Server configured with the provided
// configuration
func NewServer(config *viper.Viper, db *models.DB) (*Server, error) {

	hostAddress := config.GetString("server.host_address")
	hostPort := config.GetString("server.host_port")

	addr := net.JoinHostPort(hostAddress, hostPort)

	defaultServer := &http.Server{
		Addr:         addr,
		ReadTimeout:  Timeout,
		WriteTimeout: Timeout,
	}

	ws := &Server{
		config:  config,
		server:  defaultServer,
		session: sessions.NewCookieStore([]byte("wtf")),
		store:   db,
	}
	ws.registerRoutes()
	return ws, nil
}

// Start starts the underlying HTTP server
func (ws *Server) Start() error {
	return ws.server.ListenAndServe()
}

// Shutdown attempts to gracefully shutdown the underlying HTTP server.
func (ws *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	return ws.server.Shutdown(ctx)
}

func (ws *Server) registerRoutes() {
	router := mux.NewRouter()
	// Unauthenticated handlers for registering a new credential and logging in.

	// Get the running directory.
	runningDirectory, err := os.Getwd()
	if err != nil {
		// Handle the error.
		panic(err)
	}

	// Index route
	router.Handle("/", (http.HandlerFunc(fileHandler)))
	router.Handle("/consent", (http.HandlerFunc(ws.consentFileHandler)))
	router.Handle("/logout", http.HandlerFunc(ws.logoutHandler))
	router.Handle("/loggedout", http.HandlerFunc(ws.loggedOutFileHandler))

	// ----- oauth ------
	router.Handle("/auth/google", http.HandlerFunc(ws.GoogleLoginHandler))
	router.Handle("/auth/microsoft", http.HandlerFunc(ws.MicrosoftLoginHandler))

	router.Handle("/callback/google", http.HandlerFunc(ws.handleGoogleCallback))
	// router.Handle("/callback/microsoft", http.HandlerFunc(ws.handleMicrosoftCallback))

	// ----- protected handlers ------
	router.Handle("/status", RequestLogMiddleWare(http.HandlerFunc(StatusHandler)))
	router.Handle("/refresh-token", ValidateTokenMiddleware(http.HandlerFunc(StatusHandler)))
	router.Handle("/validate-token", ValidateTokenMiddleware(http.HandlerFunc(StatusHandler)))

	// Static file serving
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(runningDirectory+"/static"))))

	ws.server.Handler = CorrelationIDMiddleware(RequestLogMiddleWare(router))

}
