package authorization

import "gitlab.com/gilden/fortis/models"

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
func InitService() (*Auth, error) {
	return &Auth{}, nil
}
