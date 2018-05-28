package util

import (
	"sync"
	"time"
)

type TokenBucket struct {
	mutex  sync.Mutex
	ticker *time.Ticker

	interval time.Duration

	capacity   int64
	availcount int64
}

func NewTokenBucket(cap int64, interval time.Duration) *TokenBucket {
	return &TokenBucket{
		interval:   interval,
		capacity:   cap,
		availcount: cap,

		ticker: time.NewTicker(interval),
	}
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

func (tb *TokenBucket) TryTake(n int64) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	if tb.availcount >= n && n >= 0 {
		tb.availcount -= n
		return true
	}

	return false
}
