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
	"bytes"
	"errors"
	"hash/crc32"
	"time"

	byteutils "github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
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
// const
const (
	NebMessageMagicNumberEndIdx    = 4
	NebMessageChainIDEndIdx        = 8
	NebMessageReservedEndIdx       = 11
	NebMessageVersionIndex         = 11
	NebMessageVersionEndIdx        = 12
	NebMessageNameEndIdx           = 24
	NebMessageDataLengthEndIdx     = 28
	NebMessageDataCheckSumEndIdx   = 32
	NebMessageHeaderCheckSumEndIdx = 36
	NebMessageHeaderLength         = 36

	MaxNebMessageDataLength = 32 * 1024 * 1024 // 32m.
	MaxNebMessageNameLength = 24 - 12          // 12.
)

// Error types
var (
	MagicNumber     = []byte{0x4e, 0x45, 0x42, 0x31}
	DefaultReserved = []byte{0x0, 0x0, 0x0}

	ErrInsufficientMessageHeaderLength = errors.New("insufficient message header length")
	ErrInsufficientMessageDataLength   = errors.New("insufficient message data length")
	ErrInvalidMagicNumber              = errors.New("invalid magic number")
	ErrInvalidHeaderCheckSum           = errors.New("invalid header checksum")
	ErrInvalidDataCheckSum             = errors.New("invalid data checksum")
	ErrExceedMaxDataLength             = errors.New("exceed max data length")
	ErrExceedMaxMessageNameLength      = errors.New("exceed max message name length")
)

//NebMessage struct
type NebMessage struct {
	content     []byte
	messageName string

	// debug fields.
	sendMessageAt  int64
	writeMessageAt int64
}

// MagicNumber return magicNumber
func (message *NebMessage) MagicNumber() []byte {
	return message.content[0:NebMessageMagicNumberEndIdx]
}

// ChainID return chainID
func (message *NebMessage) ChainID() uint32 {
	return byteutils.Uint32(message.content[NebMessageMagicNumberEndIdx:NebMessageChainIDEndIdx])
}

// Reserved return reserved
func (message *NebMessage) Reserved() []byte {
	return message.content[NebMessageChainIDEndIdx:NebMessageReservedEndIdx]
}

// Version return version
func (message *NebMessage) Version() byte {
	return message.content[NebMessageVersionIndex]
}

// MessageName return message name
func (message *NebMessage) MessageName() string {
	if message.messageName == "" {
		data := message.content[NebMessageVersionEndIdx:NebMessageNameEndIdx]
		pos := bytes.IndexByte(data, 0)
		if pos != -1 {
			message.messageName = string(data[0:pos])
		} else {
			message.messageName = string(data)
		}
	}
	return message.messageName
}

// DataLength return dataLength
func (message *NebMessage) DataLength() uint32 {
	return byteutils.Uint32(message.content[NebMessageNameEndIdx:NebMessageDataLengthEndIdx])
}

// DataCheckSum return data checkSum
func (message *NebMessage) DataCheckSum() uint32 {
	return byteutils.Uint32(message.content[NebMessageDataLengthEndIdx:NebMessageDataCheckSumEndIdx])
}

// HeaderCheckSum return header checkSum
func (message *NebMessage) HeaderCheckSum() uint32 {
	return byteutils.Uint32(message.content[NebMessageDataCheckSumEndIdx:NebMessageHeaderCheckSumEndIdx])
}

// HeaderWithoutCheckSum return header without checkSum
func (message *NebMessage) HeaderWithoutCheckSum() []byte {
	return message.content[:NebMessageDataCheckSumEndIdx]
}

// Data return data
func (message *NebMessage) Data() []byte {
	return message.content[NebMessageHeaderLength:]
}

// Content return message content
func (message *NebMessage) Content() []byte {
	return message.content
}

// Length return message Length
func (message *NebMessage) Length() uint64 {
	return uint64(len(message.content))
}

