package filter

import (
	"net/http"
	"time"
)

// HeaderFilter rate-limiting filter
type HeaderFilter struct {
	DefaultFilter
	supportMethod map[string]interface{}
}

// Init filter initialization
func (f *HeaderFilter) Init(config string) error {
	f.supportMethod = map[string]interface{}{
		http.MethodGet:  struct{}{},
		http.MethodPost: struct{}{},
	}
	return nil
}

// Name returns HeaderFilter's name
func (f *HeaderFilter) Name() string {
	return "HeaderFilter"
}

// AsBegin execute at the beginning
func (f *HeaderFilter) AsBegin(c Context) Response {
	if _, ok := f.supportMethod[c.HTTPRequest().Method]; ok {
		return Response{
			Time: time.Now(),
			Code: FilteredPassed,
		}
	}
	return Response{
		Time:    time.Now(),
		Code:    FilteredFailed,
		Message: "unsupported http method",
	}
}

// AsEnd execute at the end
func (f *HeaderFilter) AsEnd(c Context) Response {
	return Response{
		Time: time.Now(),
		Code: FilteredPassed,
	}
}
