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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core/pb"
)

const (
	NRC20FuncTransfer     = "transfer"
	NRC20FuncTransferFrom = "transferFrom"
	NRC20FuncApprove      = "approve"
)

type Access struct {
	access *corepb.Access
}

// NewAccess returns the Access
func NewAccess(path string) (*Access, error) {
	if path != "" {
		path, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		content := string(bytes)

		access := new(corepb.Access)
		if err = proto.UnmarshalText(content, access); err != nil {
			return nil, err
		}
		return &Access{access: access}, nil
	}
	return &Access{}, nil
}

// CheckTransaction Check that the transaction meets the conditions
func (a *Access) CheckTransaction(tx *Transaction) error {
	if a.access == nil {
		// no access config need to check
		return nil
	}
	if a.access.Blacklist != nil {
		for _, addr := range a.access.Blacklist.From {
			if addr == tx.from.String() {
				return ErrRestrictedFromAddress
			}
		}
		for _, addr := range a.access.Blacklist.To {
			if addr == tx.to.String() {
				return ErrRestrictedToAddress
			}
		}

		if tx.Type() == TxPayloadDeployType || tx.Type() == TxPayloadCallType {
			for _, contract := range a.access.Blacklist.Contracts {
				match := false
				if contract.Address != "" {
					match = contract.Address == tx.to.String()
				}
				if tx.Type() == TxPayloadCallType && len(contract.Functions) > 0 {
					payload, err := tx.LoadPayload()
					callPayload := payload.(*CallPayload)
					if err != nil {
						return err
					}
					funcMatch := false
					for _, function := range contract.Functions {
						if function == callPayload.Function {
							funcMatch = true
							break
						}
					}
					match = match && funcMatch
				}
				if match {
					return ErrUnsupportedFunction
				}
				if tx.Type() == TxPayloadDeployType && len(contract.Keywords) > 0 {
					data := strings.ToLower(string(tx.Data()))
					for _, keyword := range contract.Keywords {
						keyword = strings.ToLower(keyword)
						if strings.Contains(data, keyword) {
							unsupportedKeywordError := fmt.Sprintf("transaction data has unsupported keyword(keyword: %s)", keyword)
							return errors.New(unsupportedKeywordError)
						}
					}
				}
			}
		}
	}

	if a.access.Nrc20List != nil {
		// check nrc20 security
		if err := a.nrc20SecurityCheck(tx); err != nil {
			return err
		}
	}
	return nil
}

// nrc20SecurityCheck check nrc20 contract params security
func (a *Access) nrc20SecurityCheck(tx *Transaction) error {
	if tx.Type() == TxPayloadCallType && len(a.access.Nrc20List.Contracts) > 0 {
		for _, contract := range a.access.Nrc20List.Contracts {
			// check nrc20 security
			if tx.To().String() == contract {
				payload, err := tx.LoadPayload()
				if err != nil {
					return err
				}
				call := payload.(*CallPayload)
				valueIndex := 0
				switch call.Function {
				case NRC20FuncTransfer:
					valueIndex = 1
				case NRC20FuncTransferFrom:
					valueIndex = 2
				case NRC20FuncApprove:
					valueIndex = 2
				default:
					valueIndex = -1
				}
				if valueIndex > 0 {
					var argsObj []string
					if err := json.Unmarshal([]byte(call.Args), &argsObj); err != nil {
						return ErrNrc20ArgsCheckFailed
					}
					addr := argsObj[0]
					if _, err := AddressParse(addr); err != nil {
						return ErrNrc20AddressCheckFailed
					}
					value := argsObj[valueIndex]
					if matched, err := regexp.MatchString("^[0-9]+$", value); matched == false || err != nil {
						return ErrNrc20ValueCheckFailed
					}
				}
			}
		}
	}
	return nil
}
