package api_v1

import (
	"github.com/gofiber/fiber/v2"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
)

func GetAllPausen(c *fiber.Ctx) error {
  pLs, err := crud.GetAllPausen()
  if err != nil {
    return err
  }
  if pLs == nil {
    pLs = []*sqlc.Pause{}
  }
  return c.JSON(pLs)
}

func GetPause(c *fiber.Ctx) error {
  id, err := api.GetId32FromCtx(c)
  if err != nil {
    return err
  }

  p, err := crud.GetPause(id)
  if err != nil {
    return err
  }
  return c.JSON(p)
}

func CreatePause(c *fiber.Ctx) error {
  params := new(sqlc.CreatePauseParams)
  err := c.BodyParser(params)
  if err != nil {
    return err
  }

  p, err := crud.CreatePause(*params)

  return c.JSON(p)
}

func UpdatePause(c *fiber.Ctx) error {
  params := new(sqlc.UpdatePauseParams)
  err := c.BodyParser(params)
  if err != nil {
    return err
  }

  p, err := crud.UpdatePause(*params)

  return c.JSON(p)
}
