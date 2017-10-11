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

package keystore

// Signature interface of different signature algorithm
type Signature interface {

	// Algorithm returns the standard algorithm for this key.
	Algorithm() Algorithm

	// InitSign this object for signing. If this method is called
	// again with a different argument, it negates the effect
	// of this call.
	InitSign(privateKey PrivateKey) error

	// Sign returns the signature bytes of all the data input.
	// The format of the signature depends on the underlying
	// signature scheme.
	Sign(data []byte) (out []byte, err error)

	// RecoverPublic returns a public key, which is recoverd by data and signature
	RecoverPublic(data []byte, signature []byte) (PublicKey, error)

	// InitVerify initializes this object for verification. If this method is called
	// again with a different argument, it negates the effect
	// of this call.
	InitVerify(publicKey PublicKey) error

	// Verify the passed-in signature.
	//
	// <p>A call to this method resets this signature object to the state
	// it was in when previously initialized for verification via a
	// call to <code>initVerify(PublicKey)</code>. That is, the object is
	// reset and available to verify another signature from the identity
	// whose public key was specified in the call to <code>initVerify</code>.
	Verify(data []byte, signature []byte) (bool, error)
}
