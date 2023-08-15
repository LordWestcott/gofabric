package sms

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var tss *Twilio_SMS_Service

func TestMain(m *testing.M) {
	err := godotenv.Load("../../.env")
	if err != nil {
		panic(err)
	}
	service, err := NewTwilioSMSService()
	if err != nil {
		panic(err)
	}
	tss = service
	os.Exit(m.Run())
}

// func TestTwilioSMSService_SendMessage(t *testing.T) {
// 	err := tss.SendMessage("Test Message", "+447810567513", "+447888867822")
// 	if err != nil {
// 		t.Error(err)
// 	}
// }
