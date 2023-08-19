package oauth

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var goa Google_OAuth2

func TestMain(m *testing.M) {

	godotenv.Load("../../.env")
	goa = Google_OAuth2{}
	goa.New("http://localhost:3000/oauth2/callback", os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), os.Getenv("GOOGLE_STATE"))

	code := m.Run()
	os.Exit(code)
}

func TestGoogle_OAuth2_New(t *testing.T) {
	if goa.SSOGoLang == nil {
		t.Error("SSOGoLang is nil")
	}
	if goa.RandomString == "" {
		t.Error("RandomString is empty")
	}
}

func TestGoogle_OAuth2_ShouldContainAdditionalScopes(t *testing.T) {

	ogscopes := os.Getenv("GOOGLE_SCOPES_ADDITIONAL")

	os.Setenv("GOOGLE_SCOPES_ADDITIONAL", "a|b|c")

	testgoa := Google_OAuth2{}
	testgoa.New("http://localhost:3000/oauth2/callback", os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), os.Getenv("GOOGLE_STATE"))

	if len(testgoa.SSOGoLang.Scopes) < 3 {
		t.Error("Scopes should contain at least 3 items")
	}

	a := false
	b := false
	c := false

	for _, v := range testgoa.SSOGoLang.Scopes {
		if v == "a" {
			a = true
		}
		if v == "b" {
			b = true
		}
		if v == "c" {
			c = true
		}
	}

	if !a || !b || !c {
		t.Error("Scopes should contain a, b, and c")
	}

	os.Setenv("GOOGLE_SCOPES_ADDITIONAL", ogscopes)
}
