package google

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type GoogleSignIn struct {
	clientID string
}

func (g *GoogleSignIn) New(clientID string) {
	g.clientID = clientID
}

func (g *GoogleSignIn) ValidateGoogleJWT(tokenString, clientID string) (GoogleClaims, error) {
	claimsStruct := GoogleClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) {
			pem, err := g.GetGooglePublicKey(fmt.Sprintf("%s", token.Header["kid"]))
			if err != nil {
				return nil, err
			}
			key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pem))
			if err != nil {
				return nil, err
			}
			return key, nil
		},
	)
	if err != nil {
		return GoogleClaims{}, err
	}

	claims, ok := token.Claims.(*GoogleClaims)
	if !ok {
		return GoogleClaims{}, errors.New("Invalid Google JWT")
	}

	if claims.Issuer != "accounts.google.com" && claims.Issuer != "https://accounts.google.com" {
		return GoogleClaims{}, errors.New("iss is invalid")
	}

	if claims.Audience != clientID {
		return GoogleClaims{}, errors.New("aud is invalid")
	}

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return GoogleClaims{}, errors.New("JWT is expired")
	}

	return *claims, nil
}

type GoogleClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FirstName     string `json:"given_name"`
	LastName      string `json:"family_name"`
	jwt.StandardClaims
}

func (g *GoogleSignIn) GetGooglePublicKey(keyID string) (string, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v1/certs")
	if err != nil {
		return "", err
	}
	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	myResp := map[string]string{}
	err = json.Unmarshal(dat, &myResp)
	if err != nil {
		return "", err
	}
	key, ok := myResp[keyID]
	if !ok {
		return "", errors.New("key not found")
	}
	return key, nil
}

// func (g *GoogleSignIn) LoginHandler(w http.ResponseWriter, r *http.Request) {
// 	defer r.Body.Close()

// 	// parse the GoogleJWT that was POSTed from the front-end
// 	type parameters struct {
// 		GoogleJWT *string
// 	}
// 	decoder := json.NewDecoder(r.Body)
// 	params := parameters{}
// 	err := decoder.Decode(&params)
// 	if err != nil {
// 		g.respondWithError(w, 500, "Couldn't decode parameters")
// 		return
// 	}

// 	// Validate the JWT is valid
// 	claims, err := g.ValidateGoogleJWT(*params.GoogleJWT, g.clientID)
// 	if err != nil {
// 		g.respondWithError(w, 403, "Invalid google auth")
// 		return
// 	}
// 	if claims.Email != user.Email {
// 		g.respondWithError(w, 403, "Emails don't match")
// 		return
// 	}

// 	// create a JWT for OUR app and give it back to the client for future requests
// 	tokenString, err := auth.MakeJWT(claims.Email, g.secret)
// 	if err != nil {
// 		g.respondWithError(w, 500, "Couldn't make authentication token")
// 		return
// 	}

// 	g.respondWithJSON(w, 200, struct {
// 		Token string `json:"token"`
// 	}{
// 		Token: tokenString,
// 	})
// }

// func (g *GoogleSignIn) AuthMiddleware(next http.Handler) http.Handler {
// }

// func (g *GoogleSignIn) respondWithError(w http.ResponseWriter, code int, message string) {
// 	message = fmt.Sprintf(`{"error": "%s"}`, message)
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(code)
// 	w.Write([]byte(message))
// }

// func (g *GoogleSignIn) respondWithJSON(w http.ResponseWriter, code int, message interface{}) {
// 	json, err := json.Marshal(message)
// 	if err != nil {
// 		g.respondWithError(w, 500, "Couldn't marshal response")
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(code)
// 	w.Write(json)
// }
