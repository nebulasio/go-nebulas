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

package p2p

import (
	"bytes"
	"errors"
	"hash/crc32"

	byteutils "github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

var (
	offsetChainID        = 4
	offsetReserved       = 8
	offsetVersion        = 11
	offsetMessageName    = 12
	offsetDataLength     = 24
	offsetDataCheckSum   = 28
	offsetHeaderCheckSum = 32
	offsetData           = 36
)

// Error types
var (
	ErrInvalidNebMessageHeader = errors.New("invalid neb message header")
	ErrInvalidNebMessageData   = errors.New("invalid neb message data")
)

/*
NebMessage defines protocol in Nebulas, we define our own wire protocol, as the following:

 0               1               2               3              (bytes)
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                         Magic Number                          |
+---------------------------------------------------------------+
|                         Chain ID                              |
+-----------------------------------------------+---------------+
|                         Reserved              |   Version     |
+-----------------------------------------------+---------------+
|                                                               |
+                                                               +
|                         Message Name                          |
+                                                               +
|                                                               |
+---------------------------------------------------------------+
|                         Data Length                           |
+---------------------------------------------------------------+
|                         Data Checksum                         |
+---------------------------------------------------------------+
|                         Header Checksum                       |
|---------------------------------------------------------------+
|                                                               |
+                         Data                                  +
.                                                               .
|                                                               |
+---------------------------------------------------------------+
*/
type NebMessage struct {
	magicNumber    []byte
	chainID        []byte
	version        byte
	msgName        string
	dataLength     []byte
	dataChecksum   []byte
	headerChecksum []byte
	header         []byte
	data           []byte
	reserved       []byte
}

// buildHeader build header information
func buildHeader(chainID uint32, msgName string, version byte, dataLength uint32, dataChecksum uint32, reserved []byte) []byte {

	var metaHeader = make([]byte, offsetHeaderCheckSum)
	msgNameByte := []byte(msgName)

	copy(metaHeader[:], MagicNumber)
	copy(metaHeader[offsetChainID:], byteutils.FromUint32(chainID))
	// 64-88 Reserved field
	copy(metaHeader[offsetReserved:], reserved)
	copy(metaHeader[offsetVersion:], []byte{version})
	copy(metaHeader[offsetMessageName:], msgNameByte)
	copy(metaHeader[offsetDataLength:], byteutils.FromUint32(dataLength))
	copy(metaHeader[offsetDataCheckSum:], byteutils.FromUint32(dataChecksum))

	return metaHeader
}

func (node *Node) BuildRawMessageData(data []byte, msgName string) []byte {
	dataChecksum := crc32.ChecksumIEEE(data)
	reserved := []byte{0}
	metaHeader := buildHeader(node.config.ChainID, msgName, node.config.Version, uint32(len(data)), dataChecksum, reserved)
	headerChecksum := crc32.ChecksumIEEE(metaHeader)
	metaHeader = append(metaHeader[:], byteutils.FromUint32(headerChecksum)...)
	totalData := append(metaHeader[:], data...)
	return totalData
}

func (node *Node) verifyHeader(nebMsg *NebMessage) bool {

	headerChecksum := crc32.ChecksumIEEE(nebMsg.header)

	if !byteutils.Equal(MagicNumber, nebMsg.magicNumber) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": string(MagicNumber),
			"actual": string(nebMsg.magicNumber),
		}).Error("invalid magic number")
		return false
	}

	if node.Config().ChainID != byteutils.Uint32(nebMsg.chainID) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": node.Config().ChainID,
			"actual": byteutils.Uint32(nebMsg.chainID),
		}).Error("invalid chainID")
		return false
	}

	if node.config.Version != nebMsg.version {
		logging.VLog().WithFields(logrus.Fields{
			"expect": node.config.Version,
			"actual": nebMsg.version,
		}).Error("invalid version")
		return false
	}

	if headerChecksum != byteutils.Uint32(nebMsg.headerChecksum) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": headerChecksum,
			"actual": byteutils.Uint32(nebMsg.headerChecksum),
		}).Error("invalid header checksum")
		return false
	}
	return true
}

func (node *Node) parseMsgHeader(streamBuffer []byte) (*NebMessage, error) {
	header := streamBuffer

	nebMsg := &NebMessage{}
	nebMsg.magicNumber = header[:offsetChainID]
	nebMsg.chainID = header[offsetChainID:offsetReserved]
	nebMsg.reserved = header[offsetReserved:offsetVersion]
	nebMsg.version = header[offsetVersion]
	msgName := header[offsetMessageName:offsetDataLength]
	nebMsg.dataLength = header[offsetDataLength:offsetDataCheckSum]
	nebMsg.dataChecksum = header[offsetDataCheckSum:offsetHeaderCheckSum]
	nebMsg.headerChecksum = header[offsetHeaderCheckSum:offsetData]
	nebMsg.header = header[:offsetHeaderCheckSum]

	index := bytes.IndexByte(msgName, 0)
	if index != -1 {
		msgNameByte := msgName[0:index]
		nebMsg.msgName = string(msgNameByte)
	} else {
		nebMsg.msgName = string(msgName)
	}

	if !node.verifyHeader(nebMsg) {
		return nil, ErrInvalidNebMessageHeader
	}

	logging.VLog().WithFields(logrus.Fields{
		"msgName":      nebMsg.msgName,
		"magicNumber":  string(nebMsg.magicNumber),
		"chainID":      byteutils.Uint32(nebMsg.chainID),
		"version":      nebMsg.version,
		"dataChecksum": byteutils.Uint32(nebMsg.dataChecksum),
		"dataLength":   byteutils.Uint32(nebMsg.dataLength),
	}).Info("parse neb message header success")
	return nebMsg, nil
}

func (node *Node) parseMsgData(nebMsg *NebMessage, streamBuffer []byte) error {

	dataLength := byteutils.Uint32(nebMsg.dataLength)
	nebMsg.data = streamBuffer[:dataLength]

	dataChecksum := crc32.ChecksumIEEE(nebMsg.data)
	if dataChecksum != byteutils.Uint32(nebMsg.dataChecksum) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": dataChecksum,
			"actual": byteutils.Uint32(nebMsg.dataChecksum),
		}).Error("invalid neb message data")
		return ErrInvalidNebMessageData
	}
	return nil
}
