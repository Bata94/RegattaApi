package api_v1

import (
	"strings"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/gofiber/fiber/v2"
)

func GetAllRennen(c *fiber.Ctx) error {
	withMeldStr := c.Query("withMeld", "")
	withMeld := false
	if strings.ToLower(withMeldStr) == "true" || withMeldStr == "yes" || withMeldStr == "1" {
		withMeld = true
	}

	showEmptyStr := c.Query("showEmpty", "")
	showEmpty := false
	if strings.ToLower(showEmptyStr) == "true" || showEmptyStr == "yes" || showEmptyStr == "1" {
		showEmpty = true
	}

	if withMeld {
		rLs, err := crud.GetAllRennenWithMeld(showEmpty)
		if err != nil {
			return err
		}
		return c.JSON(rLs)
	}

	rLs, err := crud.GetAllRennen()
	if err != nil {
		return err
	}
	return c.JSON(rLs)
}
