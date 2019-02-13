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

package nvm

import (
	"errors"
	"regexp"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

var (
	// StorageKeyPattern the pattern of varible key stored in stateDB
	/*
		const fieldNameRe = /^[a-zA-Z_$][a-zA-Z0-9_]+$/;
		var combineStorageMapKey = function (fieldName, key) {
			return "@" + fieldName + "[" + key + "]";
		};
	*/
	StorageKeyPattern = regexp.MustCompile("^@([a-zA-Z_$][a-zA-Z0-9_]+?)\\[(.*?)\\]$")
	// DefaultDomainKey the default domain key
	DefaultDomainKey = "_"
	// ErrInvalidStorageKey invalid storage key error
	ErrInvalidStorageKey = errors.New("invalid storage key")
)

// hashStorageKey return the key hash.
// There are two kinds of key, the one is ItemKey, the other is Map-ItemKey.
// ItemKey in SmartContract is used for object storage.
// For example, the ItemKey for the statement "token.totalSupply = 1000" is "totalSupply".
// Map-ItemKey in SmartContrat is used for Map storage.
// For example, the Map-ItemKey for the statement "token.balances.set('addr1', 100)" is "@balances[addr1]".
func parseStorageKey(key string) (string, string, error) {
	matches := StorageKeyPattern.FindAllStringSubmatch(key, -1)
	if matches == nil {
		return DefaultDomainKey, key, nil
	}

	return matches[0][1], matches[0][2], nil
}

// StorageGetFunc export StorageGetFunc
//export StorageGetFunc
func StorageGetFunc(handler uint64, k string) (string, uint64) {
	
	// calculate Gas.
	var gasCnt uint64 = 0

	_, storage := getEngineByStorageHandler(handler)
	if storage == nil {
		logging.VLog().Error("Failed to get storage handler.")
		return "", gasCnt
	}

	domainKey, itemKey, err := parseStorageKey(k)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"key":     k,
			"err":     err,
		}).Debug("Invalid storage key.")
		return "", gasCnt
	}

	val, err := storage.Get(trie.HashDomains(domainKey, itemKey))
	if err != nil {
		if err != ErrKeyNotFound {
			logging.VLog().WithFields(logrus.Fields{
				"handler": uint64(uintptr(handler)),
				"key":     k,
				"err":     err,
			}).Debug("StorageGetFunc get key failed.")
		}
		return "", gasCnt
	}

	return string(val), gasCnt
}

// StoragePutFunc export StoragePutFunc
//export StoragePutFunc
func StoragePutFunc(handler uint64, k string, value string) (int, uint64) {

	var gasCnt uint64 = 0

	_, storage := getEngineByStorageHandler(handler)
	if storage == nil {
		logging.VLog().Error("Failed to get storage handler.")
		return 1, gasCnt
	}
	v := []byte(value)

	// calculate Gas.
	gasCnt = uint64(len(k) + len(v))

	domainKey, itemKey, err := parseStorageKey(k)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"key":     k,
			"err":     err,
		}).Debug("Invalid storage key.")
		return 1, gasCnt
	}

	err = storage.Put(trie.HashDomains(domainKey, itemKey), v)
	if err != nil && err != ErrKeyNotFound {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"key":     k,
			"err":     err,
		}).Debug("StoragePutFunc put key failed.")
		return 1, gasCnt
	}

	return 0, gasCnt
}

// StorageDelFunc export StorageDelFunc
//export StorageDelFunc
func StorageDelFunc(handler uint64, k string) (int, uint64) {

	var gasCnt uint64 = 0

	_, storage := getEngineByStorageHandler(uint64(uintptr(handler)))
	if storage == nil {
		logging.VLog().Error("Failed to get storage handler.")
		return 1, gasCnt
	}

	domainKey, itemKey, err := parseStorageKey(k)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"key":     k,
			"err":     err,
		}).Debug("invalid storage key.")
		return 1, gasCnt
	}

	err = storage.Del(trie.HashDomains(domainKey, itemKey))
	if err != nil && err != ErrKeyNotFound {
		logging.VLog().WithFields(logrus.Fields{
			"handler": uint64(uintptr(handler)),
			"key":     k,
			"err":     err,
		}).Debug("StorageDelFunc del key failed.")
		return 1, gasCnt
	}

	return 0, gasCnt
}
