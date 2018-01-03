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
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

type account struct {

	// key address
	addr *core.Address

	// key save path
	path string
}

// refreshAccounts sync key files to memory
func (m *Manager) refreshAccounts() error {
	files, err := ioutil.ReadDir(m.keydir)
	if err != nil {
		return err
	}
	var (
		accounts []*account
		keyJSON  struct {
			Address string `json:"address"`
		}
	)
	for _, file := range files {
		path := filepath.Join(m.keydir, file.Name())
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") || strings.HasSuffix(file.Name(), "~") {
			logging.VLog().WithFields(logrus.Fields{
				"path": path,
			}).Warn("Skipped this key file.")
			continue
		}
		raw, err := ioutil.ReadFile(path)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":  err,
				"path": path,
			}).Error("Failed to parse the key file.")
			continue
		}
		keyJSON.Address = ""
		err = json.Unmarshal(raw, &keyJSON)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":  err,
				"path": path,
			}).Error("Failed to parse the key file.")
			continue
		}
		var (
			addr *core.Address
		)
		bytes, err := byteutils.FromHex(keyJSON.Address)
		if len(bytes) == core.AddressDataLength {
			if err == nil {
				addr, err = core.NewAddress(bytes)
			}
		} else {
			addr, err = core.AddressParse(keyJSON.Address)
		}
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":     err,
				"address": addr.String(),
			}).Error("Failed to parse the address.")
			continue
		}
		accounts = append(accounts, &account{addr, path})
	}
	m.accounts = accounts
	return nil
}

// loadFile import key to keystore in keydir
func (m *Manager) loadFile(addr *core.Address, passphrase []byte) error {
	acc := m.getAccount(addr)
	if acc == nil {
		return ErrAddrNotFind
	}
	raw, err := ioutil.ReadFile(acc.path)
	if err != nil {
		return err
	}
	_, err = m.Load(raw, passphrase)
	return err
}

func (m *Manager) exportFile(addr *core.Address, passphrase []byte) error {
	raw, err := m.Export(addr, passphrase)
	if err != nil {
		return err
	}
	var path string
	acc := m.getAccount(addr)
	if acc != nil {
		path = acc.path
	} else {
		path = filepath.Join(m.keydir, addr.String())
	}
	WriteFile(path, raw)
	return nil
}

func (m *Manager) getAccount(addr *core.Address) *account {
	for _, acc := range m.accounts {
		if acc.addr.Equals(addr) {
			return acc
		}
	}
	return nil
}

// WriteFile write file to path
func WriteFile(file string, content []byte) error {
	// Create the keystore directory with appropriate permissions
	const dirPerm = 0700
	if err := os.MkdirAll(filepath.Dir(file), dirPerm); err != nil {
		return err
	}
	f, err := ioutil.TempFile(filepath.Dir(file), "."+filepath.Base(file)+".tmp")
	if err != nil {
		return err
	}
	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}
	f.Close()
	return os.Rename(f.Name(), file)
}

func (m *Manager) deleteFile(addr *core.Address) error {
	acc := m.getAccount(addr)
	if acc != nil {
		os.Remove(acc.path)
		m.refreshAccounts()
	}
	return nil
}
