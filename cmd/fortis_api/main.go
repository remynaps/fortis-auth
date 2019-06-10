package main

import (
	"flag"
	"log"

	"github.com/joho/godotenv"
	"gitlab.com/gilden/fortis/authorization"
	"gitlab.com/gilden/fortis/cmd/fortis_api/server"
	"gitlab.com/gilden/fortis/configuration"
	"gitlab.com/gilden/fortis/logging"
	"gitlab.com/gilden/fortis/models"
)

const sessionName = "authentication"

func main() {

	var envFile string
	flag.StringVar(&envFile, "env-file", "", "Use an env file to load variables")
	flag.Parse()

	logging.Info("\n" +
		"  _ __   _____   ____ _ \n" +
		" | '_ \\ / _ \\ \\ / / _` | \n" +
		" | | | | (_) \\ V / (_| | \n" +
		" |_| |_|\\___/ \\_/ \\__,_| \n")

	if envFile != "" {
		// First, load the environment variables
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	config := configuration.New()

	err := logging.Setup(config)
	if err != nil {
		logging.Panic(err)
	}

	logging.Info("Connecting to database..")
	db, err := models.InitDB(config)
	if err != nil {
		// db connection failed. start the retry logic
		logging.Error("Failed to connect to the database." + err.Error())
	}
	logging.Info("Connected!")

	// Init services
	err = authorization.Init(config)
	if err != nil {
		logging.Panic(err)
	}

	server, err := server.NewServer(config, db)
	if err != nil {
		log.Fatal(err)
	}
	server.Start()
}
