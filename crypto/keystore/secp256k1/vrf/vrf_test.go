// Copyright (C) 2018 go-nebulas

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

package vrf

import (
	"fmt"
	"testing"

	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

const (
	// private key in hex
	privKey = `b02b430d4a9d7120b65038452a6da3f3c716829e5be3665adf934d4798d96ed7`
	// public key in hex
	pubKey = `04e4d0dde330c0b8d8d8b1b2071aa75c3e94f200a3d11ca1d908644eee50c8833a816dc0b2d003fc66187ef6750a56e1b3004d32e6159008400ab92f2ded7b4544`
)

func TestVRF(t *testing.T) {
	priv, _ := crypto.NewPrivateKey(keystore.SECP256K1, nil)
	seckey, err := priv.Encoded()
	if err != nil {
		t.Errorf("new priv err: %v", err)
	}
	seckeypub, err := priv.PublicKey().Encoded()
	if err != nil {
		t.Errorf("pub of new priv err: %v", err)
	}
	fmt.Println("1:", byteutils.Hex(seckey))
	fmt.Println("2:", byteutils.Hex(seckeypub))

}
