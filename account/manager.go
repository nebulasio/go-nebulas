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

package account

import (
	"errors"

	"path/filepath"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/cipher"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrAddrNotFind address not find.
	ErrAddrNotFind = errors.New("address not find")

	// ErrTxAddressLocked from address locked.
	ErrTxAddressLocked = errors.New("transaction from address locked")

	// ErrTxSignFrom sign addr not from
	ErrTxSignFrom = errors.New("transaction sign not use from addr")
)

// Neblet interface breaks cycle import dependency and hides unused services.
type Neblet interface {
	Config() nebletpb.Config
}

// Manager accounts manager ,handle account generate and storage
type Manager struct {

	// keystore
	ks *keystore.Keystore

	// key save path
	keydir string

	// key encrypt alg
	encryptAlg keystore.Algorithm

	// key signature alg
	signatureAlg keystore.Algorithm

	// account slice
	accounts []*account
}

// NewManager new a account manager
func NewManager(neblet Neblet) *Manager {
	m := new(Manager)
	m.ks = keystore.DefaultKS
	m.signatureAlg = keystore.SECP256K1
	m.encryptAlg = keystore.SCRYPT
	m.keydir, _ = filepath.Abs("keydir")

	if neblet != nil {
		// conf := neblet.Config().Account
		conf := neblet.Config().Chain

		keydir := conf.Keydir
		if filepath.IsAbs(keydir) {
			m.keydir = keydir
		} else {
			m.keydir, _ = filepath.Abs(keydir)
		}

		if conf.SignatureCiphers[0] > 0 {
			m.signatureAlg = keystore.Algorithm(conf.SignatureCiphers[0])
		}

		// if conf.GetSignature() > 0 {
		// 	m.signatureAlg = keystore.Algorithm(conf.GetSignature())
		// }
		// if conf.GetEncrypt() > 0 {
		// 	m.encryptAlg = keystore.Algorithm(conf.GetEncrypt())
		// }

		m.refreshAccounts()
	}
	return m
}

// NewAccount returns a new address and keep it in keystore
func (m *Manager) NewAccount(passphrase []byte) (*core.Address, error) {
	priv, err := crypto.NewPrivateKey(m.signatureAlg, nil)
	if err != nil {
		return nil, err
	}
	return m.storeAddress(priv, passphrase, true)
}

func (m *Manager) storeAddress(priv keystore.PrivateKey, passphrase []byte, writeFile bool) (*core.Address, error) {
	pub, err := priv.PublicKey().Encoded()
	if err != nil {
		return nil, err
	}
	addr, err := core.NewAddressFromPublicKey(pub)
	if err != nil {
		return nil, err
	}
	// set key to keystore
	err = m.ks.SetKey(addr.ToHex(), priv, passphrase)
	if err != nil {
		return nil, err
	}
	if writeFile {
		// export key to file in keydir
		err = m.exportFile(addr, passphrase)
		if err != nil {
			return nil, err
		}
	}
	return addr, nil
}

// Unlock unlock address with passphrase
func (m *Manager) Unlock(addr *core.Address, passphrase []byte) error {
	res, err := m.ks.ContainsAlias(addr.ToHex())
	if err != nil || res == false {
		err = m.importFile(addr, passphrase)
		if err != nil {
			return err
		}
	}
	return m.ks.Unlock(addr.ToHex(), passphrase, keystore.DefaultUnlockDuration)
}

// Lock lock address
func (m *Manager) Lock(addr *core.Address) error {
	return m.ks.Lock(addr.ToHex())
}

// Accounts returns slice of address
func (m *Manager) Accounts() []*core.Address {
	m.refreshAccounts()
	addrs := make([]*core.Address, len(m.accounts))
	for index, a := range m.accounts {
		addrs[index] = a.addr
	}
	return addrs
}

// Update update addr locked passphrase
func (m *Manager) Update(addr *core.Address, oldPassphrase, newPassphrase []byte) error {
	key, err := m.ks.GetKey(addr.ToHex(), oldPassphrase)
	if err != nil {
		err = m.importFile(addr, oldPassphrase)
		if err != nil {
			return err
		}
	}
	_, err = m.storeAddress(key.(keystore.PrivateKey), newPassphrase, true)
	return err
}

// Import import a key file to keystore, compatible ethereum keystore file
func (m *Manager) Import(keyjson, passphrase []byte) (*core.Address, error) {
	cipher := cipher.NewCipher(uint8(m.encryptAlg))
	data, err := cipher.DecryptKey(keyjson, passphrase)
	if err != nil {
		return nil, err
	}
	priv, err := crypto.NewPrivateKey(m.signatureAlg, data)
	if err != nil {
		return nil, err
	}
	return m.storeAddress(priv, passphrase, false)
}

// Export export address to key file
func (m *Manager) Export(addr *core.Address, passphrase []byte) ([]byte, error) {
	key, err := m.ks.GetKey(addr.ToHex(), passphrase)
	if err != nil {
		return nil, err
	}
	data, err := key.Encoded()
	if err != nil {
		return nil, err
	}
	cipher := cipher.NewCipher(uint8(m.encryptAlg))
	if err != nil {
		return nil, err
	}
	out, err := cipher.EncryptKey(addr.ToHex(), data, passphrase)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Delete delete address
func (m *Manager) Delete(a string, passphrase []byte) error {
	addr, err := core.AddressParse(a)
	if err != nil {
		return err
	}
	err = m.ks.Delete(a, passphrase)
	if err != nil {
		return err
	}
	//remove key file and accounts
	return m.deleteFile(addr)
}

// SignTransaction sign transaction with the specified algorithm
func (m *Manager) SignTransaction(addr *core.Address, tx *core.Transaction) error {
	// check sign addr is tx's from addr
	if !tx.From().Equals(addr) {
		return ErrTxSignFrom
	}
	key, err := m.ks.GetUnlocked(addr.ToHex())
	if err != nil {
		log.WithFields(log.Fields{
			"func": "SignTransaction",
			"err":  ErrTxAddressLocked,
			"tx":   tx,
		}).Error("transaction address locked")
		return err
	}

	signature, err := crypto.NewSignature(m.signatureAlg)
	if err != nil {
		return err
	}
	signature.InitSign(key.(keystore.PrivateKey))
	return tx.Sign(signature)
}

// SignTransactionWithPassphrase sign transaction with the from passphrase
func (m *Manager) SignTransactionWithPassphrase(addr *core.Address, tx *core.Transaction, passphrase []byte) error {
	// check sign addr is tx's from addr
	if !tx.From().Equals(addr) {
		return ErrTxSignFrom
	}
	key, err := m.ks.GetKey(addr.ToHex(), passphrase)
	if err != nil {
		log.WithFields(log.Fields{
			"func": "SignTransactionWithPassphrase",
			"err":  ErrTxAddressLocked,
			"tx":   tx,
		}).Error("transaction address get failed")
		return err
	}

	signature, err := crypto.NewSignature(m.signatureAlg)
	if err != nil {
		return err
	}
	signature.InitSign(key.(keystore.PrivateKey))
	return tx.Sign(signature)
}
