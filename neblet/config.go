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

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	log "github.com/sirupsen/logrus"
)

// LoadConfig from neblet
func LoadConfig(filename string) *nebletpb.Config {
	log.Info("Loading Neb config from file ", filename)
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	str := string(b)
	log.Info("Parsing Neb config text ", str)

	pb := new(nebletpb.Config)
	if err := proto.UnmarshalText(str, pb); err != nil {
		log.Fatal(err)
	}
	log.Info("Loaded Neb config proto ", pb)
	return pb
}
