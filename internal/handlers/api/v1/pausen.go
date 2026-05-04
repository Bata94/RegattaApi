package api_v1

import (
	"strconv"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/sqlc"
)

func GetAllPausen(c *handler.Context) error {
	pLs, err := crud.GetAllPausen()
	if err != nil {
		return err
	}
	if pLs == nil {
		pLs = []crud.Pause{}
	}
	return c.JSON(pLs)
}

func GetPause(c *handler.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return &handler.Error{StatusCode: 400, Message: "invalid id"}
	}

	p, err := crud.GetPause(id)
	if err != nil {
		return err
	}
	return c.JSON(p)
}

func DeletePause(c *handler.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return &handler.Error{StatusCode: 400, Message: "invalid id"}
	}

	err = crud.DeletePause(int32(id))
	if err != nil {
		return err
	}
	return c.JSON("Pause erfolgreich gelöscht!")
}

func CreatePause(c *handler.Context) error {
	params := new(sqlc.CreatePauseParams)
	err := c.BodyParser(params)
	if err != nil {
		return err
	}

	p, err := crud.CreatePause(*params)
	if err != nil {
		return err
	}

	return c.JSON(p)
}

func UpdatePause(c *handler.Context) error {
	params := new(sqlc.UpdatePauseParams)
	err := c.BodyParser(params)
	if err != nil {
		return err
	}

	p, err := crud.UpdatePause(*params)
	if err != nil {
		return err
	}

	return c.JSON(p)
}
