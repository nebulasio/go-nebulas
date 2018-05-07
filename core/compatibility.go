// Copyright (C) 2018 go-nebulas authors
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
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

const (
	// MainNetID mainnet id
	MainNetID uint32 = 1

	// TestNetID testnet id
	TestNetID uint32 = 1001

	//SignNetID signnet id
	SignNetID uint32 = 0
)

// mainnet/testnet
const (
	// DefaultTransferFromContractEventRecordableHeight
	DefaultTransferFromContractEventRecordableHeight uint64 = 199666

	// DefaultAcceptFuncAvailableHeight
	DefaultAcceptFuncAvailableHeight uint64 = 199666

	// DefaultRandomAvailableHeight
	DefaultRandomAvailableHeight uint64 = 199666

	// DefaultDateAvailableHeight
	DefaultDateAvailableHeight uint64 = 199666

	// DefaultRecordCallContractResultHeight
	DefaultRecordCallContractResultHeight uint64 = 199666
)

// others, e.g. local/develop
const (
	// LocalTransferFromContractEventRecordableHeight
	LocalTransferFromContractEventRecordableHeight uint64 = 2

	// LocalAcceptFuncAvailableHeight
	LocalAcceptFuncAvailableHeight uint64 = 2

	// LocalRandomAvailableHeight
	LocalRandomAvailableHeight uint64 = 2

	// LocalDateAvailableHeight
	LocalDateAvailableHeight uint64 = 2

	// LocalRecordCallContractResultHeight
	LocalRecordCallContractResultHeight uint64 = 2
)

var (
	// TransferFromContractEventRecordableHeight record event 'TransferFromContractEvent' since this height
	TransferFromContractEventRecordableHeight = DefaultTransferFromContractEventRecordableHeight

	// AcceptFuncAvailableHeight 'accept' func available since this height
	AcceptFuncAvailableHeight = DefaultAcceptFuncAvailableHeight

	// RandomAvailableHeight make 'Math.random' available in contract since this height
	RandomAvailableHeight = DefaultRandomAvailableHeight

	// DateAvailableHeight make 'Date' available in contract since this height
	DateAvailableHeight = DefaultDateAvailableHeight

	// RecordCallContractResultHeight record result of call contract to event `TopicTransactionExecutionResult` since this height
	RecordCallContractResultHeight = DefaultRecordCallContractResultHeight
)

// SetCompatibilityOptions set compatibility height according to chain_id
func SetCompatibilityOptions(chainID uint32) {

	if chainID == MainNetID || chainID == TestNetID || chainID == SignNetID {
		logging.VLog().WithFields(logrus.Fields{
			"chain_id": chainID,
			"TransferFromContractEventRecordableHeight": TransferFromContractEventRecordableHeight,
			"AcceptFuncAvailableHeight":                 AcceptFuncAvailableHeight,
			"RandomAvailableHeight":                     RandomAvailableHeight,
			"DateAvailableHeight":                       DateAvailableHeight,
			"RecordCallContractResultHeight":            RecordCallContractResultHeight,
		}).Info("Set compatibility options for mainnet/testnet.")
		return
	}

	TransferFromContractEventRecordableHeight = LocalTransferFromContractEventRecordableHeight

	AcceptFuncAvailableHeight = LocalAcceptFuncAvailableHeight

	RandomAvailableHeight = LocalRandomAvailableHeight

	DateAvailableHeight = LocalDateAvailableHeight

	RecordCallContractResultHeight = LocalRecordCallContractResultHeight

	logging.VLog().WithFields(logrus.Fields{
		"chain_id": chainID,
		"TransferFromContractEventRecordableHeight": TransferFromContractEventRecordableHeight,
		"AcceptFuncAvailableHeight":                 AcceptFuncAvailableHeight,
		"RandomAvailableHeight":                     RandomAvailableHeight,
		"DateAvailableHeight":                       DateAvailableHeight,
		"RecordCallContractResultHeight":            RecordCallContractResultHeight,
	}).Info("Set compatibility options for local.")

}
