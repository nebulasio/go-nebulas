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
	metricsRPCCounter = metrics.GetOrRegisterCounter("rpc_request", nil)

	metricsAccountStateSuccess = metrics.GetOrRegisterCounter("rpc_account_state_success", nil)
	metricsAccountStateFailed  = metrics.GetOrRegisterCounter("rpc_account_state_failed", nil)

	metricsSendTxSuccess = metrics.GetOrRegisterCounter("rpc_sendTx_success", nil)
	metricsSendTxFailed  = metrics.GetOrRegisterCounter("rpc_sendTx_failed", nil)

	metricsSendRawTxSuccess = metrics.GetOrRegisterCounter("rpc_sendRawTx_success", nil)
	metricsSendRawTxFailed  = metrics.GetOrRegisterCounter("rpc_sendRawTx_failed", nil)

	metricsSignTxSuccess = metrics.GetOrRegisterCounter("rpc_signTx_success", nil)
	metricsSignTxFailed  = metrics.GetOrRegisterCounter("rpc_signTx_failed", nil)

	metricsUnlockSuccess = metrics.GetOrRegisterCounter("rpc_unlock_success", nil)
	metricsUnlockFailed  = metrics.GetOrRegisterCounter("rpc_unlock_failed", nil)
)
