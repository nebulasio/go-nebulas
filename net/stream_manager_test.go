package net

import (
	"bytes"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	maxstreamnumber = 10
	reservednumber  = 2
)

var (
	// weight of each msg type
	msgWeight = map[string]MessageWeight{
		HELLO:      MessageWeightZero,
		OK:         MessageWeightZero,
		BYE:        MessageWeightZero,
		SYNCROUTE:  MessageWeightZero,
		ROUTETABLE: MessageWeightRouteTable,

		ChunkHeadersRequest:  MessageWeightZero,
		ChunkHeadersResponse: MessageWeightChainChunks,
		ChunkDataRequest:     MessageWeightZero,
		ChunkDataResponse:    MessageWeightChainChunkData,

		"newblock": MessageWeightNewBlock,
		"dlblock":  MessageWeightZero,
		"dlreply":  MessageWeightZero,
		"newtx":    MessageWeightZero,
	}

	MsgTypes []string
)

func TestAllMsg(t *testing.T) {
	msgtypes := []string{HELLO, OK, BYE, SYNCROUTE, ROUTETABLE,
		ChunkHeadersRequest, ChunkHeadersResponse, ChunkDataRequest, ChunkDataResponse,
		"newblock", "dlblock", "dlreply", "newtx",
	}

	MsgTypes = append(MsgTypes, msgtypes...)
	run()
}

func TestUnvaluedMsg(t *testing.T) {
	msgtypes := []string{HELLO, OK, BYE, SYNCROUTE,
		ChunkHeadersRequest, ChunkDataRequest,
		"dlblock", "dlreply", "newtx",
	}

	MsgTypes = append(MsgTypes, msgtypes...)
	run()
}

func run() {

	cleanupTicker := time.NewTicker(CleanupInterval / 12)
	stopTicker := time.NewTicker(CleanupInterval / 12)
	times := 0
	config := NewConfigFromDefaults()
	for {
		select {
		case <-stopTicker.C:
			cleanupTicker.Stop()
			stopTicker.Stop()
			return
		case <-cleanupTicker.C:
			times++
			fmt.Printf("mock %d\n: max num = %d, reserved = %d\n", times, maxstreamnumber, reservednumber)
			sm := NewStreamManager(config)
			sm.fillMockStreams(maxstreamnumber)
			cleanup(sm)
		}
	}
}

func cleanup(sm *StreamManager) {
	if sm.activePeersCount < maxstreamnumber {
		return
	}

	// total number of each msg type
	msgTotal := make(map[string]int)

	svs := make(StreamValueSlice, 0)
	sm.allStreams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)

		// t type, c count
		for t, c := range stream.msgCount {
			msgTotal[t] += c
		}

		svs = append(svs, &StreamValue{
			stream: stream,
		})

		return true
	})

	fmt.Println("total:")
	fmt.Println(orderedString(&msgTotal))

	for _, sv := range svs {
		for t, c := range sv.stream.msgCount {
			w, _ := msgWeight[t]
			sv.value += float64(c) * float64(w) / float64(msgTotal[t])
		}
	}

	sort.Sort(sort.Reverse(svs))

	fmt.Println("sorted:")
	for _, sv := range svs {
		fmt.Println(strconv.FormatFloat(sv.value, 'f', 3, 64), orderedString(&sv.stream.msgCount))
	}

	fmt.Println("eliminated:")
	eliminated := svs[maxstreamnumber-reservednumber:]
	for _, sv := range eliminated {
		fmt.Println(strconv.FormatFloat(sv.value, 'f', 3, 64), orderedString(&sv.stream.msgCount))
	}
	fmt.Println("")
}

func (sm *StreamManager) fillMockStreams(num int32) {
	sm.activePeersCount = num
	fmt.Println("details:")

	addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9999")

	for ; num > 0; num-- {
		key := "s" + strconv.FormatInt(int64(num), 10)

		rand.Seed(time.Now().UnixNano())
		msgCount := make(map[string]int)
		for _, t := range MsgTypes {
			msgCount[t] = rand.Intn(50)
		}

		pid, _ := peer.IDFromString(key)
		sm.allStreams.Store(key, &Stream{
			pid:      pid,
			addr:     addr,
			msgCount: msgCount,
		})

		fmt.Println(key, orderedString(&msgCount))
	}
}

func orderedString(mc *map[string]int) string {
	var buffer bytes.Buffer
	for _, t := range MsgTypes {
		buffer.WriteString(t + ":" + strconv.Itoa((*mc)[t]) + " ")
	}
	return buffer.String()
}
