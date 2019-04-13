package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/dchest/uniuri"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/lestrrat/go-jwx/jwk"
	"gitlab.com/gilden/fortis/authorization"
	"gitlab.com/gilden/fortis/logging"
	"gitlab.com/gilden/fortis/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// -------------------------------------
// 				Google
// -------------------------------------

var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8081/callback/google",
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	Scopes: []string{

		"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint: google.Endpoint,
}

// GoogleLoginHandler is called when the user presses the login with google button
func (server *Server) GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	logging.Debug("/auth/google called")

	session, err := server.session.Get(r, server.config.GetString("session.name"))
	if err != nil {
		logging.Debug("couldn't find existing encrypted secure cookie with name %s: %s (probably fine)", server.config.GetString("session.name"), err)
	}

	if err != nil {
		logging.Error(err)
	}

	// set the state variable in the session
	oauthStateString := uniuri.New()
	session.Values["state"] = oauthStateString
	logging.Debug("session state set to %s", session.Values["state"])

	// Store the session in the cookie
	if err := server.session.Save(r, w, session); err != nil {
		Error(w, err, "", 500, server.logger)
		return
	}

	url := googleOauthConfig.AuthCodeURL(oauthStateString)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// handle the google callback
func (server *Server) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {

	session, _ := server.session.Get(r, server.config.GetString("session.name"))

	// is the nonce "state" valid?
	queryState := r.URL.Query().Get("state")
	if session.Values["state"] != queryState {
		logging.Error("Invalid session state: stored %s, returned %s", session.Values["state"], queryState)
		http.Redirect(w, r, "/error", http.StatusUnauthorized)
		return
	}

	user, err := getGoogleUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	usr := new(models.User)

	if !server.store.UserExists(user.ID) {
		// Insert a new user
		usr.DisplayName = user.Name
		usr.ID = user.ID
		server.store.InsertUser(usr)
	}

	// retrieve the data to be shure
	usr, err = server.store.GetUserByID(user.ID)
	if err != nil {
		logging.Error(err)
	}

	// Let's create a session where we store the user id. We can ignore errors from the session store
	// as it will always return a session!
	session.Values["user"] = usr.ID

	// Store the session in the cookie
	if err := server.session.Save(r, w, session); err != nil {
		Error(w, err, "", 500, server.logger)
		return
	}

	// Redirect the user back to the consent endpoint. In a normal app, you would probably
	// add some logic here that is triggered when the user actually performs authentication and is not
	// part of the consent flow.
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

// basic method to retrieve user info using the token
func getGoogleUserInfo(state string, code string) (*authorization.TokenInfo, error) {
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}
	var user *authorization.TokenInfo
	_ = json.Unmarshal(contents, &user)
	return user, nil
}

func jsonResponse(response interface{}, w http.ResponseWriter) {

	// Create a json object from the given interface type
	json, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Json created, write success
	w.WriteHeader(http.StatusOK)
	log.Println(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

// RetrieveGoogleKeys Retieves the google public keys from the google api
func retrieveGoogleKeys(token *jwt.Token) (interface{}, error) {
	// fetch the keys and parse to a jwk
	set, err := jwk.FetchHTTP("https://www.googleapis.com/oauth2/v3/certs")
	if err != nil {
		return nil, err
	}

	// Get the key id from the header
	// This is used to determine the key to use from the jwks
	keyID, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("expecting JWT header to have string kid")
	}

	// Retrieve the acutal key
	if key := set.LookupKeyID(keyID); len(key) == 1 {
		return key[0].Materialize()
	}

	return nil, errors.New("unable to find key")
}
