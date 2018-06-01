package filter

import (
	"time"

	tbucket "github.com/moxiaomomo/token-bucket"
)

// RateLimitFilter rate-limiting filter
type RateLimitFilter struct {
	DefaultFilter
	globalRL *tbucket.TokenBucket
}

// Init filter initialization
func (f *RateLimitFilter) Init(config string) error {
	if f.globalRL == nil {
		rl, err := tbucket.NewTokenBucket(1000, time.Second)
		if err != nil {
			return err
		}
		f.globalRL = rl
	}
	return nil
}

// Name returns RateLimitFilter's name
func (f *RateLimitFilter) Name() string {
	return "ratelimitfilter"
}

// AsBegin execute at the beginning
func (f *RateLimitFilter) AsBegin(c Context) Response {
	// wait timeout for 2 seconds
	if f.globalRL.Wait(1, 2*time.Second) {
		return Response{
			Time: time.Now(),
			Code: FilteredPassed,
		}
	}
	return Response{
		Time:    time.Now(),
		Code:    FilteredFailed,
		Message: "Request failed as rate-limiting",
	}
}

// AsEnd execute at the end
func (f *RateLimitFilter) AsEnd(c Context) Response {
	return Response{
		Time: time.Now(),
		Code: FilteredPassed,
	}
}
