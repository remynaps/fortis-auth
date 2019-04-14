package models

import (
	"database/sql"
	"time"

	// postgres driver

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"gitlab.com/gilden/fortis/logging"
)

type DB struct {
	*sql.DB
}

type User struct {
	ID          string
	DisplayName string
	Email       string
	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type UserIdentity struct {
	ID          string
	UserID      string
	ExternalID  string    `json:"externalID"`
	Source      string    `json:"source"`
	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type Domain struct {
	ID          string
	DisplayName string
	ExternalID  string    `json:"externalID"`
	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type AuthClient struct {
	ID           string
	DisplayName  string
	ClientSecret string    `json:"clientSecret"`
	RedirectUris []string  `json:"redirectUris"`
	Scopes       []string  `json:"scopes"`
	Private      bool      `json:"private"`
	Created      time.Time `json:"created"`
	LastUpdated  time.Time `json:"lastUpdated"`
}

type UserStore interface {
	UserExists(id string) bool
	GetUserByID(id string) (*User, error)
	Search(query string) (*[]User, error)
	InsertUser(user *User) error
}

type DomainStore interface {
	DomainExists(id string) bool
	GetDomainByID(id string) (*Domain, error)
	SearchDomain(query string) (*[]Domain, error)
	InsertDomain(domain *Domain) error
}

type ClientStore interface {
	Clientexists(id string) bool
	GetClientByID(id string) (*User, error)
	InsertClient(client *AuthClient) error
}

func InitDB(config *viper.Viper) (*DB, error) {

	// Init the connection
	connection := config.GetString("database.data_source")
	db, err := sql.Open("postgres", connection)

	migrationsPath := config.GetString("database.migrations_path")

	// Migrate the db
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres", driver)

	// 1 step
	m.Up()

	if err != nil {
		logging.Error(err)
		return nil, err
	}
	return &DB{db}, nil
}
