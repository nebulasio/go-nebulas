package neblet

import (
	"math/rand"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	corepb "github.com/nebulasio/go-nebulas/core/pb"
	nebletpb "github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/rpc"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/stretchr/testify/assert"
)

func MockConfig() *nebletpb.Config {
	return &nebletpb.Config{
		Chain: &nebletpb.ChainConfig{
			ChainId:          100,
			Datadir:          "data.db",
			Keydir:           "keydir",
			Genesis:          "conf/default/genesis.conf",
			StartMine:        true,
			Coinbase:         "59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232",
			Miner:            "59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232",
			Passphrase:       "passphrase",
			SignatureCiphers: []string{"ECC_SECP256K1"},
		},
		Network: &nebletpb.NetworkConfig{
			Listen:     []string{"0.0.0.0:8680"},
			PrivateKey: "conf/network/ed25519key",
			NetworkId:  1,
		},
		Rpc: &nebletpb.RPCConfig{
			RpcListen:  []string{"127.0.0.1:8684"},
			HttpListen: []string{"127.0.0.1:8685"},
			HttpModule: []string{"api", "admin"},
		},
		App: &nebletpb.AppConfig{
			LogLevel:          "debug",
			LogFile:           "logs",
			EnableCrashReport: false,
			CrashReportUrl:    "https://crashreport.nebulas.io",
			Pprof:             &nebletpb.PprofConfig{HttpListen: "127.0.0.1:7777"},
		},
	}
}

const (
	TestAllElement          = 0
	TestLackChain           = 1
	TestLackNetWork         = 2
	TestLackRPC             = 3
	TestLackGenesis         = 4
	TestNewDb               = 5
	TestNewDbAndLackGenesis = 6
)

var (
	MockDynasty = []string{
		"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c",
		"2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8",
		"333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700",
		"48f981ed38910f1232c1bab124f650c482a57271632db9e3",
		"59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232",
		"75e4e5a71d647298b88928d8cb5da43d90ab1a6c52d0905f",
	}
)

// MockGenesisConf return mock genesis conf
func MockGenesisConf() *corepb.Genesis {
	return &corepb.Genesis{
		Meta: &corepb.GenesisMeta{ChainId: 100},
		Consensus: &corepb.GenesisConsensus{
			Dpos: &corepb.GenesisConsensusDpos{
				Dynasty: MockDynasty,
			},
		},
		TokenDistribution: []*corepb.GenesisTokenDistribution{
			&corepb.GenesisTokenDistribution{
				Address: "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8",
				Value:   "10000000000000000000000",
			},
		},
	}
}

