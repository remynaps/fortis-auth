package models

import (
	"database/sql"
	"time"

	// postgres driver
	_ "github.com/lib/pq"
)

var db *sql.DB

type DB struct {
	*sql.DB
}

type User struct {
	ID          string
	DisplayName string
	ExternalID  string    `json:"externalID"`
	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type UserStore interface {
	UserExists(id string) bool
	GetUserByID(id string) (*User, error)
	Search(query string) (*[]User, error)
	InsertUser(user *User) error
}

func InitDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}
