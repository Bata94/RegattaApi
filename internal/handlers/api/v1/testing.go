package api_v1

import (
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/utils"
)

func TestHandler(c *handler.Context) error {
	err := utils.SendMail(utils.SendMailParams{
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
