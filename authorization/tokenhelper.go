package authorization

import (
	"log"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"gitlab.com/gilden/fortis/models"
)

// CompleteFlow will log a user in or sign up if the user doesnt have an account yet.
// It will then generate and return a signed jwt based on the user data
func (auth *Auth) CompleteFlow(tokenInfo *TokenInfo, db models.UserStore) (*Token, error) {

	usr := new(models.User)

	if !db.UserExists(tokenInfo.ID) {
		// Insert a new user
		usr.DisplayName = tokenInfo.Name
		usr.ID = tokenInfo.ID
		db.InsertUser(usr)
	}

	// retrieve the data to be shure
	usr, err := db.GetUserByID(tokenInfo.ID)
	if err != nil {

	}

	token := auth.CreateToken(usr)

	// create the token
	return token, nil
}

// CreateToken is used to verify user login. And grant a user a token
func (auth *Auth) CreateToken(usr *models.User) *Token {

	// Generate the jwt
	token := jwt.New(jwt.SigningMethodRS256)

	// Add the required expiration and creation time claims to the token
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour).Unix()
	claims["iat"] = time.Now().Unix()
	claims["name"] = usr.DisplayName
	claims["uid"] = usr.ID
	token.Claims = claims

	// Sign the token
	tokenString, err := token.SignedString(signKey)

	log.Println(tokenString)

	if err != nil {
		return nil
	}

	// Write the token to the response
	response := Token{tokenString}

	// Send json response containing the token
	return &response
}
