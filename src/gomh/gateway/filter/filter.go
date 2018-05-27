package filter

import (
	"time"
)

type FilterResp struct {
	Code    int
	Message string
}

type Context interface {
	StartAt() time.Time
	EndAt() time.Time

	SetAttr(key string, value interface{})
	GetAttr(key string) interface{}
}

type Filter interface {
	Name() string
	Init(config string) error
	AsBegin(c Context) (FilterResp, error)
	AsEnd(c Context) (FilterResp, error)
}

type DefaultFilter struct{}

func (f DefaultFilter) Name() string {
	return "defaultfilter"
}

func (f DefaultFilter) Init() error {
	return nil
}

func (f DefaultFilter) AsBegin(c Context) (FilterResp, error) {
	return &FilterResp{}, nil
}

func (f DefaultFilter) AsEnd(c Context) (FilterResp, error) {
	return &FilterResp{}, nil
}
