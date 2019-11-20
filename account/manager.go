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

	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1/vrf/secp256k1VRF"

	"github.com/nebulasio/go-nebulas/util/byteutils"

	"path/filepath"

	"time"

	"sync"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/cipher"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/utils"
	nebletpb "github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// const SignatureCiphers
const (
	EccSecp256K1      = "ECC_SECP256K1"
	EccSecp256K1Value = 1

	DefaultKeyDir = "keydir"
)

var (
	// ErrAccountNotFound account is not found.
	ErrAccountNotFound = errors.New("account is not found")

	// ErrAccountIsLocked account locked.
	ErrAccountIsLocked = errors.New("account is locked")

	// ErrInvalidSignerAddress sign addr not from
	ErrInvalidSignerAddress = errors.New("transaction sign not use from address")
)

// Neblet interface breaks cycle import dependency and hides unused services.
type Neblet interface {
	Config() *nebletpb.Config
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

	mutex sync.Mutex
}

// NewManager new a account manager
func NewManager(neblet Neblet) (*Manager, error) {
	m := new(Manager)
	m.ks = keystore.DefaultKS
	m.signatureAlg = keystore.SECP256K1
	m.encryptAlg = keystore.SCRYPT
	tmpKeyDir, err := filepath.Abs(DefaultKeyDir)
	if err != nil {
		return nil, err
	}
	m.keydir = tmpKeyDir

	if neblet != nil && neblet.Config() != nil {
		conf := neblet.Config().Chain

		if len(conf.Keydir) > 0 {
			tmpKeyDir = conf.Keydir
			dir, err := filepath.Abs(tmpKeyDir)
			if err != nil {
				return nil, err
			}
			m.keydir = dir
		}

		if len(conf.SignatureCiphers) > 0 {
			if conf.SignatureCiphers[0] == EccSecp256K1 {
				m.signatureAlg = keystore.Algorithm(EccSecp256K1Value)
			}
		}
	}
	if err := m.refreshAccounts(); err != nil {
		return nil, err
	}

	return m, nil
}

// NewAccount returns a new address and keep it in keystore
func (m *Manager) NewAccount(passphrase []byte) (*core.Address, error) {
	priv, err := crypto.NewPrivateKey(m.signatureAlg, nil)
	if err != nil {
		return nil, err
	}

	addr, err := m.setKeyStore(priv, passphrase)
	if err != nil {
		return nil, err
	}

	path, err := m.exportFile(addr, passphrase, false)
	if err != nil {
		return nil, err
	}

	m.updateAccount(addr, path)

	return addr, nil
}

func (m *Manager) setKeyStore(priv keystore.PrivateKey, passphrase []byte) (*core.Address, error) {
	pub, err := priv.PublicKey().Encoded()
	if err != nil {
		return nil, err
	}
	addr, err := core.NewAddressFromPublicKey(pub)
	if err != nil {
		return nil, err
	}

	// set key to keystore
	err = m.ks.SetKey(addr.String(), priv, passphrase)
	if err != nil {
		return nil, err
	}

	return addr, nil
}

// Contains returns if contains address
func (m *Manager) Contains(addr *core.Address) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, acc := range m.accounts {
		if acc.addr.Equals(addr) {
			return true
		}
	}
	return false
}

// Unlock unlock address with passphrase
func (m *Manager) Unlock(addr *core.Address, passphrase []byte, duration time.Duration) error {
	res, err := m.ks.ContainsAlias(addr.String())
	if err != nil || res == false {
		err = m.loadFile(addr, passphrase)
		if err != nil {
			return err
		}
	}
	return m.ks.Unlock(addr.String(), passphrase, duration)
}

// Lock lock address
func (m *Manager) Lock(addr *core.Address) error {
	return m.ks.Lock(addr.String())
}

// Accounts returns slice of address
func (m *Manager) Accounts() []*core.Address {
	m.refreshAccounts()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	addrs := make([]*core.Address, len(m.accounts))
	for index, a := range m.accounts {
		addrs[index] = a.addr
	}
	return addrs
}

// Update update addr locked passphrase
func (m *Manager) Update(addr *core.Address, oldPassphrase, newPassphrase []byte) error {
	key, err := m.ks.GetKey(addr.String(), oldPassphrase)
	if err != nil {
		err = m.loadFile(addr, oldPassphrase)
		if err != nil {
			return err
		}
		key, err = m.ks.GetKey(addr.String(), oldPassphrase)
		if err != nil {
			return err
		}
	}
	defer key.Clear()

	if _, err := m.setKeyStore(key.(keystore.PrivateKey), newPassphrase); err != nil {
		return err
	}
	path, err := m.exportFile(addr, newPassphrase, true)
	if err != nil {
		return err
	}

	m.updateAccount(addr, path)

	return nil
}

// Load load a key file to keystore, unable to write file
func (m *Manager) Load(keyjson, passphrase []byte) (*core.Address, error) {
	cipher := cipher.NewCipher(uint8(m.encryptAlg))
	data, err := cipher.DecryptKey(keyjson, passphrase)
	if err != nil {
		return nil, err
	}
	return m.LoadPrivate(data, passphrase)
}

