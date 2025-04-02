package mailer

import (
	"errors"
	"testing"
)

func TestMail_SendSMTPMessage(t *testing.T) {
	msg := Message{
		From:        "test@example.com",
		FromName:    "Test",
		To:          "test2@example.com",
		Subject:     "Test",
		Template:    "test",
		Attachments: []string{"./testdata/mail/test.html.tmpl"},
		Data:        "Hello, world!",
	}

	err := mailer.SendSMTPMessage(msg)
	if err != nil {
		t.Error(err)
	}
}

func TestMail_SendUsingChan(t *testing.T) {
	msg := Message{
		From:     "test@example.com",
		FromName: "Test",
		To:       "test2@example.com",
		Subject:  "Test",
		Template: "test",
		Data:     "Hello, world!",
	}

	mailer.Jobs <- msg
	res := <-mailer.Results
	if res.Error != nil {
		t.Error(errors.New("error sending mail"))
	}

	msg.To = "not_an_email_address"
	mailer.Jobs <- msg
	res = <-mailer.Results
	if res.Error == nil {
		t.Error(errors.New("expected error, got nil"))
	}
}

func TestMail_SendUsing_API(t *testing.T) {
	msg := Message{
		To:          "test2@example.com",
		Subject:     "Test",
		Template:    "test",
		Attachments: []string{"./testdata/mail/test.html.tmpl"},
		Data:        "Hello, world!",
	}

	mailer.API = "unknown"
	mailer.APIKey = "1234567890"
	mailer.APIUrl = "https://fake.com"

	err := mailer.SendUsingAPI(msg, "unknown")
	if err == nil {
		t.Error(err)
	}
	mailer.API = ""
	mailer.APIKey = ""
	mailer.APIUrl = ""
}

func TestMail_buildHTMLMessage(t *testing.T) {
	msg := Message{
		From:     "test@example.com",
		FromName: "Test",
		To:       "test2@example.com",
		Subject:  "Test",
		Template: "test",
		Data:     "Hello, world!",
	}

	_, err := mailer.buildHTMLMessage(msg)
	if err != nil {
		t.Error(err)
	}
}

func TestMail_buildPlainTextMessage(t *testing.T) {
	msg := Message{
		From:     "test@example.com",
		FromName: "Test",
		To:       "test2@example.com",
		Subject:  "Test",
		Template: "test",
		Data:     "Hello, world!",
	}

	_, err := mailer.buildPlainTextMessage(msg)
	if err != nil {
		t.Error(err)
	}
}

func TestMail_Send(t *testing.T) {
	msg := Message{
		From:     "test@example.com",
		FromName: "Test",
		To:       "test2@example.com",
		Subject:  "Test",
		Template: "test",
		Data:     "Hello, world!",
	}

	err := mailer.Send(msg)
	if err != nil {
		t.Error(err)
	}

	mailer.API = "unknown"
	mailer.APIKey = "1234567890"
	mailer.APIUrl = "https://fake.com"

	err = mailer.Send(msg)
	if err == nil {
		t.Errorf("expected error:%s", err)
	}
	mailer.API = ""
	mailer.APIKey = ""
	mailer.APIUrl = ""
}

func TestMail_ChooseAPI(t *testing.T) {
	msg := Message{
		To:          "test2@example.com",
		Subject:     "Test",
		Template:    "test",
		Attachments: []string{"./testdata/mail/test.html.tmpl"},
		Data:        "Hello, world!",
	}
	mailer.API = "unknown"
	err := mailer.ChooseAPI(msg)
	if err == nil {
		t.Error(err)
	}
}
