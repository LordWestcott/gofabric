package whatsapp

type WhatsAppService interface {
	SendMessage(to, body string) error
}
