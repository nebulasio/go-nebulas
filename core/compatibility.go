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
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

/**********     js lib relative  BEGIN   **********/
const (
	// DefaultV8JSLibVersion default version
	DefaultV8JSLibVersion = "1.0.0"
)

type version struct {
	major, minor, patch int
}

type heightOfVersionSlice []*struct {
	version string
	height  uint64
}

func (h heightOfVersionSlice) String() string {
	var buf bytes.Buffer
	buf.WriteString("{")
	for _, v := range h {
		if buf.Len() > 1 {
			buf.WriteString(",")
		}
		buf.WriteString(v.version + "=" + strconv.FormatUint(v.height, 10))
	}
	buf.WriteString("}")
	return buf.String()
}
func (h heightOfVersionSlice) Len() int {
	return len(h)
}
func (h heightOfVersionSlice) Less(i, j int) bool {
	return h[i].height < h[j].height
}
func (h heightOfVersionSlice) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// var ..
var (
	// NOTE: versions should be arranged in ascending order
	// 		map[libname][versions]
	V8JSLibs = map[string][]string{
		"execution_env.js":       {"1.0.0", "1.0.5"},
		"bignumber.js":           {"1.0.0"},
		"random.js":              {"1.0.0", "1.0.5"},
		"date.js":                {"1.0.0", "1.0.5"},
		"tsc.js":                 {"1.0.0"},
		"util.js":                {"1.0.0"},
		"esprima.js":             {"1.0.0"},
		"assert.js":              {"1.0.0"},
		"instruction_counter.js": {"1.0.0"},
		"typescriptServices.js":  {"1.0.0"},
		"blockchain.js":          {"1.0.0", "1.0.5"},
		"console.js":             {"1.0.0"},
		"event.js":               {"1.0.0"},
		"storage.js":             {"1.0.0"},
		"crypto.js":              {"1.0.5"},
		"uint.js":                {"1.0.5"},
	}

	digitalized = make(map[string][]*version)
)

var (
	// ErrInvalidJSLibVersion ..
	ErrInvalidJSLibVersion = errors.New("invalid js lib version")
)

/**********     js lib relative  END   **********/

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

	//LocalNetTransferFromContractFailureEventRecordableHeight
	LocalTransferFromContractFailureEventRecordableHeight uint64 = 2

	//LocalNetNewNvmExeTimeoutConsumeGasHeight
	LocalNewNvmExeTimeoutConsumeGasHeight uint64 = 2

	//LocalNvmGasLimitWithoutTimeoutAtHeight
	LocalNvmGasLimitWithoutTimeoutAtHeight uint64 = 2
)

// var for local/develop
var (
	LocalV8JSLibVersionHeightSlice = heightOfVersionSlice{
		{"1.0.5", LocalV8JSLibVersionControlHeight},
	}
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

	// TestNetV8JSLibVersionControlHeight
	TestNetV8JSLibVersionControlHeight uint64 = 424400

	//TestNetTransferFromContractFailureEventRecordableHeight
	TestNetTransferFromContractFailureEventRecordableHeight uint64 = 424400

	//TestNetNewNvmExeTimeoutConsumeGasHeight
	TestNetNewNvmExeTimeoutConsumeGasHeight uint64 = 424400

	//TestNetNvmGasLimitWithoutTimeoutAtHeight
	TestNetNvmGasLimitWithoutTimeoutAtHeight uint64 = 600000
)

