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
// but WITHOUT ANY WARRANTY; witho
// ut even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package rpc

import (
	metrics "github.com/rcrowley/go-metrics"
)

// Metrics for rpc
var (
	metricsRPCCounter = metrics.GetOrRegisterMeter("rpc.request", nil)

	metricsAccountStateSuccess = metrics.GetOrRegisterMeter("rpc.account.success", nil)
	metricsAccountStateFailed  = metrics.GetOrRegisterMeter("rpc.account.failed", nil)

	metricsSendTxSuccess = metrics.GetOrRegisterMeter("rpc.sendTx.success", nil)
	metricsSendTxFailed  = metrics.GetOrRegisterMeter("rpc.sendTx.failed", nil)

	metricsSendRawTxSuccess = metrics.GetOrRegisterMeter("rpc.sendRawTx.success", nil)
	metricsSendRawTxFailed  = metrics.GetOrRegisterMeter("rpc.sendRawTx.failed", nil)

	metricsSignTxSuccess = metrics.GetOrRegisterMeter("rpc.signTx.success", nil)
	metricsSignTxFailed  = metrics.GetOrRegisterMeter("rpc.signTx.failed", nil)

	metricsUnlockSuccess = metrics.GetOrRegisterMeter("rpc.unlock.success", nil)
	metricsUnlockFailed  = metrics.GetOrRegisterMeter("rpc.unlock.failed", nil)
)
