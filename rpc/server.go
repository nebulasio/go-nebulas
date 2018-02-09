package rpc

import (
	"errors"
	"net"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/netutil"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	nebnet "github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"github.com/nebulasio/go-nebulas/util/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Errors
var (
	ErrEmptyRPCListenList = errors.New("empty rpc listen list")
)

// Const
const (
	DefaultConnectionLimits = 128
)

// Neblet interface breaks cycle import dependency and hides unused services.
type Neblet interface {
	Config() *nebletpb.Config
	StartPprof(string) error
	BlockChain() *core.BlockChain
	AccountManager() *account.Manager
	NetService() nebnet.Service
	EventEmitter() *core.EventEmitter
	Consensus() consensus.Consensus
}

// GRPCServer server interface for api & management etc.
type GRPCServer interface {
	// Start start server
	Start() error

	// Stop stop server
	Stop()

	// Neblet return neblet
	Neblet() Neblet

	RunGateway() error
}

// Server is the RPC server type.
type Server struct {
	neblet Neblet

	rpcServer *grpc.Server

	rpcConfig *nebletpb.RPCConfig
}

// NewServer creates a new RPC server and registers the rpc endpoints.
func NewServer(neblet Neblet) *Server {
	cfg := neblet.Config().Rpc

	rpc := grpc.NewServer(grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(loggingStream)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(loggingUnary)))

	srv := &Server{neblet: neblet, rpcServer: rpc, rpcConfig: cfg}
	api := &APIService{server: srv}
	admin := &AdminService{server: srv}

	rpcpb.RegisterApiServiceServer(rpc, api)
	rpcpb.RegisterAdminServiceServer(rpc, admin)
	// Register reflection service on gRPC server.
	// TODO: Enable reflection only for testing mode.
	reflection.Register(rpc)

	return srv
}

// Start starts the rpc server and serves incoming requests.
func (s *Server) Start() error {
	logging.CLog().Info("Starting RPC GRPCServer...")

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

func (s *Server) start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to listen to RPC GRPCServer")
		return err
	}

	logging.CLog().WithFields(logrus.Fields{
		"address": addr,
	}).Info("Started RPC GRPCServer.")

	// Limit the total number of grpc connections.
	connectionLimits := s.rpcConfig.ConnectionLimits
	if connectionLimits == 0 {
		connectionLimits = DefaultConnectionLimits
	}

	listener = netutil.LimitListener(listener, int(connectionLimits))

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
func (s *Server) RunGateway() error {
	//time.Sleep(3 * time.Second)
	rpcListen := s.rpcConfig.RpcListen[0]
	gatewayListen := s.rpcConfig.HttpListen
	httpModule := s.rpcConfig.HttpModule
	logging.CLog().WithFields(logrus.Fields{
		"rpc-server":  rpcListen,
		"http-server": gatewayListen,
	}).Info("Starting RPC Gateway GRPCServer...")

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
func (s *Server) Stop() {
	logging.CLog().WithFields(logrus.Fields{
		"listen": s.rpcConfig.RpcListen,
	}).Info("Stopping RPC GRPCServer and Gateway...")

	s.rpcServer.Stop()

	logging.CLog().Info("Stopped RPC GRPCServer and Gateway.")
}

// Neblet returns weak reference to Neblet.
func (s *Server) Neblet() Neblet {
	return s.neblet
}
