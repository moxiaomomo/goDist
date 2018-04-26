package handler

import (
	"errors"
	"gomh/config"
	"gomh/util"
	"gomh/util/logger"
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
	callCount int
	respMS    int // average response time in microsecond
	sync.Mutex
}

type Workers struct {
	Members     []Worker
	lastRRIndex int
	sync.Mutex
}

var workers Workers = NewWorkers()

type Comparable interface {
	IsEqual(a interface{}) bool
}

func NewWorker() *Worker {
	return &Worker{
		callCount: 0,
	}
}

func NewWorkers() Workers {
	return Workers{
		lastRRIndex: 0,
	}
}

func SetLBPolicy(p util.LBPolicyEnum) {
	config.SetGlobalLBConfig(map[string]interface{}{"LBPolicy": p})
}

func (w *Worker) IsEqual(nw interface{}) bool {
	if cmpnw, ok := nw.(Worker); ok {
		return w.Host == cmpnw.Host
	}
	return false
}

// get host to call, and do something extra
func (w *Worker) HostToCall() string {
	//	w.Mutex.Lock()
	//	defer w.Mutex.Unlock()

	//	w.callCount += 1
	return w.Host
}

func (w *Worker) CallCount() int {
	return w.callCount
}

func (w *Worker) AsTaskFinished(timeUsed int) {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	if timeUsed <= 0 || timeUsed > 600000 { // default 10mins timeout
		return
	}

	w.respMS = int((w.respMS*w.callCount + timeUsed) / (w.callCount + 1))
	w.callCount += 1
}

func (w *Worker) ResponseTimeUsed() int {
	return w.respMS
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
				if now-v.Heartbeat > util.HEARTBEAT_INTERVAL*2 {
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

func GetWorker() (*Worker, error) {
	switch config.GlobalLBConfig().LBPolicy {
	case util.LB_ROUNDROBIN:
		return RoundRobinWorker()
	case util.LB_FASTRESP:
		return FastResponseWorker()
	default:
		return RandomWorker()
	}
}

func RandomWorker() (*Worker, error) {
	workers.Mutex.Lock()
	defer workers.Mutex.Unlock()

	var worker = NewWorker()

	if len(workers.Members) <= 0 {
		return worker, errors.New("Empty workers")
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	worker = &workers.Members[r.Intn(len(workers.Members))]
	return worker, nil
}

func RoundRobinWorker() (*Worker, error) {
	workers.Mutex.Lock()
	defer workers.Mutex.Unlock()

	var worker = NewWorker()

	if len(workers.Members) <= 0 {
		return worker, errors.New("Empty workers")
	}

	defer func() {
		workers.lastRRIndex = (workers.lastRRIndex + 1) % len(workers.Members)
	}()
	index := workers.lastRRIndex % len(workers.Members)
	worker = &workers.Members[index]
	logger.LogDebug(index, worker.respMS)
	return worker, nil
}

func FastResponseWorker() (*Worker, error) {
	workers.Mutex.Lock()
	defer workers.Mutex.Unlock()

	var worker = NewWorker()

	if len(workers.Members) <= 0 {
		return worker, errors.New("Empty workers")
	}

	var minRespIdx int = 0
	for k, v := range workers.Members {
		if v.ResponseTimeUsed() < workers.Members[minRespIdx].ResponseTimeUsed() {
			minRespIdx = k
		}
	}
	worker = &workers.Members[minRespIdx]
	logger.LogDebug(minRespIdx, worker.respMS)
	return worker, nil
}
