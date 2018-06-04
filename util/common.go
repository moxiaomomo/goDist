package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

const (
	REG_WORKER_OK     = 0
	REG_WORKER_FAILED = -1

	HEARTBEAT_INTERVAL = 3
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
	LB_FASTRESP
)

type LogLevelEnum int

const (
	_ LogLevelEnum = iota
	LOG_DEBUG
	LOG_INFO
	LOG_WARN
	LOG_ERROR
)

func SetStructField(obj interface{}, name string, value interface{}) error {
	structObj := reflect.ValueOf(obj).Elem()
	structField := structObj.FieldByName(name)

	if !structField.IsValid() {
		return fmt.Errorf("Field of struct not found: %s", name)
	}
	if !structField.CanSet() {
		return fmt.Errorf("Field of struct cannot set: %s", name)
	}

	structFieldType := structField.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		return fmt.Errorf("Field of struct type mismatch: %s", name)
	}

	structField.Set(val)
	return nil
}

// WriteHTTPResponseAsJson write json-data to client
func WriteHTTPResponseAsJson(w http.ResponseWriter, data interface{}) error {
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Write(d)
	return nil
}
