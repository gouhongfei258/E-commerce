package adapter

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
)

type GRPC struct {
	Srv  *grpc.Server
	Addr string
}

func (a GRPC) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", a.Addr)
	if err != nil {
		return err
	}
	log.Printf("gRPC server listening on %s", a.Addr)
	return a.Srv.Serve(lis)
}

func (a GRPC) Stop(ctx context.Context) error {
	a.Srv.GracefulStop()
	return nil
}
