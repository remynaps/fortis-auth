package authorization

import (
	"log"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// Basic user info
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type Token struct {
	Token string `json:"token"`
}

// CreateToken is used to verify user login. And grant a user a token
func CreateToken() *Token {
	// Generate the jwt
	token := jwt.New(jwt.SigningMethodRS256)

	// Add the required expiration and creation time claims to the token
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour).Unix()
	claims["iat"] = time.Now().Unix()
	claims["name"] = "ben swolo"
	claims["scope"] = "get_swole"
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
