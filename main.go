package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	v1 "github.com/izaakdale/grpc-mtls-server/api/bytetransfer/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

var _ v1.RemoteServer = (*server)(nil)

type server struct {
	v1.UnimplementedRemoteServer
}

type msg struct {
	Message string
}

func (*server) Call(ctx context.Context, req *v1.Request) (*v1.Response, error) {
	return &v1.Response{Body: []byte("server says hello\n")}, nil
}

func main() {
	ls, err := net.Listen("tcp", "localhost:7777")
	if err != nil {
		panic(err)
	}

	certFile := os.Getenv("SERVER_CRT")
	keyFile := os.Getenv("SERVER_KEY")

	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		log.Fatalf("Error reading TLS cert/key: %v", err)
	}

	gsrv := grpc.NewServer(grpc.Creds(creds))
	reflection.Register(gsrv)

	srv := server{}

	v1.RegisterRemoteServer(gsrv, &srv)

	errCh := make(chan error)
	go func(ch chan error) {
		ch <- gsrv.Serve(ls)
	}(errCh)

	shCh := make(chan os.Signal, 2)
	signal.Notify(shCh, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-shCh:
			os.Exit(1)
		case err := <-errCh:
			log.Fatalf("grpc server errored: %v", err)
		}
	}
}
