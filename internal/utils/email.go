package utils

import (
	"fmt"
	"log/slog"

	"github.com/bata94/RegattaApi/internal/config"
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
	slog.Info("Init Mail")
	emailOptions = new(EMailOptions)

	emailOptions.Sender = config.C.Email.Sender
	emailOptions.PW = config.C.Email.Password
	emailOptions.SmtpHost = config.C.Email.SMTPHost
	emailOptions.SmtpPort = config.C.Email.SMTPPort
	if emailOptions.SmtpPort == 0 {
		slog.Error("Mail Init Failed: Invalid SMTP port")
		emailOptions = nil
		return
	}

	var err error
	emailClient = new(mail.Client)
	emailClient, err = mail.NewClient(
		emailOptions.SmtpHost,
		mail.WithPort(emailOptions.SmtpPort),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(emailOptions.Sender),
		mail.WithPassword(emailOptions.PW),
	)
	if err != nil {
		slog.Error("Mail Init Failed, at client creation", "err", err, "options", emailOptions)
		emailOptions = nil
		return
	}

	slog.Info("Mail Init Successfull")
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
		slog.Error(fmt.Sprintf("failed to set From address: %s", err))
		return err
	}

	if err := m.To(params.To...); err != nil {
		slog.Error(fmt.Sprintf("failed to set To address: %s", err))
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
		slog.Error(fmt.Sprintf("failed to set CC address: %s", err))
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
		slog.Error("Failed to send email", "err", err)
		return err
	}

	return nil
}
