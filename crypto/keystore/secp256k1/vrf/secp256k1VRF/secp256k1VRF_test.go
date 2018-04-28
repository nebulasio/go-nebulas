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

package secp256k1VRF

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/nebulasio/go-nebulas/util/byteutils"
	// _ "github.com/google/trillian/crypto/keys/der/proto"
)

const (
	// private key in hex
	privKey = `b02b430d4a9d7120b65038452a6da3f3c716829e5be3665adf934d4798d96ed7`
	// public key in hex
	pubKey = `04e4d0dde330c0b8d8d8b1b2071aa75c3e94f200a3d11ca1d908644eee50c8833a816dc0b2d003fc66187ef6750a56e1b3004d32e6159008400ab92f2ded7b4544`
)

func TestH1(t *testing.T) {
	for i := 0; i < 10000; i++ {
		m := make([]byte, 100)
		if _, err := rand.Read(m); err != nil {
			t.Fatalf("Failed generating random message: %v", err)
		}
		x, y := H1(m)
		if x == nil {
			t.Errorf("H1(%v)=%v, want curve point", m, x)
		}
		// attention: should not use curve.Params(), that will use elliptic.IsOnCurve()
		if got := curve.IsOnCurve(x, y); !got {
			t.Errorf("H1(%v)=[%v, %v], is not on curve", m, x, y)
		}
	}
}

func TestH2(t *testing.T) {
	l := 32
	for i := 0; i < 10000; i++ {
		m := make([]byte, 100)
		if _, err := rand.Read(m); err != nil {
			t.Fatalf("Failed generating random message: %v", err)
		}
		x := H2(m)
		if got := len(x.Bytes()); got < 1 || got > l {
			t.Errorf("len(h2(%v)) = %v, want: 1 <= %v <= %v", m, got, got, l)
		}
	}
}

func TestEmmm(t *testing.T) {
	k, pk := GenerateKey()
	m1 := []byte("da30b4ed14affb62b3719fb5e6952d3733e84e53fe6e955f8e46da503300c985")
	index1, proof1 := k.Evaluate(m1)
	fmt.Println("evaluate index = ", index1)
	fmt.Println("evaluate proof = ", proof1)

	start := time.Now()
	for i := 0; i < 100; i++ {
		_, _ = k.Evaluate(m1)
	}
	end := time.Now()
	fmt.Println("evaluate x100 total time: ", end.Sub(start))

	index, _ := pk.ProofToHash(m1, proof1)
	fmt.Println("proof index = ", index)

	start = time.Now()
	for i := 0; i < 100; i++ {
		_, _ = pk.ProofToHash(m1, proof1)
	}
	end = time.Now()
	fmt.Println("proof x100 total time: ", end.Sub(start))
}

func TestVRF(t *testing.T) {
	k, pk := GenerateKey()

	m1 := []byte("data1")
	m2 := []byte("data2")
	m3 := []byte("data2")
	index1, proof1 := k.Evaluate(m1)
	index2, proof2 := k.Evaluate(m2)
	index3, proof3 := k.Evaluate(m3)
	for _, tc := range []struct {
		m     []byte
		index [32]byte
		proof []byte
		err   error
	}{
		{m1, index1, proof1, nil},
		{m2, index2, proof2, nil},
		{m3, index3, proof3, nil},
		{m3, index3, proof2, nil},
		{m3, index3, proof1, ErrInvalidVRF},
	} {
		index, err := pk.ProofToHash(tc.m, tc.proof)
		if got, want := err, tc.err; got != want {
			t.Errorf("ProofToHash(%s, %x): %v, want %v", tc.m, tc.proof, got, want)
		}
		if err != nil {
			continue
		}
		if got, want := index, tc.index; got != want {
			t.Errorf("ProofToInex(%s, %x): %x, want %x", tc.m, tc.proof, got, want)
		}
	}
}

func TestProofToHash(t *testing.T) {
	bytes, _ := byteutils.FromHex(pubKey)
	pk, err := NewVRFVerifierFromRawKey(bytes)
	if err != nil {
		t.Errorf("NewVRFSigner failure: %v", err)
	}

	for _, tc := range []struct {
		m     []byte
		index [32]byte
		proof []byte
	}{
		{
			m:     []byte("data1"),
			index: h2i("a2f4f844d46240a86790c177f21422f430b2803c7590f32625079fc13a5fe601"),
			proof: h2b("cc23d0e1e01a20bcee479e944c94febabb8e762fa64b9443fc9dc31d3332e3a7024f4adc2cda4e8847fe67f47ab0084b677996e9325d31840531a2f91d6a5d7d04e54044c12dd5ab7b90a57117a85d6307125496ada896d9823c860c4f492c0096c714705d58ee7d66ee6cffb5f1320c5eab7f92490b0f5759145588efa0b0537d"),
		},
		{
			m:     []byte("data2"),
			index: h2i("008a288a33a2620458a26b6c995d9c16ca46c293562db76985bd1b2a159efc76"),
			proof: h2b("888e0d3191af542c40d0d8b15255e106a133ec9b219b6e26900e07a252e6ab60e510423c34bf74cc602ae2be214bffadfd639793d0a3dccd0e7303be8d0de57604322cef265dfe906cebf30de74b14aa33723435eccea3153fedb5bea70e5c58a8969af97c27e50223bc3b9a8dd8f4a60ec363a78c957f366af075cf83cc43e61c"),
		},
		{
			m:     []byte("data3"),
			index: h2i("9bb53b519519a85c8c6c6739349168c42ae208aed7dadeababf5a067a6ac1313"),
			proof: h2b("96004eb1450c68fcb1ac83e0f09c5311089829762a5e8aecdba1c51d703250d79bc9ffeb72c0c5645da6c3d2d59a5c6428b1d3a0075d75b89bae8b539453e3af044a472b26f259bd5a84f05ec8fe1d7858d6f5606adcb6febeef113a2ff4ff69d5166ebbd3c3a78c451d751490eeb37fd39358fb2fad8ae218e3fc5177fe2e9b37"),
		},
	} {

		bytes, _ := byteutils.FromHex(privKey)
		k, err := NewVRFSignerFromRawKey(bytes)
		idx, proof := k.Evaluate(tc.m)
		fmt.Println("=======")
		fmt.Println(byteutils.Hex(idx[0:]))
		fmt.Println(byteutils.Hex(proof))

		index, err := pk.ProofToHash(tc.m, tc.proof)
		if err != nil {
			t.Errorf("ProofToHash(%s, %x): %v, want nil", tc.m, tc.proof, err)
			continue
		}
		if got, want := index, tc.index; got != want {
			t.Errorf("ProofToHash(%s, %x): %x, want %x", tc.m, tc.proof, got, want)
		}
	}
}

