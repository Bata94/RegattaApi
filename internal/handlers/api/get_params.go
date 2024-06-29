package api

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

func GetId64FromCtx(c *fiber.Ctx) (int64, error) {
	idStr := c.Params("id", "")
	if idStr == "" {
		return 0, &ReqError{
			Code:       404,
			StatusCode: fiber.StatusNotFound,
			Title:      "ID not found in URL",
			Msg:        "For this Route you need to provide a ID in the URL!",
			Details:    "",
			Data:       nil,
		}
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return 0, &ReqError{
			Code:       404,
			StatusCode: fiber.StatusNotFound,
			Title:      "ID not parsable",
			Msg:        idStr,
			Details:    "",
			Data:       nil,
		}
	}

	return id, nil
}

func GetId32FromCtx(c *fiber.Ctx) (int32, error) {
	id64, err := GetId64FromCtx(c)
	if err != nil {
		return 0, err
	}

	return int32(id64), nil
}

func GetStrParamFromCtx(c *fiber.Ctx, param string) (string, error) {
	str := c.Params(param, "")
	if str == "" {
		return "", &ReqError{Code: 404, StatusCode: fiber.StatusNotFound, Title: "ID not found", Msg: "", Details: "", Data: str}
	}
	return str, nil
}

func GetUuidFromCtx(c *fiber.Ctx) (*uuid.UUID, error) {
	uuidStr := c.Params("uuid", "")
	if uuidStr == "" {
		return nil, &ReqError{Code: 404, StatusCode: fiber.StatusNotFound, Title: "ID not found", Msg: "", Details: "", Data: uuidStr}
	}

	uuid, err := uuid.Parse(uuidStr)
	if err != nil {
		retErr := BAD_REQUEST
		retErr.Msg = "UUID not parsable!"
		return nil, &retErr
	}

	return &uuid, nil
}

func GetUlidFromCtx(c *fiber.Ctx) (*ulid.ULID, error) {
	ulidStr := c.Params("ulid", "")
	if ulidStr == "" {
		return nil, &ReqError{Code: 404, StatusCode: fiber.StatusNotFound, Title: "ID not found", Msg: "", Details: "", Data: ulidStr}
	}

	ulid, err := ulid.Parse(ulidStr)
	if err != nil {
		return &ulid, &ReqError{Code: 404, StatusCode: fiber.StatusNotFound, Title: "ID not parsable", Msg: ulidStr, Details: "", Data: nil}
	}

	return &ulid, nil
}
