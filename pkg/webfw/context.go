package webfw

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const ctxKeyLocals contextKey = "webfw_locals"

func GetLocals(r *http.Request, key string) any {
	locals := r.Context().Value(ctxKeyLocals)
	if locals == nil {
		return nil
	}
	return locals.(map[string]any)[key]
}

func GetUser(r *http.Request) *jwt.Token {
	if locals := r.Context().Value(ctxKeyLocals); locals != nil {
		if m, ok := locals.(map[string]any); ok {
			if user, ok := m["user"]; ok {
				if token, ok := user.(*jwt.Token); ok {
					return token
				}
			}
		}
	}
	return nil
}

func GetUserClaims(r *http.Request) jwt.MapClaims {
	user := GetUser(r)
	if user == nil {
		return nil
	}
	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}
	return claims
}

func GetUserIDString(r *http.Request) string {
	claims := GetUserClaims(r)
	if claims == nil {
		return ""
	}
	if uuidStr, ok := claims["user_id"].(string); ok {
		return uuidStr
	}
	return ""
}

func IsLoggedIn(r *http.Request) bool {
	if locals := r.Context().Value(ctxKeyLocals); locals != nil {
		if m, ok := locals.(map[string]any); ok {
			if li, ok := m["logged_in"].(bool); ok {
				return li
			}
		}
	}
	return false
}

func GetCapabilities(r *http.Request) []string {
	if locals := r.Context().Value(ctxKeyLocals); locals != nil {
		if m, ok := locals.(map[string]any); ok {
			if caps, ok := m["capabilities"].([]string); ok {
				return caps
			}
		}
	}
	return nil
}

func HasCapability(r *http.Request, cap string) bool {
	caps := GetCapabilities(r)
	for _, c := range caps {
		if c == cap {
			return true
		}
	}
	return false
}

func HasAllCapabilities(r *http.Request, caps ...string) bool {
	for _, required := range caps {
		if !HasCapability(r, required) {
			return false
		}
	}
	return true
}

func WithAuthData(ctx context.Context, token *jwt.Token) context.Context {
	claims := token.Claims.(jwt.MapClaims)

	locals := make(map[string]any)
	locals["logged_in"] = true

	var capabilities []string
	if capsRaw, ok := claims["capabilities"].([]any); ok {
		for _, c := range capsRaw {
			if s, ok := c.(string); ok {
				capabilities = append(capabilities, s)
			}
		}
	}
	capabilities = append(capabilities, "allowed_logged_in")
	locals["capabilities"] = capabilities
	locals["user"] = token

	return context.WithValue(ctx, ctxKeyLocals, locals)
}
