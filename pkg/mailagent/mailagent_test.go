package mailagent

import (
	"fmt"
	"os"
	"testing"
)

var (
	envEmailSender, envPasswordSender, envEmailTestRecipient string
)

var tmplDir = "./test_tmpl"

func init() {
	var ok bool
	if envEmailSender, ok = os.LookupEnv("POSN_EMAIL_LOGIN"); !ok {
		panic("POSN_EMAIL_LOGIN undefined")
	}

	if envPasswordSender, ok = os.LookupEnv("POSN_EMAIL_PWD"); !ok {
		panic("POSN_EMAIL_PWD undefined")
	}

	if envEmailTestRecipient, ok = os.LookupEnv("POSN_EMAIL_TEST_RCP"); !ok {
		panic("POSN_EMAIL_TEST_RCP undefined")
	}
}

func TestInit(t *testing.T) {
	t.Run("simple init", func(t *testing.T) {
		NewMailAgent("", "")
	})
}

func TestLoadTemplatesFromDir(t *testing.T) {
	t.Run("simple load", func(t *testing.T) {
		mailagent := NewMailAgent("", "")
		if err := mailagent.LoadTemplatesFromDir(tmplDir); err != nil {
			fmt.Println(err)
		}
	})
}

func TestSend(t *testing.T) {
	mailagent := NewMailAgent(envEmailSender, envPasswordSender)
	err := mailagent.Send(envEmailTestRecipient, "hello")
	fmt.Println(err, envEmailSender, envEmailTestRecipient, envPasswordSender)
}

func TestSendTemplate(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		mailagent := NewMailAgent(envEmailSender, envPasswordSender)
		if err := mailagent.LoadTemplatesFromDir(tmplDir); err != nil {
			fmt.Println(err)
		}

		if err := mailagent.SendTemplate(envEmailTestRecipient, "code_verify.html", Value{"code": 34583958}); err != nil {
			fmt.Println(err)
		}
	})
}
