package server

import (
	"encoding/base64"
	"errors"
	"html/template"
	"net/http"

	"gitlab.com/gilden/fortis/authorization"
	"gitlab.com/gilden/fortis/correlationID"
	"gitlab.com/gilden/fortis/logging"
	"golang.org/x/crypto/bcrypt"
)

type mainTemplate struct {
	Hero string
}

type errorTemplate struct {
	Error        string
	ErrorMessage string
	ErrorHint    string
}

type consentTemplate struct {
	Hero string
}

func (server *Server) exchangeCode(w http.ResponseWriter, r *http.Request) {

	// 0. User needs to be logged in! inspect session!!!
	// 1. parse data from url (client_id, client_secret, code, redirect_url)
	// 2. Try to find client using url data
	// 2a. validate code
	// 3. if found and valid, retrieve user info.
	// 4. Generate token if everything checks out

	requestID, typeCheck := correlationID.FromContext(r.Context())
	if !typeCheck {
		logging.Error("Request id of wrong type")
	}

	// This helper checks if the user is already authenticated. If not, we
	// redirect them to the login endpoint.
	user := server.authenticated(r)
	if user == "" {

		// Error response
		Error(w, errors.New("Unauthorized"), requestID, 405, logging.Logger)
		return
	}

	// 1. Retieve required info from url
	clientID := r.URL.Query().Get("client_id")
	clientSecret := r.URL.Query().Get("client_secret")
	code := r.URL.Query().Get("code")
	redirect := r.URL.Query().Get("redirect_url")
	state := r.URL.Query().Get("state")

	// TODO: seperate checks
	if clientID == "" {
		Error(w, errors.New("No clientID supplied"), requestID, 405, logging.Logger)
		return
	}
	if clientSecret == "" {
		Error(w, errors.New("No clientSecret supplied"), requestID, 405, logging.Logger)
		return
	}
	if code == "" {
		Error(w, errors.New("No auth code supplied"), requestID, 405, logging.Logger)
		return
	}
	if redirect == "" {
		Error(w, errors.New("No redirect url supplied"), requestID, 405, logging.Logger)
		return
	}
	if state == "" {
		Error(w, errors.New("No state supplied"), requestID, 405, logging.Logger)
		return
	}

	// Client validation logic
	if server.store.ClientExists(clientID) {

		client, err := server.store.GetClientByID(clientID)

		// Redirect to the error page if the client does not exist
		if err != nil {
			Error(w, errors.New("Failed to retrieve the client"), requestID, 500, logging.Logger)
		}

		if !isValueInList(redirect, client.RedirectUris) {
			Error(w, errors.New("Invalid redirect url"), requestID, 400, logging.Logger)
		}

		session, _ := server.session.Get(r, server.config.GetString("session.name"))

		// Check if the supplied redirect url equals the url supplied in the first call
		existingRedirect := session.Values["redirect"].(string)
		if existingRedirect != redirect {
			Error(w, errors.New("Invalid redirect url"), requestID, 400, logging.Logger)
		}

		decodedSecret, err := base64.URLEncoding.DecodeString(clientSecret)
		if err != nil {
			Error(w, errors.New("The client secret does not have the correct format"), requestID, 400, logging.Logger)
			return
		}

		// Compare the secret with the stored one
		err = bcrypt.CompareHashAndPassword([]byte(client.ClientSecret), decodedSecret)
		if err != nil {
			logging.Error(err)
			Error(w, errors.New("Invalid client secret supplied"), requestID, 400, logging.Logger)
			return
		}

		// At this point we assume the user is authenticated
		userID := session.Values["user"].(string)

		// retrieve the data to be shure
		usr, err := server.store.GetUserByID(userID)
		if err != nil {
			Error(w, err, requestID, 500, logging.Logger)
		}

		// Finally, generate the jwt
		token, err := authorization.CompleteFlow(usr, server.store)

		jsonToken := Token{
			Token: token,
		}

		if err != nil {
			Error(w, err, requestID, 500, logging.Logger)
		}

		JsonResponse(jsonToken, w)
	}
}

