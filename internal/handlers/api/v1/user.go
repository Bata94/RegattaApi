package api_v1

import (
	"net/http"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
)

func GetAllUsers(c *handler.Context) error {
	uLs, err := crud.GetAllUsers(c.Request.Context(), )
	if err != nil {
		return err
	}

	return c.JSON(uLs)
}

func GetUser(c *handler.Context) error {
	uuid, err := c.GetUUID("uuid")
	if err != nil {
		return &handler.Error{StatusCode: http.StatusBadRequest, Message: err.Error()}
	}

	u, err := crud.GetUser(c.Request.Context(), uuid)
	if err != nil {
		return err
	}

	return c.JSON(*u)
}

func GetUserByName(c *handler.Context) error {
	name := c.Param("name")
	if name == "" {
		return &handler.Error{StatusCode: http.StatusBadRequest, Message: "name parameter required"}
	}

	u, err := crud.GetUserByUsername(c.Request.Context(), name)
	if err != nil {
		return err
	}

	return c.JSON(*u)
}

func CreateUser(c *handler.Context) error {
	uParams := new(crud.CreateUserParams)
	err := c.BodyParser(&uParams)
	if err != nil {
		return err
	}

	u, err := crud.CreateUser(c.Request.Context(), *uParams)
	if err != nil {
		return err
	}

	return c.JSON(u)
}

func GetAllUsersGroups(c *handler.Context) error {
	ugLs, err := crud.GetAllUsersGroups(c.Request.Context(), )
	if err != nil {
		return err
	}

	return c.JSON(ugLs)
}

func GetUsersGroup(c *handler.Context) error {
	uuid, err := c.GetUUID("uuid")
	if err != nil {
		return &handler.Error{StatusCode: http.StatusBadRequest, Message: err.Error()}
	}

	ug, err := crud.GetUsersGroup(c.Request.Context(), uuid)
	if err != nil {
		return err
	}

	return c.JSON(ug)
}

func GetUsersGroupByName(c *handler.Context) error {
	name := c.Param("name")
	if name == "" {
		return &handler.Error{StatusCode: http.StatusBadRequest, Message: "name parameter required"}
	}

	ug, err := crud.GetUsersGroupByName(c.Request.Context(), name)
	if err != nil {
		return err
	}

	return c.JSON(ug)
}
