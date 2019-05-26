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

package nvm

import (
	"fmt"
	"strings"
	"sync"
	"errors"
	"time"
	"strconv"

	"encoding/json"
	"golang.org/x/net/context"

	//lru "github.com/hashicorp/golang-lru"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	ExecutionFailedErr  = 1
	ExecutionTimeOutErr = 2

	// ExecutionTimeout max v8 execution timeout.
	ExecutionTimeout                 = 15 * 1000 * 1000
	OriginExecutionTimeout           = 5 * 1000 * 1000
	CompatibleExecutionTimeout       = 20 * 1000 * 1000
	TimeoutGasLimitCost              = 100000000
	MaxLimitsOfExecutionInstructions = 10000000 // 10,000,000

	NVMDataExchangeTypeStart = "start"
	NVMDataExchangeTypeCallBack = "callback"
	NVMDataExchangeTypeFinal = "final"
	NVMDataExchangeTypeInnerContractCall = "innercall"

	NVM_SUCCESS = 0
	NVM_EXCEPTION_ERR = -1
	NVM_MEM_LIMIT_ERR = -2
	NVM_GAS_LIMIT_ERR = -3
	NVM_UNEXPECTED_ERR = -4
	NVM_EXE_TIMEOUT_ERR = -5
	NVM_TRANSPILE_SCRIPT_ERR = -6
	NVM_INJECT_TRACING_INSTRUCTION_ERR = -7
)

// callback function names
const (
	REQUIRE_DELEGATE_FUNC = "RequireDelegateFunc"
	ATTACH_LIB_VERSION_DELEGATE_FUNC = "AttachLibVersionDelegateFunc"
	STORAGE_GET = "StorageGet"
	STORAGE_PUT = "StoragePut"
	STORAGE_DEL = "StorageDel"
	GET_TX_BY_HASH = "GetTxByHash"
	GET_ACCOUNT_STATE = "GetAccountState"
	TRANSFER = "Transfer"
	VERIFY_ADDR = "VerifyAddress"
	GET_PRE_BLOCK_HASH = "GetPreBlockHash"
	GET_PRE_BLOCK_SEED = "GetPreBlockSeed"
	EVENT_TRIGGER_FUNC = "EventTriggerFunc"
	SHA_256_FUNC = "Sha256Func"
	SHA_3256_FUNC = "Sha3256Func"
	RIPEMD_160_FUNC = "Ripemd160Func"
	RECOVER_ADDRESS_FUNC = "RecoverAddressFunc"
	MD5_FUNC = "Md5Func"
	BASE64_FUNC = "Base64Func"
	// inner contract call
	GET_CONTRACT_SRC = "GetContractSource"
	INNER_CONTRACT_CALL = "InnerContractCall"
	// nr
	GET_LATEST_NR = "GetLatestNebulasRank"
	GET_LATEST_NR_SUMMARY = "GetLatestNebulasRankSummary"
)

//engine_v8 private data
var (
//	v8engineOnce         = sync.Once{}
	storages             = make(map[uint64]*V8Engine, 1024)
	storagesIdx          = uint64(0)
	storagesLock         = sync.RWMutex{}
//	engines              = make(map[*C.V8Engine]*V8Engine, 1024)
//	enginesLock          = sync.RWMutex{}
//	sourceModuleCache, _ = lru.New(40960)
	inject               = 0
	hit                  = 0
	nvmRequestIndex uint32	 = 1

	engines 			 = make([]*V8Engine, 0)
	StartSCTime = time.Now()
)

// V8Engine v8 engine.
type V8Engine struct {
	ctx                                     *Context
	//modules                                 Modules
	//v8engine                                *C.V8Engine
	strictDisallowUsageOfInstructionCounter int
	enableLimits                            bool
	limitsOfExecutionInstructions           uint64
	limitsOfTotalMemorySize                 uint64
	actualCountOfExecutionInstructions      uint64
	actualTotalMemorySize                   uint64
	lcsHandler                              uint64
	gcsHandler                              uint64
	serverListenAddr						string
	chainID 								uint32
	startExeTime							time.Time
	executionTimeOut						uint64
	innerErrMsg                             string
	innerErr                                error
	enableInnerContract						bool
}