// var for TestNet
var (
	TestNetV8JSLibVersionHeightSlice = heightOfVersionSlice{
		{"1.0.5", TestNetV8JSLibVersionControlHeight},
	}
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
	MainNetNvmMemoryLimitWithoutInjectHeight uint64 = 306800

	//MainNetWdResetRecordDependencyHeight
	MainNetWsResetRecordDependencyHeight uint64 = 306800

	// MainNetV8JSLibVersionControlHeight
	MainNetV8JSLibVersionControlHeight uint64 = 467500

	//MainNetTransferFromContractFailureEventRecordableHeight
	MainNetTransferFromContractFailureEventRecordableHeight uint64 = 467500

	//MainNetNewNvmExeTimeoutConsumeGasHeight
	MainNetNewNvmExeTimeoutConsumeGasHeight uint64 = 467500

	//MainNetNvmGasLimitWithoutTimeoutAtHeight
	MainNetNvmGasLimitWithoutTimeoutAtHeight uint64 = 624763
)

// var for MainNet
var (
	MainNetV8JSLibVersionHeightSlice = heightOfVersionSlice{
		{"1.0.5", MainNetV8JSLibVersionControlHeight},
	}
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

	// V8JSLibVersionHeightSlice all version-height pairs
	V8JSLibVersionHeightSlice = TestNetV8JSLibVersionHeightSlice

	// TransferFromContractFailureEventRecordableHeight record event 'TransferFromContractEvent' since this height
	TransferFromContractFailureEventRecordableHeight = TestNetTransferFromContractFailureEventRecordableHeight

	//NewNvmExeTimeoutConsumeGasHeight
	NewNvmExeTimeoutConsumeGasHeight = TestNetNewNvmExeTimeoutConsumeGasHeight

	//NvmGasLimitWithoutTimeoutAtHeight
	NvmGasLimitWithoutTimeoutAtHeight = TestNetNvmGasLimitWithoutTimeoutAtHeight
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
		V8JSLibVersionHeightSlice = MainNetV8JSLibVersionHeightSlice
		TransferFromContractFailureEventRecordableHeight = MainNetTransferFromContractFailureEventRecordableHeight
		NewNvmExeTimeoutConsumeGasHeight = MainNetNewNvmExeTimeoutConsumeGasHeight
		NvmGasLimitWithoutTimeoutAtHeight = MainNetNvmGasLimitWithoutTimeoutAtHeight
	} else if chainID == TestNetID {

		TransferFromContractEventRecordableHeight = TestNetTransferFromContractEventRecordableHeight
		AcceptFuncAvailableHeight = TestNetAcceptFuncAvailableHeight
		RandomAvailableHeight = TestNetRandomAvailableHeight
		DateAvailableHeight = TestNetDateAvailableHeight
		RecordCallContractResultHeight = TestNetRecordCallContractResultHeight
		NvmMemoryLimitWithoutInjectHeight = TestNetNvmMemoryLimitWithoutInjectHeight
		WsResetRecordDependencyHeight = TestNetWsResetRecordDependencyHeight
		V8JSLibVersionControlHeight = TestNetV8JSLibVersionControlHeight
		V8JSLibVersionHeightSlice = TestNetV8JSLibVersionHeightSlice
		TransferFromContractFailureEventRecordableHeight = TestNetTransferFromContractFailureEventRecordableHeight
		NewNvmExeTimeoutConsumeGasHeight = TestNetNewNvmExeTimeoutConsumeGasHeight
		NvmGasLimitWithoutTimeoutAtHeight = TestNetNvmGasLimitWithoutTimeoutAtHeight
	} else {

		TransferFromContractEventRecordableHeight = LocalTransferFromContractEventRecordableHeight
		AcceptFuncAvailableHeight = LocalAcceptFuncAvailableHeight
		RandomAvailableHeight = LocalRandomAvailableHeight
		DateAvailableHeight = LocalDateAvailableHeight
		RecordCallContractResultHeight = LocalRecordCallContractResultHeight
		NvmMemoryLimitWithoutInjectHeight = LocalNvmMemoryLimitWithoutInjectHeight
		WsResetRecordDependencyHeight = LocalWsResetRecordDependencyHeight
		V8JSLibVersionControlHeight = LocalV8JSLibVersionControlHeight
		V8JSLibVersionHeightSlice = LocalV8JSLibVersionHeightSlice
		TransferFromContractFailureEventRecordableHeight = LocalTransferFromContractFailureEventRecordableHeight
		NewNvmExeTimeoutConsumeGasHeight = LocalNewNvmExeTimeoutConsumeGasHeight
		NvmGasLimitWithoutTimeoutAtHeight = LocalNvmGasLimitWithoutTimeoutAtHeight
	}

	// sort V8JSLibVersionHeightSlice in descending order by height
	sort.Sort(sort.Reverse(V8JSLibVersionHeightSlice))

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
		"V8JSLibVersionHeightSlice":                 V8JSLibVersionHeightSlice,
		"TransferFromContractFailureHeight":         TransferFromContractFailureEventRecordableHeight,
		"NewNvmExeTimeoutConsumeGasHeight":          NewNvmExeTimeoutConsumeGasHeight,
		"NvmGasLimitWithoutTimeoutAtHeight":         NvmGasLimitWithoutTimeoutAtHeight,
	}).Info("Set compatibility options.")

	checkJSLib()
}

