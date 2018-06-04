package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"

	"github.com/moxiaomomo/goDist/server/greeter"
	sutil "github.com/moxiaomomo/goDist/server/util"
	"github.com/moxiaomomo/goDist/util"
	"github.com/moxiaomomo/goDist/util/logger"

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

	go sutil.RegisterWithHeartbeat("/srv/hello", cfg["lbhost"].(string), cfg["listenat"].(string))

	logger.LogInfof("to run server on addr: %s\n", cfg["listenat"].(string))
	svr := grpc.NewServer()
	greeter.RegisterGreeterServer(svr)
	svr.Serve(ln)
}
