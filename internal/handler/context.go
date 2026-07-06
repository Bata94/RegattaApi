package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
	"github.com/google/uuid"
)

type Handler func(c *Context) error

type AppError struct {
	Code        int
	StatusCode  int
	Message     string
	Details     string
	FieldErrors map[string]string
	FormComp    templ.Component
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) WithForm(comp templ.Component) *AppError {
	e.FormComp = comp
	return e
}

func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
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

type statusResponseWriter struct {
	http.ResponseWriter
	headersWritten bool
}

func (s *statusResponseWriter) WriteHeader(code int) {
	if s.headersWritten {
		return
	}
	s.headersWritten = true
	s.ResponseWriter.WriteHeader(code)
}

// Write marks headersWritten and delegates to underlying writer
func (s *statusResponseWriter) Write(b []byte) (int, error) {
	if !s.headersWritten {
		s.headersWritten = true
	}
	return s.ResponseWriter.Write(b)
}

func (s *statusResponseWriter) HeadersWritten() bool {
	return s.headersWritten
}

type HeaderTracker interface {
	HeadersWritten() bool
}

type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	pathParams map[string]string
	locals     map[string]any
	statusCode int
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:     &statusResponseWriter{ResponseWriter: w},
		Request:    r,
		pathParams: make(map[string]string),
		locals:     make(map[string]any),
		statusCode: http.StatusOK,
	}
}

func (c *Context) SetPathParams(params map[string]string) {
	c.pathParams = params
}

func (c *Context) Param(key string) string {
	if v, ok := c.pathParams[key]; ok {
		return v
	}
	if p := c.Request.Context().Value("pathParams"); p != nil {
		return p.(map[string]string)[key]
	}
	return ""
}

func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Context) FormValue(key string) string {
	return c.Request.FormValue(key)
}

func (c *Context) FormFile(key string) (string, []byte, error) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return "", nil, err
	}
	file, header, err := c.Request.FormFile(key)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", nil, err
	}
	return header.Filename, content, nil
}

const (
	uploadDirPerm  os.FileMode = 0o755
	uploadFilePerm os.FileMode = 0o644
)

func (c *Context) SaveFile(filename string, data []byte) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, uploadDirPerm); err != nil {
		return err
	}
	return os.WriteFile(filename, data, uploadFilePerm)
}

func (c *Context) BodyParser(v any) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func (c *Context) JSON(data any) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(c.Writer).Encode(data)
}

func (c *Context) JSONOk(data any) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(c.Writer).Encode(data)
}

func (c *Context) Status(code int) *Context {
	c.statusCode = code
	return c
}

func (c *Context) StatusCode() int {
	return c.statusCode
}

func (c *Context) SendString(msg string) error {
	c.Writer.Header().Set("Content-Type", "text/plain")
	_, err := c.Writer.Write([]byte(msg))
	return err
}

func (c *Context) Send(data []byte) error {
	_, err := c.Writer.Write(data)
	return err
}

func (c *Context) Redirect(location string, code int) error {
	http.Redirect(c.Writer, c.Request, location, code)
	return nil
}

func (c *Context) Path() string {
	return c.Request.URL.Path
}

func (c *Context) Method() string {
	return c.Request.Method
}

func (c *Context) Headers() http.Header {
	return c.Request.Header
}

func (c *Context) IsHtmxRequest() bool {
	return c.Request.Header.Get("HX-Request") == "true"
}

func (c *Context) Locals(key string, value any) {
	c.locals[key] = value
}

func (c *Context) GetLocals(key string) any {
	return c.locals[key]
}

func (c *Context) Cookie(name string) string {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (c *Context) SetCookie(name, value string, maxAge int) {
	secure := c.Request.TLS != nil
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}

func (c *Context) DeleteCookie(name string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:   name,
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})
}

func (c *Context) IP() string {
	if forwarded := c.Request.Header.Get("X-Forwarded-For"); forwarded != "" {
		if idx := strings.Index(forwarded, ","); idx > 0 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return strings.TrimSpace(forwarded)
	}
	return c.Request.RemoteAddr
}

func (c *Context) UserAgent() string {
	return c.Request.UserAgent()
}

func (c *Context) GetUUID(key string) (uuid.UUID, error) {
	param := c.Param(key)
	if param == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(param)
}

func (c *Context) GetQueryParam(key string) string {
	return c.Query(key)
}

func (c *Context) GetUUIDFromQuery(key string) (uuid.UUID, error) {
	param := c.Query(key)
	if param == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(param)
}

func (c *Context) BaseURL() string {
	return c.Request.URL.Scheme + "://" + c.Request.Host
}

func (c *Context) Hostname() string {
	return c.Request.URL.Host
}

func (c *Context) PathEscape(s string) string {
	return url.PathEscape(s)
}

func (c *Context) JoinPath(elem ...string) string {
	return path.Join(elem...)
}
