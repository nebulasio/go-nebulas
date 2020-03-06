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

// var ..
var (
	// NOTE: versions should be arranged in ascending order
	// 		map[libname][versions]
	V8JSLibs = map[string][]string{
		"execution_env.js":       {"1.0.0", "1.0.5"},
		"bignumber.js":           {"1.0.0"},
		"random.js":              {"1.0.0", "1.0.5", "1.1.0"},
		"date.js":                {"1.0.0", "1.0.5"},
		"tsc.js":                 {"1.0.0"},
		"util.js":                {"1.0.0"},
		"esprima.js":             {"1.0.0"},
		"assert.js":              {"1.0.0"},
		"instruction_counter.js": {"1.0.0", "1.1.0"},
		"typescriptServices.js":  {"1.0.0"},
		"blockchain.js":          {"1.0.0", "1.0.5", "1.1.0"},
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

// V8JSLibVersionHeightMap key is version in string format, value is height
type V8JSLibVersionHeightMap struct {
	Data     map[string]uint64
	DescKeys []string
}

// GetHeightOfVersion ..
func (v *V8JSLibVersionHeightMap) GetHeightOfVersion(version string) uint64 {
	if r, ok := v.Data[version]; ok {
		return r
	}
	return 0
}

func (v *V8JSLibVersionHeightMap) String() string {
	var buf bytes.Buffer
	buf.WriteString("{")
	for _, ver := range v.DescKeys {
		if buf.Len() > 1 {
			buf.WriteString(",")
		}
		buf.WriteString(ver + "=" + strconv.FormatUint(v.Data[ver], 10))
	}
	buf.WriteString("}")
	return buf.String()
}

func (v *V8JSLibVersionHeightMap) validate() {
	var lastVersion *version
	for _, key := range v.DescKeys {
		cur, err := parseVersion(key)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"version": key,
				"err":     err,
			}).Fatal("parse version error.")
		}

		if lastVersion != nil {
			if compareVersion(cur, lastVersion) >= 0 || v.Data[key] >= v.Data[lastVersion.String()] {
				logging.VLog().WithFields(logrus.Fields{
					"version": key,
					"height":  v.Data[key],
				}).Fatal("non descending order version map.")
			}
		}

		lastVersion = cur
	}
}

// Compatibility ..
type Compatibility interface {
	TransferFromContractEventRecordableHeight() uint64
	AcceptFuncAvailableHeight() uint64
	RandomAvailableHeight() uint64
	DateAvailableHeight() uint64
	RecordCallContractResultHeight() uint64
	NvmMemoryLimitWithoutInjectHeight() uint64
	WsResetRecordDependencyHeight() uint64 //reserve address of to
	V8JSLibVersionControlHeight() uint64
	TransferFromContractFailureEventRecordableHeight() uint64
	NewNvmExeTimeoutConsumeGasHeight() uint64
	NvmExeTimeoutHeight() []uint64
	V8JSLibVersionHeightMap() *V8JSLibVersionHeightMap
	NvmGasLimitWithoutTimeoutHeight() uint64
	WsResetRecordDependencyHeight2() uint64 //reserve change log
	TransferFromContractFailureEventRecordableHeight2() uint64
	NvmValueCheckUpdateHeight() uint64
	NbreAvailableHeight() uint64
	Nrc20SecurityCheckHeight() uint64
	NbreSplitHeight() uint64
	NodeUpdateHeight() uint64

	NodeStartSerial() uint64
	NodeAccessContract() *Address
	NodePodContract() *Address
	NodeGovernanceContract() *Address
}

// NebCompatibility ..
var NebCompatibility = NewCompatibilityTestNet()

// SetCompatibilityOptions set compatibility height according to chain_id
func SetCompatibilityOptions(chainID uint32) {

	if chainID == MainNetID {
		NebCompatibility = NewCompatibilityMainNet()
	} else if chainID == TestNetID {
		NebCompatibility = NewCompatibilityTestNet()
	} else {
		NebCompatibility = NewCompatibilityLocal()
	}

	logging.VLog().WithFields(logrus.Fields{
		"chain_id": chainID,
		"TransferFromContractEventRecordableHeight": NebCompatibility.TransferFromContractEventRecordableHeight(),
		"AcceptFuncAvailableHeight":                 NebCompatibility.AcceptFuncAvailableHeight(),
		"RandomAvailableHeight":                     NebCompatibility.RandomAvailableHeight(),
		"DateAvailableHeight":                       NebCompatibility.DateAvailableHeight(),
		"RecordCallContractResultHeight":            NebCompatibility.RecordCallContractResultHeight(),
		"NvmMemoryLimitWithoutInjectHeight":         NebCompatibility.NvmMemoryLimitWithoutInjectHeight(),
		"WsResetRecordDependencyHeight":             NebCompatibility.WsResetRecordDependencyHeight(),
		"WsResetRecordDependencyHeight2":            NebCompatibility.WsResetRecordDependencyHeight2(),
		"V8JSLibVersionControlHeight":               NebCompatibility.V8JSLibVersionControlHeight(),
		"V8JSLibVersionHeightMap":                   NebCompatibility.V8JSLibVersionHeightMap().String(),
		"TransferFromContractFailureHeight":         NebCompatibility.TransferFromContractFailureEventRecordableHeight(),
		"TransferFromContractFailureHeight2":        NebCompatibility.TransferFromContractFailureEventRecordableHeight2(),
		"NewNvmExeTimeoutConsumeGasHeight":          NebCompatibility.NewNvmExeTimeoutConsumeGasHeight(),
		"NvmExeTimeoutHeight":                       NebCompatibility.NvmExeTimeoutHeight(),
		"NbreAvailableHeight":                       NebCompatibility.NbreAvailableHeight(),
	}).Info("Set compatibility options.")

	NebCompatibility.V8JSLibVersionHeightMap().validate()
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
				/*
					logging.VLog().WithFields(logrus.Fields{
						"libname":       libname,
						"deployVersion": deployVersion,
						"return":        libs[i],
					}).Debug("filter js lib.")
				*/
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
	m := NebCompatibility.V8JSLibVersionHeightMap()
	for _, v := range m.DescKeys {
		if blockHeight >= m.Data[v] {
			return v
		}
	}
	return ""
}

