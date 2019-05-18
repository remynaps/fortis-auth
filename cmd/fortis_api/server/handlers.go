package server

import (
	"encoding/base64"
	"errors"
	"net/http"

	"gitlab.com/gilden/fortis/authorization"
	"gitlab.com/gilden/fortis/correlationID"
	"gitlab.com/gilden/fortis/logging"
	"golang.org/x/crypto/bcrypt"
)

// authenticated checks if our cookie store has a user stored and returns the
// user's name, or an empty string if the user is not yet authenticated.
func (server *Server) authenticated(r *http.Request) string {
	session, _ := server.session.Get(r, server.config.Server.SessionName)
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

		session, _ := server.session.Get(r, server.config.Server.SessionName)

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
