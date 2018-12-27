package nvm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/nebulasio/go-nebulas/consensus/dpos"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/stretchr/testify/assert"
)

func TestNRC721Contract(t *testing.T) {

	tests := []struct {
		name         string
		contractPath string
		sourceType   string
		from         string
		to           string
		tokenID      string
	}{
		{"nrc721", "./test/NRC721BasicToken.js", "js",
			"n1FkntVUMPAsESuCAAPK711omQk19JotBjM", "n1Kjom3J4KPsHKKzZ2xtt8Lc9W5pRDjeLcW", "1001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.contractPath)
			assert.Nil(t, err, "contract path read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			assert.Nil(t, err)

			// prepare the contract.
			contract, _ := context.CreateContractAccount([]byte("account2"), nil, nil)
			contract.AddBalance(newUint128FromIntWrapper(5))

			// parepare env, block & transactions.
			tx := mockNormalTransaction(tt.from, "n1TV3sU6jyzR4rJ1D7jCAmtVGSntJagXZHC", "0")
			ctx, err := NewContext(mockBlock(), tx, contract, context)

			// execute.
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 100000000)
			args := fmt.Sprintf("[\"%s\"]", tt.name)
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

			// call mint.
			// engine = NewV8Engine(ctx)
			// engine.SetExecutionLimits(10000, 100000000)
			// mintArgs := fmt.Sprintf("[\"%s\", \"%s\"]", tt.from, tt.tokenID)
			// result, err := engine.Call(string(data), tt.sourceType, "mint", mintArgs)
			// assert.Nil(t, err)
			// assert.Equal(t, "\"\"", result)
			// engine.Dispose()

			// // call approve.
			// engine = NewV8Engine(ctx)
			// engine.SetExecutionLimits(10000, 100000000)
			// approveArgs := fmt.Sprintf("[\"%s\", \"%s\"]", tt.to, tt.tokenID)
			// result, err = engine.Call(string(data), tt.sourceType, "approve", approveArgs)
			// assert.Nil(t, err)
			// assert.Equal(t, "\"\"", result)
			// engine.Dispose()

			// // parepare env, block & transactions.
			// tx = mockNormalTransaction(tt.to, "n1TV3sU6jyzR4rJ1D7jCAmtVGSntJagXZHC", "0")
			// ctx, err = NewContext(mockBlock(), tx, contract, context)

			// // call transferFrom.
			// engine = NewV8Engine(ctx)
			// engine.SetExecutionLimits(10000, 100000000)
			// transferFromArgs := fmt.Sprintf("[\"%s\", \"%s\", \"%s\"]", tt.from, tt.to, tt.tokenID)
			// result, err = engine.Call(string(data), tt.sourceType, "transferFrom", transferFromArgs)
			// assert.Nil(t, err)
			// assert.Equal(t, "\"\"", result)
			// engine.Dispose()

		})
	}
}
