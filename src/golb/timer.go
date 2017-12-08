package golb

import (
	"time"
)

type ProcTimer struct {
	start int
	end   int
}

func (w *ProcTimer) OnStart() {
	w.start = int(time.Now().UnixNano() / 1000000)
}

func (w *ProcTimer) OnEnd() {
	w.end = int(time.Now().UnixNano() / 1000000)
}

func (w *ProcTimer) Duration() int {
	return w.end - w.start
}
