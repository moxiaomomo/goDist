package util

import (
	"sync"
	"time"
)

type TokenBucketV2 struct {
	mutex      sync.Mutex
	ticker     *time.Ticker
	interval   time.Duration
	interval2  time.Duration
	capacity   int64
	availcount int64
	multiple   int64
}

func NewTokenBucketV2(cap int64, interval time.Duration) *TokenBucketV2 {
	tb := &TokenBucketV2{
		interval:   interval,
		capacity:   cap,
		availcount: cap,
		ticker:     time.NewTicker(100 * time.Millisecond),
	}

	tb.interval2 = 100 * time.Millisecond
	tb.multiple = tb.multipleOfIntervalInFact()
	go tb.genTokenWithTicker_v2()

	return tb
}

func (tb *TokenBucketV2) genTokenWithTicker_v2() {
	for _ = range tb.ticker.C {
		tb.mutex.Lock()

		tb.availcount += tb.multiple
		if tb.availcount > tb.capacity {
			tb.availcount = tb.capacity
		}
		tb.mutex.Unlock()
	}
}

func (tb *TokenBucketV2) TryTake(n int64) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	if tb.availcount >= n && n >= 0 {
		tb.availcount -= n
		return true
	}

	return false
}

func (tb *TokenBucketV2) multipleOfIntervalInFact() int64 {
	val := tb.interval2.Nanoseconds() / tb.interval.Nanoseconds()
	if val < 1 {
		val = 1
	}
	return val
}