// FindLastNearestLibVersion ..
func FindLastNearestLibVersion(deployVersion, libname string) string {
	if len(deployVersion) == 0 || len(libname) == 0 {
		logging.VLog().WithFields(logrus.Fields{
			"libname":       libname,
			"deployVersion": deployVersion,
		}).Error("empty arguments.")
		return ""
	}

	if libs, ok := digitalized[libname]; ok {
		v, err := parseVersion(deployVersion)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":           err,
				"deployVersion": deployVersion,
				"lib":           libname,
			}).Debug("parse deploy version error.")
			return ""
		}
		for i := len(libs) - 1; i >= 0; i-- {
			if compareVersion(libs[i], v) <= 0 {
				logging.VLog().WithFields(logrus.Fields{
					"libname":       libname,
					"deployVersion": deployVersion,
					"return":        libs[i],
				}).Debug("filter js lib.")
				return V8JSLibs[libname][i]
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

func compareVersion(a, b *version) int {
	if a.major > b.major {
		return 1
	}
	if a.major < b.major {
		return -1
	}

	if a.minor > b.minor {
		return 1
	}
	if a.minor < b.minor {
		return -1
	}

	if a.patch > b.patch {
		return 1
	}
	if a.patch < b.patch {
		return -1
	}
	return 0
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

func parseVersion(ver string) (*version, error) {
	ss := strings.Split(ver, ".")
	if len(ss) != 3 {
		return nil, ErrInvalidJSLibVersion
	}

	major, err := strconv.Atoi(ss[0])
	if err != nil {
		return nil, err
	}

	minor, err := strconv.Atoi(ss[1])
	if err != nil {
		return nil, err
	}

	patch, err := strconv.Atoi(ss[2])
	if err != nil {
		return nil, err
	}
	return &version{major, minor, patch}, nil
}

func (v *version) String() string {
	return strings.Join([]string{
		strconv.Itoa(v.major),
		strconv.Itoa(v.minor),
		strconv.Itoa(v.patch),
	}, ".")
}

// convert V8JSLibs from string to type `version`
func init() {
	for lib, vers := range V8JSLibs {
		for _, ver := range vers {
			v, err := parseVersion(ver)
			if err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"err":     err,
					"lib":     lib,
					"version": ver,
				}).Fatal("parse js lib version error.")
			}

			if _, ok := digitalized[lib]; !ok {
				digitalized[lib] = make([]*version, 0)
			}
			digitalized[lib] = append(digitalized[lib], v)
		}
	}
}

// GetMaxV8JSLibVersionAtHeight ..
func GetMaxV8JSLibVersionAtHeight(blockHeight uint64) string {
	// V8JSLibVersionHeightSlice is already sorted at SetCompatibilityOptions func
	for _, v := range V8JSLibVersionHeightSlice {
		if blockHeight >= v.height {
			return v.version
		}
	}
	return ""
}
