package messaging

import (
	"os"

	"github.com/lordwestcott/gofabric/messaging/sms"
)

type Messaging struct {
	// EmailService        *EmailService
	SMSService sms.SMSService
	// VerificationService *VerificationService
	// WhatsAppService     *WhatsAppService //<- requires paid account
}

func NewMessaging() (*Messaging, error) {

	//check what services to use
	var sms_service sms.SMSService

	switch os.Getenv("SMS_SERVICE") {
	case "twilio":
		service, err := sms.NewTwilioSMSService()
		if err != nil {
			return nil, err
		}
		sms_service = service
	default:
		sms_service = nil
	}

	return &Messaging{
		// EmailService:        NewEmailService(),
		SMSService: sms_service,
		// VerificationService: NewVerificationService(),
		// WhatsAppService:     NewWhatsAppService(),
	}, nil
}
