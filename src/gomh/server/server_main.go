package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"gomh/server/greeter"
	sutil "gomh/server/util"
	"gomh/util"
	"gomh/util/logger"
	"io"
	"io/ioutil"
	"net"
	"os"

	"github.com/tidwall/gjson"
	"google.golang.org/grpc"
)

func handler(conn *net.Conn) error {
	defer (*conn).Close()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, *conn)
	if err != nil {
		return err
	}

	logger.LogInfof("readbuf:%s", buf.String())

	return nil
}

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
	cfg, _, err := LoadConfig("config/reg.conf")
	if err != nil {
		logger.LogErrorf("Program will exit while loading config failed.")
		os.Exit(1)
	}

	flag.Parse()
	logger.SetLogLevel(util.LOG_INFO)

	lip := util.GetLocalIP()
	if len(lip) == 0 {
		logger.LogError("Cannot get local ip.")
		os.Exit(-1)
	}

	ln, err := net.Listen("tcp", cfg["listenat"].(string))
	if err != nil {
		logger.LogError(err.Error())
		os.Exit(-1)
	}

	err = sutil.Register("/srv/hello", cfg["lbhost"].(string), cfg["listenat"].(string))
	if err != nil {
		logger.LogError(err)
		os.Exit(-1)
	} else {
		logger.LogInfo("Register worker succeeded.")
	}

	logger.LogInfof("to run server on addr: %s\n", cfg["listenat"].(string))
	svr := grpc.NewServer()
	greeter.RegisterGreeterServer(svr)
	svr.Serve(ln)
}
