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

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/neblet"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/urfave/cli"
)

var (
	serializeCommand = cli.Command{
		Name:     "serialize",
		Usage:    "manage serialize",
		Category: "SERIALIZE COMMANDS",
		Description: `
Manage serialize, serialize transaction, serialize block, serialize download.`,

		Subcommands: []cli.Command{
			{
				Name:      "transaction",
				Usage:     "serialize a transaction request data",
				Action:    MergeFlags(serializeTx),
				ArgsUsage: "<file>",
				Description: `
    neb serialize transaction <file>

serialize the transaction request with file path.`,
			},
			{
				Name:      "download",
				Usage:     "serialize a download request data",
				Action:    MergeFlags(serializeDownload),
				ArgsUsage: "<file>",
				Description: `
    neb serialize download <file>

serialize the download request with file path.`,
			},
		},
	}
)

type txJSON struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Nonce     uint64 `json:"nonce"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
	ChainID   uint32 `json:"chain_id"`
	GasPrice  string `json:"gas_price"`
	GasLimit  string `json:"gas_limit"`

	Contract *contractJSON `json:"contract"`

	Candidate *candidateJSON `json:"candidate"`

	Delegate *delegateJSON `json:"delegate"`

	// from key file path
	Keyfile string `json:"keyfile"`
	// from key passphrase
	Passphrase string `json:"passphrase"`
}

type contractJSON struct {
	Source     string `json:"source"`
	SourceType string `json:"source_type"`
	Function   string `json:"function"`
	Args       string `json:"args"`
}

type candidateJSON struct {
	Action string `json:"action"`
}

type delegateJSON struct {
	Action    string `json:"action"`
	Delegatee string `json:"delegatee"`
}

type blockHeaderJSON struct {
	ParentHash string `json:"parent_hash"`
	Coinbase   string `json:"coinbase"`
	Nonce      string `json:"nonce"`
	Timestamp  int64  `json:"timestamp"`
	ChainID    uint32 `json:"chain_id"`
}

type blockJSON struct {
	Header       blockHeaderJSON `json:"header"`
	Miner        string          `json:"miner"`
	Transactions txJSON          `json:"transactions"`
	Height       uint64          `json:"height"`

	// from key file path
	Keyfile string `json:"keyfile"`
	// from key passphrase
	Passphrase string `json:"passphrase"`
}

func serializeTx(ctx *cli.Context) error {
	filePath := ctx.Args().First()
	txData, err := ioutil.ReadFile(filePath)
	if err != nil {
		FatalF("serializeTx failed:%s", err)
	}

	txJSON := new(txJSON)
	err = json.Unmarshal(txData, txJSON)
	if err != nil {
		FatalF("serializeTx failed:%s", err)
	}

	neb, err := setupNeb(ctx)
	if err != nil {
		FatalF("serializeTx failed:%s", err)
	}

	tx, err := parseTransaction(neb, txJSON)
	if err != nil {
		FatalF("serializeTx failed:%s", err)
	}

	addr, err := loadAndUnlockKey(neb, txJSON.Keyfile, txJSON.Passphrase)
	if err != nil {
		FatalF("serializeTx failed:%s", err)
	}
	err = neb.AccountManager().SignTransaction(addr, tx)
	if err != nil {
		FatalF("serializeTx failed:%s", err)
	}

	pbMsg, _ := tx.ToProto()
	data, _ := proto.Marshal(pbMsg)
	data = neb.NetService().BuildRawMessageData(data, core.MessageTypeNewTx)
	fmt.Println(base64.StdEncoding.EncodeToString(data))
	return nil
}

func setupNeb(ctx *cli.Context) (*neblet.Neblet, error) {
	neb, err := makeNeb(ctx)
	if err != nil {
		return nil, err
	}

	neb.Setup()
	return neb, nil
}

func loadAndUnlockKey(neb *neblet.Neblet, keyfile, passphrase string) (*core.Address, error) {
	keyJSON, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}
	addr, err := neb.AccountManager().Load(keyJSON, []byte(passphrase))
	if err != nil {
		return nil, err
	}
	err = neb.AccountManager().Unlock(addr, []byte(passphrase), keystore.DefaultUnlockDuration)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func parseTransaction(neb *neblet.Neblet, txJSON *txJSON) (*core.Transaction, error) {
	fromAddr, err := core.AddressParse(txJSON.From)
	if err != nil {
		return nil, err
	}
	toAddr, err := core.AddressParse(txJSON.To)
	if err != nil {
		return nil, err
	}

	value := util.NewUint128FromString(txJSON.Value)
	gasPrice := util.NewUint128FromString(txJSON.GasPrice)
	gasLimit := util.NewUint128FromString(txJSON.GasLimit)

	var (
		payloadType string
		payload     []byte
	)
	if txJSON.Contract != nil && len(txJSON.Contract.Source) > 0 {
		payloadType = core.TxPayloadDeployType
		payload, err = core.NewDeployPayload(txJSON.Contract.SourceType, txJSON.Contract.Source, txJSON.Contract.Args).ToBytes()
	} else if txJSON.Contract != nil && len(txJSON.Contract.Function) > 0 {
		payloadType = core.TxPayloadCallType
		payload, err = core.NewCallPayload(txJSON.Contract.Function, txJSON.Contract.Args).ToBytes()
	} else if txJSON.Candidate != nil {
		payloadType = core.TxPayloadCandidateType
		payload, err = core.NewCandidatePayload(txJSON.Candidate.Action).ToBytes()
	} else if txJSON.Delegate != nil {
		payloadType = core.TxPayloadDelegateType
		payload, err = core.NewDelegatePayload(txJSON.Delegate.Action, txJSON.Delegate.Delegatee).ToBytes()
	} else {
		payloadType = core.TxPayloadBinaryType
	}
	if err != nil {
		return nil, err
	}

	tx := core.NewTransaction(neb.BlockChain().ChainID(), fromAddr, toAddr, value, txJSON.Nonce, payloadType, payload, gasPrice, gasLimit)
	return tx, nil
}

func serializeDownload(ctx *cli.Context) error {
	hashArg := ctx.Args().First()
	signArg := ctx.Args().Get(1)

	hash, _ := byteutils.FromHex(hashArg)
	sign, _ := byteutils.FromHex(signArg)

	neb, err := setupNeb(ctx)
	if err != nil {
		FatalF("serializeDownload failed:%s", err)
	}

	downloadMsg := &corepb.DownloadBlock{
		Hash: hash,
		Sign: sign,
	}
	data, _ := proto.Marshal(downloadMsg)
	data = neb.NetService().BuildRawMessageData(data, core.MessageTypeDownloadedBlock)
	fmt.Println(base64.StdEncoding.EncodeToString(data))
	return nil
}
