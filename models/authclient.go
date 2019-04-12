package models

import (
	"database/sql"
	"fmt"
	"log"

	uuid "github.com/satori/go.uuid"
)

// UserExists checks if a user exists and returns a simple boolean
func (db *DB) ClientExists(id string) bool {
	client := new(AuthClient)
	err := db.QueryRow("SELECT * FROM oauth_clients where externalid = $1", id).Scan(&client.ID, &client.DisplayName, &client.Created, &client.LastUpdated, &client.ClientSecret)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Printf("Name is %s\n", client.DisplayName)
	}
	return true
}

// GetUserByID retrieves one user from the database with a given id
func (db *DB) GetClientByID(id string) (*AuthClient, error) {
	client := new(AuthClient)
	err := db.QueryRow("SELECT * FROM oauth_clients where externalid = $1", id).Scan(&client.ID, &client.DisplayName, &client.Created, &client.LastUpdated, &client.ClientSecret)
	switch {
	case err == sql.ErrNoRows:
		log.Printf("No client with that ID.")
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Printf("Name is %s\n", client.DisplayName)
	}
	return client, nil
}

// InsertUser creates a new user entry in the database
// Should only be used if a user does not exists
func (db *DB) InsertClient(client *AuthClient) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	internalID := uuid.NewV4()

	stmt, err := tx.Prepare(`INSERT INTO oauth_clients (ID, DisplayName, externalId)
                     VALUES($1,$2,$3);`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(internalID, client.DisplayName, client.ID); err != nil {
		tx.Rollback() // return an error too, might need it
		return err
	}

	// Finally commit the transaction
	return tx.Commit()
}
