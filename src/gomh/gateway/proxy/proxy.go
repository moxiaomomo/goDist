package proxy

import (
	"gomh/gateway/filter"
	"sync"
)

type Proxy struct {
	mutex sync.RWMutex

	filters []filter.Filter
}
