package rpc

import (
	"encoding/json"
	"flag"
	"net/http"
	"strings"

	"github.com/nebulasio/go-nebulas/rpc/pb"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/nebulasio/grpc-gateway/runtime"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// const
const (
	API   = "api"
	Admin = "admin"
)

// Run start gateway proxy to mapping grpc to http.
func Run(rpcListen string, gatewayListen []string, httpModule []string, httpLimit int32) error {
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
		err := http.ListenAndServe(v, allowCORS(mux, httpLimit))
		if err != nil {
			return err
		}
	}

	return nil
}

func allowCORS(h http.Handler, httpLimit int32) http.Handler {
	httpCh := make(chan bool, httpLimit)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		select {
		case httpCh <- true:
			defer func() { <-httpCh }()
			if origin := r.Header.Get("Origin"); origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
					preflightHandler(w, r)
					return
				}
			}
			h.ServeHTTP(w, r)
		default:
			statusUnavailableHandler(w, r)
		}

	})
}
func statusUnavailableHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusServiceUnavailable)

	w.Write([]byte("{\"err:\",\"Sorry, we received too many simultaneous requests.\nPlease try again later.\"}"))

}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	return
}

type errorBody struct {
	Err string `json:"error,omitempty"`
}

func errorHandler(ctx context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	const fallback = "failed to marshal error message"

	w.Header().Set("Content-type", marshaler.ContentType())
	if grpc.Code(err) == codes.Unknown {
		w.WriteHeader(runtime.HTTPStatusFromCode(codes.OutOfRange))
	} else {
		w.WriteHeader(runtime.HTTPStatusFromCode(grpc.Code(err)))
	}
	jErr := json.NewEncoder(w).Encode(errorBody{
		Err: grpc.ErrorDesc(err),
	})

	if jErr != nil {
		jsonFallback, tmpErr := json.Marshal(errorBody{Err: fallback})
		if tmpErr != nil {
			logging.VLog().WithFields(logrus.Fields{
				"error":        tmpErr,
				"jsonFallback": jsonFallback,
			}).Error("fall to marshal fallback msg")
		}
		_, tmpErr = w.Write(jsonFallback)
		if tmpErr != nil {
			logging.VLog().WithFields(logrus.Fields{
				"error": tmpErr,
			}).Error("fail to write fallback msg")
		}
	}
}
