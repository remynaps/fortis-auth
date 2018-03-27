package oauth

import (
	"encoding/json"
	"log"
	"net/http"
)

type KeysResponse struct {
	Collection []GoogleKey `json:"keys"`
}

// GoogleKey describes one of the values returned in the oauth2/v3 endpoint
type GoogleKey struct {
	Kty       string `json:"kty"`
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
	Use       string `json:"use"`
	Value     string `json:"n"`
	E         string `json:"e"`
}

// RetrieveGoogleKeys Retieves the google public keys from the google api
func RetrieveGoogleKeys() *KeysResponse {
	keys := new(KeysResponse)
	r, err := http.Get("https://www.googleapis.com/oauth2/v3/certs")
	defer r.Body.Close()

	json.NewDecoder(r.Body).Decode(keys)
	log.Println(keys)

	if err != nil {
		return nil
	}
	return keys
}

func VerifyIdToken(w http.ResponseWriter, r *http.Request) bool {
	keys := RetrieveGoogleKeys()

	if keys != nil {

	}
	return true
}
