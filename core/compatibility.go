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

	//LocalNvmMemoryLimitWithoutInjectHeight
	LocalNvmMemoryLimitWithoutInjectHeight uint64 = 2

	//LocalWdResetRecordDependencyHeight
	LocalWsResetRecordDependencyHeight uint64 = 2
)

// TestNet
const (
	// TestNetTransferFromContractEventRecordableHeight
	TestNetTransferFromContractEventRecordableHeight uint64 = 199666

	// TestNetAcceptFuncAvailableHeight
	TestNetAcceptFuncAvailableHeight uint64 = 199666

	// TestNetRandomAvailableHeight
	TestNetRandomAvailableHeight uint64 = 199666

	// TestNetDateAvailableHeight
	TestNetDateAvailableHeight uint64 = 199666

	// TestNetRecordCallContractResultHeight
	TestNetRecordCallContractResultHeight uint64 = 199666

	//TestNetNvmMemoryLimitWithoutInjectHeight
	TestNetNvmMemoryLimitWithoutInjectHeight uint64 = 281800

	//TestNetWdResetRecordDependencyHeight
	TestNetWsResetRecordDependencyHeight uint64 = 281800
)

// MainNet
const (
	// MainNetTransferFromContractEventRecordableHeight
	MainNetTransferFromContractEventRecordableHeight uint64 = 225666

	// MainNetAcceptFuncAvailableHeight
	MainNetAcceptFuncAvailableHeight uint64 = 225666

	// MainNetRandomAvailableHeight
	MainNetRandomAvailableHeight uint64 = 225666

	// MainNetDateAvailableHeight
	MainNetDateAvailableHeight uint64 = 225666

	// MainNetRecordCallContractResultHeight
	MainNetRecordCallContractResultHeight uint64 = 225666

	//MainNetNvmMemoryLimitWithoutInjectHeight
	MainNetNvmMemoryLimitWithoutInjectHeight uint64 = 306700

	//MainNetWdResetRecordDependencyHeight
	MainNetWsResetRecordDependencyHeight uint64 = 306700
)

var (
	// TransferFromContractEventRecordableHeight record event 'TransferFromContractEvent' since this height
	TransferFromContractEventRecordableHeight = TestNetTransferFromContractEventRecordableHeight

	// AcceptFuncAvailableHeight 'accept' func available since this height
	AcceptFuncAvailableHeight = TestNetAcceptFuncAvailableHeight

	// RandomAvailableHeight make 'Math.random' available in contract since this height
	RandomAvailableHeight = TestNetRandomAvailableHeight

	// DateAvailableHeight make 'Date' available in contract since this height
	DateAvailableHeight = TestNetDateAvailableHeight

	// RecordCallContractResultHeight record result of call contract to event `TopicTransactionExecutionResult` since this height
	RecordCallContractResultHeight = TestNetRecordCallContractResultHeight

	// NvmMemoryLimitWithoutInjectHeight memory of nvm contract without inject code
	NvmMemoryLimitWithoutInjectHeight = TestNetNvmMemoryLimitWithoutInjectHeight

	//WdResetRecordDependencyHeight if tx execute faied, worldstate reset and need to record to address dependency
	WsResetRecordDependencyHeight = TestNetWsResetRecordDependencyHeight
)

// SetCompatibilityOptions set compatibility height according to chain_id
func SetCompatibilityOptions(chainID uint32) {

	if chainID == MainNetID {
		TransferFromContractEventRecordableHeight = MainNetTransferFromContractEventRecordableHeight
		AcceptFuncAvailableHeight = MainNetAcceptFuncAvailableHeight
		RandomAvailableHeight = MainNetRandomAvailableHeight
		DateAvailableHeight = MainNetDateAvailableHeight
		RecordCallContractResultHeight = MainNetRecordCallContractResultHeight
		NvmMemoryLimitWithoutInjectHeight = MainNetNvmMemoryLimitWithoutInjectHeight
		WsResetRecordDependencyHeight = MainNetWsResetRecordDependencyHeight
	} else if chainID == TestNetID {

		TransferFromContractEventRecordableHeight = TestNetTransferFromContractEventRecordableHeight
		AcceptFuncAvailableHeight = TestNetAcceptFuncAvailableHeight
		RandomAvailableHeight = TestNetRandomAvailableHeight
		DateAvailableHeight = TestNetDateAvailableHeight
		RecordCallContractResultHeight = TestNetRecordCallContractResultHeight
		NvmMemoryLimitWithoutInjectHeight = TestNetNvmMemoryLimitWithoutInjectHeight
		WsResetRecordDependencyHeight = TestNetWsResetRecordDependencyHeight
	} else {

		TransferFromContractEventRecordableHeight = LocalTransferFromContractEventRecordableHeight
		AcceptFuncAvailableHeight = LocalAcceptFuncAvailableHeight
		RandomAvailableHeight = LocalRandomAvailableHeight
		DateAvailableHeight = LocalDateAvailableHeight
		RecordCallContractResultHeight = LocalRecordCallContractResultHeight
		NvmMemoryLimitWithoutInjectHeight = LocalNvmMemoryLimitWithoutInjectHeight
		WsResetRecordDependencyHeight = LocalWsResetRecordDependencyHeight
	}
	logging.VLog().WithFields(logrus.Fields{
		"chain_id": chainID,
		"TransferFromContractEventRecordableHeight": TransferFromContractEventRecordableHeight,
		"AcceptFuncAvailableHeight":                 AcceptFuncAvailableHeight,
		"RandomAvailableHeight":                     RandomAvailableHeight,
		"DateAvailableHeight":                       DateAvailableHeight,
		"RecordCallContractResultHeight":            RecordCallContractResultHeight,
		"NvmMemoryLimitWithoutInjectHeight":         NvmMemoryLimitWithoutInjectHeight,
		"WsResetRecordDependencyHeight":             WsResetRecordDependencyHeight,
	}).Info("Set compatibility options.")
}
