package authorization

import (
	"log"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"gitlab.com/gilden/fortis/models"
)

// Use this code to create a token

// 			tokenData := new(authorization.TokenInfo)
// 			tokenData.ID = userID
// 			tokenData.EMail = email
// 			tokenData.Name = name

// 			// login or sign up
// 			token, err := authorization.CompleteFlow(tokenData, env.store)

// 			if err != nil {
// 				// Handler error
// 			}

// 			jsonResponse(token, w)

// CompleteFlow will log a user in or sign up if the user doesnt have an account yet.
// It will then generate and return a signed jwt based on the user data
func CompleteFlow(user *models.User, db models.UserStore) (*Token, error) {

	token := CreateToken(user)

	// create the token
	return token, nil
}

// CreateToken is used to verify user login. And grant a user a token
func CreateToken(usr *models.User) *Token {

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
