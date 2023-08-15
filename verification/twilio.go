package verification

import (
	"fmt"
	"os"

	"github.com/twilio/twilio-go"
	verify "github.com/twilio/twilio-go/rest/verify/v2"
)

type Twilio_Verification_Service struct {
	client    *twilio.RestClient
	serviceID string
}

func NewTwilioVerificationService() (*Twilio_Verification_Service, error) {

	if os.Getenv("TWILIO_ACCOUNT_SID") == "" || os.Getenv("TWILIO_AUTH_TOKEN") == "" {
		return nil, fmt.Errorf("TWILIO_ACCOUNT_SID and TWILIO_AUTH_TOKEN must be set in the environment")
	}

	if os.Getenv("TWILIO_VERIFY_SERVICE_ID") == "" {
		return nil, fmt.Errorf("TWILIO_VERIFY_SERVICE_ID must be set in the environment")
	}

	client := twilio.NewRestClient()

	return &Twilio_Verification_Service{
		client:    client,
		serviceID: os.Getenv("TWILIO_VERIFY_SERVICE_ID"),
	}, nil
}

func (tss *Twilio_Verification_Service) VerifyNumberWithSMS(to string) error {
	params := &verify.CreateVerificationParams{}
	params.SetTo(to)
	params.SetChannel("sms")

	resp, err := tss.client.VerifyV2.CreateVerification(tss.serviceID, params)
	if err != nil {
		return err
	} else {
		if resp.Sid != nil {
			fmt.Println(*resp.Sid)
		} else {
			fmt.Println("No SID")
			return fmt.Errorf("no SID")
		}
	}

	return nil
}

func (tss *Twilio_Verification_Service) VerifyWhatsAppNumber(to string) error {
	params := &verify.CreateVerificationParams{}
	params.SetTo(to)
	params.SetChannel("whatsapp")

	resp, err := tss.client.VerifyV2.CreateVerification(tss.serviceID, params)
	if err != nil {
		return err
	} else {
		if resp.Sid != nil {
			fmt.Println(*resp.Sid)
		} else {
			fmt.Println("No SID")
			return fmt.Errorf("no SID")
		}
	}

	return nil
}

func (tss *Twilio_Verification_Service) VerifyNumberWithCall(to string) error {
	params := &verify.CreateVerificationParams{}
	params.SetTo(to)
	params.SetChannel("call")

	resp, err := tss.client.VerifyV2.CreateVerification(tss.serviceID, params)
	if err != nil {
		return err
	} else {
		if resp.Sid != nil {
			fmt.Println(*resp.Sid)
		} else {
			fmt.Println("No SID")
			return fmt.Errorf("no SID")
		}
	}

	return nil
}

// VerifyNumberWithCallWithExt is the same as VerifyNumberWithCall but expects a phone extension
func (tss *Twilio_Verification_Service) VerifyNumberWithCallWithExt(to, phone_ext string) error {
	params := &verify.CreateVerificationParams{}
	params.SetSendDigits(phone_ext)
	params.SetTo(to)
	params.SetChannel("call")

	resp, err := tss.client.VerifyV2.CreateVerification(tss.serviceID, params)
	if err != nil {
		return err
	} else {
		if resp.Sid != nil {
			fmt.Println(*resp.Sid)
		} else {
			fmt.Println("No SID")
			return fmt.Errorf("no SID")
		}
	}

	return nil
}

func (tss *Twilio_Verification_Service) VerifyEmail(to string) error {
	params := &verify.CreateVerificationParams{}
	params.SetTo(to)
	params.SetChannel("email")

	resp, err := tss.client.VerifyV2.CreateVerification(tss.serviceID, params)
	if err != nil {
		return err
	} else {
		if resp.Sid != nil {
			fmt.Println(*resp.Sid)
		} else {
			fmt.Println("No SID")
			return fmt.Errorf("no SID")
		}
	}

	return nil
}

// returns: 'approved', 'pending', or 'canceled'
func (tss *Twilio_Verification_Service) CheckVerificationCodeWithPhoneNumberOrEmail(to, code string) (*string, error) {
	params := &verify.CreateVerificationCheckParams{}
	params.SetTo(to)
	params.SetCode(code)

	resp, err := tss.client.VerifyV2.CreateVerificationCheck(tss.serviceID, params)
	if err != nil {
		return nil, err
	}

	if resp.Status == nil {
		fmt.Println("No Status")
		return nil, fmt.Errorf("No Status")
	}

	fmt.Println(*resp.Status)
	return resp.Status, nil
}
