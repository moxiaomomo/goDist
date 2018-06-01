package greeter

import (
	pb "github.com/moxiaomomo/goDist/proto/greeter"
	"github.com/moxiaomomo/goDist/util/logger"
	//	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type server struct{}

var kkk int = 0

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	logger.LogInfo("SayHello Called.")
	//	kkk += 20
	// time.Sleep(time.Duration(60) * time.Second)
	return &pb.HelloResponse{Message: "Hi " + in.Name + "\n"}, nil
}

// RegisterGreeterServer register into grpc
func RegisterGreeterServer(gsvr *grpc.Server) {
	pb.RegisterGreeterServer(gsvr, &server{})
}
