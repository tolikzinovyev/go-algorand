package server

import (
	"context"
	"fmt"
	"log"

	proto "github.com/algorand/go-algorand/third_party/go-algorand-grpc"
)

type GrpcServer struct {
	proto.UnimplementedGreeterServer
}

func (s *GrpcServer) SayHello(ctx context.Context, in *proto.HelloRequest) (*proto.HelloReply, error) {
	log.Printf("Received: %v", in.Name)
	return &proto.HelloReply{Msg: fmt.Sprintf("Hello %s from algod", in.Name)}, nil
}
