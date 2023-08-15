package email

import (
	"fmt"
	"log"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGrid struct {
	client *sendgrid.Client
}

func NewSendGrid() (*SendGrid, error) {
	c := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	return &SendGrid{
		client: c,
	}, nil
}

func (sg *SendGrid) SendMail(toName, toAddress, fromName, fromAddress, subject string, plainBytes []byte, htmlBytes []byte) (*int, error) {
	from := mail.NewEmail(fromName, fromAddress)
	to := mail.NewEmail(toName, toAddress)
	plainTextContent := string(plainBytes)
	htmlContent := string(htmlBytes)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	response, err := sg.client.Send(message)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
	return &response.StatusCode, nil
}