// LoadPrivate load a private key to keystore, unable to write file
func (m *Manager) LoadPrivate(privatekey, passphrase []byte) (*core.Address, error) {
	defer utils.ZeroBytes(privatekey)
	priv, err := crypto.NewPrivateKey(m.signatureAlg, privatekey)
	if err != nil {
		return nil, err
	}
	defer priv.Clear()

	addr, err := m.setKeyStore(priv, passphrase)
	if err != nil {
		return nil, err
	}

	if _, err := m.getAccount(addr); err != nil {
		m.mutex.Lock()
		acc := &account{addr: addr}
		m.accounts = append(m.accounts, acc)
		m.mutex.Unlock()
	}
	return addr, nil
}

// Import import a key file to keystore, compatible ethereum keystore file, write to file
func (m *Manager) Import(keyjson, passphrase []byte) (*core.Address, error) {
	addr, err := m.Load(keyjson, passphrase)
	if err != nil {
		return nil, err
	}
	path, err := m.exportFile(addr, passphrase, false)
	if err != nil {
		return nil, err
	}

	m.updateAccount(addr, path)

	return addr, nil
}

// Export export address to key file
func (m *Manager) Export(addr *core.Address, passphrase []byte) ([]byte, error) {
	key, err := m.ks.GetKey(addr.String(), passphrase)
	if err != nil {
		return nil, err
	}
	defer key.Clear()

	data, err := key.Encoded()
	if err != nil {
		return nil, err
	}
	defer utils.ZeroBytes(data)

	cipher := cipher.NewCipher(uint8(m.encryptAlg))
	if err != nil {
		return nil, err
	}
	out, err := cipher.EncryptKey(addr.String(), data, passphrase)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Remove remove address and encrypted private key from keystore
func (m *Manager) Remove(addr *core.Address, passphrase []byte) error {
	err := m.ks.Delete(addr.String(), passphrase)
	if err != nil {
		return err
	}

	return nil
}

// SignHash sign hash
func (m *Manager) SignHash(addr *core.Address, hash byteutils.Hash, alg keystore.Algorithm) ([]byte, error) {
	key, err := m.ks.GetUnlocked(addr.String())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":  err,
			"addr": addr,
			"hash": hash,
		}).Error("Failed to get unlocked private key.")
		return nil, ErrAccountIsLocked
	}

	signature, err := crypto.NewSignature(alg)
	if err != nil {
		return nil, err
	}

	if err := signature.InitSign(key.(keystore.PrivateKey)); err != nil {
		return nil, err
	}

	signData, err := signature.Sign(hash)
	if err != nil {
		return nil, err
	}
	return signData, nil
}

// SignTransaction sign transaction with the specified algorithm
func (m *Manager) SignTransaction(addr *core.Address, tx *core.Transaction) error {
	// check sign addr is tx's from addr
	if !tx.From().Equals(addr) {
		return ErrInvalidSignerAddress
	}
	key, err := m.ks.GetUnlocked(addr.String())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
			"tx":  tx,
		}).Error("Failed to get unlocked private key to sign transaction.")
		return ErrAccountIsLocked
	}

	signature, err := crypto.NewSignature(m.signatureAlg)
	if err != nil {
		return err
	}
	signature.InitSign(key.(keystore.PrivateKey))
	return tx.Sign(signature)
}

// SignBlock sign block with the specified algorithm
func (m *Manager) SignBlock(addr *core.Address, block *core.Block) error {
	key, err := m.ks.GetUnlocked(addr.String())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":   err,
			"block": block,
		}).Error("Failed to get unlocked private key to sign block.")
		return ErrAccountIsLocked
	}

	signature, err := crypto.NewSignature(m.signatureAlg)
	if err != nil {
		return err
	}
	signature.InitSign(key.(keystore.PrivateKey))
	return block.Sign(signature)
}

// GenerateRandomSeed generate rand
func (m *Manager) GenerateRandomSeed(addr *core.Address, ancestorHash, parentSeed []byte) (vrfSeed, vrfProof []byte, err error) {

	key, err := m.ks.GetUnlocked(addr.String())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":  err,
			"addr": addr.String(),
		}).Error("Failed to get unlocked private key to generate block rand.")
		return nil, nil, ErrAccountIsLocked
	}

	_, err = crypto.NewSignature(m.signatureAlg)
	if err != nil {
		return nil, nil, err
	}

	seckey, err := key.(keystore.PrivateKey).Encoded()
	if err != nil {
		return nil, nil, err
	}

	signer, err := secp256k1VRF.NewVRFSignerFromRawKey(seckey)
	if err != nil {
		return nil, nil, err
	}

	data := hash.Sha3256(ancestorHash, parentSeed)

	seed, proof := signer.Evaluate(data)
	if proof == nil {
		return nil, nil, secp256k1VRF.ErrEvaluateFailed
	}
	return seed[:], proof, nil
}

// SignTransactionWithPassphrase sign transaction with the from passphrase
func (m *Manager) SignTransactionWithPassphrase(addr *core.Address, tx *core.Transaction, passphrase []byte) error {
	// check sign addr is tx's from addr
	if !tx.From().Equals(addr) {
		return ErrInvalidSignerAddress
	}
	res, err := m.ks.ContainsAlias(addr.String())
	if err != nil || res == false {
		err = m.loadFile(addr, passphrase)
		if err != nil {
			return err
		}
	}

	key, err := m.ks.GetKey(addr.String(), passphrase)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
			"tx":  tx,
		}).Error("Failed to unlock private key to sign transaction")
		return ErrAccountIsLocked
	}
	defer key.Clear()

	signature, err := crypto.NewSignature(m.signatureAlg)
	if err != nil {
		return err
	}
	signature.InitSign(key.(keystore.PrivateKey))
	return tx.Sign(signature)
}