type sourceModuleItem struct {
	source                    string
	sourceLineOffset          int
	traceableSource           string
	traceableSourceLineOffset int
}

func ResetRuntimeStatus(){
	storagesIdx = 0
	engines = nil
}

// NewV8Engine return new V8Engine instance.
func NewV8Engine(ctx *Context) *V8Engine {

	engine := &V8Engine{
		ctx:      ctx,
		//modules:  NewModules(),
		strictDisallowUsageOfInstructionCounter: 1, 	// enable by default.
		enableLimits:                            true,
		limitsOfExecutionInstructions:           0,
		limitsOfTotalMemorySize:                 0,
		actualCountOfExecutionInstructions:      0,
		actualTotalMemorySize:                   0,
		executionTimeOut:			  			 0,
		chainID:								 ctx.tx.ChainID(),		// default, set to be mainnet
	}

	(func() {
		storagesLock.Lock()
		defer storagesLock.Unlock()

		storagesIdx++
		engine.lcsHandler = storagesIdx
		storagesIdx++
		engine.gcsHandler = storagesIdx

		storages[engine.lcsHandler] = engine
		storages[engine.gcsHandler] = engine
	})()

	if core.NvmGasLimitWithoutTimeoutAtHeight(ctx.block.Height()) {
		engine.executionTimeOut = ExecutionTimeout
	} else {
		timeoutMark := core.NvmExeTimeoutAtHeight(ctx.block.Height())
		if timeoutMark {
			engine.executionTimeOut = OriginExecutionTimeout
		} else {
			engine.executionTimeOut = CompatibleExecutionTimeout
		}
	}

	if core.EnableInnerContractAtHeight(ctx.block.Height()) {
		engine.enableInnerContract = true;
	}
	return engine
}

// SetEnableLimit eval switch
func (e *V8Engine) SetEnableLimit(isLimit bool) {
	e.enableLimits = isLimit
}

// Dispose dispose all resources.
func (e *V8Engine) Dispose() {
	storagesLock.Lock()
	delete(storages, e.lcsHandler)
	delete(storages, e.gcsHandler)
	logging.CLog().WithFields(logrus.Fields{
			"lcsHandler": e.lcsHandler,
			"gcsHandler": e.gcsHandler,
		}).Error("------ Storages are deleted")
	storagesLock.Unlock()
}

// Context returns engine context
func (e *V8Engine) Context() *Context {
	return e.ctx
}

// SetExecutionLimits set execution limits of V8 Engine, prevent Halting Problem.
func (e *V8Engine) SetExecutionLimits(limitsOfExecutionInstructions, limitsOfTotalMemorySize uint64) error {
	logging.VLog().WithFields(logrus.Fields{
		"limits_of_executed_instructions": limitsOfExecutionInstructions,
		"limits_of_total_memory_size":     limitsOfTotalMemorySize,
	}).Debug("set execution limits.")

	e.limitsOfExecutionInstructions = limitsOfExecutionInstructions
	e.limitsOfTotalMemorySize = limitsOfTotalMemorySize

	if limitsOfExecutionInstructions == 0 || limitsOfTotalMemorySize == 0 {
		logging.VLog().Debugf("limit args has empty. limitsOfExecutionInstructions:%v,limitsOfTotalMemorySize:%d", limitsOfExecutionInstructions, limitsOfTotalMemorySize)
		return ErrLimitHasEmpty
	}
	// V8 needs at least 6M heap memory.
	if limitsOfTotalMemorySize < 6000000 {
		logging.VLog().Debugf("V8 needs at least 6M (6000000) heap memory, your limitsOfTotalMemorySize (%d) is too low.", limitsOfTotalMemorySize)
		return ErrSetMemorySmall
	}
	return nil
}

