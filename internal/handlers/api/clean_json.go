package api

import (
	"github.com/gofiber/fiber/v2"
)

func JSON(c *fiber.Ctx, data interface{}) error {
	// jsonData, err := json.Marshal(data)
	// if err != nil {
	// 	return err
	// }
	//
	// log.Debug("Returning JSON: ", string(jsonData))
	// TODO: rm null key, value pairs

	// probably doubling JSON Encoding
	return c.JSON(data)
}
