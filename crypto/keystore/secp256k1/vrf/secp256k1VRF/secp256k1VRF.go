// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Implements a verifiable random function using curve secp256k1.

package secp256k1VRF

// Discrete Log based VRF from Appendix A of CONIKS:
// http://www.jbonneau.com/doc/MBBFF15-coniks.pdf
// based on "Unique Ring Signatures, a Practical Construction"
// http://fc13.ifca.ai/proc/5-1.pdf

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"io"
	"math/big"

	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1/bitelliptic"

	"github.com/google/keytransparency/core/crypto/vrf"
)

var (
	curve  = bitelliptic.S256()
	params = curve.Params()

	// ErrPointNotOnCurve occurs when a public key is not on the curve.
	ErrPointNotOnCurve = errors.New("point is not on the P256 curve")
	// ErrWrongKeyType occurs when a key is not an ECDSA key.
	ErrWrongKeyType = errors.New("not an ECDSA key")
	// ErrNoPEMFound occurs when attempting to parse a non PEM data structure.
	ErrNoPEMFound = errors.New("no PEM block found")
	// ErrInvalidVRF occurs when the VRF does not validate.
	ErrInvalidVRF = errors.New("invalid VRF proof")
)

// PublicKey holds a public VRF key.
type PublicKey struct {
	*ecdsa.PublicKey
}

// PrivateKey holds a private VRF key.
type PrivateKey struct {
	*ecdsa.PrivateKey
}

// GenerateKey generates a fresh keypair for this VRF
func GenerateKey() (vrf.PrivateKey, vrf.PublicKey) {
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil
	}

	return &PrivateKey{key}, &PublicKey{&key.PublicKey}
}

// H1 hashes m to a curve point
func H1(m []byte) (x, y *big.Int) {
	h := sha512.New()
	var i uint32
	byteLen := (params.BitSize + 7) >> 3
	for x == nil && i < 100 {
		// TODO: Use a NIST specified DRBG.
		h.Reset()
		binary.Write(h, binary.BigEndian, i)
		h.Write(m)
		r := []byte{2} // Set point encoding to "compressed", y=0.
		r = h.Sum(r)
		x, y = Unmarshal(curve, r[:byteLen+1])
		i++
	}
	return
}

var one = big.NewInt(1)

// H2 hashes to an integer [1,N-1]
func H2(m []byte) *big.Int {
	// NIST SP 800-90A § A.5.1: Simple discard method.
	byteLen := (params.BitSize + 7) >> 3
	h := sha512.New()
	for i := uint32(0); ; i++ {
		// TODO: Use a NIST specified DRBG.
		h.Reset()
		binary.Write(h, binary.BigEndian, i)
		h.Write(m)
		b := h.Sum(nil)
		k := new(big.Int).SetBytes(b[:byteLen])
		if k.Cmp(new(big.Int).Sub(params.N, one)) == -1 {
			return k.Add(k, one)
		}
	}
}

