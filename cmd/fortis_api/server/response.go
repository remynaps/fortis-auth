package server

// Json wrapper function
import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"
)

func JsonResponse(response interface{}, w http.ResponseWriter) {

	// Create a json object from the given interface type
	json, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Json created, write success
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func Error(w http.ResponseWriter, err error, requestId string, code int, logger *logrus.Logger) {
	// Hide error from client if it is internal.
	if code == http.StatusInternalServerError {
		err = errors.New("An internal server error occured")
	}

	result := &ErrorResponseData{
		APIVersion: "beta", // change this
		Method:     "GET",  // and this
		RequestID:  requestId,
		Error: ErrorData{
			Code:    code,
			Message: err.Error(),
		},
	}

	// Write generic error response.
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(result)
}
