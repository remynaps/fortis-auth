package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/lestrrat/go-jwx/jwk"
	"gitlab.com/gilden/fortis/authorization"
	"gitlab.com/gilden/fortis/logging"
)

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

func (server *Server) MicrosoftLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Try to parse the token
	claims := jwt.MapClaims{}
	token, err := request.ParseFromRequestWithClaims(r, request.AuthorizationHeaderExtractor,
		claims, retrieveMicrosofteKeys)

	// There should be no error if the token is parsed
	if err == nil {
		if token.Valid {

			// Sub is the unique ms id key
			// We will use it to query the user in our database
			var userID = claims["sub"].(string)
			var name = claims["name"].(string)
			var email = ""

			// Only include the mail if it has been verified
			// TODO: ask the user for email if it isn't?
			if claims["email_verified"] == true {
				email = claims["email"].(string)
			}

			tokenData := new(authorization.TokenInfo)
			tokenData.ID = userID
			tokenData.EMail = email
			tokenData.Name = name

			// retrieve the data to be shure
			usr, err := server.store.GetUserByID(userID)
			if err != nil {
				logging.Error(err)
			}

			// Let's create a session where we store the user id. We can ignore errors from the session store
			// as it will always return a session!
			// session.Values["user"] = usr.ID

			// Store the session in the cookie
			// if err := server.session.Save(r, w, session); err != nil {
			// 	Error(w, err, "", 500, server.logger)
			// 	return
			// }

			token, err := authorization.CompleteFlow(usr, server.store)
			if token != "nil" {

			}
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
}
