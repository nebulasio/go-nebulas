package rpc

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Run start gateway proxy to mapping grpc to http.
func Run(apiPort uint32, gatewayPort uint32) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	echoEndpoint := flag.String("api_service_endpoint", apiAddress(int(apiPort)), "endpoint of api_service")
	err := rpcpb.RegisterAPIServiceHandlerFromEndpoint(ctx, mux, *echoEndpoint, opts)
	if err != nil {
		return err
	}

	return http.ListenAndServe(gateWayAddress(int(gatewayPort)), mux)
}

func apiAddress(port int) string {
	addr := fmt.Sprintf("localhost:%d", port)
	return addr
}

func gateWayAddress(port int) string {
	addr := fmt.Sprintf(":%d", port)
	return addr
}
