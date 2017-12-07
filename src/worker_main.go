package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"logger"
	"net"
	"os"
	"strconv"
	"util"
	"worker"

	"google.golang.org/grpc"
)

var (
	port = flag.String("port", "9000", "listen port")
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

func main() {
	logger.SetLogLevel(logger.LOG_INFO)

	portInt, err := strconv.Atoi(*port)
	if err != nil {
		logger.LogError(err.Error())
		os.Exit(-1)
	}

	lip := util.GetLocalIP()
	if len(lip) == 0 {
		logger.LogError("Cannot get local ip.")
		os.Exit(-1)
	}
	err = worker.Register(lip, portInt)
	if err != nil {
		logger.LogError(err.Error())
		os.Exit(-1)
	} else {
		logger.LogInfo("Register worker succeeded.")
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", portInt))
	if err != nil {
		logger.LogError(err.Error())
		os.Exit(-1)
	}

	logger.LogInfof("to run server on port: %d\n", portInt)
	svr := grpc.NewServer()
	worker.RegisterGreeterServer(svr)
	svr.Serve(ln)
}
