package api_v1

import (
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/mailer"
)

func TestHandler(c *handler.Context) error {
	err := mailer.Enqueue(c.Request.Context(), mailer.Params{
		To: []string{
			"bastian.sievers@gmail.com",
			"bastian.sievers+test@gmail.com",
		},
		CC:      []string{},
		Subject: "Test Mail",
		Body:    "Dies ist eine TestMail",
		Files:   []string{},
	})

	if err != nil {
		return err
	}

	return c.JSON("success")
}
