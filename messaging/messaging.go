package messaging

import (
	"os"

	"github.com/lordwestcott/gofabric/messaging/email"
	"github.com/lordwestcott/gofabric/messaging/sms"
	"github.com/lordwestcott/gofabric/messaging/verification"
	"github.com/lordwestcott/gofabric/messaging/whatsapp"
)

type Messaging struct {
	EmailService        email.EmailService
	SMSService          sms.SMSService
	VerificationService verification.VerificationService
	WhatsAppService     whatsapp.WhatsAppService //<- requires paid account
}

func NewMessaging() (*Messaging, error) {

	var messaging *Messaging

	switch os.Getenv("SMS_SERVICE") {
	case "twilio":
		service, err := sms.NewTwilioSMSService()
		if err != nil {
			return nil, err
		}
		messaging.SMSService = service
	default:
		messaging.SMSService = nil
	}

	switch os.Getenv("EMAIL_SERVICE") {
	case "sendgrid":
		service, err := email.NewSendGrid()
		if err != nil {
			return nil, err
		}
		messaging.EmailService = service
	default:
		messaging.EmailService = nil
	}

	switch os.Getenv("WHATSAPP_SERVICE") {
	case "twilio":
		service, err := whatsapp.NewTwilioWhatsAppService()
		if err != nil {
			return nil, err
		}
		messaging.WhatsAppService = service
	default:
		messaging.WhatsAppService = nil
	}

	switch os.Getenv("VERIFICATION_SERVICE") {
	case "twilio":
		service, err := verification.NewTwilioVerificationService()
		if err != nil {
			return nil, err
		}
		messaging.VerificationService = service
	default:
		messaging.VerificationService = nil
	}

	return messaging, nil
}
