package api_v1

import (
	"net/http"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/golang-jwt/jwt/v5"
)

func Login(c *handler.Context) error {
	loginParams := new(crud.LoginParams)
	if err := c.BodyParser(loginParams); err != nil {
		return &handler.Error{StatusCode: http.StatusBadRequest, Message: err.Error()}
	}

	u, err := crud.AuthLogin(c.Request.Context(), *loginParams)
	if err != nil {
		return err
	}

	c.SetCookie("auth_token", u.Jwt.Token, 72*60*60)
	c.Writer.Header().Set("HX-Redirect", "/")
	return c.JSON(u)
}

func Logout(c *handler.Context) error {
	return c.JSON("Logout successful!")
}

func AuthValidate(c *handler.Context) error {
	return c.JSON("Auth successful!")
}

func AuthMe(c *handler.Context) error {
	user := c.GetLocals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	uuidStr := claims["user_id"].(string)

	userID, err := crud.ParseUUID(uuidStr)
	if err != nil {
		return &handler.Error{StatusCode: http.StatusUnauthorized, Message: "Invalid token"}
	}

	u, err := crud.GetUser(c.Request.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(*u)
}
