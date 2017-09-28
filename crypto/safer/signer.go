package safer

import (
	"errors"

	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/ecdsa"
)

const (
	// ECDSA a type of signer
	ECDSA uint8 = iota
)

var (
	// ErrAlgorithmInvalid invalid AlgorithmType for sign.
	ErrAlgorithmInvalid = errors.New("invalid AlgorithmType for sign")
)

// Sign signs with a specified algorithm ，returns the signature
func Sign(alg uint8, priv keystore.PrivateKey, data []byte) ([]byte, error) {
	switch alg {
	case ECDSA:
		return ecdsaSign(priv, data)
	default:
		return nil, ErrAlgorithmInvalid
	}
}

func ecdsaSign(priv keystore.PrivateKey, data []byte) ([]byte, error) {
	signer := &ecdsa.Signature{}
	signer.InitSign(priv)
	signature, err := signer.Sign(data)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

// RecoverPub returns a publickey, which recover from data and signature
func RecoverPub(alg uint8, data []byte, signature []byte) ([]byte, error) {
	switch alg {
	case ECDSA:
		return ecdsaRecover(data, signature)
	default:
		return nil, ErrAlgorithmInvalid
	}
}

func ecdsaRecover(data []byte, signature []byte) ([]byte, error) {
	pub, err := ecdsa.RecoverPublicKey(data, signature)
	if err != nil {
		return nil, err
	}
	return ecdsa.FromPublicKey(pub)
}

// Verify verify signature
func Verify(alg uint8, data []byte, signature []byte, pub []byte) (bool, error) {
	switch alg {
	case ECDSA:
		return ecdsaVerify(data, signature, pub)
	default:
		return false, ErrAlgorithmInvalid
	}
}

func ecdsaVerify(data []byte, signature []byte, pub []byte) (bool, error) {
	publickey, err := ecdsa.ToPublicKey(pub)
	if err != nil {
		return false, err
	}

	pubStoreKey := ecdsa.NewPublicStoreKey(*publickey)
	signer := &ecdsa.Signature{}
	signer.InitVerify(pubStoreKey)
	return signer.Verify(data, signature)
}
