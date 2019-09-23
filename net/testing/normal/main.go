package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"log"
	mrand "math/rand"
	"time"

	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multicodec"
	json "github.com/multiformats/go-multicodec/json"
)

const proto = "/example/1.0.0"
const letterBytes = "0123456789ABCDEF0123456789ABCDE10123456789ABCDEF0123456789ABCDEF"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// Message is a serializable/encodable object that we will send
// on a Stream.
type Message struct {
	Msg    string
	Index  int
	HangUp bool
}

// WrappedStream wraps a libp2p stream. We encode/decode whenever we
// write/read from a stream, so we can just carry the encoders
// and bufios with us
type WrappedStream struct {
	stream network.Stream
	enc    multicodec.Encoder
	dec    multicodec.Decoder
	w      *bufio.Writer
	r      *bufio.Reader
}

// WrapStream takes a stream and complements it with r/w bufios and
// decoder/encoder. In order to write raw data to the stream we can use
// wrap.w.Write(). To encode something into it we can wrap.enc.Encode().
// Finally, we should wrap.w.Flush() to actually send the data. Handling
// incoming data works similarly with wrap.r.Read() for raw-reading and
// wrap.dec.Decode() to decode.
func WrapStream(s network.Stream) *WrappedStream {
	reader := bufio.NewReader(s)
	writer := bufio.NewWriter(s)
	// This is where we pick our specific multicodec. In order to change the
	// codec, we only need to change this place.
	// See https://godoc.org/github.com/multiformats/go-multicodec/json
	dec := json.Multicodec(false).Decoder(reader)
	enc := json.Multicodec(false).Encoder(writer)
	return &WrappedStream{
		stream: s,
		r:      reader,
		w:      writer,
		enc:    enc,
		dec:    dec,
	}
}

// messages that will be sent between the hosts.
var conversationMsgs = []string{
	"Hello!",
	"Hey!",
	"How are you doing?",
	"Very good! It is great that you can send data on a stream to me!",
	"Not only that, the data is encoded in a JSON object.",
	"Yeah, and we are using the multicodecs interface to encode and decode.",
	"This way we could swap it easily for, say, cbor, or msgpack!",
	"Let's leave that as an excercise for the reader...",
	"Agreed, our last message should activate the HangUp flag",
	"Yes, and the example code will close streams. So sad :(. Bye!",
}

func randSeed(n int) string {
	var src = mrand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = mrand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func makeRandomHost(port int) host.Host {
	// Ignoring most errors for brevity
	// See echo example for more details and better implementation
	// priv, pub, _ := crypto.GenerateKeyPair(crypto.RSA, 2048)
	randseedstr := randSeed(64)
	randseed, _ := hex.DecodeString(randseedstr)
	priv, pub, _ := crypto.GenerateEd25519Key(
		bytes.NewReader(randseed),
	)
	pid, _ := peer.IDFromPublicKey(pub)
	log.Println(pid.Pretty())
	listen, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
	opts := []libp2p.Option{
		libp2p.ListenAddrs(listen),
		libp2p.Identity(priv),
	}
	host, _ := libp2p.New(context.Background(), opts...)

	return host
}

func main() {
	// port := flag.Int("p", 0, "po")
	// flag.Parse()

	// // Choose random ports between 10000-10100
	// // rand.Seed(666)
	// // port1 := rand.Intn(100) + 10000
	// // port2 := port1 + 1

	// // Make 2 hosts
	// // h1 := makeRandomHost(port1)
	// h := makeRandomHost(*port)
	// address, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/51413/ipfs/QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN")
	// bootAddr, bootID, err := p2p.ParseAddressFromMultiaddr(address)
	// // // h1.Peerstore().AddAddrs(h2.ID(), h2.Addrs(), ps.PermanentAddrTTL)
	// log.Info(bootAddr, "|", bootID)
	// h.Peerstore().AddAddrs(bootID, []ma.Multiaddr{bootAddr}, ps.PermanentAddrTTL)

	// // log.Printf("This is a conversation between %s and %s\n", h1.ID(), h2.ID())

	// // Define a stream handler for host number 2
	// // h2.SetStreamHandler(proto, func(stream inet.Stream) {
	// // 	log.Printf("%s: Received a stream", h2.ID())
	// // 	wrappedStream := WrapStream(stream)
	// // 	defer stream.Close()
	// // 	handleStream(wrappedStream)
	// // })

	// // Create new stream from h1 to h2 and start the conversation
	// start := time.Now().Unix()
	// stream, err := h.NewStream(context.Background(), bootID, proto)

	// log.Error("HelloPerf ", time.Now().Unix()-start, "  err", err)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // wrappedStream := WrapStream(stream)
	// // // // This sends the first message
	// // // sendMessage(0, wrappedStream)
	// // // // We keep the conversation on the created stream so we launch
	// // // // this to handle any responses
	// // handleStream(wrappedStream)
	// err = stream.Close()
	// log.Error(err)
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// go func() {
	// 	<-c
	// 	// When we are done, close the stream on our side and exit.
	// 	stream.Close()

	// 	// TODO: remove this once p2pManager handles stop properly.
	// 	os.Exit(1)
	// }()

}

// receiveMessage reads and decodes a message from the stream
func receiveMessage(ws *WrappedStream) (*Message, error) {
	var msg Message
	err := ws.dec.Decode(&msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// sendMessage encodes and writes a message to the stream
func sendMessage(index int, ws *WrappedStream) error {
	msg := &Message{
		Msg:    conversationMsgs[index],
		Index:  index,
		HangUp: index >= len(conversationMsgs)-1,
	}

	err := ws.enc.Encode(msg)
	// Because output is buffered with bufio, we need to flush!
	ws.w.Flush()
	return err
}

// handleStream is a for loop which receives and then sends a message
// an artificial delay of 500ms happens in-between.
// When Message.HangUp is true, it exists. This will close the stream
// on one of the sides. The other side's receiveMessage() will error
// with EOF, thus also breaking out from the loop.
func handleStream(ws *WrappedStream) {
	for {
		// Read
		msg, err := receiveMessage(ws)
		if err != nil {
			break
		}
		pid := ws.stream.Conn().LocalPeer()
		log.Printf("%s says: %s\n", pid, msg.Msg)
		time.Sleep(500 * time.Millisecond)
		if msg.HangUp {
			break
		}
		// Send response
		err = sendMessage(msg.Index+1, ws)
		if err != nil {
			break
		}
	}
}
