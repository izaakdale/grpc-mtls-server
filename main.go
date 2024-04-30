package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "github.com/izaakdale/grpc-mtls-server/api/bytetransfer/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var _ v1.RemoteServer = (*server)(nil)

type server struct {
	v1.UnimplementedRemoteServer
}

func (*server) Call(ctx context.Context, req *v1.Request) (*v1.Response, error) {
	return &v1.Response{Body: []byte("server says hello\n")}, nil
}

// Stream implements v1.RemoteServer.
func (s *server) Stream(req *v1.Request, st v1.Remote_StreamServer) error {
	for {
		if err := st.Send(&v1.Response{Body: []byte("server says hello\n")}); err != nil {
			return err
		}
		time.Sleep(time.Second * 3)
	}
}

func main() {
	ls, err := net.Listen("tcp", fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT")))
	if err != nil {
		panic(err)
	}

	// certFile := os.Getenv("SERVER_CRT")
	// keyFile := os.Getenv("SERVER_KEY")

	// creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	// if err != nil {
	// 	log.Fatalf("Error reading TLS cert/key: %v", err)
	// }

	gsrv := grpc.NewServer()
	// gsrv := grpc.NewServer(grpc.Creds(creds))
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
