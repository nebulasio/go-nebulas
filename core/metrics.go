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
	metrics "github.com/nebulasio/go-nebulas/metrics"
)

// Metrics for core
var (
	// block metrics
	metricsBlockHeightGauge      = metrics.NewGauge("neb.block.height")
	metricsBlocktailHashGauge    = metrics.NewGauge("neb.block.tailhash")
	metricsBlockRevertTimesGauge = metrics.NewGauge("neb.block.revertcount")
	metricsBlockRevertMeter      = metrics.NewMeter("neb.block.revert")
	metricsBlockOnchainTimer     = metrics.NewTimer("neb.block.onchain")
	metricsTxOnchainTimer        = metrics.NewTimer("neb.transaction.onchain")
	metricsBlockPackTxTime       = metrics.NewGauge("neb.block.packtx")

	// block_pool metrics
	metricsCachedNewBlock      = metrics.NewGauge("neb.block.new.cached")
	metricsCachedDownloadBlock = metrics.NewGauge("neb.block.download.cached")
	metricsDuplicatedBlock     = metrics.NewCounter("neb.block.duplicated")
	metricsInvalidBlock        = metrics.NewCounter("neb.block.invalid")
	metricsTxsInBlock          = metrics.NewGauge("neb.block.txs")
	metricsBlockVerifiedTime   = metrics.NewGauge("neb.block.executed")
	metricsTxVerifiedTime      = metrics.NewGauge("neb.tx.executed")
	metricsTxPackedCount       = metrics.NewGauge("neb.tx.packed")
	metricsTxUnpackedCount     = metrics.NewGauge("neb.tx.unpacked")
	metricsTxGivebackCount     = metrics.NewGauge("neb.tx.giveback")

	// txpool metrics
	metricsCachedTx            = metrics.NewGauge("neb.txpool.cached")
	metricsInvalidTx           = metrics.NewCounter("neb.txpool.invalid")
	metricsDuplicateTx         = metrics.NewCounter("neb.txpool.duplicate")
	metricsTxPoolBelowGasPrice = metrics.NewCounter("neb.txpool.below_gas_price")
	metricsTxPoolOutOfGasLimit = metrics.NewCounter("neb.txpool.out_of_gas_limit")

	// transaction metrics
	metricsTxSubmit     = metrics.NewMeter("neb.transaction.submit")
	metricsTxExecute    = metrics.NewMeter("neb.transaction.execute")
	metricsTxExeSuccess = metrics.NewMeter("neb.transaction.execute.success")
	metricsTxExeFailed  = metrics.NewMeter("neb.transaction.execute.failed")

	// event metrics
	metricsCachedEvent = metrics.NewGauge("neb.event.cached")
)