func TestReadFromOpenSSL(t *testing.T) {
	for _, tc := range []struct {
		priv string
		pub  string
	}{
		{privKey, pubKey},
	} {
		// Private VRF Key
		bytes, _ := byteutils.FromHex(tc.priv)
		signer, err := NewVRFSignerFromRawKey(bytes)
		if err != nil {
			t.Errorf("NewVRFSigner failure: %v", err)
		}

		// Public VRF key
		bytes, _ = byteutils.FromHex(tc.pub)
		verifier, err := NewVRFVerifierFromRawKey(bytes)
		if err != nil {
			t.Errorf("NewVRFSigner failure: %v", err)
		}

		// Evaluate and verify.
		m := []byte("M")
		_, proof := signer.Evaluate(m)
		if _, err := verifier.ProofToHash(m, proof); err != nil {
			t.Errorf("Failed verifying VRF proof")
		}
	}
}

func TestRightTruncateProof(t *testing.T) {
	k, pk := GenerateKey()

	data := []byte("data")
	_, proof := k.Evaluate(data)
	proofLen := len(proof)
	for i := 0; i < proofLen; i++ {
		proof = proof[:len(proof)-1]
		if _, err := pk.ProofToHash(data, proof); err == nil {
			t.Errorf("Verify unexpectedly succeeded after truncating %v bytes from the end of proof", i)
		}
	}
}

func TestLeftTruncateProof(t *testing.T) {
	k, pk := GenerateKey()

	data := []byte("data")
	_, proof := k.Evaluate(data)
	proofLen := len(proof)
	for i := 0; i < proofLen; i++ {
		proof = proof[1:]
		if _, err := pk.ProofToHash(data, proof); err == nil {
			t.Errorf("Verify unexpectedly succeeded after truncating %v bytes from the beginning of proof", i)
		}
	}
}

func TestBitFlip(t *testing.T) {
	k, pk := GenerateKey()

	data := []byte("data")
	_, proof := k.Evaluate(data)
	for i := 0; i < len(proof)*8; i++ {
		// Flip bit in position i.
		if _, err := pk.ProofToHash(data, flipBit(proof, i)); err == nil {
			t.Errorf("Verify unexpectedly succeeded after flipping bit %v of vrf", i)
		}
	}
}

func flipBit(a []byte, pos int) []byte {
	index := int(math.Floor(float64(pos) / 8))
	b := byte(a[index])
	b ^= (1 << uint(math.Mod(float64(pos), 8.0)))

	var buf bytes.Buffer
	buf.Write(a[:index])
	buf.Write([]byte{b})
	buf.Write(a[index+1:])
	return buf.Bytes()
}

func TestVectors(t *testing.T) {
	bytes, _ := byteutils.FromHex(privKey)
	k, err := NewVRFSignerFromRawKey(bytes)
	if err != nil {
		t.Errorf("NewVRFSigner failure: %v", err)
	}
	bytes, _ = byteutils.FromHex(pubKey)
	pk, err := NewVRFVerifierFromRawKey(bytes)
	if err != nil {
		t.Errorf("NewVRFSigner failure: %v", err)
	}
	for _, tc := range []struct {
		m     []byte
		index [32]byte
	}{
		{
			m:     []byte("test"),
			index: h2i("c095a258b89a5fbf0790e45cd2b1a31c1723f0f99c7df3df98e03eef2a4a25af"),
		},
		{
			m:     nil,
			index: h2i("19a1da136a3dadfd4ffb2e95d0b236b72e7bd448541a46fde595acd5052775cb"),
		},
	} {
		index, proof := k.Evaluate(tc.m)
		if got, want := index, tc.index; got != want {
			t.Errorf("Evaluate(%s).Index: %x, want %x", tc.m, got, want)
		}
		index2, err := pk.ProofToHash(tc.m, proof)
		if err != nil {
			t.Errorf("ProofToHash(%s): %v", tc.m, err)
		}
		if got, want := index2, index; got != want {
			t.Errorf("ProofToHash(%s): %x, want %x", tc.m, got, want)
		}
	}
}

func h2i(h string) [32]byte {
	b, err := hex.DecodeString(h)
	if err != nil {
		panic("Invalid hex")
	}
	var i [32]byte
	copy(i[:], b)
	return i
}

func h2b(h string) []byte {
	b, err := hex.DecodeString(h)
	if err != nil {
		panic("Invalid hex")
	}
	return b
}