// ExecutionInstructions returns the execution instructions
func (e *V8Engine) ExecutionInstructions() uint64 {
	return e.actualCountOfExecutionInstructions
}

func (e *V8Engine) CheckTimeout() bool {
	elapsedTime := time.Since(e.startExeTime)

	if elapsedTime.Nanoseconds()/1000 > ExecutionTimeout{
		logging.CLog().WithFields(logrus.Fields{
				"elapsedTime": elapsedTime,
			}).Error("NVM execution timed out.")
		return true
	}

	return false
}

// Call function in a script
func (e *V8Engine) Call(config *core.NVMConfig) (string, error) {
	e.serverListenAddr = config.ListenAddr
	e.chainID = config.ChainID

	function := config.FunctionName
	if core.PublicFuncNameChecker.MatchString(function) == false {
		logging.VLog().Debugf("Invalid function: %v", function)
		return "", ErrDisallowCallNotStandardFunction
	}
	if strings.EqualFold(core.ContractInitFunc, function) == true {
		return "", ErrDisallowCallPrivateFunction
	}
	return e.RunScriptSource(config)
}

func (e *V8Engine) DeployAndInit(config *core.NVMConfig) (string, error){
	e.serverListenAddr = config.ListenAddr
	e.chainID = config.ChainID
	config.FunctionName = core.ContractInitFunc
	return e.RunScriptSource(config)
}

// GetNVMLeftResources return current NVM verb total resource
func (e *V8Engine) GetNVMLeftResources(engine *V8Engine, finalResponse *NVMFinalResponse) (uint64, uint64) {
	if finalResponse == nil {
		return engine.limitsOfExecutionInstructions, engine.limitsOfTotalMemorySize
	}

	statsBundle := finalResponse.StatsBundle
	instruction := uint64(0)
	mem := uint64(0)
	if engine.limitsOfExecutionInstructions >= statsBundle.ActualCountOfExecutionInstruction {
		instruction = engine.limitsOfExecutionInstructions - statsBundle.ActualCountOfExecutionInstruction
	}

	if engine.limitsOfTotalMemorySize >= statsBundle.ActualUsedMemSize {
		mem = engine.limitsOfTotalMemorySize - statsBundle.ActualUsedMemSize
	}
	return instruction, mem
}

func (e *V8Engine) GetCurrentV8Engine() *V8Engine{
	if len(engines) > 0{
		return engines[len(engines)-1]
	}
	return e
}

