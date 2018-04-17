package models

import (
	"database/sql"
	"fmt"
	"log"
)

// User is a simple displayname, id struct
type User struct {
	ID          string
	DisplayName string
}

// UserExists checks if a user exists and returns a simple bool
func UserExists(id string) bool {
	usr := new(User)
	err := db.QueryRow("SELECT * FROM users where id = $1", id).Scan(&usr.ID, &usr.DisplayName)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Printf("Username is %s\n", usr.DisplayName)
	}
	return true
}

// GetUserByID retrieves one user from the database with a given id
func GetUserByID(id string) (*User, error) {
	usr := new(User)
	err := db.QueryRow("SELECT * FROM users where id = $1", id).Scan(&usr.ID, &usr.DisplayName)
	switch {
	case err == sql.ErrNoRows:
		log.Printf("No user with that ID.")
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Printf("Username is %s\n", usr.DisplayName)
	}
	return usr, nil
}

// Search queries the database for users with the specified display name
func Search(query string) (*[]User, error) {

	var users []User

	rows, err := db.Query("SELECT * FROM users where displayname = $1", query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Start iterating over the retrieved rows
	for rows.Next() {
		var usr User
		// scan the row and set the vars in usr
		err := rows.Scan(&usr.ID, &usr.DisplayName)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, usr)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return &users, nil
}

// InsertUser creates a new user entry in the database
// Should only be used if a user does not exists
func InsertUser(user *User) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO users (ID, DisplayName)
                     VALUES($1,$2);`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(user.ID, user.DisplayName); err != nil {
		tx.Rollback() // return an error too, might need it
		return err
	}

	// Finally commit the transaction
	return tx.Commit()
}
