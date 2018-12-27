package nvm

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/consensus/dpos"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/stretchr/testify/assert"
)

func TestBlockChain(t *testing.T) {
	tests := []struct {
		filepath    string
		expectedErr error
	}{
		{"test/test_blockchain.js", nil},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			addr, _ := core.AddressParse("n1FkntVUMPAsESuCAAPK711omQk19JotBjM")
			owner, err := context.GetOrCreateUserAccount(addr.Bytes())
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(1000000000))
			addr, _ = core.AddressParse("n1JNHZJEUvfBYfjDRD14Q73FX62nJAzXkMR")
			contract, err := context.CreateContractAccount(addr.Bytes(), nil, nil)
			assert.Nil(t, err)

			ctx, err := NewContext(mockBlock(), mockTransaction(), contract, context)
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(100000, 10000000)
			_, err = engine.RunScriptSource(string(data), 0)
			assert.Equal(t, tt.expectedErr, err)
			engine.Dispose()
		})
	}
}

func TestContractFeatureGetAccountState(t *testing.T) {
	type fields struct {
		function string
		args     string
		result   string
		error    string
	}
	tests := []struct {
		contract   string
		sourceType string
		initArgs   string
		calls      []fields
	}{
		{
			"./test/test_contract_features.js",
			"js",
			"[]",
			[]fields{
				{"testGetAccountState", "[]", "\"1000000000000\"", ""},
				{"testGetAccountStateWrongAddr", "[]", "\"0\"", ""},
			},
		},
	}

	account1 := "n1FkntVUMPAsESuCAAPK711omQk19JotBjM"
	account2 := "n1JNHZJEUvfBYfjDRD14Q73FX62nJAzXkMR"

	for _, tt := range tests {
		t.Run(tt.contract, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.contract)
			assert.Nil(t, err, "contract path read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			add1, _ := core.AddressParse(account1)
			owner, err := context.GetOrCreateUserAccount(add1.Bytes())
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(1000000000000))
			add2, _ := core.AddressParse(account2)
			contract, err := context.CreateContractAccount(add2.Bytes(), nil, &corepb.ContractMeta{Version: "1.0.5"})
			assert.Nil(t, err)
			tx := mockNormalTransaction("n1FkntVUMPAsESuCAAPK711omQk19JotBjM", "n1JNHZJEUvfBYfjDRD14Q73FX62nJAzXkMR", "0")
			ctx, err := NewContext(mockBlockForLib(2000000), tx, contract, context)

			// deploy and init.
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(100000, 10000000)
			_, err = engine.DeployAndInit(string(data), tt.sourceType, tt.initArgs)
			assert.Nil(t, err)
			engine.Dispose()

			// call.
			for _, fields := range tt.calls {
				state, _ := ctx.state.GetOrCreateUserAccount([]byte(account1))
				fmt.Println("===", state)
				engine = NewV8Engine(ctx)
				engine.SetExecutionLimits(100000, 10000000)
				result, err := engine.Call(string(data), tt.sourceType, fields.function, fields.args)
				assert.Equal(t, fields.result, result)
				assert.Nil(t, err)
				engine.Dispose()
			}
		})
	}
}

