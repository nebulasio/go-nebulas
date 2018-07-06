package nvm

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/nebulasio/go-nebulas/consensus/dpos"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/stretchr/testify/assert"
)

func TestBankVaultContract(t *testing.T) {
	type TakeoutTest struct {
		args          string
		expectedErr   error
		beforeBalance string
		afterBalance  string
	}

	tests := []struct {
		name         string
		contractPath string
		sourceType   string
		saveValue    string
		saveArgs     string
		takeoutTests []TakeoutTest
	}{
		{"deploy bank_vault_contract.js", "./test/bank_vault_contract.js", "js", "5", "[0]",
			[]TakeoutTest{
				{"[1]", nil, "5", "4"},
				{"[5]", core.ErrExecutionFailed, "4", "4"},
				{"[4]", nil, "4", "0"},
				{"[1]", core.ErrExecutionFailed, "0", "0"},
			},
		},
		{"deploy bank_vault_contract.ts", "./test/bank_vault_contract.ts", "ts", "5", "[0]",
			[]TakeoutTest{
				{"[1]", nil, "5", "4"},
				{"[5]", core.ErrExecutionFailed, "4", "4"},
				{"[4]", nil, "4", "0"},
				{"[1]", core.ErrExecutionFailed, "0", "0"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.contractPath)
			assert.Nil(t, err, "contract path read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			owner, err := context.GetOrCreateUserAccount([]byte("account1"))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(10000000))

			// prepare the contract.
			addr, err := core.NewContractAddressFromData([]byte("n1FkntVUMPAsESuCAAPK711omQk19JotBjM"), byteutils.FromUint64(1))
			assert.Nil(t, err)
			contract, _ := context.CreateContractAccount(addr.Bytes(), nil, nil)
			contract.AddBalance(newUint128FromIntWrapper(5))

			// parepare env, block & transactions.
			tx := mockNormalTransaction("n1FkntVUMPAsESuCAAPK711omQk19JotBjM", "n1JNHZJEUvfBYfjDRD14Q73FX62nJAzXkMR", tt.saveValue)
			ctx, err := NewContext(mockBlock(), tx, contract, context)

			// execute.
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 100000000)
			_, err = engine.DeployAndInit(string(data), tt.sourceType, "")
			assert.Nil(t, err)
			engine.Dispose()

			// call save.
			engine = NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 100000000)
			_, err = engine.Call(string(data), tt.sourceType, "save", tt.saveArgs)
			assert.Nil(t, err)
			engine.Dispose()

			var (
				bal struct {
					Balance string `json:"balance"`
				}
			)

			// call takeout.
			for _, tot := range tt.takeoutTests {
				// call balanceOf.
				engine = NewV8Engine(ctx)
				engine.SetExecutionLimits(10000, 100000000)
				balance, err := engine.Call(string(data), tt.sourceType, "balanceOf", "")
				assert.Nil(t, err)
				bal.Balance = ""
				err = json.Unmarshal([]byte(balance), &bal)
				assert.Nil(t, err)
				assert.Equal(t, tot.beforeBalance, bal.Balance)
				engine.Dispose()

				engine = NewV8Engine(ctx)
				engine.SetExecutionLimits(10000, 100000000)
				_, err = engine.Call(string(data), tt.sourceType, "takeout", tot.args)
				assert.Equal(t, err, tot.expectedErr)
				engine.Dispose()

				// call balanceOf.
				engine = NewV8Engine(ctx)
				engine.SetExecutionLimits(10000, 100000000)
				balance, err = engine.Call(string(data), tt.sourceType, "balanceOf", "")
				assert.Nil(t, err)
				bal.Balance = ""
				err = json.Unmarshal([]byte(balance), &bal)
				assert.Nil(t, err)
				assert.Equal(t, tot.afterBalance, bal.Balance)
				engine.Dispose()
			}
		})
	}
}
