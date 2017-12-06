package common

const (
	REG_WORKER_OK     = 0
	REG_WORKER_FAILED = -1
)

type CommonResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
