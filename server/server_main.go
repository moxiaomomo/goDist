package main

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/moxiaomomo/goDist/server/handler"
	"github.com/moxiaomomo/goDist/util/logger"

	"github.com/tidwall/gjson"
)

// LoadConfig LoadConfig
func LoadConfig(confPath string) (map[string]interface{}, string, error) {
	cfg, err := ioutil.ReadFile(confPath)
	if err != nil {
		fmt.Printf("LoadConfig failed, err:%s\n", err.Error())
		return nil, "", err
	}
	m, ok := gjson.Parse(string(cfg)).Value().(map[string]interface{})
	if !ok {
		return nil, "", errors.New("Parse config failed.")
	}
	logger.LogInfof("config:%+v\n", m)
	return m, string(cfg), nil
}

func main() {
	handler.StartServer()
}
