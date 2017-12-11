package rpc

import (
	"flag"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// const
const (
	API   = "api"
	Admin = "admin"
)

// Run start gateway proxy to mapping grpc to http.
func Run(rpcListen string, gatewayListen []string, httpModule []string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	echoEndpoint := flag.String("rpc", rpcListen, "")
	for _, v := range httpModule {
		switch v {
		case API:
			rpcpb.RegisterApiServiceHandlerFromEndpoint(ctx, mux, *echoEndpoint, opts)
		case Admin:
			rpcpb.RegisterAdminServiceHandlerFromEndpoint(ctx, mux, *echoEndpoint, opts)
		}
	}

	for _, v := range gatewayListen {
		err := http.ListenAndServe(v, allowCORS(mux))
		if err != nil {
			return err
		}
	}

	return nil
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
	return
}
