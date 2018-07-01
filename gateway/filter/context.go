package filter

import (
	"fmt"
	"net/http"
	"time"
)

// Context filter context interface
type Context interface {
	StartAt() time.Time
	EndAt() time.Time

	SetAttr(key string, value interface{})
	GetAttr(key string) interface{}

	HTTPRequest() *http.Request
}

type context struct {
	attrs map[string]interface{}
	req   *http.Request
}

// NewContext NewContext
func NewContext(req *http.Request) Context {
	ctx := &context{
		req:   req,
		attrs: make(map[string]interface{}),
	}
	ctx.attrs["remoteURL"] = fmt.Sprintf("%s%s", req.Host, req.RequestURI)
	return ctx
}

func (c *context) HTTPRequest() *http.Request {
	return c.req
}

func (c *context) StartAt() time.Time {
	return time.Now()
}

func (c *context) EndAt() time.Time {
	return time.Now()
}

func (c *context) SetAttr(key string, value interface{}) {

}

func (c *context) GetAttr(key string) interface{} {
	if val, ok := c.attrs[key]; ok {
		return val
	}
	return nil
}
