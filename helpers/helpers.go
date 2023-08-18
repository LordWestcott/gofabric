package helpers

import "crypto/rand"

type Helpers struct{}

const (
	randomString = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"
)

func (h *Helpers) RandomString(length int) string {
	s, r := make([]rune, length), []rune(randomString)

	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}

	return string(s)
}
