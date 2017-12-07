package golb

import (
	"common"
	"errors"
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

var workers []Worker

type Comparable interface {
	IsEqual(a interface{}) bool
}

func (w *Worker) IsEqual(nw interface{}) bool {
	if cmpnw, ok := nw.(Worker); ok {
		return w.Host == cmpnw.Host
	}
	return false
}

func Workers() []Worker {
	return workers
}

func AddWorker(w Worker) error {
	for i := 0; i < len(workers); i++ {
		if workers[i].IsEqual(w) {
			workers[i].Heartbeat = w.Heartbeat
			return ERR_WORKER_EXISTS
		}
	}
	workers = append(workers, w)
	return nil
}

func RemoveWorker(w Worker) error {
	for k, v := range workers {
		if v.IsEqual(w) {
			if k == len(workers)-1 {
				workers = workers[:k]
			} else {
				workers = append(workers[:k], workers[k+1:]...)
			}
			return nil
		}
	}
	return ERR_WORKER_NOT_EXISTS
}

func RemoveWorkerAsTimeout() {
	for {
		now := time.Now().Unix()
		for k, v := range workers {
			// timeout after twice heartbeat interval
			if now-v.Heartbeat > common.HEARTBEAT_INTERVAL*2 {
				if k == len(workers)-1 {
					workers = workers[:k]
				} else {
					workers = append(workers[:k], workers[k+1:]...)
				}
			}
		}
		time.Sleep(time.Second)
	}
}
