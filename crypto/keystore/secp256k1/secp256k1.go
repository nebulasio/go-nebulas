// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package secp256k1

/*
#cgo CFLAGS: -I./libsecp256k1
#cgo CFLAGS: -I./libsecp256k1/src/
#define USE_NUM_NONE
#define USE_FIELD_10X26
#define USE_FIELD_INV_BUILTIN
#define USE_SCALAR_8X32
#define USE_SCALAR_INV_BUILTIN
#define ENABLE_MODULE_RECOVERY
#define NDEBUG
#include "./libsecp256k1/src/secp256k1.c"
*/
import "C"

//#cgo CFLAGS: -std=gnu99 -Wno-error
//#cgo LDFLAGS: -lgmp
//#cgo CFLAGS: -Wno-error

import (
	"crypto/ecdsa"
	"errors"
	"unsafe"
)

var (
	// ErrInvalidMsgLen invalid message length
	ErrInvalidMsgLen = errors.New("invalid message length, need 32 bytes")

	// ErrInvalidSignature invalid signature length
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrInvalidPrivateKey invalid private key
	ErrInvalidPrivateKey = errors.New("invalid private key")

	// ErrInvalidPublicKey invalid public key
	ErrInvalidPublicKey = errors.New("invalid public key")

	// ErrSignFailed sign failed
	ErrSignFailed = errors.New("sign failed")

	// ErrRecoverFailed recover failed
	ErrRecoverFailed = errors.New("recovery failed")
)

var ctx *C.secp256k1_context

// use bitcoin's libsecp256k1 library
// use like https://github.com/btccom/secp256k1-go/blob/master/secp256k1/secp256k1.go

func init() {
	ctx = C.secp256k1_context_create(C.SECP256K1_CONTEXT_SIGN | C.SECP256K1_CONTEXT_VERIFY)
}

// SeckeyVerify check private is ok for secp256k1
func SeckeyVerify(priv *ecdsa.PrivateKey) bool {
	seckey, err := FromECDSAPrivateKey(priv)
	if err != nil {
		return false
	}
	return C.secp256k1_ec_seckey_verify(ctx, cBuf(seckey)) == 1
}

// RecoverECDSAPublicKey recover verifies the compact signature "signature" of "hash"
func RecoverECDSAPublicKey(msg []byte, signature []byte) (*ecdsa.PublicKey, error) {
	if len(msg) != 32 {
		return nil, ErrInvalidMsgLen
	}
	if len(signature) != 65 {
		return nil, ErrInvalidSignature
	}
	var (
		sig    C.secp256k1_ecdsa_recoverable_signature
		pubkey C.secp256k1_pubkey
	)

	result := int(C.secp256k1_ecdsa_recoverable_signature_parse_compact(ctx, &sig, (*C.uchar)(unsafe.Pointer(&signature[0])), (C.int(signature[64]))))
	if result != 1 {
		return nil, ErrRecoverFailed
	}
	if int(C.secp256k1_ecdsa_recover(ctx, &pubkey, &sig, cBuf(msg))) != 1 {
		return nil, ErrRecoverFailed
	}
	output := make([]C.uchar, 65)
	outputLen := C.size_t(65)
	result = int(C.secp256k1_ec_pubkey_serialize(ctx, &output[0], &outputLen, &pubkey, C.SECP256K1_EC_UNCOMPRESSED))
	if result != 1 {
		return nil, ErrRecoverFailed
	}
	return ToECDSAPublicKey(goBytes(output, C.int(outputLen)))
}

// Sign sign hash with private key
func Sign(msg []byte, priv *ecdsa.PrivateKey) ([]byte, error) {
	if len(msg) != 32 {
		return nil, ErrInvalidMsgLen
	}
	seckey, err := FromECDSAPrivateKey(priv)
	if err != nil {
		return nil, err
	}
	if C.secp256k1_ec_seckey_verify(ctx, cBuf(seckey)) != 1 {
		return nil, ErrInvalidPrivateKey
	}

	var (
		noncefunc = C.secp256k1_nonce_function_rfc6979
		sigstruct C.secp256k1_ecdsa_recoverable_signature
	)
	if C.secp256k1_ecdsa_sign_recoverable(ctx, &sigstruct, cBuf(msg), cBuf(seckey), noncefunc, nil) == 0 {
		return nil, ErrSignFailed
	}

	var (
		sig   = make([]byte, 65)
		recid C.int
	)
	C.secp256k1_ecdsa_recoverable_signature_serialize_compact(ctx, cBuf(sig), &recid, &sigstruct)
	sig[64] = byte(recid) // add back recid to get 65 bytes sig
	return sig, nil
}

// Verify verify with public key
func Verify(msg []byte, signature []byte, pub *ecdsa.PublicKey) (bool, error) {
	if len(msg) != 32 {
		return false, ErrInvalidMsgLen
	}
	pubdata, err := FromECDSAPublicKey(pub)
	if err != nil {
		return false, err
	}
	var (
		sig    C.secp256k1_ecdsa_signature
		pubkey C.secp256k1_pubkey
	)
	result := int(C.secp256k1_ec_pubkey_parse(ctx, &pubkey, cBuf(pubdata), C.size_t(len(pubdata))))
	if result != 1 {
		return false, ErrInvalidPublicKey
	}
	result = int(C.secp256k1_ecdsa_signature_parse_compact(ctx, &sig, cBuf(signature)))
	if result != 1 {
		return false, ErrInvalidSignature
	}
	result = int(C.secp256k1_ecdsa_verify(ctx, &sig, cBuf(msg), &pubkey))
	return result == 1, nil
}

func cBuf(goSlice []byte) *C.uchar {
	return (*C.uchar)(unsafe.Pointer(&goSlice[0]))
}

func goBytes(cSlice []C.uchar, size C.int) []byte {
	return C.GoBytes(unsafe.Pointer(&cSlice[0]), size)
}