// NewNebMessage new neb message
func NewNebMessage(chainID uint32, reserved []byte, version byte, messageName string, data []byte) (*NebMessage, error) {
	if len(data) > MaxNebMessageDataLength {
		logging.VLog().WithFields(logrus.Fields{
			"messageName": messageName,
			"dataLength":  len(data),
			"limits":      MaxNebMessageDataLength,
		}).Error("Exceeded max data length.")
		return nil, ErrExceedMaxDataLength
	}

	if len(messageName) > MaxNebMessageNameLength {
		logging.VLog().WithFields(logrus.Fields{
			"messageName":      messageName,
			"len(messageName)": len(messageName),
			"limits":           MaxNebMessageNameLength,
		}).Error("Exceeded max message name length.")
		return nil, ErrExceedMaxMessageNameLength

	}

	dataCheckSum := crc32.ChecksumIEEE(data)

	message := &NebMessage{
		content: make([]byte, NebMessageHeaderLength+len(data)),
	}

	// copy fields.
	copy(message.content[0:NebMessageMagicNumberEndIdx], MagicNumber)
	copy(message.content[NebMessageMagicNumberEndIdx:NebMessageChainIDEndIdx], byteutils.FromUint32(chainID))
	copy(message.content[NebMessageChainIDEndIdx:NebMessageReservedEndIdx], reserved)
	message.content[NebMessageVersionIndex] = version
	copy(message.content[NebMessageVersionEndIdx:NebMessageNameEndIdx], []byte(messageName))
	copy(message.content[NebMessageNameEndIdx:NebMessageDataLengthEndIdx], byteutils.FromUint32(uint32(len(data))))
	copy(message.content[NebMessageDataLengthEndIdx:NebMessageDataCheckSumEndIdx], byteutils.FromUint32(dataCheckSum))

	// header checksum.
	headerCheckSum := crc32.ChecksumIEEE(message.HeaderWithoutCheckSum())
	copy(message.content[NebMessageDataCheckSumEndIdx:NebMessageHeaderCheckSumEndIdx], byteutils.FromUint32(headerCheckSum))

	// copy data.
	copy(message.content[NebMessageHeaderCheckSumEndIdx:], data)

	return message, nil
}

// ParseNebMessage parse neb message
func ParseNebMessage(data []byte) (*NebMessage, error) {
	if len(data) < NebMessageHeaderLength {
		return nil, ErrInsufficientMessageHeaderLength
	}

	message := &NebMessage{
		content: make([]byte, NebMessageHeaderLength),
	}
	copy(message.content, data)

	if err := message.VerifyHeader(); err != nil {
		return nil, err
	}

	return message, nil
}

// ParseMessageData parse neb message data
func (message *NebMessage) ParseMessageData(data []byte) error {
	if uint32(len(data)) < message.DataLength() {
		return ErrInsufficientMessageDataLength
	}

	message.content = append(message.content, data[:message.DataLength()]...)
	return message.VerifyData()
}

// VerifyHeader verify message header
func (message *NebMessage) VerifyHeader() error {
	if !byteutils.Equal(MagicNumber, message.MagicNumber()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": MagicNumber,
			"actual": message.MagicNumber(),
			"err":    "invalid magic number",
		}).Debug("Failed to verify header.")
		return ErrInvalidMagicNumber
	}

	expectedCheckSum := crc32.ChecksumIEEE(message.HeaderWithoutCheckSum())
	if expectedCheckSum != message.HeaderCheckSum() {
		logging.VLog().WithFields(logrus.Fields{
			"expect": expectedCheckSum,
			"actual": message.HeaderCheckSum(),
			"err":    "invalid header checksum",
		}).Debug("Failed to verify header.")
		return ErrInvalidHeaderCheckSum
	}

	if message.DataLength() > MaxNebMessageDataLength {
		logging.VLog().WithFields(logrus.Fields{
			"messageName": message.MessageName(),
			"dataLength":  message.DataLength(),
			"limit":       MaxNebMessageDataLength,
			"err":         "exceeded max data length",
		}).Debug("Failed to verify header.")
		return ErrExceedMaxDataLength
	}

	return nil
}

// VerifyData verify message data
func (message *NebMessage) VerifyData() error {
	expectedCheckSum := crc32.ChecksumIEEE(message.Data())
	if expectedCheckSum != message.DataCheckSum() {
		logging.VLog().WithFields(logrus.Fields{
			"expect": expectedCheckSum,
			"actual": message.DataCheckSum(),
			"err":    "invalid data checksum",
		}).Debug("Failed to verify data")
		return ErrInvalidDataCheckSum
	}
	return nil
}

// FlagSendMessageAt flag of send message time
func (message *NebMessage) FlagSendMessageAt() {
	message.sendMessageAt = time.Now().UnixNano()
}

// FlagWriteMessageAt flag of write message time
func (message *NebMessage) FlagWriteMessageAt() {
	message.writeMessageAt = time.Now().UnixNano()
}

// LatencyFromSendToWrite latency from sendMessage to writeMessage
func (message *NebMessage) LatencyFromSendToWrite() int64 {
	if message.sendMessageAt == 0 {
		return -1
	} else if message.writeMessageAt == 0 {
		message.FlagWriteMessageAt()
	}

	// convert from nano to millisecond.
	return (message.writeMessageAt - message.sendMessageAt) / int64(time.Millisecond)
}
