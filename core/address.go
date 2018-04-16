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
	"github.com/btcsuite/btcutil/base58"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

// AddressType address type
type AddressType byte

// address type enum
const (
	AccountAddress AddressType = 0x57 + iota
	ContractAddress
	UnkownAddress AddressType = 0x00
)

// const
const (
	Padding byte = 0x19

	NebulasFaith = 'n'
)

const (
	// AddressPaddingLength the length of headpadding in byte
	AddressPaddingLength = 1
	// AddressPaddingIndex the index of headpadding bytes
	AddressPaddingIndex = 0

	// AddressTypeLength the length of address type in byte
	AddressTypeLength = 1
	// AddressTypeIndex the index of address type bytes
	AddressTypeIndex = 1

	// AddressDataLength the length of data of address in byte.
	AddressDataLength = 20

	// AddressChecksumLength the checksum of address in byte.
	AddressChecksumLength = 4

	// AddressLength the length of address in byte.
	AddressLength = AddressPaddingLength + AddressTypeLength + AddressDataLength + AddressChecksumLength
	// AddressDataEnd the end of the address data
	AddressDataEnd = 22

	// AddressBase58Length length of base58(Address.address)
	AddressBase58Length = 35
	// PublicKeyDataLength length of public key
	PublicKeyDataLength = 65
)

// Address design of nebulas address
/*
[Account Address]
Similar to Bitcoin and Ethereum, Nebulas also adopts elliptic curve algorithm as its basic encryption algorithm for Nebulas accounts.
The address is derived from **public key**, which is in turn derived from the **private key** that encrypted with user's **passphrase**.
Also we have the checksum design aiming to prevent a user from sending _Nas_ to a wrong user account accidentally due to entry of several incorrect characters.

The specific calculation formula is as follows:

	Content = ripemd160( sha3_256( Public Key ) )
	CheckSum = sha3_256( 0x19 + 0x57 + Content )[0:4]
	Address = base58( 0x19 + 0x57 + Content + CheckSum )

 	0x57 is a one-byte "type code" for account address, 0x19 is a one-byte fixed "padding"

The ripemd160 digest of SHA3-256 digest of a public key serve as the major component of an address,
for which another SHA3-256 digest should be conducted and the first 4 bytes should be used as a checksum. For example:
The final address of Nebulas Wallet should be:  n1TV3sU6jyzR4rJ1D7jCAmtVGSntJagXZHC

[Smart Contract Address]
Calculating contract address differs slightly from account, passphrase of contract sender is not required but address & nonce.
For more information, plz check (https://github.com/nebulasio/wiki/blob/master/tutorials/%5BEnglish%5D%20Nebulas%20101%20-%2003%20Smart%20Contracts%20JavaScript.md) and [rpc.sendTransaction](https://github.com/nebulasio/wiki/blob/master/rpc.md#sendtransaction).
Calculation formula is as follows:

	Content = ripemd160( sha3_256( tx.from, tx.nonce ) )
	CheckSum = sha3_256( 0x19 + 0x58 + Content )[0:4]
	Address = base58( 0x19 + 0x58 + Content + CheckSum )

	0x58 is a one-byte "type code" for smart contract address, 0x19 is a one-byte fixed "padding"


[TODO]
In addition to standard address with 50 characters, we also support extended address in order to ensure the security of transfers conducted by users.
The traditional bank transfer design is used for reference:
In the process of a bank transfer, bank card number of the remittee should be verified, in addition to which the remitter must enter the name of the remittee.
The transfer can be correctly processed only when the bank card number and the name match each other.
The generating algorithm for extended address is described as follows:

  ExtData = Utf8Bytes({Nickname or any string})
  ExtHash = sha3_256(Data + ExtData)[0:2]
  ExtAddress = Account Address + Hex(ExtHash)

An extended address is generated through addition of 2-byte extended verification to the end of a standard address and contains a total of 54 characters.
Addition of extended information allows the addition of another element verification to the Nebulas Wallet APP. For example:

	The standard address of Alice’s wallet is  0xdf4d22611412132d3e9bd322f82e2940674ec1bc03b20e40, and the extended address after addition of the nickname "alice" should be 0xdf4d22611412132d3e9bd322f82e2940674ec1bc03b20e40e345.
	Alice tells Bob the extended address 0xdf4d22611412132d3e9bd322f82e2940674ec1bc03b20e40e345 and her nickname alice.
	Bob enters 0xdf4d22611412132d3e9bd322f82e2940674ec1bc03b20e40e345 and alice in the Wallet App.
	The Wallet App verifies the consistency between the wallet address and the nickname in order to avoid the circumstance that Bob enters the account number of another user by mistake.
*/
type Address struct {
	address byteutils.Hash
}

// ContractTxFrom tx from
type ContractTxFrom []byte

// ContractTxNonce tx nonce
type ContractTxNonce []byte

// Bytes returns address bytes
func (a *Address) Bytes() []byte {
	return a.address
}

// String returns address string
func (a *Address) String() string {
	return base58.Encode(a.address)
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

// Type return the type of address.
func (a *Address) Type() AddressType {
	if len(a.address) <= AddressTypeIndex {
		return UnkownAddress
	}
	return AddressType(a.address[AddressTypeIndex])
}

// NewAddress create new #Address according to data bytes.
func newAddress(t AddressType, args ...[]byte) (*Address, error) {
	if len(args) == 0 {
		return nil, ErrInvalidArgument
	}

	switch t {
	case AccountAddress, ContractAddress:
	default:
		return nil, ErrInvalidArgument
	}

	buffer := make([]byte, AddressLength)
	buffer[AddressPaddingIndex] = Padding
	buffer[AddressTypeIndex] = byte(t)

	sha := hash.Sha3256(args...)
	content := hash.Ripemd160(sha)
	copy(buffer[AddressTypeIndex+1:AddressDataEnd], content)

	cs := checkSum(buffer[:AddressDataEnd])
	copy(buffer[AddressDataEnd:], cs)

	return &Address{address: buffer}, nil
}

// NewAddressFromPublicKey return new address from publickey bytes
func NewAddressFromPublicKey(s []byte) (*Address, error) {
	if len(s) != PublicKeyDataLength {
		return nil, ErrInvalidArgument
	}
	return newAddress(AccountAddress, s)
}

// NewContractAddressFromData return new contract address from bytes.
func NewContractAddressFromData(from ContractTxFrom, nonce ContractTxNonce) (*Address, error) {
	if len(from) == 0 || len(nonce) == 0 {
		return nil, ErrInvalidArgument
	}
	return newAddress(ContractAddress, from, nonce)
}

// AddressParse parse address string.
func AddressParse(s string) (*Address, error) {
	if len(s) != AddressBase58Length || s[0] != NebulasFaith {
		return nil, ErrInvalidAddressFormat
	}

	return AddressParseFromBytes(base58.Decode(s))
}

// AddressParseFromBytes parse address from bytes.
func AddressParseFromBytes(b []byte) (*Address, error) {
	if len(b) != AddressLength || b[AddressPaddingIndex] != Padding {
		return nil, ErrInvalidAddressFormat
	}

	switch AddressType(b[AddressTypeIndex]) {
	case AccountAddress, ContractAddress:
	default:
		return nil, ErrInvalidAddressType
	}

	if !byteutils.Equal(checkSum(b[:AddressDataEnd]), b[AddressDataEnd:]) {
		return nil, ErrInvalidAddressChecksum
	}

	return &Address{address: b}, nil
}

func checkSum(data []byte) []byte {
	return hash.Sha3256(data)[:AddressChecksumLength]
}
