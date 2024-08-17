package api_v1

import (
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/oklog/ulid/v2"

	// jwtware "github.com/gofiber/contrib/jwt"
	"github.com/golang-jwt/jwt/v5"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func Login(c *fiber.Ctx) error {
	loginParams := new(crud.LoginParams)
	if err := c.BodyParser(&loginParams); err != nil {
		retErr := api.BAD_REQUEST
		retErr.Details = err.Error()
		return &retErr
	}

	log.Debug("Login attempt, Username: ", loginParams.Username, " Password: ", loginParams.Password)

	u, err := crud.AuthLogin(*loginParams)
	if err != nil {
		return err
	}

	return api.JSON(c, u)
}

// TODO: Implement
func Logout(c *fiber.Ctx) error {
	return api.JSON(c, "Logout successful!")
}

func AuthValidate(c *fiber.Ctx) error {
	return api.JSON(c, "Auth successful!")
}

func AuthMe(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	ulidStr := claims["user_id"].(string)

	ulid, err := ulid.Parse(ulidStr)
	if err != nil {
		return &api.UNAUTHORIZED
	}

	userRaw, err := crud.GetUser(ulid)

	u := crud.ReturnUser{
		Ulid:      userRaw.Ulid,
		Username:  userRaw.Username,
		UserGroup: userRaw.UserGroup,
	}

	return api.JSON(c, u)
}
