package verification

// for twilio testing, you need to have a verify service id.
// you can do this here: https://console.twilio.com/us1/service/verify

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var tvs *Twilio_Verification_Service

func TestMain(m *testing.M) {
	// setup
	err := godotenv.Load("../../.env")
	if err != nil {
		panic(err)
	}
	service, err := NewTwilioVerificationService()
	if err != nil {
		panic(err)
	}
	tvs = service

	// run tests
	code := m.Run()
	// teardown
	os.Exit(code)
}

// func TestTwilio_VerifyNumberWithSMS(t *testing.T) {
// 	// arrange
// 	phoneNumber := "+447810567513"

// 	// act
// 	err := tvs.VerifyNumberWithSMS(phoneNumber)
// 	if err != nil {
// 		t.Error(err)
// 	}
// }

// func TestTwilio_CheckVerificationCodeWithPhoneNumberOrEmail_Phone(t *testing.T) {
// 	// arrange
// 	phoneNumber := "+447810567513"
// 	code := "137689"

// 	// act
// 	status, err := tvs.CheckVerificationCodeWithPhoneNumberOrEmail(phoneNumber, code)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if *status != "approved" || *status != "pending" || *status != "canceled" {
// 		t.Error("status is not approved, pending or canceled")
// 	}
// }

// A Mailer is required to send emails
// func TestTwilio_VerifyEmail(t *testing.T) {
// 	// arrange
// 	email := "olly.filmer@outlook.com"

// 	err := tvs.VerifyEmail(email)
// 	if err != nil {
// 		t.Error(err)
// 	}
// }

// func TestTwilio_CheckVerificationCodeWithPhoneNumberOrEmail_Email(t *testing.T) {
// 	// arrange
// 	email := "olly.filmer@outlook.com"
// 	code := "123456"

// 	// act
// 	err := tvs.CheckVerificationCodeWithPhoneNumberOrEmail(email, code)
// 	if err != nil {
// 		t.Error(err)
// 	}
// }

// func TestTwilio_VerifyWhatsAppNumber(t *testing.T) {
// 	// arrange
// 	phoneNumber := "+447810567513"

// 	// act
// 	err := tvs.VerifyNumberWithCall(phoneNumber)
// 	if err != nil {
// 		t.Error(err)
// 	}
// }
