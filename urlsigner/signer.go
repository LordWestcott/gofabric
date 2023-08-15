package urlsigner

import (
	"fmt"
	"strings"
	"time"

	goalone "github.com/bwmarrin/go-alone"
)

type Signer struct {
	Secret []byte
}

func (s *Signer) New(secret string) error {

	if len(secret) < 32 {
		return fmt.Errorf("secret must be at least 32 characters")
	}

	s.Secret = []byte(secret)
	return nil
}

// Generates a signed token from a url string.
// The token is appended to the url as a query string parameter `hash`.
func (s *Signer) GenerateTokenFromString(data string) string {
	var urlToSign string

	crypt := goalone.New(s.Secret, goalone.Timestamp)
	if strings.Contains(data, "?") {
		urlToSign = fmt.Sprintf("%s&hash=", data)
	} else {
		urlToSign = fmt.Sprintf("%s?hash=", data)
	}

	tokenBytes := crypt.Sign([]byte(urlToSign))
	token := string(tokenBytes)

	return token
}

func (s *Signer) VerifyToken(token string) bool {
	crypt := goalone.New(s.Secret, goalone.Timestamp)
	_, err := crypt.Unsign([]byte(token))
	return err == nil
}

func (s *Signer) Expired(token string, minsUntilExpired int) bool {
	crypt := goalone.New(s.Secret, goalone.Timestamp)
	ts := crypt.Parse([]byte(token))
	return time.Since(ts.Timestamp) > time.Duration(minsUntilExpired)*time.Minute
}
