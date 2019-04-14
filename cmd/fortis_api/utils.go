package main

import (
	"html/template"
	"net/http"
)

func isValueInList(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func renderError(w http.ResponseWriter, errorText string, statusCode int) {
	t := template.Must(template.New("error.html").ParseFiles("./templates/error.html")) // Create a template.

	template := mainTemplate{}

	template.Hero = errorText

	w.WriteHeader(statusCode)

	t.Execute(w, template) // merge.
}
