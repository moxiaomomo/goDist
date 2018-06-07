package handler

import (
	"net"
	"os"
	"time"

	"github.com/moxiaomomo/goDist/server/config"
	"github.com/moxiaomomo/goDist/server/greeter"
	"github.com/moxiaomomo/goDist/util"
	"github.com/moxiaomomo/goDist/util/logger"
	"google.golang.org/grpc"
)

func RegisterWithHeartbeat(conf *config.ServerConfig) {
	err := util.HeartbeatForRegistry(conf.LBHost, conf.SvrAddr, "", conf.URIPath)
	if err != nil {
		logger.LogErrorf("failed to register server, err:%s\n", err)
	}

	t := time.NewTicker(time.Second * util.HEARTBEAT_INTERVAL)
	for range t.C {
		err := util.HeartbeatForRegistry(conf.LBHost, conf.SvrAddr, "", conf.URIPath)
		if err != nil {
			logger.LogErrorf("failed to register server, err:%s\n", err)
		}
	}
}

// StartServer StartServer
func StartServer(conf *config.ServerConfig, grpcOpt []grpc.ServerOption) {
	go RegisterWithHeartbeat(conf)

	logger.LogInfof("to run server on addr: %s\n", conf.SvrAddr)

	ln, err := net.Listen("tcp", conf.SvrAddr)
	if err != nil {
		logger.LogError(err.Error())
		os.Exit(-1)
	}
	svr := grpc.NewServer(grpcOpt...)
	greeter.RegisterGreeterServer(svr)
	svr.Serve(ln)
}
