package api

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type ReqError struct {
	Code       int
	StatusCode int
	Title      string
	Msg        string
	Details    string
	Data       interface{}
}

func (r *ReqError) Error() string {
	return fmt.Sprintf("statusCode: %d | code: %d | message: %v", r.StatusCode, r.Code, r.Msg)
}

// Error from var Error but pass details
func ReqErrorFrom(err *ReqError, msg, detail string) *ReqError {
	err.Msg = msg
	err.Details = detail
	return err
}

var (
	BAD_REQUEST           = ReqError{Code: 400, StatusCode: fiber.StatusBadRequest, Title: "Missing params/body"}
	NOT_FOUND             = ReqError{Code: 404, StatusCode: fiber.StatusNotFound, Title: "Not found"}
	UNAUTHORIZED          = ReqError{Code: 401, StatusCode: fiber.StatusUnauthorized, Title: "Unauthorized"}
	FORBIDDEN             = ReqError{Code: 403, StatusCode: fiber.StatusForbidden, Title: "Forbidden"}
	NOT_ACCEPTABLE        = ReqError{Code: 406, StatusCode: fiber.StatusNotAcceptable, Title: "Not acceptable"}
	INTERNAL_SERVER_ERROR = ReqError{Code: 500, StatusCode: fiber.StatusInternalServerError, Title: "Internal Server error"}

	// The validation error can be used but the message should be overwritten
	VALIDATION_ERROR = ReqError{Code: 1000, StatusCode: fiber.StatusBadRequest, Title: "Validation error"}
	TIME_PARSE_ERROR = ReqError{Code: 1001, StatusCode: fiber.StatusInternalServerError, Title: ""}

	AUTH_LOGIN_WRONG_PASSWORD = ReqError{Code: 1050, StatusCode: fiber.StatusUnauthorized, Title: "Wrong password"}
	WRONG_REFRESH_TOKEN       = ReqError{Code: 1051, StatusCode: fiber.StatusUnauthorized, Title: "Wrong refresh token"}
	TOKEN_GENERATION_ERROR    = ReqError{Code: 1052, StatusCode: fiber.StatusInternalServerError, Title: "Failed to generate token"}

	ACCOUNT_WITH_EMAIL_ALREADY_EXISTS = ReqError{Code: 1100, StatusCode: fiber.StatusBadRequest, Title: "An account with this email already exists"}
)
