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

package cipher

// Cipher encrypt cipher
type Cipher struct {
	encrypt Encrypt
}

// NewCipher returns a new cipher
func NewCipher(alg uint8) *Cipher {
	c := new(Cipher)
	switch alg {
	case 1 << 4: //keysotore.SCRYPT
		c.encrypt = new(Scrypt)
	default:
		panic("cipher not support the algorithm")
	}
	return c
}

// Encrypt scrypt encrypt
func (c *Cipher) Encrypt(data []byte, passphrase []byte) ([]byte, error) {
	return c.encrypt.Encrypt(data, passphrase)
}

// EncryptKey encrypt key with address
func (c *Cipher) EncryptKey(address string, data []byte, passphrase []byte) ([]byte, error) {
	return c.encrypt.EncryptKey(address, data, passphrase)
}

// Decrypt decrypts data, returning the origin data
func (c *Cipher) Decrypt(data []byte, passphrase []byte) ([]byte, error) {
	return c.encrypt.Decrypt(data, passphrase)
}

// DecryptKey decrypts a key, returning the private key itself.
func (c *Cipher) DecryptKey(keyjson []byte, passphrase []byte) ([]byte, error) {
	return c.encrypt.DecryptKey(keyjson, passphrase)
}
