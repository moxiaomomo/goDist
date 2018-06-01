package filter

import (
	"time"

	"github.com/moxiaomomo/goDist/util/logger"
)

// TimeUsedFilter rate-limiting filter
type TimeUsedFilter struct {
	DefaultFilter
	start time.Time
	end   time.Time
	ts    time.Duration
}

// Init filter initialization
func (f *TimeUsedFilter) Init(config string) error {
	return nil
}

// Name returns TimeUsedFilter's name
func (f *TimeUsedFilter) Name() string {
	return "TimeUsedFilter"
}

// AsBegin execute at the beginning
func (f *TimeUsedFilter) AsBegin(c Context) Response {
	f.start = time.Now()
	return Response{
		Time: f.start,
		Code: FilteredPassed,
	}
}

// AsEnd execute at the end
func (f *TimeUsedFilter) AsEnd(c Context) Response {
	f.end = time.Now()

	logger.LogInfof("uri: %s, timeused: %s\n",
		c.HTTPRequest().RequestURI,
		(f.end.Sub(f.start)).String())
	return Response{
		Time: f.end,
		Code: FilteredPassed,
	}
}
