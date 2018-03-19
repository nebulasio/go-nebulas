package rpc

import (
	"flag"
	"io"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
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

	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard,
		&runtime.JSONPb{OrigName: true, EmitDefaults: true}),
		runtime.WithProtoErrorHandler(errorHandler))
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
			w.Header().Set("Access-Control-Allow-Origin", "*")
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
func errorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	// return Internal when Marshal failed
	const fallback = `{"code": 13, "message": "failed to marshal error message"}`

	w.Header().Del("Trailer")
	w.Header().Set("Content-Type", marshaler.ContentType())

	s, ok := status.FromError(err)
	if !ok {
		s = status.New(codes.Unknown, err.Error())
	}

	buf, merr := marshaler.Marshal(s.Proto())
	if merr != nil {
		grpclog.Printf("Failed to marshal error message %q: %v", s.Proto(), merr)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, fallback); err != nil {
			grpclog.Printf("Failed to write response: %v", err)
		}
		return
	}

	//md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		grpclog.Printf("Failed to extract ServerMetadata from context")
	}

	//	runtime.handleForwardResponseServerMetadata(w, mux, md)
	//	runtime.handleForwardResponseTrailerHeader(w, md)
	if s.Code() == codes.Unknown {
		st := http.StatusBadRequest
		w.WriteHeader(st)
	} else {
		st := runtime.HTTPStatusFromCode(s.Code())
		w.WriteHeader(st)
	}
	if _, err := w.Write(buf); err != nil {
		grpclog.Printf("Failed to write response: %v", err)
	}

	//	handleForwardResponseTrailer(w, md)
}
