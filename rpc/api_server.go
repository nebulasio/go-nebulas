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
	logging.CLog().Info("Starting RPC Server")
	if len(s.rpcConfig.RpcListen) > 0 {
		for _, v := range s.rpcConfig.RpcListen {
			err := s.start(v)
			if err != nil {
				logging.CLog().WithFields(logrus.Fields{
					"err": err,
				}).Error("Failed to start RPC Server")
				return errors.New("parse rpc-config rpc-listen occurs error")
			}
		}
	} else {
		logging.CLog().WithFields(logrus.Fields{
			"err": "empty rpc listen list",
		}).Error("Failed to start RPC Server")
		return errors.New("parse rpc-config rpc-listen occurs error")
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
	logging.CLog().Info("Launched RPC server at: ", addr)
	if err := s.rpcServer.Serve(listener); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to serve RPC Server")
		return err
	}
	return nil
}

// RunGateway run grpc mapping to http after apiserver have started.
func (s *APIServer) RunGateway() error {
	//todo make sure rpc server has run before gateway start.
	//time.Sleep(3 * time.Second)
	rpcListen := s.rpcConfig.RpcListen[0]
	gatewayListen := s.rpcConfig.HttpListen
	httpModule := s.rpcConfig.HttpModule
	logging.CLog().Info("Starting api gateway server bind rpc-server: ", rpcListen, " to:", gatewayListen)
	if err := Run(rpcListen, gatewayListen, httpModule); err != nil {
		logging.CLog().Error("RPC server gateway failed to serve: ", err)
		return err
	}
	return nil
}

// Stop stops the rpc server and closes listener.
func (s *APIServer) Stop() {
	logging.CLog().Info("Stopping RPC server at: ", s.rpcConfig.RpcListen)
	s.rpcServer.Stop()
}

// Neblet returns weak reference to Neblet.
func (s *APIServer) Neblet() Neblet {
	return s.neblet
}