// Prepare V8 data exchange request, if it's inner contract call, then the innercall flag is true
func PrepareLaunchRequest(e *V8Engine, config *core.NVMConfig, innerCallFlag bool) (*NVMDataRequest, error) {

	// check source type
	sourceType := config.PayloadSourceType
	if sourceType != core.SourceTypeJavaScript && sourceType != core.SourceTypeTypeScript {
		return nil, ErrUnsupportedSourceType
	}

	// prepare for execute.
	block := toSerializableBlock(e.ctx.block)
	blockJSON, err := json.Marshal(block)
	if err != nil {
		return nil, errors.New("Failed to serialize block")
	}
	tx := toSerializableTransaction(e.ctx.tx)
	txJSON, err := json.Marshal(tx)
	if err != nil {
		return nil, errors.New("Failed to serialize transaction")
	}

	//var runnableSource string
	var argsInput []byte
	args := config.GetContractArgs()
	if len(args) > 0 {
		var argsObj []interface{}
		if err := json.Unmarshal([]byte(args), &argsObj); err != nil {
			return nil, errors.New("Arguments error")
		}
		if argsInput, err = json.Marshal(argsObj); err != nil {
			return nil, errors.New("Arguments error")
		}
	} else {
		argsInput = []byte("[]")
	}

	var moduleID string
	if innerCallFlag {
		engineIndx := fmt.Sprintf("%v", e.ctx.index);
		moduleID = "contract" + engineIndx + ".js"
	}else{
		moduleID = "contract.js"
	}
	
	var metaVersion string = "";
	if e.ctx.contract.ContractMeta() != nil {
		metaVersion = e.ctx.contract.ContractMeta().Version
		if len(metaVersion) == 0 {
			logging.VLog().WithFields(logrus.Fields{
				"height":  e.ctx.block.Height(),
			}).Error("contract deploy lib version is empty.")
			metaVersion = ""
		}
	}

	logging.CLog().WithFields(logrus.Fields{
		"metaversion": metaVersion,
	}).Error(">>>>> Meta VERSION!!!!!")

	runnableSource := fmt.Sprintf(`Blockchain.blockParse("%s");
		Blockchain.transactionParse("%s");
		var __contract = require("%s");
		var __instance = new __contract();
		__instance["%s"].apply(__instance, JSON.parse("%s"));`,
			formatArgs(string(blockJSON)), formatArgs(string(txJSON)),
			moduleID, config.FunctionName, formatArgs(string(argsInput)))

	logging.CLog().WithFields(logrus.Fields{
		"runnable": runnableSource,
	}).Info("------->>>>>> Runnable Source")

	if core.NvmGasLimitWithoutTimeoutAtHeight(e.ctx.block.Height()) {
		if e.limitsOfExecutionInstructions > MaxLimitsOfExecutionInstructions {
			e.limitsOfExecutionInstructions = MaxLimitsOfExecutionInstructions
		}
	}

	sourceHash := byteutils.Hex(hash.Sha3256([]byte(config.PayloadSource)))
	configBundle := &NVMConfigBundle{
		ScriptSrc:config.PayloadSource, 
		ScriptType:config.PayloadSourceType, 
		ScriptHash: sourceHash, 
		EnableLimits: true, 
		RunnableSrc: runnableSource, 
		MaxLimitsOfExecutionInstruction:MaxLimitsOfExecutionInstructions, 
		DefaultLimitsOfTotalMemSize:core.DefaultLimitsOfTotalMemorySize,
		LimitsExeInstruction: e.limitsOfExecutionInstructions, 
		LimitsTotalMemSize: e.limitsOfTotalMemorySize, 
		ExecutionTimeout: e.executionTimeOut,
		BlockJson:formatArgs(string(blockJSON)), 
		TxJson: formatArgs(string(txJSON)), 
		ModuleId: moduleID, 
		ChainId: e.chainID, 
		BlockHeight: e.ctx.block.Height(), 
		MetaVersion: metaVersion}

	callbackResult := &NVMCallbackResult{Result:""}

	// for call request, the metadata is nil
	requestType := NVMDataExchangeTypeStart
	if innerCallFlag {
		requestType = NVMDataExchangeTypeInnerContractCall
	}
	request := &NVMDataRequest{
		RequestType: requestType, 
		RequestIndx: nvmRequestIndex,
		ConfigBundle: configBundle,
		LcsHandler: e.lcsHandler, 
		GcsHandler: e.gcsHandler,
		CallbackResult: callbackResult}
	
	return request, nil
}

