package worker

import (
	"fmt"
	pb "proto/greeter"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type server struct{}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	fmt.Println("SayHello Called.")
	return &pb.HelloResponse{Message: "Hi " + in.Name}, nil
}

func RegisterGreeterServer(gsvr *grpc.Server) {
	pb.RegisterGreeterServer(gsvr, &server{})
}
