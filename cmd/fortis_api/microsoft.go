package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/dchest/uniuri"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/lestrrat/go-jwx/jwk"
	"gitlab.com/gilden/fortis/authorization"
	"gitlab.com/gilden/fortis/logging"
	"gitlab.com/gilden/fortis/models"
	"golang.org/x/oauth2"
)

var microsoftOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8081/callback/microsoft",
	ClientSecret: os.Getenv("MICROSOFT_CLIENT_SECRET"),
	Scopes:       []string{"user.read"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
		TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
	},
}

// the main login handler for microsoft oauth
func (server *Server) MicrosoftLoginHandler(w http.ResponseWriter, r *http.Request) *RequestError {
	session, err := server.session.Get(r, server.config.GetString("session.name"))
	if err != nil {
		logging.Debug("couldn't find existing encrypted secure cookie with name %s: %s (probably fine)", server.config.GetString("session.name"), err)
	}

	// set the state variable in the session
	oauthStateString := uniuri.New()
	session.Values["state"] = oauthStateString
	logging.Debug("session state set to %s", session.Values["state"])

	// Store the session in the cookie
	if err := server.session.Save(r, w, session); err != nil {
		return &RequestError{err, 500, "Can't display record"}
	}

	url := microsoftOauthConfig.AuthCodeURL(oauthStateString)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)

	return nil
}

// handle the microsoft callback
func (server *Server) handleMicrosoftCallback(w http.ResponseWriter, r *http.Request) *RequestError {
	session, _ := server.session.Get(r, server.config.GetString("session.name"))

	// is the nonce "state" valid?
	queryState := r.URL.Query().Get("state")
	if session.Values["state"] != queryState {
		logging.Error("Invalid session state: stored %s, returned %s", session.Values["state"], queryState)
		return &RequestError{errors.New("Invalid session state"), 405, "Can't display record"}
	}

	user, err := getMicrosoftUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusFound)
		return &RequestError{err, 405, "Code exchange failed"}
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
		return &RequestError{err, 500, "Failed to retrieve user"}
	}

	// Let's create a session where we store the user id. We can ignore errors from the session store
	// as it will always return a session!
	session.Values["user"] = usr.ID

	// Store the session in the cookie
	if err := server.session.Save(r, w, session); err != nil {
		Error(w, err, "", 500, server.logger)
		return &RequestError{err, 500, "Failed to save cookie"}
	}

	// Finally, generate the jwt
	token, err := authorization.CompleteFlow(usr, server.store)

	if err != nil {
		return &RequestError{err, 500, "Failed to create token"}
	}

	redirectUrl := session.Values["redirect"]

	if redirectUrl == "" {
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		http.Redirect(w, r, redirectUrl.(string)+"?token="+token, http.StatusFound)
	}
	return nil
}

// get basic user info from microsoft
func getMicrosoftUserInfo(state string, code string) (*authorization.TokenInfo, error) {
	token, err := microsoftOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	response, err := http.Get("https://login.microsoftonline.com/common/v2.0/openid/userinfo?access_token=" + token.AccessToken)
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

// RetrieveGoogleKeys Retieves the google public keys from the google api
func retrieveMicrosofteKeys(token *jwt.Token) (interface{}, error) {
	// fetch the keys and parse to a jwk
	set, err := jwk.FetchHTTP("https://login.microsoftonline.com/common/discovery/v2.0/keys")
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

// func (server *Server) MicrosoftLoginHandler(w http.ResponseWriter, r *http.Request) {
// 	// Try to parse the token
// 	claims := jwt.MapClaims{}
// 	token, err := request.ParseFromRequestWithClaims(r, request.AuthorizationHeaderExtractor,
// 		claims, retrieveMicrosofteKeys)

// 	// There should be no error if the token is parsed
// 	if err == nil {
// 		if token.Valid {

// 			// Sub is the unique ms id key
// 			// We will use it to query the user in our database
// 			var userID = claims["sub"].(string)
// 			var name = claims["name"].(string)
// 			var email = ""

// 			// Only include the mail if it has been verified
// 			// TODO: ask the user for email if it isn't?
// 			if claims["email_verified"] == true {
// 				email = claims["email"].(string)
// 			}

// 			tokenData := new(authorization.TokenInfo)
// 			tokenData.ID = userID
// 			tokenData.EMail = email
// 			tokenData.Name = name

// 			// retrieve the data to be shure
// 			usr, err := server.store.GetUserByID(userID)
// 			if err != nil {
// 				logging.Error(err)
// 			}

// 			// Let's create a session where we store the user id. We can ignore errors from the session store
// 			// as it will always return a session!
// 			// session.Values["user"] = usr.ID

// 			// Store the session in the cookie
// 			// if err := server.session.Save(r, w, session); err != nil {
// 			// 	Error(w, err, "", 500, server.logger)
// 			// 	return
// 			// }

// 			token, err := authorization.CompleteFlow(usr, server.store)
// 			if token != "nil" {

// 			}
// 		} else {

// 			// Notify the client about the invalid token
// 			w.WriteHeader(http.StatusUnauthorized)
// 			fmt.Fprint(w, "Token is not valid")
// 		}
// 	} else {
// 		log.Println(err)
// 		// Client isnt authorized
// 		w.WriteHeader(http.StatusUnauthorized)
// 		fmt.Fprint(w, "Unauthorized access to this resource")
// 	}
// }
