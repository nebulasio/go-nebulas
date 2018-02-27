// Copyright (C) 2018 go-nebulas authors
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

package net

import (
	"fmt"

	metrics "github.com/nebulasio/go-nebulas/metrics"
)

// Metrics map for different in/out network msg types
var (
	metricsPacketsIn = metrics.NewMeter("neb.net.packets.in")
	metricsBytesIn   = metrics.NewMeter("neb.net.bytes.in")

	metricsPacketsOut = metrics.NewMeter("neb.net.packets.out")
	metricsBytesOut   = metrics.NewMeter("neb.net.bytes.out")
)

func metricsPacketsInByMessageName(messageName string, size uint64) {
	meter := metrics.NewMeter(fmt.Sprintf("neb.net.packets.in.%s", messageName))
	meter.Mark(1)

	meter = metrics.NewMeter(fmt.Sprintf("neb.net.bytes.in.%s", messageName))
	meter.Mark(int64(size))
}

func metricsPacketsOutByMessageName(messageName string, size uint64) {
	meter := metrics.NewMeter(fmt.Sprintf("neb.net.packets.out.%s", messageName))
	meter.Mark(1)

	meter = metrics.NewMeter(fmt.Sprintf("neb.net.bytes.out.%s", messageName))
	meter.Mark(int64(size))
}
