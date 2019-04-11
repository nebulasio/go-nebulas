package nvm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/nebulasio/go-nebulas/consensus/dpos"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/stretchr/testify/assert"
)

func TestERC20(t *testing.T) {
	tests := []struct {
		name         string
		contractPath string
		sourceType   string
		initArgs     string
		totalSupply  string
	}{
		{"deploy ERC20.js", "./test/ERC20.js", "js", "[\"TEST001\", \"TEST\", 1000000000]", "1000000000"},
	}

	// TODO: Addd more test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.contractPath)
			assert.Nil(t, err, "contract path read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			owner, err := context.GetOrCreateUserAccount([]byte("account1"))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(10000000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil, nil)

			ctx, err := NewContext(mockBlock(), mockTransaction(), contract, context)
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 10000000)
			_, err = engine.DeployAndInit(string(data), tt.sourceType, tt.initArgs)
			assert.Nil(t, err)
			engine.Dispose()

			engine = NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 10000000)
			_, err = engine.Call(string(data), tt.sourceType, "totalSupply", "[]")
			assert.Nil(t, err)
			engine.Dispose()

		})
	}
}

func TestNRC20Contract(t *testing.T) {
	type TransferTest struct {
		to     string
		result bool
		value  string
	}

	tests := []struct {
		test          string
		contractPath  string
		sourceType    string
		name          string
		symbol        string
		decimals      int
		totalSupply   string
		from          string
		transferTests []TransferTest
	}{
		{"nrc20", "./test/NRC20.js", "js", "StandardToken标准代币", "ST", 18, "1000000000",
			"n1FkntVUMPAsESuCAAPK711omQk19JotBjM",
			[]TransferTest{
				{"n1FkntVUMPAsESuCAAPK711omQk19JotBjM", true, "5"},
				{"n1JNHZJEUvfBYfjDRD14Q73FX62nJAzXkMR", true, "10"},
				{"n1Kjom3J4KPsHKKzZ2xtt8Lc9W5pRDjeLcW", true, "15"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.contractPath)
			assert.Nil(t, err, "contract path read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			owner, err := context.GetOrCreateUserAccount([]byte(tt.from))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(10000000))

			// prepare the contract.
			contractAddr, err := core.AddressParse(contractStr)
			contract, _ := context.CreateContractAccount(contractAddr.Bytes(), nil, nil)
			contract.AddBalance(newUint128FromIntWrapper(5))

			// parepare env, block & transactions.
			tx := mockNormalTransaction(tt.from, "n1TV3sU6jyzR4rJ1D7jCAmtVGSntJagXZHC", "0")
			ctx, err := NewContext(mockBlock(), tx, contract, context)

			// execute.
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 100000000)
			args := fmt.Sprintf("[\"%s\", \"%s\", %d, \"%s\"]", tt.name, tt.symbol, tt.decimals, tt.totalSupply)
			_, err = engine.DeployAndInit(string(data), tt.sourceType, args)
			assert.Nil(t, err)
			engine.Dispose()

			// call name.
			engine = NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 100000000)
			name, err := engine.Call(string(data), tt.sourceType, "name", "")
			assert.Nil(t, err)
			var nameStr string
			err = json.Unmarshal([]byte(name), &nameStr)
			assert.Nil(t, err)
			assert.Equal(t, tt.name, nameStr)
			engine.Dispose()

			// call symbol.
			engine = NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 100000000)
			symbol, err := engine.Call(string(data), tt.sourceType, "symbol", "")
			assert.Nil(t, err)
			var symbolStr string
			err = json.Unmarshal([]byte(symbol), &symbolStr)
			assert.Nil(t, err)
			assert.Equal(t, tt.symbol, symbolStr)
			assert.Nil(t, err)
			engine.Dispose()

			// call decimals.
			engine = NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 100000000)
			decimals, err := engine.Call(string(data), tt.sourceType, "decimals", "")
			assert.Nil(t, err)
			var decimalsInt int
			err = json.Unmarshal([]byte(decimals), &decimalsInt)
			assert.Nil(t, err)
			assert.Equal(t, tt.decimals, decimalsInt)
			assert.Nil(t, err)
			engine.Dispose()

			// call totalSupply.
			engine = NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 100000000)
			totalSupply, err := engine.Call(string(data), tt.sourceType, "totalSupply", "")
			assert.Nil(t, err)
			var totalSupplyStr string
			err = json.Unmarshal([]byte(totalSupply), &totalSupplyStr)
			assert.Nil(t, err)
			expect, _ := big.NewInt(0).SetString(tt.totalSupply, 10)
			expect = expect.Mul(expect, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(tt.decimals)), nil))
			assert.Equal(t, expect.String(), totalSupplyStr)
			assert.Nil(t, err)
			engine.Dispose()

			// call takeout.
			for _, tot := range tt.transferTests {
				// call balanceOf.
				ctx.tx = mockNormalTransaction(tt.from, "n1TV3sU6jyzR4rJ1D7jCAmtVGSntJagXZHC", "0")
				engine = NewV8Engine(ctx)
				engine.SetExecutionLimits(10000, 100000000)
				balArgs := fmt.Sprintf("[\"%s\"]", tt.from)
				_, err := engine.Call(string(data), tt.sourceType, "balanceOf", balArgs)
				assert.Nil(t, err)
				engine.Dispose()

				engine = NewV8Engine(ctx)
				engine.SetExecutionLimits(10000, 100000000)
				transferArgs := fmt.Sprintf("[\"%s\", \"%s\"]", tot.to, tot.value)
				result, err := engine.Call(string(data), tt.sourceType, "transfer", transferArgs)
				assert.Nil(t, err)
				assert.Equal(t, "\"\"", result)
				engine.Dispose()

				engine = NewV8Engine(ctx)
				engine.SetExecutionLimits(10000, 100000000)
				approveArgs := fmt.Sprintf("[\"%s\", \"0\", \"%s\"]", tot.to, tot.value)
				result, err = engine.Call(string(data), tt.sourceType, "approve", approveArgs)
				assert.Nil(t, err)
				assert.Equal(t, "\"\"", result)
				engine.Dispose()

				engine = NewV8Engine(ctx)
				engine.SetExecutionLimits(10000, 100000000)
				allowanceArgs := fmt.Sprintf("[\"%s\", \"%s\"]", tt.from, tot.to)
				amount, err := engine.Call(string(data), tt.sourceType, "allowance", allowanceArgs)
				assert.Nil(t, err)
				var amountStr string
				err = json.Unmarshal([]byte(amount), &amountStr)
				assert.Nil(t, err)
				assert.Equal(t, tot.value, amountStr)
				engine.Dispose()

				ctx.tx = mockNormalTransaction(tot.to, "n1TV3sU6jyzR4rJ1D7jCAmtVGSntJagXZHC", "0")
				engine = NewV8Engine(ctx)
				engine.SetExecutionLimits(10000, 100000000)
				transferFromArgs := fmt.Sprintf("[\"%s\", \"%s\", \"%s\"]", tt.from, tot.to, tot.value)
				result, err = engine.Call(string(data), tt.sourceType, "transferFrom", transferFromArgs)
				assert.Nil(t, err)
				assert.Equal(t, "\"\"", result)
				engine.Dispose()

				ctx.tx = mockNormalTransaction(tot.to, "n1TV3sU6jyzR4rJ1D7jCAmtVGSntJagXZHC", "0")
				engine = NewV8Engine(ctx)
				engine.SetExecutionLimits(10000, 100000000)
				transferFromArgs = fmt.Sprintf("[\"%s\", \"%s\", \"%s\"]", tt.from, tot.to, tot.value)
				_, err = engine.Call(string(data), tt.sourceType, "transferFrom", transferFromArgs)
				assert.NotNil(t, err)
				engine.Dispose()
			}
		})
	}
}

func TestNRC20ContractMultitimes(t *testing.T) {
	for i := 0; i < 5; i++ {
		TestNRC20Contract(t)
	}
}
