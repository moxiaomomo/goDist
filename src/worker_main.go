package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
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

	fmt.Printf("[Info]readbuf:%s", buf.String())

	return nil
}

func main() {
	portInt, err := strconv.Atoi(*port)
	if err != nil {
		fmt.Printf("[Error]%s", err.Error())
		os.Exit(-1)
	}

	lip := util.GetLocalIP()
	if len(lip) == 0 {
		fmt.Printf("[Error]Cannot get local ip.")
		os.Exit(-1)
	}
	err = worker.Register(lip, portInt)
	if err != nil {
		fmt.Printf("[Error]%s", err.Error())
		os.Exit(-1)
	} else {
		fmt.Printf("[Info]Register worker succeeded.")
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", portInt))
	if err != nil {
		fmt.Printf("[Error]%s", err.Error())
		os.Exit(-1)
	}

	fmt.Printf("to run server on port: %d\n", portInt)
	svr := grpc.NewServer()
	worker.RegisterGreeterServer(svr)
	svr.Serve(ln)
	//	for {
	//		conn, err := ln.Accept()
	//		if err != nil {
	//			fmt.Printf("[Error]%s", err.Error())
	//		}

	//		go handler(&conn)
	//	}
}