func (e *V8Engine) RunScriptSource(config *core.NVMConfig) (string, error){

	conn, err := grpc.Dial(e.serverListenAddr, grpc.WithInsecure())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to connect with V8 server")

		//TODO: try to re-launch the process		
	}
	defer conn.Close()

	logging.CLog().Info("NVM client is trying to connect the server")
	
	v8Client := NewNVMServiceClient(conn)
	var timeOut time.Duration = 15000   // Set execution timeout to be 15s'
	ctx, cancel := context.WithTimeout(context.Background(), timeOut*time.Second)
	defer cancel()
	
	logging.CLog().Info(">>>>>>>>>Now started call request!, the listener address is: ", e.serverListenAddr)


	// For debugging
	StartSCTime = time.Now()

	engines = append(engines, e)
	request, err := PrepareLaunchRequest(e, config, false)
	if err != nil {
		return "", err
	}

	stream, err := v8Client.SmartContractCall(ctx); if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
			"module": "nvm",
		}).Error("Failed to get streaming object")
		return "", ErrRPCConnection
	}

	logging.CLog().Info(">>>>>>>>>Before sending request to V8 process")

	err = stream.Send(request); if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
			"module": "nvm",
		}).Error("Failed to send out initial request")
		return "", ErrRPCConnection
	}

	// start counting for execution
	e.startExeTime = time.Now()
	for {
		dataResponse, err := stream.Recv()
		if err != nil{
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
				"module": "nvm",
			}).Error("Failed to receive data response from server")
			return "", ErrRPCConnection
		}

		// check the result here
		logging.CLog().Error(">>>>GOLANG received response from C++!!!!")

		if(dataResponse.GetResponseType() == NVMDataExchangeTypeFinal && dataResponse.GetFinalResponse() != nil){

			stream.CloseSend()
			finalResponse := dataResponse.GetFinalResponse()
			ret := finalResponse.Result
			result := finalResponse.Msg
			stats := finalResponse.StatsBundle
			notNil := finalResponse.NotNull

			elapsed := time.Since(StartSCTime)
			logging.CLog().WithFields(logrus.Fields{
				"Elapsed Time": elapsed.Nanoseconds()/1000,
			}).Error("%%%%%%%%%% SC Exe time!!!!!!!!")

			// check the result here
			logging.CLog().WithFields(
				logrus.Fields{
					"result": result,                                                                                                                                                                                                                                                                                                     
					"ret": ret,
				}).Info(">>>>The contract execution result")
		
			// TODO, collect tracing stats
			//e.CollectTracingStats()
			actualCountOfExecutionInstructions := stats.ActualCountOfExecutionInstruction
			actualUsedMemSize := stats.ActualUsedMemSize
			e.actualCountOfExecutionInstructions = actualCountOfExecutionInstructions
			e.actualTotalMemorySize = actualUsedMemSize

			logging.CLog().WithFields(logrus.Fields{
				"actualAcountOfExecutionInstructions": actualCountOfExecutionInstructions,
				"actualUsedMemSize": actualUsedMemSize,
				"finalresult": ret,
				"tx hash": e.ctx.tx.Hash(),
				"tx height": e.ctx.block.Height(),
			}).Info(">>>>Got stats info")
		
			if core.NvmGasLimitWithoutTimeoutAtHeight(e.ctx.block.Height()) {
				if e.limitsOfExecutionInstructions == MaxLimitsOfExecutionInstructions && err == ErrInsufficientGas {
				  err = ErrExecutionTimeout
				  result = "\"null\""
				}
			}


			//set err
			if ret == NVM_TRANSPILE_SCRIPT_ERR {
				err = ErrTranspileTypeScriptFailed

			} else if ret == NVM_INJECT_TRACING_INSTRUCTION_ERR {
				err = ErrInjectTracingInstructionFailed

			} else if ret == NVM_EXE_TIMEOUT_ERR {
				err = ErrExecutionTimeout
				if core.NvmGasLimitWithoutTimeoutAtHeight(e.ctx.block.Height()) {
					err = core.ErrUnexpected
				} else if core.NewNvmExeTimeoutConsumeGasAtHeight(e.ctx.block.Height()) {
					if TimeoutGasLimitCost > e.limitsOfExecutionInstructions {
						e.actualCountOfExecutionInstructions = e.limitsOfExecutionInstructions
						//actualCountOfExecutionInstructions = e.limitsOfExecutionInstructions

					} else {
						e.actualCountOfExecutionInstructions = TimeoutGasLimitCost
						//actualCountOfExecutionInstructions = TimeoutGasLimitCost
					}
				}
			} else if ret == NVM_UNEXPECTED_ERR {
				err = core.ErrUnexpected

			} else {
				if ret != NVM_SUCCESS {
					err = core.ErrExecutionFailed
				}

				if e.limitsOfExecutionInstructions > 0 &&
					e.limitsOfExecutionInstructions < e.actualCountOfExecutionInstructions {
					// Reach instruction limits.
					err = ErrInsufficientGas
					e.actualCountOfExecutionInstructions = e.limitsOfExecutionInstructions
				} else if e.limitsOfTotalMemorySize > 0 && e.limitsOfTotalMemorySize < e.actualTotalMemorySize {
					// reach memory limits.
					err = ErrExceedMemoryLimits
					e.actualCountOfExecutionInstructions = e.limitsOfExecutionInstructions
				}
			}

			//set result			
			if !notNil && ret == NVM_SUCCESS {
				result = "\"\"" // default JSON String.
			}
			
			logging.CLog().WithFields(logrus.Fields{
				"result": result,
			}).Info(">>>>>>> The contract execution result!")

			return result, err

		}else{
			serverLcsHandler := dataResponse.GetLcsHandler()
			serverGcsHandler := dataResponse.GetGcsHandler()
			callbackResponse := dataResponse.GetCallbackResponse()
			responseFuncName := callbackResponse.GetFuncName()
			responseFuncParams := callbackResponse.GetFuncParams()

			var newRequest *NVMDataRequest

			if responseFuncName == INNER_CONTRACT_CALL {
				logging.CLog().WithFields(logrus.Fields{
					"responseFuncName": responseFuncName,
					"serverLcsHandler": serverLcsHandler,
					"param0": responseFuncParams[0],
					"param1": responseFuncParams[1],
					"param2": responseFuncParams[2],
					"param3": responseFuncParams[3],
				}).Error("INNER CONTRACT CALL callback on Golang side!")

				// check remaining mem and instructions
				currEngine := e.GetCurrentV8Engine()  			// The new engine's parent engine
				remainCountOfInstructions, remainMemSize := e.GetNVMLeftResources(currEngine, dataResponse.GetFinalResponse())
				engineNew, nvmConfigNew, gasCnt, innerCallErr := InnerContractFunc(
											serverLcsHandler, 
											responseFuncParams[0],
											responseFuncParams[1],
											responseFuncParams[2],
											responseFuncParams[3],
											remainCountOfInstructions,
											remainMemSize)

				nvmRequestIndex+=1
				if innerCallErr != nil {
					logging.CLog().Error("Failed to create inner contract engine")
					innerCallbackResult := &NVMCallbackResult{Result:""}
					innerCallbackResult.NotNull = false
					innerCallbackResult.FuncName = responseFuncName
					innerCallbackResult.Extra = append(innerCallbackResult.Extra, fmt.Sprintf("%v", gasCnt))
					// for call request, the metadata is nil
					requestType := NVMDataExchangeTypeInnerContractCall
					newRequest = &NVMDataRequest{
						RequestType: requestType,
						LcsHandler: serverLcsHandler, 
						GcsHandler: serverGcsHandler,
						RequestIndx: nvmRequestIndex,
						CallbackResult: innerCallbackResult}

				}else{
					engines = append(engines, engineNew)
					logging.CLog().Info("Successfully created inner contract engine")
					newRequest, err = PrepareLaunchRequest(engineNew, nvmConfigNew, true)
					newRequest.CallbackResult.Result = ""
					newRequest.LcsHandler = engineNew.lcsHandler
					newRequest.GcsHandler = engineNew.gcsHandler
					newRequest.CallbackResult.NotNull = true
					newRequest.CallbackResult.FuncName = responseFuncName
					newRequest.CallbackResult.Extra = append(newRequest.CallbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				}

			} else {
				callbackResult := &NVMCallbackResult{}

				switch responseFuncName{
				case ATTACH_LIB_VERSION_DELEGATE_FUNC:
					pathName := AttachLibVersionDelegateFunc(serverLcsHandler, responseFuncParams[0])
					callbackResult.Result = pathName
				case STORAGE_GET:
					value, gasCnt, notNil := StorageGetFunc(serverLcsHandler, responseFuncParams[0])
					callbackResult.Result = value
					callbackResult.NotNull = notNil
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				case STORAGE_PUT:
					resCode, gasCnt := StoragePutFunc(serverLcsHandler, responseFuncParams[0], responseFuncParams[1])
					callbackResult.Result = fmt.Sprintf("%v", resCode)
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				case STORAGE_DEL:
					resCode, gasCnt := StorageDelFunc(serverLcsHandler, responseFuncParams[0])
					callbackResult.Result = fmt.Sprintf("%v", resCode)
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				case GET_TX_BY_HASH:
					resStr, gasCnt, notNil := GetTxByHashFunc(serverLcsHandler, responseFuncParams[0])
					callbackResult.Result = resStr
					callbackResult.NotNull = notNil
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				case GET_ACCOUNT_STATE:
					resCode, resStr, exceptionInfo, gasCnt, notNil := GetAccountStateFunc(serverLcsHandler, responseFuncParams[0])
					callbackResult.Result = fmt.Sprintf("%v", resCode)
					callbackResult.NotNull = notNil
					callbackResult.Extra = append(callbackResult.Extra, resStr)
					callbackResult.Extra = append(callbackResult.Extra, exceptionInfo)
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				case TRANSFER:
					resCode, gasCnt := TransferFunc(serverLcsHandler, responseFuncParams[0], responseFuncParams[1])
					callbackResult.Result = fmt.Sprintf("%v", resCode)
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				case VERIFY_ADDR:
					resCode, gasCnt := VerifyAddressFunc(serverLcsHandler, responseFuncParams[0])
					callbackResult.Result = fmt.Sprintf("%v", resCode)
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				case GET_PRE_BLOCK_HASH:
					offset, _ := strconv.ParseInt(responseFuncParams[0], 10, 64)
					resCode, resStr, exceptionInfo, gasCnt, notNil := GetPreBlockHashFunc(serverLcsHandler, uint64(offset))
					callbackResult.Result = fmt.Sprintf("%v", resCode)
					callbackResult.NotNull = notNil
					callbackResult.Extra = append(callbackResult.Extra, resStr)
					callbackResult.Extra = append(callbackResult.Extra, exceptionInfo)
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				case GET_PRE_BLOCK_SEED:
					offset, _ := strconv.ParseInt(responseFuncParams[0], 10, 64)
					resCode, resStr, exceptionInfo, gasCnt, notNil := GetPreBlockSeedFunc(serverLcsHandler, uint64(offset))
					callbackResult.Result = fmt.Sprintf("%v", resCode)
					callbackResult.NotNull = notNil
					callbackResult.Extra = append(callbackResult.Extra, resStr)
					callbackResult.Extra = append(callbackResult.Extra, exceptionInfo)
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				case EVENT_TRIGGER_FUNC:
					gasCnt := EventTriggerFunc(serverLcsHandler, responseFuncParams[0], responseFuncParams[1])
					callbackResult.Result = fmt.Sprintf("%v", gasCnt)
				case SHA_256_FUNC:
					resStr, gasCnt, notNil := Sha256Func(responseFuncParams[0])
					callbackResult.Result = resStr
					callbackResult.NotNull = notNil
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				case SHA_3256_FUNC:
					resStr, gasCnt, notNil := Sha3256Func(responseFuncParams[0])
					callbackResult.Result = resStr
					callbackResult.NotNull = notNil
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				case RIPEMD_160_FUNC:
					resStr, gasCnt, notNil := Ripemd160Func(responseFuncParams[0])
					callbackResult.Result = resStr
					callbackResult.NotNull = notNil
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				case RECOVER_ADDRESS_FUNC:
					alg, _ := strconv.Atoi(responseFuncParams[0])
					resStr, gasCnt, notNil := RecoverAddressFunc(alg, responseFuncParams[1], responseFuncParams[2])
					callbackResult.Result = resStr
					callbackResult.NotNull = notNil
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				case MD5_FUNC:
					resStr, gasCnt, notNil := Md5Func(responseFuncParams[0])
					callbackResult.Result = resStr
					callbackResult.NotNull = notNil
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				case BASE64_FUNC:
					resStr, gasCnt, notNil := Base64Func(responseFuncParams[0])
					callbackResult.Result = resStr
					callbackResult.NotNull = notNil
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				case GET_CONTRACT_SRC:
					resStr, gasCnt, notNil := GetContractSourceFunc(serverLcsHandler, responseFuncParams[0])
					callbackResult.Result = resStr
					callbackResult.NotNull = notNil
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				case GET_LATEST_NR:
					resCode, resStr, exceptionInfo, gasCnt, notNil := GetLatestNebulasRankFunc(serverLcsHandler, responseFuncParams[0])
					callbackResult.Result = fmt.Sprintf("%v", resCode)
					callbackResult.NotNull = notNil
					callbackResult.Extra = append(callbackResult.Extra, resStr)
					callbackResult.Extra = append(callbackResult.Extra, exceptionInfo)
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				case GET_LATEST_NR_SUMMARY:
					resCode, resStr, exceptionInfo, gasCnt, notNil := GetLatestNebulasRankSummaryFunc(serverLcsHandler)
					callbackResult.Result = fmt.Sprintf("%v", resCode)
					callbackResult.NotNull = notNil
					callbackResult.Extra = append(callbackResult.Extra, resStr)
					callbackResult.Extra = append(callbackResult.Extra, exceptionInfo)
					callbackResult.Extra = append(callbackResult.Extra, fmt.Sprintf("%v", gasCnt))
				default:
					logging.CLog().WithFields(logrus.Fields{
						"func": responseFuncName,
						"params": responseFuncParams,
					}).Error("Invalid callback function name")
				}

				// stream.Send()
				nvmRequestIndex += 1
				// check the callback type
				callbackResult.FuncName = responseFuncName
				newRequest = &NVMDataRequest{
					RequestType:NVMDataExchangeTypeCallBack, 
					RequestIndx:dataResponse.GetResponseIndx(),
					LcsHandler:serverLcsHandler,
					GcsHandler:serverGcsHandler,
					CallbackResult:callbackResult,
				}
			}

			stream.Send(newRequest)
		}
		
		if(e.CheckTimeout()){
			stream.CloseSend()
			return "", ErrExecutionTimeout
		}
	}

	return "", ErrUnexpected
}

