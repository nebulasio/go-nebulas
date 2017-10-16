package rpc

import (
	"net"

	"fmt"

	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	host = "127.0.0.1"
	port = uint32(51510)
)

// Server is the RPC server type.
type Server struct {
	neblet Neblet

	rpcServer *grpc.Server
}

// Neblet interface breaks cycle import dependency and hides unused services.
type Neblet interface {
	Config() nebletpb.Config
	BlockChain() *core.BlockChain
	AccountManager() *account.Manager
}

// NewServer creates a new RPC server and registers the API endpoints.
func NewServer(neblet Neblet) *Server {
	cfg := neblet.Config()
	port = cfg.Rpc.Port

	rpc := grpc.NewServer()
	srv := &Server{neblet: neblet, rpcServer: rpc}
	api := &APIService{srv}

	rpcpb.RegisterAPIServiceServer(rpc, api)
	// Register reflection service on gRPC server.
	// TODO: Enable reflection only for testing mode.
	reflection.Register(rpc)

	return srv
}

// Start starts the rpc server and serves incoming requests.
func (s *Server) Start() error {
	log.Info("Starting RPC server at: ", Address())
	listener, err := net.Listen("tcp", Address())
	if err != nil {
		log.Error("RPC server failed to listen: ", err)
		return err
	}
	if err := s.rpcServer.Serve(listener); err != nil {
		log.Error("RPC server failed to serve: ", err)
		return err
	}
	return nil
}

// Stop stops the rpc server and closes listener.
func (s *Server) Stop() {
	log.Info("Stopping RPC server at: ", Address())
	s.rpcServer.Stop()
}

// Neblet returns weak reference to Neblet.
func (s *Server) Neblet() Neblet {
	return s.neblet
}

// Address returns the RPC server address.
func Address() string {
	addr := fmt.Sprintf("%s:%d", host, port)
	return addr
}
