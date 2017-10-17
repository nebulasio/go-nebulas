package rpc

import (
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Dial returns a client connection.
func Dial(target string) (*grpc.ClientConn, error) {
	// TODO: support secure connection.
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		log.Warn("rpc.Dial() failed: ", err)
	}
	return conn, err
}
