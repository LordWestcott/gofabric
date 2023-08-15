package sms

type SMSService interface {
	SendMessage(body, to, from string) error
}
