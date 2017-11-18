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

package core

import (
	"encoding/json"
	"errors"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	log "github.com/sirupsen/logrus"
)

// Action types
const (
	LoginAction    = "login"
	LogoutAction   = "logout"
	WithdrawAction = "withdraw"
)

// Withdraw works after logout > 1000 blocks
const (
	WithdrawExpireBlocks = 1000
)

// Errors constants
var (
	ErrInvalidElectAction   = errors.New("invalid elect action")
	ErrChargeTooManyDeposit = errors.New("charged too many deposit")
)

// Deposit 4000 NAS
var (
	StandardDeposit = util.NewUint128FromInt(4000)
)

// ElectPayload carry election information
type ElectPayload struct {
	Action string
}

// LoadElectPayload from bytes
func LoadElectPayload(bytes []byte) (*ElectPayload, error) {
	payload := &ElectPayload{}
	if err := json.Unmarshal(bytes, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// NewElectPayload with function & args
func NewElectPayload(action string) *ElectPayload {
	return &ElectPayload{
		Action: action,
	}
}

// ToBytes serialize payload
func (payload *ElectPayload) ToBytes() ([]byte, error) {
	return json.Marshal(payload)
}

func (payload *ElectPayload) calDiffBetweenCurrentAndStandardDeposit(from []byte, block *Block) (*util.Uint128, error) {
	diff := util.NewUint128FromBigInt(StandardDeposit.Int)
	bytes, err := block.depositTrie.Get(from)
	if err == nil {
		// already charged some deposit
		current, err := util.NewUint128FromFixedSizeByteSlice(bytes)
		if err != nil {
			return nil, err
		}
		if diff.Cmp(current.Int) < 0 {
			log.WithFields(log.Fields{
				"func":     "ElectPayload.Execute",
				"from":     from,
				"original": current.Int64(),
				"standard": diff.Int64(),
			}).Error("charged too many deposit")
			return nil, ErrChargeTooManyDeposit
		}
		diff.Sub(diff.Int, current.Int)
		log.WithFields(log.Fields{
			"func":     "ElectPayload.Execute",
			"from":     from,
			"original": current.Int64(),
			"diff":     diff.Int64(),
		}).Debug("fill deposit")
		return diff, nil
	} else if err == storage.ErrKeyNotFound {
		return diff, nil
	} else {
		return nil, err
	}
}

func (payload *ElectPayload) login(from []byte, block *Block) error {
	exist, err := checkExistanceInDynasty(from, block.dynastyCandidatesTrie)
	if err != nil {
		return err
	}
	if exist {
		log.WithFields(log.Fields{
			"func": "ElectPayload.Execute",
			"from": from,
		}).Warn("dup login")
		return nil
	}
	// login, charge deposit
	diff, err := payload.calDiffBetweenCurrentAndStandardDeposit(from, block)
	if err != nil {
		return err
	}
	account := block.accState.GetOrCreateUserAccount(from)
	if account.Balance().Cmp(diff.Int) < 0 {
		return ErrInsufficientBalance
	}
	account.SubBalance(diff)
	bytes, err := StandardDeposit.ToFixedSizeByteSlice()
	if err != nil {
		return err
	}
	if _, err = block.depositTrie.Put(from, bytes); err != nil {
		return err
	}
	if _, err = block.dynastyCandidatesTrie.Put(from, from); err != nil {
		return err
	}
	return nil
}

func (payload *ElectPayload) logout(from []byte, block *Block) error {
	_, err := block.dynastyCandidatesTrie.Get(from)
	if err == nil {
		block.dynastyCandidatesTrie.Del(from)
		log.WithFields(log.Fields{
			"func": "ElectPayload.Execute",
			"from": from,
		}).Debug("logout.")
		return nil
	}
	if err != storage.ErrKeyNotFound {
		return err
	}
	log.WithFields(log.Fields{
		"func": "ElectPayload.Execute",
		"from": from,
	}).Warn("logout before login")
	return nil
}

func checkExistanceInDynasty(address []byte, dynasty *trie.BatchTrie) (bool, error) {
	_, err := dynasty.Get(address)
	if err == nil {
		return true, nil
	}
	if err != storage.ErrKeyNotFound {
		return false, err
	}
	return false, nil
}

func (payload *ElectPayload) checkWithdrawExpired(from []byte, block *Block) (bool, error) {
	// not in candidates, should logout at first
	exist, err := checkExistanceInDynasty(from, block.dynastyCandidatesTrie)
	if err != nil {
		return false, err
	}
	if exist {
		log.WithFields(log.Fields{
			"func":  "ElectPayload.Execute",
			"from":  from,
			"block": block,
		}).Warn("cannot withdraw. not expired. exist in candidates, pls logout at first.")
		return false, nil
	}
	// not in next dynasty
	exist, err = checkExistanceInDynasty(from, block.nextDynastyTrie)
	if err != nil {
		return false, err
	}
	if exist {
		log.WithFields(log.Fields{
			"func":  "ElectPayload.Execute",
			"from":  from,
			"block": block,
		}).Warn("cannot withdraw. not expired. exist in next dynasty.")
		return false, nil
	}
	// not in previous 1000 blocks' dynasties
	curBlock := block
	for i := 0; i < WithdrawExpireBlocks; i++ {
		exist, err := checkExistanceInDynasty(from, curBlock.curDynastyTrie)
		if err != nil {
			return false, err
		}
		if exist {
			log.WithFields(log.Fields{
				"func":     "ElectPayload.Execute",
				"from":     from,
				"current":  block,
				"ancestor": curBlock,
			}).Warn("cannot withdraw. not expired. found in recent ancestors.")
			return false, nil
		}
		curBlock, err = LoadBlockFromStorage(block.ParentHash(), block.storage, block.txPool)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (payload *ElectPayload) refundDeposit(from []byte, block *Block) error {
	bytes, err := block.depositTrie.Get(from)
	if err != nil {
		return err
	}
	deposit, err := util.NewUint128FromFixedSizeByteSlice(bytes)
	if err != nil {
		return err
	}
	account := block.accState.GetOrCreateUserAccount(from)
	account.AddBalance(deposit)
	block.depositTrie.Del(from)
	log.WithFields(log.Fields{
		"func":    "ElectPayload.Execute",
		"from":    from,
		"deposit": deposit.Int64(),
	}).Warn("withdraw")
	return nil
}

func (payload *ElectPayload) withdraw(from []byte, block *Block) error {
	// check expired
	expired, err := payload.checkWithdrawExpired(from, block)
	if expired {
		// refund deposit
		return payload.refundDeposit(from, block)
	}
	return err
}

// Execute the call payload in tx, call a function
func (payload *ElectPayload) Execute(tx *Transaction, block *Block) error {
	from := tx.from.Bytes()
	switch payload.Action {
	case LoginAction:
		return payload.login(from, block)
	case LogoutAction:
		return payload.logout(from, block)
	case WithdrawAction:
		return payload.withdraw(from, block)
	default:
		return ErrInvalidElectAction
	}
}
