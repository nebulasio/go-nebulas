package rpc

import (
	"fmt"
	"net"

	"github.com/nebulasio/go-nebulas/rpc/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	// TODO: Load from config.
	address = "127.0.0.1:50007"
)

// NewServer creates a new RPC server and register API endpoints.
func NewServer() *grpc.Server {
	server := grpc.NewServer()
	rpcpb.RegisterAccountServer(server, &AccountServer{})
	// Register reflection service on gRPC server.
	// TODO: Enable reflection only for testing mode.
	reflection.Register(server)
	return server
}

// StartServer starts the RPC server and serves incoming requests.
func StartServer(server *grpc.Server) {
	fmt.Println("Starting RPC server at: ", Address())
	listener, err := net.Listen("tcp", Address())
	if err != nil {
		log.Panic("RPC server failed to listen: ", err)
	}
	if err := server.Serve(listener); err != nil {
		log.Panic("RPC server failed to serve: ", err)
	}
}

// Address returns the RPC server address.
func Address() string {
	return address
}
