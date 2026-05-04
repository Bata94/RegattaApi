package api_v1

import (
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/utils"
	"github.com/google/uuid"
	"log"
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

func TestHandlerUUID(c *handler.Context) error {
	revUUID := uuid.MustParse("018a65b6-36fc-7112-96a1-d0b0aac587e2")
	newUUID, _ := uuid.NewV7()

	log.Println(revUUID, " ", newUUID)
	log.Println(revUUID.ClockSequence(), " ", newUUID.ClockSequence())

	log.Println(revUUID.ClockSequence() < newUUID.ClockSequence())

	return c.JSON("success")
}
