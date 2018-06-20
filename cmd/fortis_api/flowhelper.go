package main

import (
	"github.com/remynaps/fortis/authorization"
	"github.com/remynaps/fortis/models"
)

type TokenInfo struct {
	ID    string
	Name  string
	EMail string
}

// CompleteFlow will log a user in or sign up if the user doesnt have an account yet.
// It will then generate and return a signed jwt based on the user data
func CompleteFlow(tokenInfo *TokenInfo, db models.UserStore) (*authorization.Token, error) {
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

	token := authorization.CreateToken()

	// create the token
	return token, nil
}
