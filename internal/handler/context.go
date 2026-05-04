package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/google/uuid"
)

type Handler func(c *Context) error

type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string {
	return e.Message
}

type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	pathParams map[string]string
	locals     map[string]interface{}
	statusCode int
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:     w,
		Request:    r,
		pathParams: make(map[string]string),
		locals:     make(map[string]interface{}),
		statusCode: http.StatusOK,
	}
}

func (c *Context) SetPathParams(params map[string]string) {
	c.pathParams = params
}

func (c *Context) Param(key string) string {
	return c.pathParams[key]
}

func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Context) FormValue(key string) string {
	return c.Request.FormValue(key)
}

func (c *Context) BodyParser(v interface{}) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func (c *Context) JSON(data interface{}) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(c.Writer).Encode(data)
}

func (c *Context) JSONOk(data interface{}) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(http.StatusOK)
	return json.NewEncoder(c.Writer).Encode(data)
}

func (c *Context) Status(code int) *Context {
	c.statusCode = code
	c.Writer.WriteHeader(code)
	return c
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

func (c *Context) Locals(key string, value interface{}) {
	c.locals[key] = value
}

func (c *Context) GetLocals(key string) interface{} {
	return c.locals[key]
}

func (c *Context) Cookie(name string) string {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (c *Context) IP() string {
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
