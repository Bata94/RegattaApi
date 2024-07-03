package api_v1

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

func TestHandler(c *fiber.Ctx) error {
	revUUID := uuid.MustParse("018a65b6-36fc-7112-96a1-d0b0aac587e2")
	newUUID, _ := uuid.NewV7()

	log.Debug(revUUID, " ", newUUID)
	log.Debug(revUUID.ClockSequence(), " ", newUUID.ClockSequence())

	log.Debug(revUUID.ClockSequence() < newUUID.ClockSequence())

	time.Sleep(time.Second * 60)

	return c.JSON("success")
}