// Evaluate returns the verifiable unpredictable function evaluated at m
func (k PrivateKey) Evaluate(m []byte) (index [32]byte, proof []byte) {
	nilIndex := [32]byte{}
	// Prover chooses r <-- [1,N-1]
	r, _, _, err := generateKeyFromCurve(curve, rand.Reader)
	if err != nil {
		return nilIndex, nil
	}
	ri := new(big.Int).SetBytes(r)

	// H = H1(m)
	Hx, Hy := H1(m)

	// VRF_k(m) = [k]H
	sHx, sHy := curve.ScalarMult(Hx, Hy, k.D.Bytes())

	// vrf := elliptic.Marshal(curve, sHx, sHy) // 65 bytes.
	vrf := curve.Marshal(sHx, sHy) // 65 bytes.

	// G is the base point
	// s = H2(G, H, [k]G, VRF, [r]G, [r]H)
	rGx, rGy := curve.ScalarBaseMult(r)
	rHx, rHy := curve.ScalarMult(Hx, Hy, r)
	var b bytes.Buffer
	b.Write(curve.Marshal(params.Gx, params.Gy))
	b.Write(curve.Marshal(Hx, Hy))
	b.Write(curve.Marshal(k.PublicKey.X, k.PublicKey.Y))
	b.Write(vrf)
	b.Write(curve.Marshal(rGx, rGy))
	b.Write(curve.Marshal(rHx, rHy))
	s := H2(b.Bytes())

	// t = r−s*k mod N
	t := new(big.Int).Sub(ri, new(big.Int).Mul(s, k.D))
	t.Mod(t, params.N)

	// Index = H(vrf)
	index = sha256.Sum256(vrf)

	// Write s, t, and vrf to a proof blob. Also write leading zeros before s and t
	// if needed.
	var buf bytes.Buffer
	buf.Write(make([]byte, 32-len(s.Bytes())))
	buf.Write(s.Bytes())
	buf.Write(make([]byte, 32-len(t.Bytes())))
	buf.Write(t.Bytes())
	buf.Write(vrf)

	return index, buf.Bytes()
}

// ProofToHash asserts that proof is correct for m and outputs index.
func (pk *PublicKey) ProofToHash(m, proof []byte) (index [32]byte, err error) {
	nilIndex := [32]byte{}
	// verifier checks that s == H2(m, [t]G + [s]([k]G), [t]H1(m) + [s]VRF_k(m))
	if got, want := len(proof), 64+65; got != want {
		return nilIndex, ErrInvalidVRF
	}

	// Parse proof into s, t, and vrf.
	s := proof[0:32]
	t := proof[32:64]
	vrf := proof[64 : 64+65]

	// uHx, uHy := elliptic.Unmarshal(curve, vrf)
	uHx, uHy := curve.Unmarshal(vrf) //////???
	if uHx == nil {
		return nilIndex, ErrInvalidVRF
	}

	// [t]G + [s]([k]G) = [t+ks]G
	tGx, tGy := curve.ScalarBaseMult(t)
	ksGx, ksGy := curve.ScalarMult(pk.X, pk.Y, s)
	tksGx, tksGy := curve.Add(tGx, tGy, ksGx, ksGy)

	// H = H1(m)
	// [t]H + [s]VRF = [t+ks]H
	Hx, Hy := H1(m)
	tHx, tHy := curve.ScalarMult(Hx, Hy, t)
	sHx, sHy := curve.ScalarMult(uHx, uHy, s)
	tksHx, tksHy := curve.Add(tHx, tHy, sHx, sHy)

	//   H2(G, H, [k]G, VRF, [t]G + [s]([k]G), [t]H + [s]VRF)
	// = H2(G, H, [k]G, VRF, [t+ks]G, [t+ks]H)
	// = H2(G, H, [k]G, VRF, [r]G, [r]H)
	var b bytes.Buffer
	b.Write(curve.Marshal(params.Gx, params.Gy))
	b.Write(curve.Marshal(Hx, Hy))
	b.Write(curve.Marshal(pk.X, pk.Y))
	b.Write(vrf)
	b.Write(curve.Marshal(tksGx, tksGy))
	b.Write(curve.Marshal(tksHx, tksHy))
	h2 := H2(b.Bytes())

	// Left pad h2 with zeros if needed. This will ensure that h2 is padded
	// the same way s is.
	var buf bytes.Buffer
	buf.Write(make([]byte, 32-len(h2.Bytes())))
	buf.Write(h2.Bytes())

	if !hmac.Equal(s, buf.Bytes()) {
		return nilIndex, ErrInvalidVRF
	}
	return sha256.Sum256(vrf), nil
}

// NewFromWrappedKey creates a VRF signer object from an encrypted private key.
// The opaque private key must resolve to an `ecdsa.PrivateKey` in order to work.
// func NewFromWrappedKey(ctx context.Context, wrapped proto.Message) (vrf.PrivateKey, error) {
// 	// Unwrap.
// 	signer, err := keys.NewSigner(ctx, wrapped)
// 	if err != nil {
// 		return nil, err
// 	}

