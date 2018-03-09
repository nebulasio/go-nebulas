package neblet

import (
	"testing"

	/* "github.com/nebulasio/go-nebulas/core" */
	"github.com/nebulasio/go-nebulas/neblet/pb"
	/* 	nebnet "github.com/nebulasio/go-nebulas/net"
	   	"github.com/nebulasio/go-nebulas/storage"
	   	"github.com/stretchr/testify/assert" */)

func MockConfig() *nebletpb.Config {
	return &nebletpb.Config{
		Chain: &nebletpb.ChainConfig{
			ChainId:          100,
			Datadir:          "/Users/tangtangshouxin/workspace/blockchain/src/github.com/nebulasio/go-nebulas/data.db",
			Keydir:           "keydir",
			Genesis:          "/Users/tangtangshouxin/workspace/blockchain/src/github.com/nebulasio/go-nebulas/conf/default/genesis.conf",
			StartMine:        true,
			Coinbase:         "59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232",
			Miner:            "59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232",
			Passphrase:       "passphrase",
			SignatureCiphers: []string{"ECC_SECP256K1"},
		},
		Network: &nebletpb.NetworkConfig{
			Listen:     []string{"0.0.0.0:8680"},
			PrivateKey: "/Users/tangtangshouxin/workspace/blockchain/src/github.com/nebulasio/go-nebulas/conf/network/ed25519key",
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
	TEST_ALL_ELEMENT             = 0
	TEST_LACK_CHAIN              = 1
	TEST_LACK_NETWORK            = 2
	TEST_LACK_RPC                = 3
	TEST_LACK_GENESIS            = 4
	TEST_NEW_DB                  = 5
	TEST_NEW_DB_AND_LACK_GENESIS = 6
)

func makeNebT(cmd int) (*Neblet, error) {
	//var conf *nebletpb.Config = new(nebletpb.Config)
	conf := MockConfig()
	switch cmd {
	case TEST_ALL_ELEMENT:
		{
			break
		}
	case TEST_LACK_CHAIN:
		{
			conf.Chain = nil
		}
	case TEST_LACK_NETWORK:
		{
			conf.Network = nil
		}
	case TEST_LACK_RPC:
		{
			conf.Rpc = nil
		}
	case TEST_LACK_GENESIS:
		{
			conf.Chain.Genesis = "/Users/tangtangshouxin/workspace/blockchain/src/github.com/nebulasio/go-nebulas/conf/default/x.conf"
		}
	case TEST_NEW_DB:
		{
			conf.Chain.Datadir = "x.db"
		}
	case TEST_NEW_DB_AND_LACK_GENESIS:
		{
			conf.Chain.Genesis = "/Users/tangtangshouxin/workspace/blockchain/src/github.com/nebulasio/go-nebulas/conf/default/x.conf"
			conf.Chain.Datadir = "x.db"
		}
	}
	n, err := New(conf)
	if err != nil {
		return nil, err
	}
	return n, nil
}

/*
func TestConfig(t *testing.T) {
	neb, err := makeNebT(TEST_ALL_ELEMENT)
	assert.Nil(t, err)

	neb.storage, err = storage.NewDiskStorage(neb.config.Chain.Datadir)
	assert.Nil(t, err)

	_, err = nebnet.NewNetService(neb)
	assert.Nil(t, err)

	_, err = core.NewBlockChain(neb)
	assert.Nil(t, err)

	neb.rpcServer = rpc.NewServer(neb)
	assert.NotNil(t, neb.rpcServer)

	defer func() {
		if err := recover(); err != nil {
			println(err.(string))
		}
	}()
}*/
/*
func TestConfigLackChain(t *testing.T) {
	_, err := makeNebT(TEST_LACK_CHAIN)
	assert.NotNil(t, err)
}*/
/*
func TestConfigLackNetWork(t *testing.T) {
	neb, err := makeNebT(TEST_LACK_NETWORK)
	assert.Nil(t, err)
	//fmt.Printf("chain:%v\n", neb.config.GetChain())
	neb.storage, err = storage.NewDiskStorage(neb.config.Chain.Datadir)
	assert.Nil(t, err)

	_, err = nebnet.NewNetService(neb) //FATA
	assert.NotNil(t, err)
	//assert.Panics(t, err)
}*/
/*func TestConfigLackRpc(t *testing.T) {
	neb, err := makeNebT(TEST_LACK_RPC)
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
func TestConfigLackGenesisConf(t *testing.T) {
	{
		//init not genesis db and not genesis.conf
		/*neb, err := makeNebT(TEST_NEW_DB_AND_LACK_GENESIS)
		assert.Nil(t, err)

		neb.storage, err = storage.NewDiskStorage(neb.config.Chain.Datadir)
		assert.NotNil(t, err)

		_, err = nebnet.NewNetService(neb)
		assert.Nil(t, err)

		_, err = core.NewBlockChain(neb)
		assert.NotNil(t, err)*/
	}
	{
		//first start db is new
		/*neb, err := makeNebT(TEST_NEW_DB)
		assert.Nil(t, err)

		neb.storage, err = storage.NewDiskStorage(neb.config.Chain.Datadir)
		assert.Nil(t, err)

		_, err = nebnet.NewNetService(neb)
		assert.Nil(t, err)

		_, err = core.NewBlockChain(neb)
		assert.Nil(t, err)*/
	}
	{
		/*
			//second start and lack config
			neb, err := makeNebT(TEST_NEW_DB_AND_LACK_GENESIS)
			assert.Nil(t, err)

			neb.storage, err = storage.NewDiskStorage(neb.config.Chain.Datadir)
			assert.Nil(t, err)

			_, err = nebnet.NewNetService(neb)
			assert.Nil(t, err)

			_, err = core.NewBlockChain(neb)
			assert.Nil(t, err)*/
	}
	{
		//genesis db exist and genesis.conf is exist
		/* 		neb, err := makeNebT(TEST_NEW_DB)
		   		assert.Nil(t, err)

		   		neb.storage, err = storage.NewDiskStorage(neb.config.Chain.Datadir)
		   		assert.Nil(t, err)

		   		_, err = nebnet.NewNetService(neb)
		   		assert.Nil(t, err)

		   		_, err = core.NewBlockChain(neb)
		   		assert.Nil(t, err) */
	}
}
