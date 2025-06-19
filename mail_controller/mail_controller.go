package mailcontroller

import (
	"os"
	"strconv"

	gomail "gopkg.in/mail.v2"
)

var MailerConfig struct {
	host        string
	port        int
	username    string
	password    string
	from        string
	contentType string
	dialer      *gomail.Dialer
}

func InitMailSystem() {
	MailerConfig.host = os.Getenv("MAIL_HOST")
	MailerConfig.port, _ = strconv.Atoi(os.Getenv("MAIL_PORT"))
	MailerConfig.username = os.Getenv("MAIL_USERNAME")
	MailerConfig.password = os.Getenv("MAIL_PASSWORD")
	MailerConfig.from = os.Getenv("MAIL_FROM")
	MailerConfig.contentType = os.Getenv("MAIL_CONTENT_TYPE")
	MailerConfig.dialer = gomail.NewDialer(MailerConfig.host, MailerConfig.port, MailerConfig.username, MailerConfig.password)
}

func SendMail(to, subject, body string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", MailerConfig.from)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody(MailerConfig.contentType, body)
	return MailerConfig.dialer.DialAndSend(message)
}
