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
		createFile(filename)
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

func createFile(filename string) {
	content := `
	  p2p {
		port: 51413
		chain_id: 100
		version: 1
	  }
	  rpc {
		api_port: 51510
		management_port: 52520
		api_http_port: 8090
		management_http_port: 8191
	  }
	  pow {
		coinbase: "8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf"
	  }
	  
	  storage {
		location: "seed.db"
	  }
	  
	  account {
		# keystore.SECP256K1 = 1
		signature: 1
	  
		# keystore.SCRYPT = 1 << 4
		encrypt: 16
	  
		key_dir: "keydir"
	  
		test_passphrase: "passphrase"
	  }
	  
	  influxdb {
		host: "http://localhost:8086"
		db: "nebulas"
		username: "admin"
		password: "admin"
	  }
	  
	  metrics {
		enable: false
	  }`

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