// 	switch key := signer.(type) {
// 	case *ecdsa.PrivateKey:
// 		return NewVRFSigner(key)
// 	default:
// 		return nil, fmt.Errorf("NewSigner().type: %T, want ecdsa.PrivateKey", key)
// 	}
// }

// NewVRFSigner creates a signer object from a private key.
func NewVRFSigner(key *ecdsa.PrivateKey) (vrf.PrivateKey, error) {
	if *(key.Params()) != *curve.Params() {
		return nil, ErrPointNotOnCurve
	}
	if !curve.IsOnCurve(key.X, key.Y) {
		return nil, ErrPointNotOnCurve
	}
	return &PrivateKey{key}, nil
}

// Public returns the corresponding public key as bytes.
func (k PrivateKey) Public() crypto.PublicKey {
	return &k.PublicKey
}

// NewVRFVerifier creates a verifier object from a public key.
func NewVRFVerifier(pubkey *ecdsa.PublicKey) (vrf.PublicKey, error) {
	if *(pubkey.Params()) != *curve.Params() {
		return nil, ErrPointNotOnCurve
	}
	if !curve.IsOnCurve(pubkey.X, pubkey.Y) {
		return nil, ErrPointNotOnCurve
	}
	return &PublicKey{pubkey}, nil
}

// NewVRFSignerFromPEM creates a vrf private key from a PEM data structure.
// func NewVRFSignerFromPEM(b []byte) (vrf.PrivateKey, error) {
// 	p, _ := pem.Decode(b)
// 	if p == nil {
// 		return nil, ErrNoPEMFound
// 	}
// 	return NewVRFSignerFromRawKey(p.Bytes)
// }

// NewVRFSignerFromRawKey returns the private key from a raw private key bytes.
func NewVRFSignerFromRawKey(b []byte) (vrf.PrivateKey, error) {
	k, err := secp256k1.ToECDSAPrivateKey(b)
	if err != nil {
		return nil, err
	}
	return NewVRFSigner(k)
}

// NewVRFVerifierFromPEM creates a vrf public key from a PEM data structure.
// func NewVRFVerifierFromPEM(b []byte) (vrf.PublicKey, error) {
// 	p, _ := pem.Decode(b)
// 	if p == nil {
// 		return nil, ErrNoPEMFound
// 	}
// 	return NewVRFVerifierFromRawKey(p.Bytes)
// }

// NewVRFVerifierFromRawKey returns the public key from a raw public key bytes.
func NewVRFVerifierFromRawKey(b []byte) (vrf.PublicKey, error) {
	k, err := secp256k1.ToECDSAPublicKey(b)
	if err != nil {
		return nil, err
	}
	return NewVRFVerifier(k)
}

var mask = []byte{0xff, 0x1, 0x3, 0x7, 0xf, 0x1f, 0x3f, 0x7f}

func generateKeyFromCurve(curve *bitelliptic.BitCurve, rand io.Reader) (priv []byte, x, y *big.Int, err error) {
	N := curve.Params().N
	bitSize := N.BitLen()
	byteLen := (bitSize + 7) >> 3
	priv = make([]byte, byteLen)

	for x == nil {
		_, err = io.ReadFull(rand, priv)
		if err != nil {
			return
		}
		// We have to mask off any excess bits in the case that the size of the
		// underlying field is not a whole number of bytes.
		priv[0] &= mask[bitSize%8]
		// This is because, in tests, rand will return all zeros and we don't
		// want to get the point at infinity and loop forever.
		priv[1] ^= 0x42

		// If the scalar is out of range, sample another random number.
		if new(big.Int).SetBytes(priv).Cmp(N) >= 0 {
			continue
		}

		x, y = curve.ScalarBaseMult(priv)
	}
	return
}
