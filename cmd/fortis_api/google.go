package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/lestrrat/go-jwx/jwk"
)

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

func (env *Env) GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Try to parse the token
	claims := jwt.MapClaims{}
	token, err := request.ParseFromRequestWithClaims(r, request.AuthorizationHeaderExtractor, claims,
		retrieveGoogleKeys)

	// There should be no error if the token is parsed
	if err == nil {
		if token.Valid {
			// Gather information from the token

			// Sub is the unique google id key
			// We will use it to query the user in our database
			var userID = claims["sub"].(string)
			var name = "swolo"
			var email = ""

			// Only include the mail if it has been verified
			// TODO: ask the user for email if it isn't?
			if claims["email_verified"] == true {
				email = claims["email"].(string)
			}

			tokenData := new(TokenInfo)
			tokenData.ID = userID
			tokenData.EMail = email
			tokenData.Name = name

			// login or sign up
			token, err := CompleteFlow(tokenData, env.db)

			if err != nil {
				// Handler error
			}

			jsonResponse(token, w)
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