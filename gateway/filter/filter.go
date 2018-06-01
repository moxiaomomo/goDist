package filter

import (
	"time"
)

const (
	// FilteredPassed request can pass to server
	FilteredPassed = 0
	// FilteredFailed request cannot pass to server
	FilteredFailed = 1
)

// Response filter response
type Response struct {
	Time    time.Time
	Code    int
	Message string
}

// Context filter context interface
type Context interface {
	StartAt() time.Time
	EndAt() time.Time

	SetAttr(key string, value interface{})
	GetAttr(key string) interface{}
}

// Filter base filter interface
type Filter interface {
	Name() string
	Init(config string) error
	AsBegin(c Context) Response
	AsEnd(c Context) Response
}

// DefaultFilter base filter
type DefaultFilter struct{}

// Name filter's name
func (f *DefaultFilter) Name() string {
	return "defaultfilter"
}

// Init filter initialization
func (f *DefaultFilter) Init() error {
	return nil
}

// AsBegin execute at the beginning
func (f *DefaultFilter) AsBegin(c Context) Response {
	return Response{
		Time: time.Now(),
		Code: FilteredPassed,
	}
}

// AsEnd execute at the end
func (f *DefaultFilter) AsEnd(c Context) Response {
	return Response{
		Time: time.Now(),
		Code: FilteredPassed,
	}
}
