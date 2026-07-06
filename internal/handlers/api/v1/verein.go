package api_v1

import (
	"net/http"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
)

func GetAllVerein(c *handler.Context) error {
	vLs, err := crud.GetAllVerein(c.Request.Context(), )
	if err != nil {
		return err
	}

	return c.JSON(vLs)
}

func GetVerein(c *handler.Context) error {
	uuid, err := c.GetUUID("uuid")
	if err != nil {
		return &handler.Error{StatusCode: http.StatusBadRequest, Message: err.Error()}
	}

	v, err := crud.GetVerein(c.Request.Context(), uuid)
	if err != nil {
		return err
	}

	return c.JSON(v)
}

func GetAllAthletenForVerein(c *handler.Context) error {
	uuid, err := c.GetUUID("uuid")
	if err != nil {
		return &handler.Error{StatusCode: http.StatusBadRequest, Message: err.Error()}
	}

	aLs, err := crud.GetAllAthletenForVerein(c.Request.Context(), uuid)
	if err != nil {
		return err
	}

	return c.JSON(aLs)
}

func GetAllAthletenForVereinMissStartber(c *handler.Context) error {
	uuid, err := c.GetUUID("uuid")
	if err != nil {
		return &handler.Error{StatusCode: http.StatusBadRequest, Message: err.Error()}
	}

	aLs, err := crud.GetAllAthletenForVereinMissStartber(c.Request.Context(), uuid)
	if err != nil {
		return err
	}

	return c.JSON(aLs)
}

func GetAllAthletenForVereinWaage(c *handler.Context) error {
	uuid, err := c.GetUUID("uuid")
	if err != nil {
		return &handler.Error{StatusCode: http.StatusBadRequest, Message: err.Error()}
	}

	aLs, err := crud.GetAllAthletenForVereinWaage(c.Request.Context(), uuid)
	if err != nil {
		return err
	}

	return c.JSON(aLs)
}