func NewT(config *nebletpb.Config, cmd int) (*Neblet, error) {
	//var err error
	n := &Neblet{
		config:     config,
		netService: mockNetService{},
	}

	// try enable profile.
	n.TryStartProfiling()

	//n.genesis, _ = core.LoadGenesisConf(config.Chain.Genesis)
	n.genesis = MockGenesisConf()
	switch cmd {
	case TestNewDbAndLackGenesis:
		{
			n.genesis = nil
		}
	}
	//n.accountManager = account.NewManager(n)

	// init random seed.
	rand.Seed(time.Now().UTC().UnixNano())

	return n, nil
}
func makeNebT(cmd int) (*Neblet, error) {
	//var conf *nebletpb.Config = new(nebletpb.Config)
	conf := MockConfig()
	switch cmd {
	case TestAllElement:
		{
			break
		}
	case TestLackChain:
		{
			conf.Chain = nil
		}
	case TestLackNetWork:
		{
			conf.Network = nil
		}
	case TestLackRPC:
		{
			conf.Rpc = nil
		}
	case TestLackGenesis:
		{
			break
		}
	case TestNewDb:
		{
			conf.Chain.Datadir = "x2.db"
		}
	case TestNewDbAndLackGenesis:
		{
			conf.Chain.Datadir = "x2.db"
		}
	}
	n, err := NewT(conf, cmd)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func TestConfig(t *testing.T) {
	var neb *Neblet
	var err error
	neb, err = makeNebT(TestAllElement)
	assert.Nil(t, err)

	neb.storage, err = storage.NewDiskStorage(neb.config.Chain.Datadir)
	assert.Nil(t, err)

	// _, err := nebnet.NewNetService(neb)
	//assert.Nil(t, err)

	_, err = core.NewBlockChain(neb)
	assert.Nil(t, err)

	neb.rpcServer = rpc.NewServer(neb)
	assert.NotNil(t, neb.rpcServer)

}

/*func TestConfigLackChain(t *testing.T) {
	_, err := makeNebT(TestLackChain)
	assert.NotNil(t, err)
}*/

/*
func TestConfigLackNetWork(t *testing.T) {
	neb, err := makeNebT(TestLackNetWork)
	assert.Nil(t, err)
	//fmt.Printf("chain:%v\n", neb.config.GetChain())
	neb.storage, err = storage.NewDiskStorage(neb.config.Chain.Datadir)
	assert.Nil(t, err)

	_, err = nebnet.NewNetService(neb) //FATA
	assert.NotNil(t, err)
	//assert.Panics(t, err)
}*/
/*func TestConfigLackRpc(t *testing.T) {
	neb, err := makeNebT(TestLackRpc)
	assert.Nil(t, err)

	neb.storage, err = storage.NewDiskStorage(neb.config.Chain.Datadir)
	assert.Nil(t, err)

	_, err = nebnet.NewNetService(neb)
	assert.Nil(t, err)

	_, err = core.NewBlockChain(neb)
	assert.Nil(t, err)

	neb.rpcServer = rpc.NewServer(neb) //fatal
	assert.NotNil(t, neb.rpcServer)
}*/
func TestConfigBlockChain(t *testing.T) {
	{
		//init not genesis db and not genesis.conf
		/*var neb *Neblet
		var err error
		neb, err = makeNebT(TestNewDb)
		assert.Nil(t, err)

		neb.storage, err = storage.NewDiskStorage(neb.config.Chain.Datadir)
		assert.Nil(t, err)

		_, err = core.NewBlockChain(neb)
		assert.Nil(t, err)*/

		//neb.storage.Close()
	}
	{
		//db exist and lack genesis.conf
		/*var neb *Neblet
		var err error
		neb, err = makeNebT(TestNewDb)
		assert.Nil(t, err)

		neb.storage, err = storage.NewDiskStorage(neb.config.Chain.Datadir)
		assert.Nil(t, err)

		_, err = core.NewBlockChain(neb)
		assert.Nil(t, err)*/
	}
	{
		//db exist and genesis.conf exist
		/*neb, err := makeNebT(TestNewDb)
		assert.Nil(t, err)

		neb.storage, err = storage.NewDiskStorage(neb.config.Chain.Datadir)
		assert.Nil(t, err)

		_, err = core.NewBlockChain(neb)
		assert.Nil(t, err)*/
	}
	{
		//genesis db exist and genesis.conf is exist
		/*neb, err := makeNebT(TestNewDb)
		assert.Nil(t, err)

		neb.storage, err = storage.NewDiskStorage(neb.config.Chain.Datadir)
		assert.Nil(t, err)

		_, err = nebnet.NewNetService(neb)
		assert.Nil(t, err)

		_, err = core.NewBlockChain(neb)
		assert.Nil(t, err)*/
	}
}

var (
	received = []byte{}
)

type mockNetService struct{}

func (n mockNetService) Start() error { return nil }
func (n mockNetService) Stop()        {}

func (n mockNetService) Node() *net.Node { return nil }

func (n mockNetService) Sync(net.Serializable) error { return nil }

func (n mockNetService) Register(...*net.Subscriber)   {}
func (n mockNetService) Deregister(...*net.Subscriber) {}

func (n mockNetService) Broadcast(name string, msg net.Serializable, priority int) {
	pb, _ := msg.ToProto()
	bytes, _ := proto.Marshal(pb)
	received = bytes
}
func (n mockNetService) Relay(name string, msg net.Serializable, priority int) {
	pb, _ := msg.ToProto()
	bytes, _ := proto.Marshal(pb)
	received = bytes
}
func (n mockNetService) SendMsg(name string, msg []byte, target string, priority int) error {
	received = msg
	return nil
}

func (n mockNetService) SendMessageToPeers(messageName string, data []byte, priority int, filter net.PeerFilterAlgorithm) []string {
	return make([]string, 0)
}
func (n mockNetService) SendMessageToPeer(messageName string, data []byte, priority int, peerID string) error {
	return nil
}

func (n mockNetService) ClosePeer(peerID string, reason error) {}

func (n mockNetService) BroadcastNetworkID([]byte) {}
