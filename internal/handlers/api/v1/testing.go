package api_v1

import (
	"time"

	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/wneessen/go-mail"
)

func TestHandler(c *fiber.Ctx) error {
	// Sender data.
	from := ""
	password := ""

	// Receiver email address.
	to := ""

	// smtp server configuration.
	smtpHost := ""
	smtpPort := 0

	// Message.
	subject := "Test Mail Meldeergebnis"
	message := "This is a test email message. Sending Meldeergebnis"

	m := mail.NewMsg()
	if err := m.From(from); err != nil {
		log.Fatalf("failed to set From address: %s", err)
		return err
	}
	if err := m.To(to); err != nil {
		log.Fatalf("failed to set To address: %s", err)
		return err
	}

	files, err := utils.GetFilenames("meldeergebnis")
	if err != nil {
		return err
	}
	log.Debug(files)

	if len(files) != 0 {
		log.Warn("Files found attaching ", files[len(files)-1])
		m.AttachFile("/opt/app/files/meldeergebnis/" + files[len(files)-1])
	}

	m.Subject(subject)
	m.SetBodyString(mail.TypeTextPlain, message)
	client, err := mail.NewClient(
		smtpHost,
		mail.WithPort(smtpPort),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(from),
		mail.WithPassword(password))
	if err != nil {
		log.Fatalf("failed to create mail client: %s", err)
		return err
	}
	if err := client.DialAndSend(m); err != nil {
		log.Fatalf("failed to send mail: %s", err)
		return err
	}

	return api.JSON(c, "success")
}

func TestHandlerUUID(c *fiber.Ctx) error {
	revUUID := uuid.MustParse("018a65b6-36fc-7112-96a1-d0b0aac587e2")
	newUUID, _ := uuid.NewV7()

	log.Debug(revUUID, " ", newUUID)
	log.Debug(revUUID.ClockSequence(), " ", newUUID.ClockSequence())

	log.Debug(revUUID.ClockSequence() < newUUID.ClockSequence())

	time.Sleep(time.Second * 60)

	return api.JSON(c, "success")
}
