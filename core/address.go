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
	"crypto/ecdsa"

	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/utils/bytes"
)

const (
	CheckSumLength = 4
	AddressLength  = 20
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
	address string
}

// NewAddress return new @Address instance.
func NewAddress(address string) *Address {
	addr := &Address{address: address}
	return addr
}

// NewAddressWithPrivateKey generate Address from private key
func NewAddressWithPrivateKey(privateKey *ecdsa.PrivateKey) *Address {
	publicKeyBytes := crypto.FromECDSAPub(&privateKey.PublicKey)
	return NewAddressWithPublicKey(publicKeyBytes)
}

// NewAddressWithPublicKey generate Address from public key
func NewAddressWithPublicKey(publicKeyBytes []byte) *Address {
	data := hash.Sha3256(publicKeyBytes)[len(publicKeyBytes)-AddressLength:]
	checkSum := hash.Sha3256(data)[:CheckSumLength]
	address := bytes.Hex(append(data, checkSum...))
	addr := &Address{address: address}
	return addr
}

type ExtAddress struct {
	nick       string // nick or some comment for address
	address    Address
	extAddress string
}

// NewExtAddress return new @ExtAddress instance.
func NewExtAddress(nick string, publicKeyBytes []byte) *ExtAddress {
	data := hash.Sha3256(publicKeyBytes)[len(publicKeyBytes)-AddressLength:]
	addr := NewAddressWithPublicKey(publicKeyBytes)
	extHash := hash.Sha3256(append(data, []byte(nick)...))[:2]
	extAddress := addr.address + bytes.Hex(extHash)
	extAddr := &ExtAddress{
		nick:       nick,
		address:    *addr,
		extAddress: extAddress}
	return extAddr
}
