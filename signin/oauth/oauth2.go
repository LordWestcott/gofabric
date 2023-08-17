package oauth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Google_OAuth2 struct {
	SSOGoLang    *oauth2.Config
	RandomString string
}

func (o *Google_OAuth2) New(redirect, googleClientID, googleClientSecret, state string) error {
	o.SSOGoLang = &oauth2.Config{
		RedirectURL:  redirect,
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}

	o.RandomString = state
	return nil
}

func (o *Google_OAuth2) SignIn(w http.ResponseWriter, r *http.Request) {
	url := o.SSOGoLang.AuthCodeURL(o.RandomString)
	fmt.Println(url) //for testing purposes
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)

}

func (o *Google_OAuth2) CallBack(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	state := r.FormValue("state")
	code := r.FormValue("code")

	return o.getUserData(state, code)
}

func (o *Google_OAuth2) getUserData(state, code string) ([]byte, error) {
	if state != o.RandomString {
		return nil, errors.New("invalid user state")
	}

	token, err := o.SSOGoLang.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}

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
