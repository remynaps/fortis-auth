package oauth

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

var (
	oauthConf = &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		RedirectURL:  "http://localhost:8080/login/facebook/callback",
		Scopes:       []string{"public_profile"},
		Endpoint:     facebook.Endpoint,
	}
	oauthStateString = "thisshouldberandom"
)

func FacebookLoginHandler(w http.ResponseWriter, r *http.Request) {
	Url, err := url.Parse(oauthConf.Endpoint.AuthURL)
	if err != nil {
		log.Fatal("Parse: ", err)
	}
	parameters := url.Values{}
	parameters.Add("client_id", oauthConf.ClientID)
	parameters.Add("scope", strings.Join(oauthConf.Scopes, " "))
	parameters.Add("redirect_uri", oauthConf.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", oauthStateString)
	Url.RawQuery = parameters.Encode()
	url := Url.String()
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleFacebookCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		fmt.Fprint(w, "invalid oauth state")
		return
	}
	log.Println("recieved token")

	code := r.FormValue("code")

	token, err := oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Fprint(w, "Exchange failed")
		return
	}

	log.Println(token)

}
