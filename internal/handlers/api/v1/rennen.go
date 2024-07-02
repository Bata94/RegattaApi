package api_v1

import (
	"strings"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetRennen(c *fiber.Ctx) error {
	uuid, err := api.GetUuidFromCtx(c)
	if err != nil {
		return err
	}

	r, err := crud.GetRennen(*uuid)
	if err != nil {
		return err
	}

	return c.JSON(r)
}

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

func GetAllRennenByWettkampf(c *fiber.Ctx) error {
	wettkampfStr, err := api.GetStrParamFromCtx(c, "wettkampf")
	if err != nil {
		retErr := api.BAD_REQUEST
		retErr.Details = err.Error()
		return &retErr
	}
	caser := cases.Title(language.German)
	wettkampfStr = caser.String(wettkampfStr)
	wettkampf := sqlc.Wettkampf(wettkampfStr)

	showEmptyStr := c.Query("showEmpty", "")
	showEmpty := false
	if strings.ToLower(showEmptyStr) == "true" || showEmptyStr == "yes" || showEmptyStr == "1" {
		showEmpty = true
	}

	showStartedStr := c.Query("showStarted", "")
	showStarted := false
	if strings.ToLower(showStartedStr) == "true" || showStartedStr == "yes" || showStartedStr == "1" {
		showStarted = true
	}

	rLs, err := crud.GetAllRennenByWettkampf(wettkampf, showStarted, showEmpty)
	if err != nil {
		return err
	}

	return c.JSON(rLs)
}
