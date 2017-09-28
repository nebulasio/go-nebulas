package cipher

import (
	"errors"

	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/ecdsa"
)

// Algorithm type alias
type Algorithm uint8

const (
	// SECP256K1 a type of signer
	SECP256K1 Algorithm = iota
)

var (
	// ErrAlgorithmInvalid invalid Algorithm for sign.
	ErrAlgorithmInvalid = errors.New("invalid Algorithm")
)

// GetSignature returns the specified algorithm Signature
func GetSignature(alg Algorithm) (keystore.Signature, error) {
	switch alg {
	case SECP256K1:
		secp256k1 := &ecdsa.Signature{}
		return secp256k1, nil
	default:
		return nil, ErrAlgorithmInvalid
	}
}
