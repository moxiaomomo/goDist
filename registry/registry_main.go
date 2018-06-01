package registry

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/moxiaomomo/goDist/registry/handler"
	"github.com/moxiaomomo/goDist/util/logger"
	"github.com/tidwall/gjson"
)

var (
	_confPath = flag.String("_confpath", "raft.cfg", "configuration filepath")
)

func loadConfig(confPath string) (map[string]interface{}, error) {
	cfg, err := ioutil.ReadFile(confPath)
	if err != nil {
		fmt.Printf("LoadConfig failed, err:%s\n", err.Error())
		return nil, err
	}
	m, ok := gjson.Parse(string(cfg)).Value().(map[string]interface{})
	if !ok {
		return nil, errors.New("Parse config failed.")
	}
	logger.LogInfof("config:%+v\n", m)
	return m, nil
}

func _main() {
	flag.Parse()

	//	cfg, err := loadConfig("config/reg.conf")
	//	if err != nil {
	//		logger.LogErrorf("Program will exit while loading config failed.")
	//		os.Exit(1)
	//	}

	sv, err := handler.NewService(*_confPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = sv.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
