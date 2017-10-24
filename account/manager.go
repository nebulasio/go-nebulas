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

	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/cipher"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	log "github.com/sirupsen/logrus"
)

var (
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
	ks *keystore.Keystore

	encryptAlg keystore.Algorithm

	signatureAlg keystore.Algorithm
}

// NewManager new a account manager
func NewManager(neblet Neblet) *Manager {
	m := new(Manager)
	m.ks = keystore.DefaultKS
	m.signatureAlg = keystore.SECP256K1
	m.encryptAlg = keystore.SCRYPT

	if neblet != nil {
		conf := neblet.Config().Account
		if conf.GetSignature() > 0 {
			m.signatureAlg = keystore.Algorithm(conf.GetSignature())
		}
		if conf.GetEncrypt() > 0 {
			m.encryptAlg = keystore.Algorithm(conf.GetEncrypt())
		}

		// TODO(larry.wang): test keys load from keyDir, latter remove
		if len(conf.GetTestPassphrase()) > 0 && len(conf.GetKeyDir()) > 0 {
			m.loadTestKey(conf.GetKeyDir(), []byte(conf.GetTestPassphrase()))
		}
	}
	return m
}

// load test key files
func (m *Manager) loadTestKey(keyDir string, passphrase []byte) {
	keyDir, _ = filepath.Abs(keyDir)
	log.Debug("load test keys form:", keyDir)
	files, _ := ioutil.ReadDir(keyDir)
	for _, fi := range files {
		path := filepath.Join(keyDir, fi.Name())

		raw, err := ioutil.ReadFile(path)
		if err != nil {
			log.Error("read key file failed", err)
			continue
		}
		addr, err := m.Import(raw, passphrase)
		if err != nil {
			log.Error("load key failed", err)
		}
		// TODO(larry.wang):test key file auto unlock 1 year
		m.ks.Unlock(addr.ToHex(), passphrase, time.Second*60*60*24*365)
		log.Debug("load test addr:", addr.ToHex())
	}
}

// NewAccount returns a new address and keep it in keystore
func (m *Manager) NewAccount(passphrase []byte) (*core.Address, error) {

	priv, err := crypto.NewPrivateKey(m.signatureAlg, nil)
	if err != nil {
		return nil, err
	}
	return m.storeAddress(priv, passphrase)
}

func (m *Manager) storeAddress(priv keystore.PrivateKey, passphrase []byte) (*core.Address, error) {
	pub, err := priv.PublicKey().Encoded()
	if err != nil {
		return nil, err
	}
	addr, err := core.NewAddressFromPublicKey(pub)
	if err != nil {
		return nil, err
	}

	err = m.ks.SetKey(addr.ToHex(), priv, passphrase)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

// Unlock unlock address with passphrase
func (m *Manager) Unlock(addr *core.Address, passphrase []byte) error {
	return m.ks.Unlock(addr.ToHex(), passphrase, keystore.DefaultUnlockDuration)
}

// Lock lock address
func (m *Manager) Lock(addr *core.Address) error {
	return m.ks.Lock(addr.ToHex())
}

// Accounts returns slice of address
func (m *Manager) Accounts() []*core.Address {
	aliases := m.ks.Aliases()
	addres := make([]*core.Address, len(aliases))
	for index, a := range aliases {
		addr, err := core.AddressParse(a)
		if err == nil {
			// currently keystore only storage address as alias
			addres[index] = addr
		}
	}
	return addres
}

// Update update addr locked passphrase
func (m *Manager) Update(addr *core.Address, oldPassphrase, newPassphrase []byte) error {
	return m.ks.Update(addr.ToHex(), oldPassphrase, newPassphrase)
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
	return m.storeAddress(priv, passphrase)
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

// SignTransaction sign transaction with the specified algorithm
func (m *Manager) SignTransaction(addr *core.Address, tx *core.Transaction) error {
	// check sign addr is tx's from addr
	if !byteutils.Equal(tx.From(), addr.Bytes()) {
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
