package neblet

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/nebulasio/go-nebulas/cmd/console"

	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/consensus/dpos"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/metrics"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net/p2p"
	"github.com/nebulasio/go-nebulas/rpc"
	"github.com/nebulasio/go-nebulas/storage"
	nsync "github.com/nebulasio/go-nebulas/sync"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	m "github.com/rcrowley/go-metrics"
)

var (
	// ErrNebletAlreadyRunning throws when the neblet is already running.
	ErrNebletAlreadyRunning = errors.New("neblet is already running")

	// ErrIncompatibleStorageSchemeVersion throws when the storage schema has been changed
	ErrIncompatibleStorageSchemeVersion = errors.New("incompatible storage schema version, pls migrate your storage")
)

var (
	storageSchemeVersionKey = []byte("scheme")
	storageSchemeVersionVal = []byte("0.5.0")
	metricsNebstartGauge    = m.GetOrRegisterGauge("neb.start", nil)
)

// Neblet manages ldife cycle of blockchain services.
type Neblet struct {
	config nebletpb.Config

	genesis *corepb.Genesis

	accountManager *account.Manager

	netService p2p.Manager

	consensus consensus.Consensus

	storage storage.Storage

	blockChain *core.BlockChain

	syncService *nsync.SyncService

	apiServer rpc.Server

	managementServer rpc.Server

	lock sync.RWMutex

	eventEmitter *core.EventEmitter

	running bool
}

// New returns a new neblet.
func New(config nebletpb.Config) (*Neblet, error) {
	var err error
	n := &Neblet{config: config}
	n.genesis, err = core.LoadGenesisConf(config.Chain.Genesis)
	if err != nil {
		return nil, err
	}
	n.accountManager = account.NewManager(n)
	return n, nil
}

// Setup setup neblet
func (n *Neblet) Setup() error {
	var err error

	// init random seed.
	rand.Seed(time.Now().UTC().UnixNano())

	// storage
	n.storage, err = storage.NewDiskStorage(n.config.Chain.Datadir)
	if err != nil {
		return err
	}
	if err = n.checkSchemeVersion(n.storage); err != nil {
		return err
	}

	// net
	n.netService, err = p2p.NewNetService(n)
	if err != nil {
		return err
	}

	// core
	n.eventEmitter = core.NewEventEmitter(1024)
	n.blockChain, err = core.NewBlockChain(n)
	if err != nil {
		return err
	}
	gasPrice := util.NewUint128FromString(n.config.Chain.GasPrice)
	gasLimit := util.NewUint128FromString(n.config.Chain.GasLimit)
	n.blockChain.TransactionPool().SetGasConfig(gasPrice, gasLimit)
	n.blockChain.BlockPool().RegisterInNetwork(n.netService)
	n.blockChain.TransactionPool().RegisterInNetwork(n.netService)

	// consensus
	n.consensus, err = dpos.NewDpos(n)
	if err != nil {
		return err
	}
	n.blockChain.SetConsensusHandler(n.consensus)

	// sync
	n.syncService = nsync.NewSyncService(n.blockChain, n.netService)

	// api
	n.apiServer = rpc.NewAPIServer(n)
	return nil
}

// Start starts the services of the neblet.
func (n *Neblet) Start() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	logging.CLog().Info("Starting Neblet...")

	if n.running {
		return ErrNebletAlreadyRunning
	}
	n.running = true

	if n.config.Stats.EnableMetrics {
		metrics.Start(n)
	}

	if err := n.netService.Start(); err != nil {
		return err
	}

	if err := n.apiServer.Start(); err != nil {
		return err
	}

	if err := n.apiServer.RunGateway(); err != nil {
		return err
	}

	n.blockChain.BlockPool().Start()
	n.blockChain.TransactionPool().Start()
	n.eventEmitter.Start()
	n.syncService.Start()

	// start consensus
	chainConf := n.config.Chain
	n.consensus.Start()
	if chainConf.StartMine {
		passphrase := n.config.Chain.Passphrase
		if len(passphrase) == 0 {
			fmt.Println("***********************************************")
			fmt.Println("miner address:" + n.config.Chain.Miner)
			prompt := console.Stdin
			passphrase, _ = prompt.PromptPassphrase("Enter the miner's passphrase:")
			fmt.Println("***********************************************")
		}
		err := n.consensus.EnableMining(chainConf.Passphrase)
		if err != nil {
			return err
		}
	}

	metricsNebstartGauge.Update(1)
	return nil
}

// Stop stops the services of the neblet.
func (n *Neblet) Stop() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	logging.CLog().Info("Stopping neblet...")

	if n.consensus != nil {
		n.consensus.Stop()
		n.consensus = nil
	}

	if n.blockChain != nil {
		n.blockChain.BlockPool().Stop()
		n.blockChain = nil
	}

	if n.eventEmitter != nil {
		n.eventEmitter.Stop()
		n.eventEmitter = nil
	}

	if n.netService != nil {
		n.netService.Stop()
		n.netService = nil
	}

	if n.apiServer != nil {
		n.apiServer.Stop()
		n.apiServer = nil
	}

	if n.managementServer != nil {
		n.managementServer.Stop()
		n.managementServer = nil
	}

	if n.config.Stats.EnableMetrics {
		metrics.Stop()
	}

	n.accountManager = nil

	n.running = false

	return nil
}

// SetGenesis set genesis conf
func (n *Neblet) SetGenesis(g *corepb.Genesis) {
	n.genesis = g
}

// Genesis returns genesis conf.
func (n *Neblet) Genesis() *corepb.Genesis {
	return n.genesis
}

// Config returns neblet configuration.
func (n *Neblet) Config() nebletpb.Config {
	return n.config
}

// Storage returns storage reference.
func (n *Neblet) Storage() storage.Storage {
	return n.storage
}

// BlockChain returns block chain reference.
func (n *Neblet) BlockChain() *core.BlockChain {
	return n.blockChain
}

// EventEmitter returns eventEmitter reference.
func (n *Neblet) EventEmitter() *core.EventEmitter {
	return n.eventEmitter
}

// AccountManager returns account manager reference.
func (n *Neblet) AccountManager() *account.Manager {
	return n.accountManager
}

// NetManager returns p2p manager reference.
func (n *Neblet) NetManager() p2p.Manager {
	return n.netService
}

// Consensus returns consensus reference.
func (n *Neblet) Consensus() consensus.Consensus {
	return n.consensus
}

// StartActiveSync start active sync from peers.
func (n *Neblet) StartActiveSync() {
	n.syncService.StartActiveSync()
}

// checks if the storage scheme version is compatiable
func (n *Neblet) checkSchemeVersion(stor storage.Storage) error {
	version, err := stor.Get(storageSchemeVersionKey)
	if err != nil && err != storage.ErrKeyNotFound {
		return err
	}
	if err == storage.ErrKeyNotFound {
		stor.Put(storageSchemeVersionKey, storageSchemeVersionVal)
		return nil
	}
	if !byteutils.Equal(version, storageSchemeVersionVal) {
		return ErrIncompatibleStorageSchemeVersion
	}
	return nil
}