func getEngineByStorageHandler(handler uint64) (*V8Engine, Account) {
	storagesLock.RLock()
	engine := storages[handler]
	storagesLock.RUnlock()

	if engine == nil {
		logging.VLog().WithFields(logrus.Fields{
			"wantedHandler": handler,
		}).Error("wantedHandler is not found.")
		return nil, nil
	}

	if engine.lcsHandler == handler {
		return engine, engine.ctx.contract
	} else if engine.gcsHandler == handler {
		// disable gcs according to issue https://github.com/nebulasio/go-nebulas/issues/23.
		return nil, nil
		// return engine, engine.ctx.owner
	} else {
		logging.VLog().WithFields(logrus.Fields{
			"lcsHandler":    engine.lcsHandler,
			"gcsHandler":    engine.gcsHandler,
			"wantedHandler": handler,
		}).Error("in-consistent storage handler.")
		return nil, nil
	}
}

// Still use the storage maps to get the v8 engine
func getEngineByEngineHandler(handler uint64) *V8Engine {
	storagesLock.RLock()
	defer storagesLock.RUnlock()

	engine := storages[handler]
	if engine == nil {
		logging.VLog().WithFields(logrus.Fields{
			"wantedHandler": handler,
		}).Error("wantedHandler is not found.")
		return nil
	}

	// only use the lcs handler to check
	if engine.lcsHandler == handler {
		return engine
	} else {
		return nil
	}
}

func formatArgs(s string) string {
	s = strings.Replace(s, "\\", "\\\\", -1)
	s = strings.Replace(s, "\n", "\\n", -1)
	s = strings.Replace(s, "\r", "\\r", -1)
	s = strings.Replace(s, "\"", "\\\"", -1)
	return s
}