func TestContractsFeatureGetBlockHashAndSeed(t *testing.T) {
	type fields struct {
		function string
		args     string
		result   string
		err      error
	}
	tests := []struct {
		contract   string
		sourceType string
		initArgs   string
		calls      []fields
	}{
		{
			"./test/test_contract_features.js",
			"js",
			"[]",
			[]fields{
				{"testGetPreBlockHash1", "[1]", "\"" + byteutils.Hex([]byte("blockHash")) + "\"", nil},
				{"testGetPreBlockHash1", "[0]", "getPreBlockHash: invalid offset", core.ErrExecutionFailed},
				{"testGetPreBlockHashByNativeBlock", "[1.1]", "Blockchain.GetPreBlockHash(), argument must be integer", core.ErrExecutionFailed},
				{"testGetPreBlockSeedByNativeBlock", "[1.1]", "Blockchain.GetPreBlockSeed(), argument must be integer", core.ErrExecutionFailed},
				{"testGetPreBlockHash1", "[1111111111111111111]", "getPreBlockHash: block not exist", core.ErrExecutionFailed},
				{"testGetPreBlockSeed1", "[1]", "\"" + byteutils.Hex([]byte("randomSeed")) + "\"", nil},
			},
		},
	}

	account1 := "n1FkntVUMPAsESuCAAPK711omQk19JotBjM"
	account2 := "n1JNHZJEUvfBYfjDRD14Q73FX62nJAzXkMR"

	for _, tt := range tests {
		t.Run(tt.contract, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.contract)
			assert.Nil(t, err, "contract path read error")

			mem, _ := storage.NewMemoryStorage()
			curBlock := mockBlockForLib(2000000)

			preBlock := &corepb.Block{
				Header: &corepb.BlockHeader{
					Random: &corepb.Random{
						VrfSeed: []byte("randomSeed"),
					},
				},
			}
			preBlockHash := []byte("blockHash")
			preBlockHeight := curBlock.Height() - 1
			blockBytes, err := proto.Marshal(preBlock)
			assert.Nil(t, err)

			mem.Put(byteutils.FromUint64(preBlockHeight), preBlockHash)
			mem.Put(preBlockHash, blockBytes)

			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			add1, _ := core.AddressParse(account1)
			owner, err := context.GetOrCreateUserAccount(add1.Bytes())
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(1000000000000))
			add2, _ := core.AddressParse(account2)
			contract, err := context.CreateContractAccount(add2.Bytes(), nil, &corepb.ContractMeta{Version: "1.0.5"})
			assert.Nil(t, err)
			tx := mockNormalTransaction("n1FkntVUMPAsESuCAAPK711omQk19JotBjM", "n1JNHZJEUvfBYfjDRD14Q73FX62nJAzXkMR", "0")
			ctx, err := NewContext(mockBlockForLib(2000000), tx, contract, context)

			// deploy and init.
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(100000, 10000000)
			_, err = engine.DeployAndInit(string(data), tt.sourceType, tt.initArgs)
			assert.Nil(t, err)
			engine.Dispose()

			// call.
			for _, fields := range tt.calls {
				engine = NewV8Engine(ctx)
				engine.SetExecutionLimits(100000, 10000000)
				result, err := engine.Call(string(data), tt.sourceType, fields.function, fields.args)
				fmt.Println("result", result)
				assert.Equal(t, fields.result, result)
				assert.Equal(t, fields.err, err)
				engine.Dispose()
			}
		})
	}
}
func TestTransferValueFromContracts(t *testing.T) {
	type fields struct {
		function string
		args     string
	}
	tests := []struct {
		contract   string
		sourceType string
		initArgs   string
		calls      []fields
		value      string
		success    bool
	}{
		{
			"./test/transfer_value_from_contract.js",
			"js",
			"",
			[]fields{
				{"transfer", "[\"n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17\"]"},
			},
			"100",
			true,
		},
		{
			"./test/transfer_value_from_contract.js",
			"js",
			"",
			[]fields{
				{"transfer", "[\"n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17\"]"},
			},
			"101",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.contract, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.contract)
			assert.Nil(t, err, "contract path read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)

			owner, err := context.GetOrCreateUserAccount([]byte("account1"))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(10000000))
			addr, err := core.NewContractAddressFromData([]byte("n1FkntVUMPAsESuCAAPK711omQk19JotBjM"), byteutils.FromUint64(1))
			assert.Nil(t, err)
			contract, err := context.CreateContractAccount(addr.Bytes(), nil, nil)
			assert.Nil(t, err)

			contract.AddBalance(newUint128FromIntWrapper(100))
			mockTx := mockNormalTransaction("n1FkntVUMPAsESuCAAPK711omQk19JotBjM", "n1FkntVUMPAsESuCAAPK711omQk19JotBjM", tt.value)
			ctx, err := NewContext(mockBlock(), mockTx, contract, context)

			// deploy and init.
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(1000, 10000000)
			_, err = engine.DeployAndInit(string(data), tt.sourceType, tt.initArgs)
			assert.Nil(t, err)
			engine.Dispose()

			// call.
			for _, fields := range tt.calls {
				engine = NewV8Engine(ctx)
				engine.SetExecutionLimits(10000, 10000000)
				result, err := engine.Call(string(data), tt.sourceType, fields.function, fields.args)
				if tt.success {
					assert.Equal(t, result, "\""+fmt.Sprint(tt.value)+"\"")
					assert.Nil(t, err)
				} else {
					assert.NotNil(t, err)
				}
				engine.Dispose()
			}
		})
	}
}
