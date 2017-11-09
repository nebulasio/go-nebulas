package rpc

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
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

	return http.ListenAndServe(gateWayAddress(int(gatewayPort)), allowCORS(mux))
}

func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	glog.Infof("preflight request for %s", r.URL.Path)
	return
}

func apiAddress(port int) string {
	addr := fmt.Sprintf("localhost:%d", port)
	return addr
}

func gateWayAddress(port int) string {
	addr := fmt.Sprintf(":%d", port)
	return addr
}
