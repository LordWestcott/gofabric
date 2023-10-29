package oauth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Google_OAuth2 struct {
	SSOGoLang    *oauth2.Config
	RandomString string
}

func (o *Google_OAuth2) New(redirect, googleClientID, googleClientSecret, state string) error {

	scopes := []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	}

	additional := os.Getenv("GOOGLE_SCOPES_ADDITIONAL")
	if additional != "" {
		as := strings.Split(additional, "|")
		scopes = append(scopes, as...)
	}

	o.SSOGoLang = &oauth2.Config{
		RedirectURL:  redirect,
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}

	o.RandomString = state
	return nil
}

func (o *Google_OAuth2) SignIn(w http.ResponseWriter, r *http.Request) {
	url := o.SSOGoLang.AuthCodeURL(o.RandomString, oauth2.AccessTypeOffline)
	fmt.Println(url) //for testing purposes
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (o *Google_OAuth2) CallBack(code, state string) (*oauth2.Token, []byte, error) {

	if code == "" {
		return nil, nil, errors.New("empty user code")
	}

	if state == "" {
		return nil, nil, errors.New("empty user state")
	}

	if state != o.RandomString {
		return nil, nil, errors.New("invalid user state")
	}

	token, err := o.SSOGoLang.Exchange(context.Background(), code)
	if err != nil {
		return nil, nil, err
	}

	userData, err := getUserDataFromToken(token)
	if err != nil {
		return token, nil, err
	}

	return token, userData, nil
}

func getUserDataFromToken(token *oauth2.Token) ([]byte, error) {
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
