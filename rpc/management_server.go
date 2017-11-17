package rpc

import (
	"net"

	"fmt"

	"time"

	"github.com/nebulasio/go-nebulas/rpc/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// ManagementServer is the management RPC server type.
type ManagementServer struct {
	neblet Neblet

	rpcServer *grpc.Server

	port uint32

	gatewayPort uint32
}

// NewManagementServer creates a new management RPC server and registers the API endpoints.
func NewManagementServer(neblet Neblet) *ManagementServer {
	cfg := neblet.Config()

	rpc := grpc.NewServer()
	srv := &ManagementServer{neblet: neblet, rpcServer: rpc, port: cfg.Rpc.ManagementPort, gatewayPort: cfg.Rpc.ManagementHttpPort}

	// register api to management server
	api := &APIService{srv}
	rpcpb.RegisterAPIServiceServer(rpc, api)

	// register management api to management server
	managementAPI := &ManagementService{srv}
	rpcpb.RegisterManagementServiceServer(rpc, managementAPI)

	// Register reflection service on gRPC server.
	// TODO: Enable reflection only for testing mode.
	reflection.Register(rpc)

	return srv
}

// Start starts the rpc server and serves incoming requests.
func (s *ManagementServer) Start() error {
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

// Stop stops the rpc server and closes listener.
func (s *ManagementServer) Stop() {
	log.Info("Stopping RPC server at: ", s.Address())
	s.rpcServer.Stop()
}

// Neblet returns weak reference to Neblet.
func (s *ManagementServer) Neblet() Neblet {
	return s.neblet
}

// Address returns the RPC server address.
func (s *ManagementServer) Address() string {
	addr := fmt.Sprintf("%s:%d", host, s.port)
	return addr
}

// RunGateway run grpc mapping to http after apiserver have started.
func (s *ManagementServer) RunGateway() error {
	time.Sleep(3 * time.Second)
	log.Info("Starting management gateway server bind port: ", s.port, " to:", s.gatewayPort)
	if err := Run(GatewayManagementServiceKey, s.port, s.gatewayPort); err != nil {
		log.Error("management server gateway failed to serve: ", err)
		return err
	}
	return nil
}
