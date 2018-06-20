package authorization

import (
	"crypto/rsa"
	"io/ioutil"
	"log"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	VerifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

// InitKeys
func Init(path string) {

	privKeyPath := path + "/jwt-privatekey"
	pubKeyPath := path + "/jwt-publickey"

	log.Println("Getting private key...")

	// Read the bytes of the private key
	signBytes, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		log.Fatal(err)
	}
	// get the actual key
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Getting public key...")

	// Read the bytes of the public key
	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		log.Fatal(err)
	}
	// Get the actual public key
	VerifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Keys retrieved")

}
