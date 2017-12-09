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
	log "github.com/sirupsen/logrus"
)

// LoadConfig loads configuration from the file.
func LoadConfig(filename string) *nebletpb.Config {
	//log.Info("Loading Neb config from file ", filename)
	if !pathExist(filename) {
		CreateDefaultConfigFile(filename)
	}
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	str := string(b)
	//log.Info("Parsing Neb config text ", str)

	pb := new(nebletpb.Config)
	if err := proto.UnmarshalText(str, pb); err != nil {
		log.Fatal(err)
	}
	//log.Info("Loaded Neb config proto ", pb)
	return pb
}

// CreateDefaultConfigFile create a default config file.
func CreateDefaultConfigFile(filename string) {
	content := `
	network {
		listen: ["127.0.0.1:51413"]
	}
	
	chain {
		chain_id: 100
		datadir: "seed.db"
		keydir: "keydir"
		coinbase: "000000000000000000000000000000000000000000000000"
		signature_ciphers: [0]
	}
	
	rpc {
			rpc_listen: ["127.0.0.1:51510"]
			http_listen: ["127.0.0.1:8090"]
			http_module: [0,1]
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

	if err := ioutil.WriteFile(filename, []byte(content), 0644); err != nil {
		log.Fatal(err)
	}
}

func pathExist(_path string) bool {
	_, err := os.Stat(_path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
