package main

import (
	"bytes"
	"flag"
	"fmt"
	"gomh/server"
	"gomh/util"
	"gomh/util/logger"
	"io"
	"net"
	"os"
	"strconv"

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
	flag.Parse()
	logger.SetLogLevel(util.LOG_INFO)

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

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", portInt))
	if err != nil {
		logger.LogError(err.Error())
		os.Exit(-1)
	}

	err = server.Register(lip, portInt)
	if err != nil {
		logger.LogError(err.Error())
		os.Exit(-1)
	} else {
		logger.LogInfo("Register worker succeeded.")
	}

	logger.LogInfof("to run server on port: %d\n", portInt)
	svr := grpc.NewServer()
	server.RegisterGreeterServer(svr)
	svr.Serve(ln)
}
