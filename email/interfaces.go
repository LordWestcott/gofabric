package email

type EmailService interface {
	SendMail(toName, toAddress, fromName, fromAddress, subject string, plainBytes, htmlBytes []byte) (*int, error)
}
