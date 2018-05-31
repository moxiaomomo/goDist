package handler

import (
	"errors"
	"fmt"
	//	"fmt"
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
	Heartbeat int64  // last heartbeat timestamp
	UriPath   string // api request uripath
	Host      string
	callCount int
	respMS    int // average response time in microsecond
	sync.Mutex
}

type Workers struct {
	Members     map[string][]Worker
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
		Members:     make(map[string][]Worker, 0),
	}
}

func SetLBPolicy(p util.LBPolicyEnum) {
	config.SetGlobalLBConfig(map[string]interface{}{"LBPolicy": p})
}

func (w *Worker) IsEqual(nw interface{}) bool {
	if cmpnw, ok := nw.(Worker); ok {
		return w.Host == cmpnw.Host && w.UriPath == cmpnw.UriPath
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

	mlist, ok := workers.Members[w.UriPath]
	if ok {
		for i := 0; i < len(mlist); i++ {
			if mlist[i].IsEqual(w) {
				mlist[i].Heartbeat = w.Heartbeat
				return ERR_WORKER_EXISTS
			}
		}
	} else {
		workers.Members[w.UriPath] = make([]Worker, 0)
	}

	workers.Members[w.UriPath] = append(workers.Members[w.UriPath], w)
	fmt.Printf("%+v\n", workers.Members)
	return nil
}

func RemoveWorker(w Worker) error {
	workers.Mutex.Lock()
	defer workers.Mutex.Unlock()

	mlist, ok := workers.Members[w.UriPath]
	if !ok {
		return ERR_WORKER_NOT_EXISTS
	}

	for k, v := range mlist {
		if v.IsEqual(w) {
			if k == len(mlist)-1 {
				workers.Members[w.UriPath] = workers.Members[w.UriPath][:k]
			} else {
				workers.Members[w.UriPath] = append(workers.Members[w.UriPath][:k], workers.Members[w.UriPath][k+1:]...)
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
			for k1, v1 := range workers.Members {
				for k, v := range v1 {
					//fmt.Printf("%+v\n", v1)
					// timeout after twice heartbeat interval
					if now-v.Heartbeat > util.HEARTBEAT_INTERVAL*2 {
						if k == len(workers.Members[k1])-1 {
							workers.Members[k1] = workers.Members[k1][:k]
						} else {
							workers.Members[k1] = append(workers.Members[k1][:k], workers.Members[k1][k+1:]...)
						}
						logger.LogWarnf("Lost heartbeat from worker: %s\n", v.Host)
					}
				}
			}
		}()

		time.Sleep(time.Second)
	}
}

func GetWorker(uripath string) (*Worker, error) {
	switch config.GlobalLBConfig().LBPolicy {
	case util.LB_ROUNDROBIN:
		return RoundRobinWorker(uripath)
	case util.LB_FASTRESP:
		return FastResponseWorker(uripath)
	default:
		return RandomWorker(uripath)
	}
}

func RandomWorker(uripath string) (*Worker, error) {
	workers.Mutex.Lock()
	defer workers.Mutex.Unlock()

	mlist, ok := workers.Members[uripath]
	var worker = NewWorker()
	//	fmt.Printf("%s %d %d\n", uripath, mlist, ok)
	if !ok || len(mlist) <= 0 {
		return worker, errors.New("Empty workers")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	worker = &mlist[r.Intn(len(mlist))]
	return worker, nil
}

func RoundRobinWorker(uripath string) (*Worker, error) {
	workers.Mutex.Lock()
	defer workers.Mutex.Unlock()

	mlist, ok := workers.Members[uripath]
	var worker = NewWorker()
	if !ok || len(mlist) <= 0 {
		return worker, errors.New("Empty workers")
	}

	defer func() {
		workers.lastRRIndex = (workers.lastRRIndex + 1) % len(mlist)
	}()
	index := workers.lastRRIndex % len(mlist)
	worker = &mlist[index]
	logger.LogDebug(index, worker.respMS)
	return worker, nil
}

func FastResponseWorker(uripath string) (*Worker, error) {
	workers.Mutex.Lock()
	defer workers.Mutex.Unlock()

	mlist, ok := workers.Members[uripath]
	var worker = NewWorker()
	if !ok || len(mlist) <= 0 {
		return worker, errors.New("Empty workers")
	}

	var minRespIdx int = 0
	for k, v := range mlist {
		if v.ResponseTimeUsed() < mlist[minRespIdx].ResponseTimeUsed() {
			minRespIdx = k
		}
	}
	worker = &mlist[minRespIdx]
	logger.LogDebug(minRespIdx, worker.respMS)
	return worker, nil
}
