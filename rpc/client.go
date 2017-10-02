package rpc

import (
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Dial returns a client connection.
func Dial() (*grpc.ClientConn, error) {
	// TODO: support secure connection.
	conn, err := grpc.Dial(Address(), grpc.WithInsecure())
	if err != nil {
		log.Warn("rpc.Dial() failed: ", err)
	}
	return conn, err
}
