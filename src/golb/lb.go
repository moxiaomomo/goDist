package golb

import (
	"common"
	"config"
	"errors"
	"logger"
	"math/rand"
	"sync"
	"time"
)

var (
	ERR_WORKER_EXISTS     = errors.New("ERR_WORKER_EXISTS")
	ERR_WORKER_NOT_EXISTS = errors.New("ERR_WORKER_NOT_EXISTS")
)

type Worker struct {
	Heartbeat int64 // last heartbeat timestamp
	Host      string
}

type Workers struct {
	Members     []Worker
	lastRRIndex int
	sync.Mutex
}

var workers Workers = Workers{lastRRIndex: 0}

type Comparable interface {
	IsEqual(a interface{}) bool
}

func SetLBPolicy(p common.LBPolicyEnum) {
	config.SetGlobalLBConfig(map[string]interface{}{"LBPolicy": p})
}

func (w *Worker) IsEqual(nw interface{}) bool {
	if cmpnw, ok := nw.(Worker); ok {
		return w.Host == cmpnw.Host
	}
	return false
}

func AddWorker(w Worker) error {
	workers.Mutex.Lock()
	defer workers.Mutex.Unlock()

	for i := 0; i < len(workers.Members); i++ {
		if workers.Members[i].IsEqual(w) {
			workers.Members[i].Heartbeat = w.Heartbeat
			return ERR_WORKER_EXISTS
		}
	}
	workers.Members = append(workers.Members, w)
	return nil
}

func RemoveWorker(w Worker) error {
	workers.Mutex.Lock()
	defer workers.Mutex.Unlock()

	for k, v := range workers.Members {
		if v.IsEqual(w) {
			if k == len(workers.Members)-1 {
				workers.Members = workers.Members[:k]
			} else {
				workers.Members = append(workers.Members[:k], workers.Members[k+1:]...)
			}
			return nil
		}
	}
	return ERR_WORKER_NOT_EXISTS
}

func RemoveWorkerAsTimeout() {
	for {
		func() {
			workers.Mutex.Lock()
			defer workers.Mutex.Unlock()

			now := time.Now().Unix()
			for k, v := range workers.Members {
				// timeout after twice heartbeat interval
				if now-v.Heartbeat > common.HEARTBEAT_INTERVAL*2 {
					if k == len(workers.Members)-1 {
						workers.Members = workers.Members[:k]
					} else {
						workers.Members = append(workers.Members[:k], workers.Members[k+1:]...)
					}
					logger.LogWarnf("Lost heartbeat from worker: %s", v.Host)
				}
			}
		}()

		time.Sleep(time.Second)
	}
}

func GetWorker() (Worker, error) {
	if config.GlobalLBConfig().LBPolicy == common.LB_ROUNDROBIN {
		return RoundRobinWorker()
	}
	return RandomWorker()
}

func RandomWorker() (Worker, error) {
	workers.Mutex.Lock()
	defer workers.Mutex.Unlock()

	if len(workers.Members) <= 0 {
		return Worker{}, errors.New("Empty workers")
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return workers.Members[r.Intn(len(workers.Members))], nil
}

func RoundRobinWorker() (Worker, error) {
	workers.Mutex.Lock()
	defer workers.Mutex.Unlock()

	if len(workers.Members) <= 0 {
		return Worker{}, errors.New("Empty workers")
	}

	defer func() {
		workers.lastRRIndex = (workers.lastRRIndex + 1) % len(workers.Members)
	}()
	return workers.Members[workers.lastRRIndex%len(workers.Members)], nil
}
