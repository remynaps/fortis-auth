package configuration

import (
	"os"
)

type ServerConfig struct {
	HostPort    string
	HostAddress string
	SessionName string
}

type KeyConfig struct {
	KeyPath    string
	PublicKey  string
	PrivateKey string
}

type DatabaseConfig struct {
	DatabasePath   string
	DatabasePort   string
	MigrationsPath string
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
}

type MicrosoftConfig struct {
	ClientID     string
	ClientSecret string
}

type LoggingConfig struct {
	File string
}

type Config struct {
	Server    ServerConfig
	Keys      KeyConfig
	Database  DatabaseConfig
	Google    GoogleConfig
	Microsoft MicrosoftConfig
	Logging   LoggingConfig
}

// New returns a new Config struct
func New() *Config {
	return &Config{
		Server: ServerConfig{
			HostAddress: getEnv("FORTIS_HOST_ADDRESS", ""),
			HostPort:    getEnv("FORTIS_HOST_PORT", "8081"),
			SessionName: getEnv("FORTIS_SESSION_NAME", "fortis_auth"),
		},
		Keys: KeyConfig{
			KeyPath:    getEnv("FORTIS_KEY_PATH", "./config/jwt/"),
			PublicKey:  getEnv("FORTIS_PUBLIC_KEY", "app.rsa.pub"),
			PrivateKey: getEnv("FORTIS_PRIVATE_KEY", "app.rsa"),
		},
		Database: DatabaseConfig{
			DatabasePath:   getEnv("FORTIS_DATABASE_PATH", ""),
			DatabasePort:   getEnv("FORTIS_DATABASE_PORT", ""),
			MigrationsPath: getEnv("FORTIS_MIGRATIONS_PATH", ""),
		},
		Google: GoogleConfig{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		},
		Microsoft: MicrosoftConfig{
			ClientID:     getEnv("MICROSOFT_CLIENT_ID", ""),
			ClientSecret: getEnv("MICROSOFT_CLIENT_SECRET", ""),
		},
		Logging: LoggingConfig{
			File: getEnv("LOGGING_FILE_PATH", ""),
		},
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
