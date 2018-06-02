package handler

import (
	"net/http"
	"time"

	"github.com/moxiaomomo/goDist/util"
)

// AsyncHealthCheck check works with heartbeat
func (workers *Workers) AsyncHealthCheck() {
	t := time.NewTicker(workers.healthcheckInterval)
	go func() {
		for range t.C {
			workers.Mutex.Lock()
			now := time.Now().Unix()
			for k := range workers.Members {
				for idx := range workers.Members[k] {
					if now-workers.Members[k][idx].Heartbeat > util.HEARTBEAT_INTERVAL*3 {
						if idx == len(workers.Members[k])-1 {
							workers.Members[k] = workers.Members[k][:idx]
						} else {
							workers.Members[k] = append(workers.Members[k][:idx], workers.Members[k][idx+1:]...)
						}
					} else {
						workers.Members[k][idx].AsyncHealthCheck()
					}
				}
			}
			workers.Mutex.Unlock()
		}
	}()
}

// AsyncHealthCheck helthcheck for one node
func (worker *Worker) AsyncHealthCheck() {
	url := worker.HealthCheckURL
	client := &http.Client{
		Timeout: time.Duration(3 * time.Second),
	}
	go func() {
		_, err := client.Get(url)
		if err != nil {
			worker.Heartbeat = time.Now().Unix()
		}
	}()
}
