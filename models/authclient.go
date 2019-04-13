package models

import (
	"database/sql"
	"log"

	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"gitlab.com/gilden/fortis/logging"
)

// ClientExists checks if a user exists and returns a simple boolean
func (db *DB) ClientExists(id string) bool {

	parsedId, err := uuid.FromString(id)

	if err != nil {
		logging.Error("The id does not have the correct format: " + id)
		return false
	}

	client := new(AuthClient)
	err = db.QueryRow("SELECT * FROM oauth_clients where client_id = $1", parsedId).Scan(&client.ID, &client.DisplayName, &client.ClientSecret, pq.Array(&client.RedirectUris), pq.Array(&client.Scopes), &client.Private, &client.Created, &client.LastUpdated)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		log.Fatal(err)
	}
	return true
}

// GetClientByID retrieves one user from the database with a given id
func (db *DB) GetClientByID(id string) (*AuthClient, error) {

	parsedId, err := uuid.FromString(id)

	if err != nil {
		logging.Error("The id does not have the correct format: " + id)
		return nil, err
	}

	client := new(AuthClient)
	err = db.QueryRow("SELECT * FROM oauth_clients where client_id = $1", parsedId).Scan(&client.ID, &client.DisplayName, &client.ClientSecret, pq.Array(&client.RedirectUris), pq.Array(&client.Scopes), &client.Private, &client.Created, &client.LastUpdated)
	switch {
	case err == sql.ErrNoRows:
		log.Printf("No client with that ID.")
	case err != nil:
		log.Fatal(err)
	}
	return client, nil
}

// InsertClient creates a new client entry in the database
// Should only be used if a client does not exists
func (db *DB) InsertClient(client *AuthClient) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO oauth_clients (client_id, displayname, client_secret, redirect_uris, scopes, is_private)
                     VALUES($1,$2,$3,$4,$5);`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(client.ID, client.DisplayName, client.ClientSecret, client.RedirectUris, client.Scopes, client.Private); err != nil {
		tx.Rollback() // return an error too, might need it
		return err
	}

	// Finally commit the transaction
	return tx.Commit()
}
