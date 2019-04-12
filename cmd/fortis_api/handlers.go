package main

import (
	"html/template"
	"net/http"
)

type mainTemplate struct {
	hero string
}

type consentTemplate struct {
	hero string
}

// Simple status handler to call to validate api
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("API is up and running"))
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("login.html").ParseFiles("./templates/login.html")) // Create a template.

	template := new(consentTemplate)

	template.hero = "This is where the fun begins"

	t.Execute(w, template) // merge.
}

func consentFileHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("consent.html").ParseFiles("./templates/consent.html")) // Create a template.

	template := new(mainTemplate)

	template.hero = "This is where the fun begins"

	t.Execute(w, template) // merge.
}
