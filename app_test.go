package gofabric

import (
	"testing"

	"github.com/joho/godotenv"
)

func TestApp(t *testing.T) {
	godotenv.Load(".env")

	_, err := InitApp()
	if err != nil {
		t.Error(err)
	}
}
