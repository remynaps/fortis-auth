package main

import (
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/remynaps/fortis/oauth"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/lestrrat/go-jwx/jwk"
)

const (
	privKeyPath = "./config/jwt/app.rsa"
	pubKeyPath  = "./config/jwt/app.rsa.pub"
)

var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
	googleKey string
)

type Env struct {
	db        *sql.DB
	logger    *log.Logger
	templates *template.Template
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

func main() {

	// Get the rsa keys from the file system.
	initKeys()

	// Init the http multiplexer
	r := mux.NewRouter()

	log.Println("starting server..")

	//add handlers to the multiplexer
	// ----- public handlers ------
	r.Handle("/login", http.HandlerFunc(LoginHandler))

	// ----- oauth ------
	r.Handle("/login/facebook", http.HandlerFunc(oauth.FacebookLoginHandler))
	r.Handle("/login/facebook/callback", http.HandlerFunc(oauth.HandleFacebookCallback))
	r.Handle("/login/google", http.HandlerFunc(oauth.GoogleLoginHandler))
	r.Handle("/login/microsoft", http.HandlerFunc(oauth.MicrosoftLoginHandler))
	r.Handle("/login/microsoft/callback", http.HandlerFunc(oauth.HandleMicrosoftCallback))

	// ----- protected handlers ------
	r.Handle("/status", ValidateTokenMiddleware(http.HandlerFunc(StatusHandler)))
	r.Handle("/refresh-token", ValidateTokenMiddleware(http.HandlerFunc(StatusHandler)))
	r.Handle("/validate-token", ValidateTokenMiddleware(http.HandlerFunc(StatusHandler)))
	r.Handle("/logout", ValidateTokenMiddleware(http.HandlerFunc(StatusHandler)))

	log.Println("Init complete")

	//use the default servemux(nil)
	err := http.ListenAndServe(":8081", handlers.LoggingHandler(os.Stdout, r))
	fatal(err)

}

func getKey(token *jwt.Token) (interface{}, error) {

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

func initKeys() {

	log.Println("Getting private key...")

	// Read the bytes of the private key
	signBytes, err := ioutil.ReadFile(privKeyPath)
	fatal(err)

	// get the actual key
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	fatal(err)

	log.Println("Getting public key...")

	// Read the bytes of the public key
	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	fatal(err)

	// Get the actual public key
	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	fatal(err)

	log.Println("Keys retrieved")

}

func getJson(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
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
			getKey)
		log.Println(token)

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

// LoginHandler is used to verify user login. And grant a user a token
func LoginHandler(w http.ResponseWriter, r *http.Request) {

	user := User{} //initialize empty user

	//Parse json request body and use it to set fields on user
	//Note that user is passed as a pointer variable so that it's fields can be modified
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		panic(err)
	}

	if user.Name == "palprotein" {
		// Generate the jwt
		token := jwt.New(jwt.SigningMethodRS256)

		// Add the required expiration and creation time claims to the token
		claims := make(jwt.MapClaims)
		claims["exp"] = time.Now().Add(time.Hour).Unix()
		claims["iat"] = time.Now().Unix()
		claims["name"] = "ben swolo"
		token.Claims = claims

		// Sign the token
		tokenString, err := token.SignedString(signKey)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Error while signing the token")
			fatal(err)
		}

		// Write the token to the response
		response := Token{tokenString}

		// Send json response containing the token
		JsonResponse(response, w)
	}
}
