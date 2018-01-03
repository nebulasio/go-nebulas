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

package neblet

import (
	"io/ioutil"
	"os"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/util/logging"
)

// LoadConfig loads configuration from the file.
func LoadConfig(file string) *nebletpb.Config {
	//logging.VLog().Info("Loading Neb config from file ", file)

	var content string
	if len(file) > 0 {
		if !pathExist(file) {
			CreateDefaultConfigFile(file)
		}
		b, err := ioutil.ReadFile(file)
		if err != nil {
			logging.VLog().Fatal(err)
		}

		content = string(b)
	} else {
		content = defaultConfig()
	}
	//logging.VLog().Info("Parsing Neb config text ", content)

	pb := new(nebletpb.Config)
	if err := proto.UnmarshalText(content, pb); err != nil {
		logging.VLog().Fatal(err)
	}
	//logging.VLog().Info("Loaded Neb config proto ", pb)
	return pb
}

func defaultConfig() string {
	content := `
	network {
		listen: ["127.0.0.1:51413"]
	}

	chain {
		chain_id: 100
		datadir: "seed.db"
		keydir: "keydir"
		coinbase: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8"
		signature_ciphers: ["ECC_SECP256K1"]
	}

	rpc {
		rpc_listen: ["127.0.0.1:51510"]
		http_listen: ["127.0.0.1:8090"]
		http_module: ["api","admin"]
	}

  app {
    enable_crash_report: true
  }

	stats {
		enable_metrics: false
		influxdb: {
			host: "http://localhost:8086"
			db: "nebulas"
			user: "admin"
			password: "admin"
		}
	}
	`
	return content
}

// CreateDefaultConfigFile create a default config file.
func CreateDefaultConfigFile(filename string) {
	if err := ioutil.WriteFile(filename, []byte(defaultConfig()), 0644); err != nil {
		logging.VLog().Fatal(err)
	}
}

func pathExist(_path string) bool {
	_, err := os.Stat(_path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
