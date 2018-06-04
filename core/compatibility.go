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
	"os"
	"path/filepath"
	"strings"

	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

const (
	// MainNetID mainnet id
	MainNetID uint32 = 1

	// TestNetID testnet id
	TestNetID uint32 = 1001
)

const (
	// DefaultV8JSLibVersion default version
	DefaultV8JSLibVersion = "1.0.0"

	// CurrentV8JSLibVersion current js lib version
	CurrentV8JSLibVersion = "1.0.1"
)

// var ..
var (
	// NOTE: versions should be arranged in ascending order
	// 		map[libname][versions]
	V8JSLibs = map[string][]string{
		"execution_env.js":       {"1.0.0", "1.0.1"},
		"bignumber.js":           {"1.0.0"},
		"random.js":              {"1.0.0", "1.0.1"},
		"date.js":                {"1.0.0", "1.0.1"},
		"tsc.js":                 {"1.0.0"},
		"util.js":                {"1.0.0"},
		"esprima.js":             {"1.0.0"},
		"assert.js":              {"1.0.0"},
		"instruction_counter.js": {"1.0.0"},
		"typescriptServices.js":  {"1.0.0"},
		"blockchain.js":          {"1.0.0"},
		"console.js":             {"1.0.0"},
		"event.js":               {"1.0.0"},
		"storage.js":             {"1.0.0"},
		"crypto.js":              {"1.0.1"},
		"uint.js":                {"1.0.1"},
	}
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

	// LocalV8JSLibVersionControlHeight
	LocalV8JSLibVersionControlHeight uint64 = 2
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
	TestNetNvmMemoryLimitWithoutInjectHeight uint64 = 281600

	//TestNetWdResetRecordDependencyHeight
	TestNetWsResetRecordDependencyHeight uint64 = 281600

	// TestNetV8JSLibVersionControlHeight
	TestNetV8JSLibVersionControlHeight uint64 = 400000
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
	MainNetNvmMemoryLimitWithoutInjectHeight uint64 = 325666

	//MainNetWdResetRecordDependencyHeight
	MainNetWsResetRecordDependencyHeight uint64 = 325666

	// MainNetV8JSLibVersionControlHeight
	MainNetV8JSLibVersionControlHeight uint64 = 400000
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

	//WsResetRecordDependencyHeight if tx execute faied, worldstate reset and need to record to address dependency
	WsResetRecordDependencyHeight = TestNetWsResetRecordDependencyHeight

	// V8JSLibVersionControlHeight enable v8 js lib version control
	V8JSLibVersionControlHeight = TestNetV8JSLibVersionControlHeight
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
		V8JSLibVersionControlHeight = MainNetV8JSLibVersionControlHeight
	} else if chainID == TestNetID {

		TransferFromContractEventRecordableHeight = TestNetTransferFromContractEventRecordableHeight
		AcceptFuncAvailableHeight = TestNetAcceptFuncAvailableHeight
		RandomAvailableHeight = TestNetRandomAvailableHeight
		DateAvailableHeight = TestNetDateAvailableHeight
		RecordCallContractResultHeight = TestNetRecordCallContractResultHeight
		NvmMemoryLimitWithoutInjectHeight = TestNetNvmMemoryLimitWithoutInjectHeight
		WsResetRecordDependencyHeight = TestNetWsResetRecordDependencyHeight
		V8JSLibVersionControlHeight = TestNetV8JSLibVersionControlHeight
	} else {

		TransferFromContractEventRecordableHeight = LocalTransferFromContractEventRecordableHeight
		AcceptFuncAvailableHeight = LocalAcceptFuncAvailableHeight
		RandomAvailableHeight = LocalRandomAvailableHeight
		DateAvailableHeight = LocalDateAvailableHeight
		RecordCallContractResultHeight = LocalRecordCallContractResultHeight
		NvmMemoryLimitWithoutInjectHeight = LocalNvmMemoryLimitWithoutInjectHeight
		WsResetRecordDependencyHeight = LocalWsResetRecordDependencyHeight
		V8JSLibVersionControlHeight = LocalV8JSLibVersionControlHeight
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
		"V8JSLibVersionControlHeight":               V8JSLibVersionControlHeight,
	}).Info("Set compatibility options.")

	checkJSLib()
}

// FilterLibVersion ..
func FilterLibVersion(deployVersion, libname string) string {
	if len(deployVersion) == 0 || len(libname) == 0 {
		logging.VLog().WithFields(logrus.Fields{
			"libname":       libname,
			"deployVersion": deployVersion,
		}).Error("empty arguments.")
		return ""
	}
	if libs, ok := V8JSLibs[libname]; ok {
		for i := len(libs) - 1; i >= 0; i-- {
			// TODO: check comparison
			if strings.Compare(libs[i], deployVersion) <= 0 {
				logging.VLog().WithFields(logrus.Fields{
					"libname":       libname,
					"deployVersion": deployVersion,
					"return":        libs[i],
				}).Debug("filter js lib.")
				return libs[i]
			}
		}
	} else {
		logging.VLog().WithFields(logrus.Fields{
			"libname":       libname,
			"deployVersion": deployVersion,
		}).Debug("js lib not configured.")
	}
	return ""
}

func checkJSLib() {
	for lib, vers := range V8JSLibs {
		for _, ver := range vers {
			p := filepath.Join("lib", ver, lib)
			fi, err := os.Stat(p)
			if os.IsNotExist(err) {
				logging.VLog().WithFields(logrus.Fields{
					"path": p,
				}).Fatal("lib file not exist.")
			}
			if fi.IsDir() {
				logging.VLog().WithFields(logrus.Fields{
					"path": p,
				}).Fatal("directory already exists with the same name.")
			}

			logging.VLog().WithFields(logrus.Fields{
				"path": p,
			}).Debug("check js lib.")
		}
	}
}
