package verification

type VerificationService interface {
	VerifyNumberWithSMS(to string) error
	VerifyWhatsAppNumber(to string) error
	VerifyNumberWithCall(to string) error
	VerifyNumberWithCallWithExt(to, phone_ext string) error
	VerifyEmail(to string) error
	CheckVerificationCodeWithPhoneNumberOrEmail(to, code string) error
}
