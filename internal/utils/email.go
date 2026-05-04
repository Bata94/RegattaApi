package utils

import (
	"log"
	"os"
	"strconv"

	"github.com/wneessen/go-mail"
)

var (
	emailOptions *EMailOptions
	emailClient  *mail.Client
)

type EMailOptions struct {
	Sender   string
	PW       string
	SmtpHost string
	SmtpPort int
}

func InitEmail() {
	log.Println("Init Mail")
	var err error
	emailOptions = new(EMailOptions)

	emailOptions.Sender = os.Getenv("EMAIL_SENDER")
	emailOptions.PW = os.Getenv("EMAIL_PW")
	emailOptions.SmtpHost = os.Getenv("EMAIL_SMTP_HOST")
	emailOptions.SmtpPort, err = strconv.Atoi(os.Getenv("EMAIL_SMTP_PORT"))
	if err != nil {
		log.Println("Mail Init Failed, at reading ENVs: ", err, "Mail Options: ", &emailOptions)
		emailOptions = nil
		return
	}

	emailClient = new(mail.Client)
	emailClient, err = mail.NewClient(
		emailOptions.SmtpHost,
		mail.WithPort(emailOptions.SmtpPort),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(emailOptions.Sender),
		mail.WithPassword(emailOptions.PW),
	)
	if err != nil {
		log.Println("Mail Init Failed, at client creation: ", err, "Mail Options: ", &emailOptions)
		emailOptions = nil
		return
	}

	log.Println("Mail Init Successfull")
}

type SendMailParams struct {
	To      []string
	CC      []string
	Subject string
	Body    string
	Files   []string
}

func SendMail(params SendMailParams) error {
	m := mail.NewMsg()

	if err := m.From(emailOptions.Sender); err != nil {
		log.Println("failed to set From address: %s", err)
		return err
	}

	if err := m.To(params.To...); err != nil {
		log.Println("failed to set To address: %s", err)
		return err
	}

	senderInCC := false
	for _, cc := range params.CC {
		if cc == emailOptions.Sender {
			senderInCC = true
			break
		}
	}
	if !senderInCC {
		params.CC = append(params.CC, emailOptions.Sender)
	}

	if err := m.Cc(params.CC...); err != nil {
		log.Println("failed to set CC address: %s", err)
		return err
	}

	if len(params.Files) != 0 {
		for _, file := range params.Files {
			m.AttachFile(file)
		}
	}

	m.Subject("[SYSTEM] " + params.Subject)
	m.SetBodyString(mail.TypeTextPlain, params.Body)

	if err := emailClient.DialAndSend(m); err != nil {
		log.Println("Failed to send email: ", err)
		return err
	}

	return nil
}
