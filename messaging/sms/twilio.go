package sms

import (
	"fmt"
	"os"

	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

type Twilio_SMS_Service struct {
	client *twilio.RestClient
}

func NewTwilioSMSService() (*Twilio_SMS_Service, error) {

	if os.Getenv("TWILIO_ACCOUNT_SID") == "" || os.Getenv("TWILIO_AUTH_TOKEN") == "" {
		return nil, fmt.Errorf("TWILIO_ACCOUNT_SID and TWILIO_AUTH_TOKEN must be set in the environment")
	}

	client := twilio.NewRestClient()

	return &Twilio_SMS_Service{
		client: client,
	}, nil
}

func (tss *Twilio_SMS_Service) SendMessage(body, to, from string) error {
	params := &api.CreateMessageParams{}

	params.SetBody(body)
	params.SetTo(to)
	params.SetFrom(from)

	resp, err := tss.client.Api.CreateMessage(params)
	if err != nil {
		return err
	}

	if resp.Sid != nil {
		fmt.Println(*resp.Sid)
	} else {
		fmt.Println("No SID")
		return fmt.Errorf("no SID")
	}

	return nil
}
