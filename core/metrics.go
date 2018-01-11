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
	metrics "github.com/rcrowley/go-metrics"
)

// Metrics for core
var (
	// block_pool metrics
	metricsDuplicatedBlock    = metrics.GetOrRegisterCounter("neb.block.duplicated", nil)
	metricsInvalidBlock       = metrics.GetOrRegisterCounter("neb.block.invalid", nil)
	metricsBlockExecutedTimer = metrics.GetOrRegisterTimer("neb.block.executed", nil)
	metricsTxExecutedTimer    = metrics.GetOrRegisterTimer("neb.tx.executed", nil)

	// txpool metrics
	metricsInvalidTx           = metrics.GetOrRegisterCounter("txpool_invalid", nil)
	metricsDuplicateTx         = metrics.GetOrRegisterCounter("txpool_duplicate", nil)
	metricsTxPoolBelowGasPrice = metrics.GetOrRegisterCounter("txpool_below_gas_price", nil)
	metricsTxPoolOutOfGasLimit = metrics.GetOrRegisterCounter("txpool_out_of_gas_limit", nil)

	// transaction metrics
	metricsTxSubmit     = metrics.GetOrRegisterMeter("tx_submit", nil)
	metricsTxExecute    = metrics.GetOrRegisterMeter("tx_execute", nil)
	metricsTxExeSuccess = metrics.GetOrRegisterMeter("tx_execute_success", nil)
	metricsTxExeFailed  = metrics.GetOrRegisterMeter("tx_execute_failed", nil)
)
