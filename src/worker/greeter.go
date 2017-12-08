package worker

import (
	"logger"
	pb "proto/greeter"
	//	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type server struct{}

var kkk int = 0

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	logger.LogInfo("SayHello Called.")
	//	kkk += 20
	//	time.Sleep(time.Duration(kkk) * time.Millisecond)
	return &pb.HelloResponse{Message: "Hi " + in.Name}, nil
}

func RegisterGreeterServer(gsvr *grpc.Server) {
	pb.RegisterGreeterServer(gsvr, &server{})
}
