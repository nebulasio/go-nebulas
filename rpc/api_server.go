package rpc

import (
	"net"
	"time"

	"fmt"

	"github.com/nebulasio/go-nebulas/rpc/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// APIServer is the RPC server type.
type APIServer struct {
	neblet Neblet

	rpcServer *grpc.Server

	port uint32

	gatewayPort uint32
}

// NewAPIServer creates a new RPC server and registers the API endpoints.
func NewAPIServer(neblet Neblet) *APIServer {
	cfg := neblet.Config()

	rpc := grpc.NewServer()
	srv := &APIServer{neblet: neblet, rpcServer: rpc, port: cfg.Rpc.RpcListen, gatewayPort: cfg.Rpc.HttpListen}
	api := &APIService{srv}

	rpcpb.RegisterAPIServiceServer(rpc, api)
	// Register reflection service on gRPC server.
	// TODO: Enable reflection only for testing mode.
	reflection.Register(rpc)

	return srv
}

// Start starts the rpc server and serves incoming requests.
func (s *APIServer) Start() error {
	log.Info("Starting RPC server at: ", s.Address())
	listener, err := net.Listen("tcp", s.Address())
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

// RunGateway run grpc mapping to http after apiserver have started.
func (s *APIServer) RunGateway() error {
	time.Sleep(3 * time.Second)
	log.Info("Starting api gateway server bind port: ", s.port, " to:", s.gatewayPort)
	if err := Run(GatewayAPIServiceKey, s.port, s.gatewayPort); err != nil {
		log.Error("RPC server gateway failed to serve: ", err)
		return err
	}
	return nil
}

// Stop stops the rpc server and closes listener.
func (s *APIServer) Stop() {
	log.Info("Stopping RPC server at: ", s.Address())
	s.rpcServer.Stop()
}

// Neblet returns weak reference to Neblet.
func (s *APIServer) Neblet() Neblet {
	return s.neblet
}

// Address returns the RPC server address.
func (s *APIServer) Address() string {
	addr := fmt.Sprintf("%s:%d", host, s.port)
	return addr
}
