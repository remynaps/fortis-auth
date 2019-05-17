package main

import (
	"log"

	"gitlab.com/gilden/fortis/authorization"
	"gitlab.com/gilden/fortis/cmd/fortis_api/server"
	"gitlab.com/gilden/fortis/configuration"
	"gitlab.com/gilden/fortis/logging"
	"gitlab.com/gilden/fortis/models"
)

const sessionName = "authentication"

func main() {

	logging.Info("\n" +
		"  _ __   _____   ____ _ \n" +
		" | '_ \\ / _ \\ \\ / / _` | \n" +
		" | | | | (_) \\ V / (_| | \n" +
		" |_| |_|\\___/ \\_/ \\__,_| \n")

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
