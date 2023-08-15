package email

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var sg *SendGrid

func TestMain(m *testing.M) {
	// setup
	godotenv.Load("../../.env")
	sendgrid, err := NewSendGrid()
	if err != nil {
		panic(err)
	}

	sg = sendgrid

	code := m.Run()
	// teardown
	os.Exit(code)
}

// func TestSendGrid_SendMail(t *testing.T) {

// 	fa := "noreply@lodgino.com"
// 	fn := "Test"
// 	ta := "olly.filmer@outlook.com"
// 	tn := "Olly"
// 	ptc := "and easy to do anywhere, even with Go"
// 	html := "<strong>and easy to do anywhere, even with Go</strong>"
// 	sub := "Sending with SendGrid is Fun"
// 	resp, err := sg.SendMail(tn, ta, fn, fa, sub, []byte(ptc), []byte(html))
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	fmt.Printf("response :%v\n", resp)
// }
