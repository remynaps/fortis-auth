package models

import (
	"database/sql"
	"log"

	uuid "github.com/satori/go.uuid"
	"gitlab.com/remynaps/fortis-hydra/logging"
)

// UserExists checks if a user exists and returns a simple boolean
func (db *DB) UserExists(id string) bool {

	usr := new(User)
	err := db.QueryRow("SELECT * FROM users where email = $1", id).Scan(&usr.ID, &usr.DisplayName, &usr.Email, &usr.Created, &usr.LastUpdated)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		log.Fatal(err)
	}
	return true
}

// GetUserByID retrieves one user from the database with a given id
func (db *DB) GetUserByID(id string) (*User, error) {
	usr := new(User)
	err := db.QueryRow("SELECT * FROM users where id = $1", id).Scan(&usr.ID, &usr.DisplayName, &usr.Email, &usr.Created, &usr.LastUpdated)
	switch {
	case err == sql.ErrNoRows:
		logging.Error("No user with that id")
	case err != nil:
		logging.Panic(err)
	}
	return usr, nil
}

// GetUserByID retrieves one user from the database with a given id
func (db *DB) GetUserByExternalID(id string) (*User, error) {
	usr := new(User)
	err := db.QueryRow("SELECT * FROM users where email = $1", id).Scan(&usr.ID, &usr.DisplayName, &usr.Email, &usr.Created, &usr.LastUpdated)
	switch {
	case err == sql.ErrNoRows:
		logging.Error("No user with that id")
	case err != nil:
		logging.Panic(err)
	}
	return usr, nil
}

// Search queries the database for users with the specified display name
func (db *DB) Search(query string) (*[]User, error) {

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
func (db *DB) InsertUser(user *User) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	internalID := uuid.NewV4()

	stmt, err := tx.Prepare(`INSERT INTO users (ID, DisplayName, email)
                     VALUES($1,$2,$3);`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(internalID, user.DisplayName, user.ID); err != nil {
		tx.Rollback() // return an error too, might need it
		return err
	}

	// Finally commit the transaction
	return tx.Commit()
}
