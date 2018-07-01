package filter

import (
	"time"

	cbreaker "github.com/moxiaomomo/circuitbreaker"
)

// CircuitBreakerFilter circuit-breaker filter
type CircuitBreakerFilter struct {
	DefaultFilter
	CBreaker *cbreaker.Circuits
}

// Init filter initialization
func (c *CircuitBreakerFilter) Init(config string) error {
	if c.CBreaker == nil {
		c.CBreaker = cbreaker.NewCirucuitBreaker(time.Second, 1000, 10)
	}
	return nil
}

// Name returns CircuitBreakerFilter's name
func (c *CircuitBreakerFilter) Name() string {
	return "CircuitBreakerFilter"
}

// AsBegin execute at the beginning
func (c *CircuitBreakerFilter) AsBegin(ctx Context) Response {
	rurl := ctx.GetAttr("RemoteURL")
	if rurl == nil {
		rurl = "DEFAULTCOMMAND"
	}
	c.CBreaker.RegisterCommandAsDefault(rurl.(string))

	if c.CBreaker.AllowExec(rurl.(string)) {
		return Response{
			Time: time.Now(),
			Code: FilteredPassed,
		}
	}
	return Response{
		Time:    time.Now(),
		Code:    FilteredFailed,
		Message: "Request failed as circuit-breaker open",
	}
}

// AsEnd execute at the end
func (c *CircuitBreakerFilter) AsEnd(ctx Context) Response {
	rurl := ctx.GetAttr("RemoteURL")
	if rurl == nil {
		rurl = "DEFAULTCOMMAND"
	}

	c.CBreaker.Report(rurl.(string), ctx.GetAttr("ReverseRes").(bool))
	return Response{
		Time: time.Now(),
		Code: FilteredPassed,
	}
}
