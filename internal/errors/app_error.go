package apierr

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
)

type AppError struct {
	Code        int
	StatusCode  int
	Message     string
	Details     string
	FieldErrors map[string]string
	FormComp    templ.Component
	Data        any
}

func (e *AppError) Error() string {
	return fmt.Sprintf("statusCode: %d | code: %d | message: %s", e.StatusCode, e.Code, e.Message)
}

func (e *AppError) WithForm(comp templ.Component) *AppError {
	e.FormComp = comp
	return e
}

func (e *AppError) WithDetails(details string) *AppError {
	cp := *e
	cp.Details = details
	return &cp
}

func NotFound(msg string) *AppError {
	return &AppError{Code: 404, StatusCode: http.StatusNotFound, Message: msg}
}

func BadRequest(msg string) *AppError {
	return &AppError{Code: 400, StatusCode: http.StatusBadRequest, Message: msg}
}

func Unauthorized(msg string) *AppError {
	return &AppError{Code: 401, StatusCode: http.StatusUnauthorized, Message: msg}
}

func Forbidden(msg string) *AppError {
	return &AppError{Code: 403, StatusCode: http.StatusForbidden, Message: msg}
}

func NotAcceptable(msg string) *AppError {
	return &AppError{Code: 406, StatusCode: http.StatusNotAcceptable, Message: msg}
}

func InternalError(msg string) *AppError {
	return &AppError{Code: 500, StatusCode: http.StatusInternalServerError, Message: msg}
}

func ValidationError(fieldErrors map[string]string) *AppError {
	return &AppError{Code: 1000, StatusCode: http.StatusBadRequest, Message: "Validation error", FieldErrors: fieldErrors}
}

func OK(msg string) *AppError {
	return &AppError{Code: 200, StatusCode: http.StatusOK, Message: msg}
}

var (
	ErrNotFound      = &AppError{Code: 404, StatusCode: http.StatusNotFound, Message: "Ressource nicht gefunden"}
	ErrBadRequest    = &AppError{Code: 400, StatusCode: http.StatusBadRequest, Message: "Ungültige Anfrage"}
	ErrUnauthorized  = &AppError{Code: 401, StatusCode: http.StatusUnauthorized, Message: "Unautorisierter Zugriff"}
	ErrForbidden     = &AppError{Code: 403, StatusCode: http.StatusForbidden, Message: "Zugriff verweigert"}
	ErrNotAcceptable = &AppError{Code: 406, StatusCode: http.StatusNotAcceptable, Message: "Nicht akzeptabel"}
	ErrInternal      = &AppError{Code: 500, StatusCode: http.StatusInternalServerError, Message: "Interner Serverfehler"}

	ErrAuthLoginUserNotActive = &AppError{Code: 1054, StatusCode: http.StatusUnauthorized, Message: "User Account is disabled, please contact the Admin!"}
	ErrAuthLoginWrongPassword = &AppError{Code: 1050, StatusCode: http.StatusUnauthorized, Message: "Wrong password"}
	ErrTokenGeneration        = &AppError{Code: 1052, StatusCode: http.StatusInternalServerError, Message: "Failed to generate token"}
	ErrTokenInvalid           = &AppError{Code: 1053, StatusCode: http.StatusUnauthorized, Message: "Failed to validate token"}
	ErrWrongRefreshToken      = &AppError{Code: 1051, StatusCode: http.StatusUnauthorized, Message: "Wrong refresh token"}
	ErrAccountAlreadyExists   = &AppError{Code: 1100, StatusCode: http.StatusBadRequest, Message: "An account with this email already exists"}
	ErrTimeParse              = &AppError{Code: 1001, StatusCode: http.StatusInternalServerError, Message: ""}
)