// authenticated checks if our cookie store has a user stored and returns the
// user's name, or an empty string if the user is not yet authenticated.
func (server *Server) authenticated(r *http.Request) string {
	session, _ := server.session.Get(r, server.config.GetString("session.name"))
	if u, ok := session.Values["user"]; !ok {
		return ""
	} else if user, ok := u.(string); !ok {
		return ""
	} else {
		return user
	}
}

// Simple status handler to call to validate api
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("API is up and running"))
}

func (server *Server) fileHandler(w http.ResponseWriter, r *http.Request) *RequestError {

	// TODO: check if user has authenticated befire. probably other middleware

	// Search for the client id and redirect url in the session and url
	// Needs to be saved in the session to be used later on
	clientID := r.URL.Query().Get("client_id")
	redirect := r.URL.Query().Get("redirect_url")
	state := r.URL.Query().Get("state")

	session, err := server.session.Get(r, server.config.GetString("session.name"))
	if err != nil {
		logging.Warning("couldn't find existing encrypted secure cookie with name %s: %s (probably fine)", server.config.GetString("session.name"), err)
	}

	if err != nil {
		logging.Error(err)
	}

	if clientID == "" {
		return &RequestError{err, 405, "No ClientID supplied"}
	}
	if redirect == "" {
		return &RequestError{err, 405, "No redirect url supplied"}
	}
	if state == "" {
		return &RequestError{err, 405, "No state supplied"}
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
		session.Values["state"] = clientID

		// Store the session in the cookie
		if err := server.session.Save(r, w, session); err != nil {
			return &RequestError{err, 500, "Failed to save session"}
		}

	} else {
		return &RequestError{err, 405, "The client does not exist"}
	}

	t := template.Must(template.New("login.html").ParseFiles("./templates/login.html")) // Create a template.

	template := new(consentTemplate)

	template.Hero = "This is where the fun begins"

	t.Execute(w, template) // merge.

	return nil
}

func (server *Server) consentFileHandler(w http.ResponseWriter, r *http.Request) {

	// This helper checks if the user is already authenticated. If not, we
	// redirect them to the login endpoint.
	user := server.authenticated(r)
	if user == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	t := template.Must(template.New("consent.html").ParseFiles("./templates/consent.html")) // Create a template.

	template := new(mainTemplate)

	template.Hero = "This is where the fun begins"

	t.Execute(w, template) // merge.
}

func (server *Server) loggedOutFileHandler(w http.ResponseWriter, r *http.Request) {

	t := template.Must(template.New("logout.html").ParseFiles("./templates/logout.html")) // Create a template.

	template := new(mainTemplate)

	template.Hero = "This is where the fun begins"

	t.Execute(w, template) // merge.
}

func (server *Server) errorFileHandler(w http.ResponseWriter, r *http.Request) {

	error := r.URL.Query().Get("Error")
	errorMessage := r.URL.Query().Get("Error_description")
	errorHint := r.URL.Query().Get("Error_hint")

	t := template.Must(template.New("error.html").ParseFiles("./templates/error.html")) // Create a template.

	template := new(errorTemplate)

	template.Error = error
	template.ErrorMessage = errorMessage
	template.ErrorHint = errorHint

	t.Execute(w, template) // merge.
}

// LogoutHandler /logout
// currently performs a 302 redirect to Google
func (server *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	logging.Debug("/logout")

	logging.Debug("saving session")
	server.session.MaxAge(-1)
	session, err := server.session.Get(r, server.config.GetString("session.name"))
	if err != nil {
		logging.Error(err)
	}
	session.Save(r, w)
	server.session.MaxAge(300)

	var requestedURL = r.URL.Query().Get("url")
	if requestedURL != "" {
		http.Redirect(w, r, requestedURL, http.StatusFound)
	} else {
		http.Redirect(w, r, "/loggedout", http.StatusFound)
	}
}
