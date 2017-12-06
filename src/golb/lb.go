package golb

import (
	"errors"
)

var (
	ERR_WORKER_EXISTS     = errors.New("ERR_WORKER_EXISTS")
	ERR_WORKER_NOT_EXISTS = errors.New("ERR_WORKER_NOT_EXISTS")
)

type Worker struct {
	Host string
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
	for _, v := range workers {
		if v.IsEqual(w) {
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
