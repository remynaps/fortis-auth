package authorization

import (
	"crypto/rsa"
	"io/ioutil"

	jwt "github.com/dgrijalva/jwt-go"
	"gitlab.com/gilden/fortis/configuration"
	"gitlab.com/gilden/fortis/logging"
	"gitlab.com/gilden/fortis/models"
)

// The Auth calling struct
type Auth struct{}

type AuthorizationService interface {
	CompleteFlow(tokenInfo *TokenInfo, db models.UserStore) (*Token, error)
	CreateToken(usr *models.User) *Token
}

// Basic user info
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type Token struct {
	Token string `json:"token"`
}

type TokenInfo struct {
	ID    string
	Name  string
	EMail string
}

// Handle more complex init
func Init(config *configuration.Config) error {
	path := config.Keys.KeyPath

	private_name := config.Keys.PrivateKey
	public_name := config.Keys.PublicKey

	privKeyPath := path + private_name
	pubKeyPath := path + public_name

	logging.Info("Getting private key...")

	// Read the bytes of the private key
	signBytes, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		logging.Panic(err)
	}
	// get the actual key
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		logging.Panic(err)
	}
	logging.Info("Getting public key...")

	// Read the bytes of the public key
	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		logging.Panic(err)
	}
	// Get the actual public key
	VerifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		logging.Panic(err)
	}
	logging.Info("Keys retrieved")
	return err
}

var (
	VerifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)
