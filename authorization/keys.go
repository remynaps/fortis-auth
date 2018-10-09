package authorization

import (
	"crypto/rsa"
	"io/ioutil"

	"github.com/sirupsen/logrus"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	VerifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

// InitKeys
func Init(path string, logger *logrus.Logger) {

	privKeyPath := path + "/app.rsa"
	pubKeyPath := path + "/app.rsa.pub"

	logger.Info("Getting private key...")

	// Read the bytes of the private key
	signBytes, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		logger.Fatal(err)
	}
	// get the actual key
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("Getting public key...")

	// Read the bytes of the public key
	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		logger.Fatal(err)
	}
	// Get the actual public key
	VerifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("Keys retrieved")

}
