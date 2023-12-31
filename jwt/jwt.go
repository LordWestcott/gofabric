package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID   int64    `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Scope    []string `json:"scope"`
	jwt.RegisteredClaims
}

type JWT struct {
	Secret []byte
}

// GenerateJWT generates a JWT token.
// If expMins is 0, then the token will never expire.
func (j *JWT) GenerateJWT(tokenClaims *Claims, expSecs int) (string, error) {
	if len(j.Secret) < 1 {
		return "", errors.New("No secret provided")
	}

	if tokenClaims == nil {
		tokenClaims = &Claims{}
	}

	if expSecs > 0 {
		tokenClaims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(expSecs)))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)

	tokenString, err := token.SignedString(j.Secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// If the token hasn't expired or if expiry on token is unset, the Expired will be false.
type JWTDetails struct {
	Claims *Claims
	Valid  bool
}

func (j *JWT) VerifyJWT(tokenString string) (*JWTDetails, error) {
	details := &JWTDetails{}

	if tokenString == "" {
		return nil, errors.New("No JWT token provided")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return j.Secret, nil
		},
	)

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, errors.New("Unauthorized Token.")
		}
		return nil, err
	}

	details.Claims = claims
	details.Valid = token.Valid

	return details, nil
}
