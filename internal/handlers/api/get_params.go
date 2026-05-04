package api

import (
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/google/uuid"
)

func GetQueryParamBoolFromCtx(c *handler.Context, param string, def bool) bool {
	str := c.Query(param)
	if str == "true" || str == "yes" || str == "1" {
		return true
	} else if str == "false" || str == "no" || str == "0" {
		return false
	}
	return def
}

func GetUuidFromCtx(c *handler.Context) (*uuid.UUID, error) {
	uuidStr := c.Param("uuid")
	if uuidStr == "" {
		return nil, &ReqError{Code: 404, StatusCode: 404, Title: "ID not found", Msg: "", Details: "", Data: uuidStr}
	}

	u, err := uuid.Parse(uuidStr)
	if err != nil {
		retErr := BAD_REQUEST
		retErr.Msg = "UUID not parsable!"
		return nil, &retErr
	}

	return &u, nil
}
