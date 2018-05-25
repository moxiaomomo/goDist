package handler

import (
	"fmt"
	pb "gomh/proto/greeter"
	"gomh/util/logger"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/tidwall/gjson"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func InitHandlers() {
	http.HandleFunc("/hello", HelloHandler)
}

func StartGatewayServer(listenHost string) {
	InitHandlers()

	logger.LogInfof("to start server on port: %s\n", listenHost)
	http.ListenAndServe(listenHost, nil)
}
