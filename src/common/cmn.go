package common

const (
	REG_WORKER_OK     = 0
	REG_WORKER_FAILED = -1

	HEARTBEAT_INTERVAL = 5
)

type CommonResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type LBPolicyEnum int

const (
	_ LBPolicyEnum = iota
	LB_RANDOM
	LB_ROUNDROBIN
)
