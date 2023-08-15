package whatsapp

import (
	"fmt"
	"os"

	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

type Twilio_Whatsapp_Service struct {
	client         *twilio.RestClient
	whatsappNumber string
}

func NewTwilioWhatsAppService() (*Twilio_Whatsapp_Service, error) {
	if os.Getenv("TWILIO_ACCOUNT_SID") == "" || os.Getenv("TWILIO_AUTH_TOKEN") == "" {
		return nil, fmt.Errorf("TWILIO_ACCOUNT_SID and TWILIO_AUTH_TOKEN must be set in the environment")
	}

	client := twilio.NewRestClient()
	whatsappNumber := os.Getenv("TWILIO_WHATSAPP_NUMBER")

	return &Twilio_Whatsapp_Service{
		client:         client,
		whatsappNumber: whatsappNumber,
	}, nil
}

func (tws *Twilio_Whatsapp_Service) SendMessage(to, body string) error {
	params := &api.CreateMessageParams{}
	params.SetFrom("whatsapp:" + tws.whatsappNumber)
	params.SetBody(body)
	params.SetTo("whatsapp:" + to)

	_, err := tws.client.Api.CreateMessage(params)
	if err != nil {
		return err
	}

	return nil
}
