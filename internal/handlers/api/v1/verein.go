package api_v1

import (
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
)

func GetAllVerein(c *handler.Context) error {
	vLs, err := crud.GetAllVerein()
	if err != nil {
		return err
	}

	return c.JSON(vLs)
}

func GetVerein(c *handler.Context) error {
	uuid, err := c.GetUUID("uuid")
	if err != nil {
		return &handler.Error{StatusCode: 400, Message: err.Error()}
	}

	v, err := crud.GetVerein(uuid)
	if err != nil {
		return err
	}

	return c.JSON(v)
}

func GetAllAthletenForVerein(c *handler.Context) error {
	uuid, err := c.GetUUID("uuid")
	if err != nil {
		return &handler.Error{StatusCode: 400, Message: err.Error()}
	}

	aLs, err := crud.GetAllAthletenForVerein(uuid)
	if err != nil {
		return err
	}

	return c.JSON(aLs)
}

func GetAllAthletenForVereinMissStartber(c *handler.Context) error {
	uuid, err := c.GetUUID("uuid")
	if err != nil {
		return &handler.Error{StatusCode: 400, Message: err.Error()}
	}

	aLs, err := crud.GetAllAthletenForVereinMissStartber(uuid)
	if err != nil {
		return err
	}

	return c.JSON(aLs)
}

func GetAllAthletenForVereinWaage(c *handler.Context) error {
	uuid, err := c.GetUUID("uuid")
	if err != nil {
		return &handler.Error{StatusCode: 400, Message: err.Error()}
	}

	aLs, err := crud.GetAllAthletenForVereinWaage(uuid)
	if err != nil {
		return err
	}

	return c.JSON(aLs)
}
