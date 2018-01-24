package rpc

import (
	"errors"
	"net"

	"github.com/sirupsen/logrus"

	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"github.com/nebulasio/go-nebulas/util/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Errors
var (
	ErrEmptyRPCListenList = errors.New("empty rpc listen list")
)

// APIServer is the RPC server type.
type APIServer struct {
	neblet Neblet

	rpcServer *grpc.Server

	rpcConfig *nebletpb.RPCConfig
}

// NewAPIServer creates a new RPC server and registers the API endpoints.
func NewAPIServer(neblet Neblet) *APIServer {
	cfg := neblet.Config().Rpc

	rpc := grpc.NewServer()

	srv := &APIServer{neblet: neblet, rpcServer: rpc, rpcConfig: cfg}
	api := &APIService{srv}

	rpcpb.RegisterApiServiceServer(rpc, api)
	rpcpb.RegisterAdminServiceServer(rpc, api)
	// Register reflection service on gRPC server.
	// TODO: Enable reflection only for testing mode.
	reflection.Register(rpc)

	return srv
}

// Start starts the rpc server and serves incoming requests.
func (s *APIServer) Start() error {
	logging.CLog().Info("Starting RPC Server...")

	if len(s.rpcConfig.RpcListen) == 0 {
		return ErrEmptyRPCListenList
	}

	for _, v := range s.rpcConfig.RpcListen {
		if err := s.start(v); err != nil {
			return err
		}
	}

	return nil
}

func (s *APIServer) start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to listen to RPC Server")
		return err
	}

	logging.CLog().WithFields(logrus.Fields{
		"address": addr,
	}).Info("Started RPC Server.")

	go func() {
		if err := s.rpcServer.Serve(listener); err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"err": err,
			}).Info("RPC server exited.")
		}
	}()

	return nil
}

// RunGateway run grpc mapping to http after apiserver have started.
func (s *APIServer) RunGateway() error {
	//time.Sleep(3 * time.Second)
	rpcListen := s.rpcConfig.RpcListen[0]
	gatewayListen := s.rpcConfig.HttpListen
	httpModule := s.rpcConfig.HttpModule
	logging.CLog().WithFields(logrus.Fields{
		"rpc-server":  rpcListen,
		"http-server": gatewayListen,
	}).Info("Starting RPC Gateway Server...")

	go (func() {
		if err := Run(rpcListen, gatewayListen, httpModule); err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Failed to start RPC Gateway.")
		}
	})()
	return nil
}

// Stop stops the rpc server and closes listener.
func (s *APIServer) Stop() {
	logging.CLog().WithFields(logrus.Fields{
		"listen": s.rpcConfig.RpcListen,
	}).Info("Stopping RPC Server and Gateway...")

	s.rpcServer.Stop()

	logging.CLog().Info("Stopped RPC Server and Gateway.")
}

// Neblet returns weak reference to Neblet.
func (s *APIServer) Neblet() Neblet {
	return s.neblet
}
