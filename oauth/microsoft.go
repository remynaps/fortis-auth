package oauth

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

var (
	microsoftConfig = &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		RedirectURL:  "http://localhost:8080/login/microsoft/callback",
		Scopes:       []string{"user.read"},
		Endpoint:     microsoft.AzureADEndpoint(""), // empty = common
	}
	microsoftOauthStateString = "thisshouldberandom"
)

func MicrosoftLoginHandler(w http.ResponseWriter, r *http.Request) {
	Url, err := url.Parse(microsoftConfig.Endpoint.AuthURL)
	if err != nil {
		log.Fatal("Parse: ", err)
	}
	parameters := url.Values{}
	parameters.Add("client_id", microsoftConfig.ClientID)
	parameters.Add("scope", strings.Join(microsoftConfig.Scopes, " "))
	parameters.Add("redirect_uri", microsoftConfig.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", microsoftOauthStateString)
	Url.RawQuery = parameters.Encode()
	url := Url.String()
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleMicrosoftCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != microsoftOauthStateString {
		fmt.Fprint(w, "invalid oauth state")
		return
	}
	log.Println("recieved token")

	code := r.FormValue("code")

	token, err := microsoftConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Fprint(w, "Exchange failed")
		return
	}

	log.Println(token)

}
