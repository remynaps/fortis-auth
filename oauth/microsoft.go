package oauth

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/lestrrat/go-jwx/jwk"
	"github.com/remynaps/fortis/authorization"
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

func MicrosoftLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Try to parse the token
	token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor,
		retrieveMicrosofteKeys)

	// There should be no error if the token is parsed
	if err == nil {
		if token.Valid {
			// login or sign up
			jsonResponse(authorization.CreateToken(), w)
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
