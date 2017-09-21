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

package core

import (
	"errors"

	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/util/byteutils"

	"crypto/rand"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/ecdsa"
	"github.com/nebulasio/go-nebulas/crypto/keystore/key"
	log "github.com/sirupsen/logrus"
)

const (
	// AddressDataLength the length of data of address in byte.
	AddressDataLength = 20

	// AddressChecksumLength the checksum of address in byte.
	AddressChecksumLength = 4

	// AddressLength the length of address in byte.
	AddressLength = AddressDataLength + AddressChecksumLength
)

var (
	// ErrInvalidAddress invalid address error.
	ErrInvalidAddress = errors.New("address: invalid address")

	// ErrInvalidAddressDataLength invalid data length error.
	ErrInvalidAddressDataLength = errors.New("address: invalid address data length")
)

/*
Address Similar to Bitcoin and Ethereum, Nebulas also adopts elliptic curve algorithm as its basic encryption algorithm for Nebulas accounts. A user’s private key is a randomly generated 256-bit binary number, based on which a 64-byte public key can be generated via elliptic curve multiplication. Bitcoin and Ethereum addresses are computed by public key via the deterministic Hash algorithm, and the difference between them lies in: Bitcoin address has the checksum design aiming to prevent a user from sending Bitcoins to a wrong user account accidentally due to entry of several incorrect characters; while Ethereum doesn’t have such checksum design.

We believe that checksum design is reasonable from the perspective of users, so Nebulas address also includes checksum, for which the specific calculation method is provided as follows:

  Data = sha3_256(Public Key)[-20:]
  CheckSum = sha3_256(Data)[0:4]
  Address = "0x" + Hex(Data + CheckSum)

The last 20 bytes of SHA3-256 digest of a public key serve as the major component of an address, for which another SHA3-256 digest should be conducted and the first 4 bytes should be used as a checksum, which is equivalent to the practice of adding a 4-byte checksum to the end of an Ethereum address. For example:

The standard address of Alice’s Ethereum wallet is 0xdf4d22611412132d3e9bd322f82e2940674ec1bc;
The final address of Nebulas Wallet should be:  0xdf4d22611412132d3e9bd322f82e2940674ec1bc03b20e40

In addition to standard address with 50 characters, we also support extended address in order to ensure the security of transfers conducted by users. The traditional bank transfer design is used for reference: In the process of a bank transfer, bank card number of the remittee should be verified, in addition to which the remitter must enter the name of the remittee. The transfer can be correctly processed only when the bank card number and the name match each other. The generating algorithm for extended address is described as follows:

  Data = sha3_256(Public Key)[-20:]
  CheckSum = sha3_256(Data)[0:4]
  Address = "0x" + Hex(Data + CheckSum)

  ExtData = Utf8Bytes({Nickname or any string})
  ExtHash = sha3_256(Data + ExtData)[0:2]
  ExtAddress = Address + Hex(ExtHash)

An extended address is generated through addition of 2-byte extended verification to the end of a standard address and contains a total of 54 characters. Addition of extended information allows the addition of another element verification to the Nebulas Wallet APP. For example:

	The standard address of Alice’s wallet is  0xdf4d22611412132d3e9bd322f82e2940674ec1bc03b20e40, and the extended address after addition of the nickname "alice" should be 0xdf4d22611412132d3e9bd322f82e2940674ec1bc03b20e40e345.
	Alice tells Bob the extended address 0xdf4d22611412132d3e9bd322f82e2940674ec1bc03b20e40e345 and her nickname alice.
	Bob enters 0xdf4d22611412132d3e9bd322f82e2940674ec1bc03b20e40e345 and alice in the Wallet App.
	The Wallet App verifies the consistency between the wallet address and the nickname in order to avoid the circumstance that Bob enters the account number of another user by mistake.
*/
type Address struct {
	address []byte
}

// ToHex convert address to hex
func (a *Address) ToHex() string {
	return byteutils.Hex(a.address)
}

// TestKS return a test keystore
func TestKS() *keystore.Keystore {
	ks := keystore.NewKeystore()
	p1, _ := ecdsa.NewPrivateKey(rand.Reader)
	ps1 := ecdsa.NewPrivateStoreKey(p1)
	addr1, _ := NewAddressFromKey(ps1)
	p2, _ := ecdsa.NewPrivateKey(rand.Reader)
	ps2 := ecdsa.NewPrivateStoreKey(p2)
	addr2, _ := NewAddressFromKey(ps2)
	p3, _ := ecdsa.NewPrivateKey(rand.Reader)
	ps3 := ecdsa.NewPrivateStoreKey(p3)
	addr3, _ := NewAddressFromKey(ps3)

	ks.SetKeyPassphrase(key.Alias(addr1.ToHex()), ps1, []byte("test"))
	ks.SetKeyPassphrase(key.Alias(addr2.ToHex()), ps2, []byte("test"))
	ks.SetKeyPassphrase(key.Alias(addr3.ToHex()), ps3, []byte("test"))

	pass1, _ := key.NewPassphrase([]byte("test"))
	ks.Unlock(key.Alias(addr1.ToHex()), pass1)

	pass2, _ := key.NewPassphrase([]byte("test"))
	ks.Unlock(key.Alias(addr2.ToHex()), pass2)

	pass3, _ := key.NewPassphrase([]byte("test"))
	ks.Unlock(key.Alias(addr3.ToHex()), pass3)

	return ks
}

// NewAddressFromKey create new #Address according to data key.
func NewAddressFromKey(k key.Key) (*Address, error) {
	pbyte := []byte{}
	switch k.(type) {
	case *ecdsa.PrivateStoreKey:
		pub, err := k.(*ecdsa.PrivateStoreKey).EncodedPub()
		if err != nil {
			return nil, err
		}
		pbyte = pub[len(pub)-AddressDataLength:]
	case *ecdsa.PublicStoreKey:
		pub, err := k.Encoded()
		if err != nil {
			return nil, err
		}
		pbyte = pub[len(pub)-AddressDataLength:]
	default:
		return nil, errors.New("Address: key type not support")
	}
	return NewAddress(pbyte)
}

// NewAddress create new #Address according to data bytes.
func NewAddress(s []byte) (*Address, error) {
	if len(s) != AddressDataLength {
		log.Errorf("invalid address data: length of s is %d, expected to %d.", len(s), AddressDataLength)
		return nil, ErrInvalidAddressDataLength
	}

	cs := checkSum(s)
	return &Address{address: append(s, cs...)}, nil
}

// Parse parse address string.
func Parse(s string) (*Address, error) {
	r, err := byteutils.FromHex(s)
	if err != nil {
		log.WithFields(log.Fields{
			"s": s, "err": err,
		}).Error("invalid address: string should be encoded in Hexadecimal.")
		return nil, ErrInvalidAddress
	}

	return ParseFromBytes(r)
}

// ParseFromBytes parse address from bytes.
func ParseFromBytes(s []byte) (*Address, error) {
	if len(s) != AddressLength {
		log.Errorf("invalid address: length of s is %d, expected to %d.", len(s), AddressLength)
		return nil, ErrInvalidAddress
	}

	data := s[:AddressDataLength]
	cs := s[AddressDataLength:AddressLength]
	dcs := checkSum(data)

	for i := 0; i < AddressChecksumLength; i++ {
		if dcs[i] != cs[i] {
			log.Errorf("invalid address: checksum is %s, expected to %s.", cs, dcs)
			return nil, ErrInvalidAddress
		}
	}

	return &Address{address: s}, nil
}

func checkSum(data []byte) []byte {
	return hash.Sha3256(data)[:AddressChecksumLength]
}
