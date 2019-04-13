package main

import (
	"html/template"
	"net/http"

	"gitlab.com/gilden/fortis/logging"
)

type mainTemplate struct {
	Hero string
}

type consentTemplate struct {
	Hero string
}

// authenticated checks if our cookie store has a user stored and returns the
// user's name, or an empty string if the user is not yet authenticated.
func (server *Server) authenticated(r *http.Request) string {
	session, _ := server.session.Get(r, sessionName)
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

func (server *Server) fileHandler(w http.ResponseWriter, r *http.Request) {

	// TODO: check if user has authenticated befire. probably other middleware

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

	if clientID == "" {
		logging.Error("No ClientID supplied")
		renderError(w, "No ClientID supplied")
		return
	}
	if redirect == "" {
		logging.Error("No redirect url supplied")
		renderError(w, "No redirect url supplied")
		return
	}

	// Client validation logic
	if server.store.ClientExists(clientID) {

		client, err := server.store.GetClientByID(clientID)

		// Redirect to the error page if the client does not exist
		if err != nil {
			logging.Info("The client does not exist")
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

	} else {
		renderError(w, "The client does not exist")
		return
	}

	t := template.Must(template.New("login.html").ParseFiles("./templates/login.html")) // Create a template.

	template := new(consentTemplate)

	template.Hero = "This is where the fun begins"

	t.Execute(w, template) // merge.
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

	t := template.Must(template.New("error.html").ParseFiles("./templates/error.html")) // Create a template.

	template := new(mainTemplate)

	template.Hero = "This is where the fun begins"

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
