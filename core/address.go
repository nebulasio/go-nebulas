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
	"math/big"

	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

// AddressType address type
type AddressType byte

// address type enum
const (
	AccountAddress AddressType = iota
	ContractAddress
)

// const
const (
	HeadPadding byte = 1
)

const (
	// AddressPaddingLength the length of headpadding in byte
	AddressPaddingLength = 1

	// AddressTypeLength the length of address type in byte
	AddressTypeLength = 1
	// AddressDataLength the length of data of address in byte.
	AddressDataLength = 20

	// AddressChecksumLength the checksum of address in byte.
	AddressChecksumLength = 4

	// AddressLength the length of address in byte.
	AddressLength = AddressPaddingLength + AddressTypeLength + AddressDataLength + AddressChecksumLength
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
	address byteutils.Hash
}

// Bytes returns address bytes
func (a *Address) Bytes() []byte {
	return a.address
}

// String returns address string
func (a *Address) String() string {
	if a == nil || len(a.address) != AddressLength {
		return ""
	}

	n := new(big.Int)
	n.SetBytes(a.address)
	n.Mul(n, hash.Base58BigRadix)
	n.Add(n, hash.Base58Nebulas)

	return hash.EncodeBase58(n.Bytes())
}

// Equals compare two Address. True is equal, otherwise false.
func (a *Address) Equals(b *Address) bool {
	if a == nil {
		return b == nil
	}
	if b == nil {
		return false
	}
	return a.address.Equals(b.address)
}

// NewAddress create new #Address according to data bytes.
func NewAddress(t AddressType, args ...[]byte) (*Address, error) {
	if len(args) == 0 {
		return nil, ErrInvalidArgument
	}

	for _, v := range args {
		if len(v) == 0 {
			return nil, ErrInvalidArgument
		}
	}

	switch t {
	case AccountAddress, ContractAddress:
	default:
		return nil, ErrInvalidArgument
	}

	buffer := make([]byte, AddressLength)
	buffer[0] = HeadPadding // 1 padding
	buffer[1] = byte(t)     // address type

	sha := hash.Sha3256(args...)
	content := hash.Ripemd160(sha)
	copy(buffer[2:22], content)

	cs := checkSum(buffer[1:22])
	copy(buffer[22:], cs)

	return &Address{address: buffer}, nil
}

// NewAddressFromPublicKey return new address from publickey bytes
func NewAddressFromPublicKey(s []byte) (*Address, error) {
	return NewAddress(AccountAddress, s)
}

// NewContractAddressFromData return new contract address from bytes.
func NewContractAddressFromData(args ...[]byte) (*Address, error) {
	return NewAddress(ContractAddress, args...)
}

// AddressParse parse address string.
func AddressParse(s string) (*Address, error) {
	if len(s) != 36 || s[0] != 'n' {
		return nil, ErrInvalidAddressFormat
	}

	b := hash.DecodeBase58(s)
	if len(b) == 0 {
		return nil, ErrInvalidAddressCode
	}

	n := new(big.Int)
	n.SetBytes(b)
	n.Sub(n, hash.Base58Nebulas)
	if n.Sign() <= 0 {
		return nil, ErrInvalidAddress
	}
	n.Div(n, hash.Base58BigRadix)
	if n.Sign() <= 0 {
		return nil, ErrInvalidAddress
	}

	return AddressParseFromBytes(n.Bytes())
}

// AddressParseFromBytes parse address from bytes.
func AddressParseFromBytes(b []byte) (*Address, error) {
	if len(b) != AddressLength || b[0] != HeadPadding {
		return nil, ErrInvalidAddressFormat
	}

	switch AddressType(b[1]) {
	case AccountAddress, ContractAddress:
	default:
		return nil, ErrInvalidAddressType
	}

	if !byteutils.Equal(checkSum(b[1:22]), b[22:]) {
		return nil, ErrInvalidAddressChecksum
	}

	return &Address{address: b}, nil
}

func checkSum(data []byte) []byte {
	return hash.Sha3256(data)[:AddressChecksumLength]
}
