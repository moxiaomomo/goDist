package util

import (
	"sync"
	"time"
)

type TokenBucketV1 struct {
	mutex      sync.Mutex
	ticker     *time.Ticker
	interval   time.Duration
	capacity   int64
	availcount int64
}

func NewTokenBucketV1(cap int64, interval time.Duration) *TokenBucketV1 {
	tb := &TokenBucketV1{
		interval:   interval,
		capacity:   cap,
		availcount: cap,
		ticker:     time.NewTicker(interval),
	}

	go tb.genTokenWithTicker_v1()

	return tb
}

func (tb *TokenBucketV1) genTokenWithTicker_v1() {
	for _ = range tb.ticker.C {
		tb.mutex.Lock()
		if tb.availcount < tb.capacity {
			tb.availcount++
		}
		tb.mutex.Unlock()
	}
}

func (tb *TokenBucketV1) TryTake(n int64) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	if tb.availcount >= n && n >= 0 {
		tb.availcount -= n
		return true
	}

	return false
}
