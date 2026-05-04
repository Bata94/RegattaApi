package api_v1

import (
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
)

func GetAllUsers(c *handler.Context) error {
	uLs, err := crud.GetAllUsers()
	if err != nil {
		return err
	}

	return c.JSON(uLs)
}

func GetUser(c *handler.Context) error {
	uuid, err := c.GetUUID("ulid")
	if err != nil {
		return &handler.Error{StatusCode: 400, Message: err.Error()}
	}

	u, err := crud.GetUser(uuid)
	if err != nil {
		return err
	}

	return c.JSON(u.ToReturnUser())
}

func GetUserByName(c *handler.Context) error {
	name := c.Param("name")
	if name == "" {
		return &handler.Error{StatusCode: 400, Message: "name parameter required"}
	}

	u, err := crud.GetUserByUsername(name)
	if err != nil {
		return err
	}

	return c.JSON(u.ToReturnUser())
}

func CreateUser(c *handler.Context) error {
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

func GetAllUsersGroups(c *handler.Context) error {
	ugLs, err := crud.GetAllUsersGroups()
	if err != nil {
		return err
	}

	return c.JSON(ugLs)
}

func GetUsersGroup(c *handler.Context) error {
	uuid, err := c.GetUUID("ulid")
	if err != nil {
		return &handler.Error{StatusCode: 400, Message: err.Error()}
	}

	ug, err := crud.GetUsersGroup(uuid)
	if err != nil {
		return err
	}

	return c.JSON(ug)
}

func GetUsersGroupByName(c *handler.Context) error {
	name := c.Param("name")
	if name == "" {
		return &handler.Error{StatusCode: 400, Message: "name parameter required"}
	}

	ug, err := crud.GetUsersGroupByName(name)
	if err != nil {
		return err
	}

	return c.JSON(ug)
}
