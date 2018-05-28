package util

import (
	"sync"
	"time"
)

type TokenBucket struct {
	mutex       sync.Mutex
	ticker      *time.Ticker
	ticker2     *time.Ticker
	lastAjustAt time.Time

	interval  time.Duration
	interval2 time.Duration

	capacity   int64
	availcount int64
	multiple   int64
}

func NewTokenBucket(cap int64, interval time.Duration) *TokenBucket {
	tb := &TokenBucket{
		interval:   interval,
		capacity:   cap,
		availcount: cap,

		ticker2: time.NewTicker(100 * time.Millisecond),
	}

	tb.interval2 = 100 * time.Millisecond
	tb.multiple = tb.multipleOfIntervalInFact()
	tb.lastAjustAt = time.Now()
	go tb.genTokenWithTicker_v2()

	return tb
}

func (tb *TokenBucket) genTokenWithTicker() {
	for _ = range tb.ticker.C {
		tb.mutex.Lock()
		if tb.availcount < tb.capacity {
			tb.availcount++
		}
		tb.mutex.Unlock()
	}
}

func (tb *TokenBucket) genTokenWithTicker_v2() {
	for _ = range tb.ticker2.C {
		tb.mutex.Lock()

		tb.availcount += tb.multiple
		if tb.availcount > tb.capacity {
			tb.availcount = tb.capacity
		}
		tb.mutex.Unlock()
	}
}

func (tb *TokenBucket) adjust() {

}

func (tb *TokenBucket) TryTake(n int64) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	if tb.availcount >= n && n >= 0 {
		tb.availcount -= n
		return true
	}

	return false
}

func (tb *TokenBucket) multipleOfIntervalInFact() int64 {
	val := tb.interval2.Nanoseconds() / tb.interval.Nanoseconds()
	if val < 1 {
		val = 1
	}
	return val
}
