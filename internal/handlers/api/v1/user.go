package api_v1

import (
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/gofiber/fiber/v2"
)

func GetAllUsers(c *fiber.Ctx) error {
	uLs, err := crud.GetAllUsers()
	if err != nil {
		return err
	}

	return c.JSON(uLs)
}

func GetUser(c *fiber.Ctx) error {
	ulid, err := api.GetUlidFromCtx(c)
	if err != nil {
		return err
	}

	u, err := crud.GetUser(*ulid)
	if err != err {
		return err
	}

	return c.JSON(u.ToReturnUser())
}

func GetUserByName(c *fiber.Ctx) error {
	name, err := api.GetStrParamFromCtx(c, "name")
	if err != nil {
		return err
	}

	u, err := crud.GetUserByUsername(name)
	if err != err {
		return err
	}

	return c.JSON(u.ToReturnUser())
}

func CreateUser(c *fiber.Ctx) error {
	uParams := new(crud.CreateUserParams)
	err := c.BodyParser(&uParams)
	if err != nil {
		return err
	}

	u, err := crud.CreateUser(*uParams)
	if err != nil {
		return err
	}

	return c.JSON(u)
}

func GetAllUsersGroups(c *fiber.Ctx) error {
	ugLs, err := crud.GetAllUsersGroups()
	if err != nil {
		return err
	}

	return c.JSON(ugLs)
}

func GetUsersGroup(c *fiber.Ctx) error {
	ulid, err := api.GetUlidFromCtx(c)
	if err != nil {
		return err
	}

	ug, err := crud.GetUsersGroup(*ulid)
	if err != err {
		return err
	}

	return c.JSON(ug)
}

func GetUsersGroupByName(c *fiber.Ctx) error {
	name, err := api.GetStrParamFromCtx(c, "name")
	if err != nil {
		return err
	}

	ug, err := crud.GetUsersGroupByName(name)
	if err != err {
		return err
	}

	return c.JSON(ug)
}
