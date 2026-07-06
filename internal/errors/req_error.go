package apierr

import (
	"fmt"
	"net/http"
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

func ReqErrorFrom(err *ReqError, msg, detail string) *ReqError {
	err.Msg = msg
	err.Details = detail
	return err
}

func NewReqError(title, msg string, statusCode int) *ReqError {
	err := new(ReqError)
	err.Title = title
	err.Msg = msg
	err.StatusCode = statusCode
	return err
}

var (
	BAD_REQUEST           = ReqError{Code: 400, StatusCode: http.StatusBadRequest, Title: "Missing params/body"}
	NOT_FOUND             = ReqError{Code: 404, StatusCode: http.StatusNotFound, Title: "Not found"}
	UNAUTHORIZED          = ReqError{Code: 401, StatusCode: http.StatusUnauthorized, Title: "Unauthorized"}
	FORBIDDEN             = ReqError{Code: 403, StatusCode: http.StatusForbidden, Title: "Forbidden"}
	NOT_ACCEPTABLE        = ReqError{Code: 406, StatusCode: http.StatusNotAcceptable, Title: "Not acceptable"}
	INTERNAL_SERVER_ERROR = ReqError{Code: 500, StatusCode: http.StatusInternalServerError, Title: "Internal Server error"}

	VALIDATION_ERROR = ReqError{Code: 1000, StatusCode: http.StatusBadRequest, Title: "Validation error"}
	TIME_PARSE_ERROR = ReqError{Code: 1001, StatusCode: http.StatusInternalServerError, Title: ""}

	AUTH_LOGIN_WRONG_PASSWORD = ReqError{Code: 1050, StatusCode: http.StatusUnauthorized, Title: "Wrong password"}
	WRONG_REFRESH_TOKEN       = ReqError{Code: 1051, StatusCode: http.StatusUnauthorized, Title: "Wrong refresh token"}
	TOKEN_GENERATION_ERROR    = ReqError{Code: 1052, StatusCode: http.StatusInternalServerError, Title: "Failed to generate token"}
	TOKEN_INVALID             = ReqError{Code: 1053, StatusCode: http.StatusUnauthorized, Title: "Failed to validate token"}
	AUTH_LOGIN_USER_NOT_ACTIVE= ReqError{Code: 1054, StatusCode: http.StatusUnauthorized, Title: "User Account is disabled, please contact the Admin!"}

	ACCOUNT_WITH_EMAIL_ALREADY_EXISTS = ReqError{Code: 1100, StatusCode: http.StatusBadRequest, Title: "An account with this email already exists"}
)
