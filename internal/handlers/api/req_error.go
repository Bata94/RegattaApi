package api

import apierr "github.com/bata94/RegattaApi/internal/errors"

type ReqError = apierr.ReqError

var (
	BAD_REQUEST                    = apierr.BAD_REQUEST
	NOT_FOUND                      = apierr.NOT_FOUND
	UNAUTHORIZED                   = apierr.UNAUTHORIZED
	FORBIDDEN                      = apierr.FORBIDDEN
	NOT_ACCEPTABLE                 = apierr.NOT_ACCEPTABLE
	INTERNAL_SERVER_ERROR          = apierr.INTERNAL_SERVER_ERROR
	VALIDATION_ERROR               = apierr.VALIDATION_ERROR
	TIME_PARSE_ERROR               = apierr.TIME_PARSE_ERROR
	AUTH_LOGIN_WRONG_PASSWORD      = apierr.AUTH_LOGIN_WRONG_PASSWORD
	WRONG_REFRESH_TOKEN            = apierr.WRONG_REFRESH_TOKEN
	TOKEN_GENERATION_ERROR         = apierr.TOKEN_GENERATION_ERROR
	TOKEN_INVALID                  = apierr.TOKEN_INVALID
	AUTH_LOGIN_USER_NOT_ACTIVE     = apierr.AUTH_LOGIN_USER_NOT_ACTIVE
	ACCOUNT_WITH_EMAIL_ALREADY_EXISTS = apierr.ACCOUNT_WITH_EMAIL_ALREADY_EXISTS
)

func ReqErrorFrom(err *ReqError, msg, detail string) *ReqError {
	return apierr.ReqErrorFrom(err, msg, detail)
}

func NewReqError(title, msg string, statusCode int) *ReqError {
	return apierr.NewReqError(title, msg, statusCode)
}
