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

package metrics

import (
	"time"

	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net/p2p"
	metrics "github.com/rcrowley/go-metrics"
	influxdb "github.com/vrischmann/go-metrics-influxdb"
)

const (
	duration = 2 * time.Second
	tagName  = "nodeID"
)

// Neblet interface breaks cycle import dependency.
type Neblet interface {
	Config() nebletpb.Config
	NetService() *p2p.NetService
}

// Start metrics mornitor
func Start(neb Neblet) {
	tags := make(map[string]string)
	tags[tagName] = neb.NetService().Node().ID()
	influxdb.InfluxDBWithTags(metrics.DefaultRegistry, duration, neb.Config().Influxdb.Host, neb.Config().Influxdb.Db, neb.Config().Influxdb.Username, neb.Config().Influxdb.Password, tags)
}