// V8BlockSeedAvailableAtHeight ..
func V8BlockSeedAvailableAtHeight(blockHeight uint64) bool {
	/*  For old contract, Blockchain.block.seed should not be disable
	 * 		return blockHeight >= NebCompatibility.RandomAvailableHeight() &&
	 * 		blockHeight < NebCompatibility.V8JSLibVersionHeightMap().GetHeightOfVersion("1.1.0")
	 */

	return blockHeight >= NebCompatibility.RandomAvailableHeight()
}

// V8JSLibVersionControlAtHeight ..
func V8JSLibVersionControlAtHeight(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.V8JSLibVersionControlHeight()
}

// RandomAvailableAtHeight ..
func RandomAvailableAtHeight(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.RandomAvailableHeight()
}

// DateAvailableAtHeight ..
func DateAvailableAtHeight(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.DateAvailableHeight()
}

// AcceptAvailableAtHeight ..
func AcceptAvailableAtHeight(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.AcceptFuncAvailableHeight()
}

// WsResetRecordDependencyAtHeight ..
func WsResetRecordDependencyAtHeight(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.WsResetRecordDependencyHeight()
}

// WsResetRecordDependencyAtHeight2 ..
func WsResetRecordDependencyAtHeight2(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.WsResetRecordDependencyHeight2()
}

// RecordCallContractResultAtHeight ..
func RecordCallContractResultAtHeight(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.RecordCallContractResultHeight()
}

// NvmMemoryLimitWithoutInjectAtHeight ..
func NvmMemoryLimitWithoutInjectAtHeight(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.NvmMemoryLimitWithoutInjectHeight()
}

// NewNvmExeTimeoutConsumeGasAtHeight ..
func NewNvmExeTimeoutConsumeGasAtHeight(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.NewNvmExeTimeoutConsumeGasHeight()
}

// TransferFromContractEventRecordableAtHeight ..
func TransferFromContractEventRecordableAtHeight(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.TransferFromContractEventRecordableHeight()
}

// TransferFromContractFailureEventRecordableAtHeight ..
func TransferFromContractFailureEventRecordableAtHeight(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.TransferFromContractFailureEventRecordableHeight()
}

// TransferFromContractFailureEventRecordableAtHeight2 ..
func TransferFromContractFailureEventRecordableAtHeight2(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.TransferFromContractFailureEventRecordableHeight2()
}

// NvmGasLimitWithoutTimeoutAtHeight ..
func NvmGasLimitWithoutTimeoutAtHeight(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.NvmGasLimitWithoutTimeoutHeight()
}

// NvmExeTimeoutAtHeight ..
func NvmExeTimeoutAtHeight(blockHeight uint64) bool {
	for _, height := range NebCompatibility.NvmExeTimeoutHeight() {
		if blockHeight == height {
			return true
		}
	}
	return false
}

// GetNearestInstructionCounterVersionAtHeight ..
func GetNearestInstructionCounterVersionAtHeight(blockHeight uint64) string {
	m := NebCompatibility.V8JSLibVersionHeightMap()
	for _, v := range m.DescKeys {
		if v == "1.1.0" && blockHeight >= m.Data[v] {
			return v
		}
	}
	return "1.0.0"
}

// EnableInnerContractAtHeight ..
func EnableInnerContractAtHeight(blockHeight uint64) bool {
	m := NebCompatibility.V8JSLibVersionHeightMap()
	return blockHeight >= m.Data["1.1.0"]
}

// NvmValueCheckUpdateHeight ..
func NvmValueCheckUpdateHeight(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.NvmValueCheckUpdateHeight()
}

// NbreAvailableHeight ..
func NbreAvailableHeight(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.NbreAvailableHeight()
}

// Nrc20SecurityCheckAtHeight ..
func Nrc20SecurityCheckAtHeight(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.Nrc20SecurityCheckHeight()
}

// NbreSplitAtHeight ..
func NbreSplitAtHeight(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.NbreSplitHeight()
}

// NodeUpdateAtHeight ..
func NodeUpdateAtHeight(blockHeight uint64) bool {
	return blockHeight >= NebCompatibility.NodeUpdateHeight()
}

// NodeStartSerial ..
func NodeStartSerial() uint64 {
	return NebCompatibility.NodeStartSerial()
}

// NodeAccessContract ..
func NodeAccessContract() *Address {
	return NebCompatibility.NodeAccessContract()
}

// NodePodContract ..
func NodePodContract() *Address {
	return NebCompatibility.NodePodContract()
}

// NodeGovernanceContract ..
func NodeGovernanceContract() *Address {
	return NebCompatibility.NodeGovernanceContract()
}
