package server

import (
	"log"
	"net/http"
	"net/url"

	"gitlab.com/gilden/fortis/logging"
)

func isValueInList(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func renderError(w http.ResponseWriter, r *http.Request, errorText string, errorDescription string, errorHint string) {

	relativeURL := "/error"
	u, err := url.Parse(relativeURL)
	if err != nil {
		log.Fatal(err)
	}

	queryString := u.Query()
	queryString.Set("Error", errorText)
	queryString.Set("Error_description", errorDescription)
	queryString.Set("Error_hint", "")
	u.RawQuery = queryString.Encode()
	logging.Info(u.String())

	http.Redirect(w, r, u.String(), http.StatusFound)
}
