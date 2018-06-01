package filter

import (
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
	req *http.Request
}

// NewContext NewContext
func NewContext(req *http.Request) Context {
	return &context{req: req}
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
	return nil
}
