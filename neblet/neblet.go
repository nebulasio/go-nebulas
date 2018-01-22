package neblet

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

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
	metricsNebstartGauge = m.GetOrRegisterGauge("neb.start", nil)
)

// Neblet manages ldife cycle of blockchain services.
type Neblet struct {
	config *nebletpb.Config

	genesis *corepb.Genesis

	accountManager *account.Manager

	netService p2p.Manager

	consensus consensus.Consensus

	storage storage.Storage

	blockChain *core.BlockChain

	syncService *nsync.Service

	apiServer rpc.Server

	managementServer rpc.Server

	lock sync.RWMutex

	eventEmitter *core.EventEmitter

	running bool
}

// New returns a new neblet.
func New(config *nebletpb.Config) (*Neblet, error) {
	var err error
	n := &Neblet{config: config}
	n.genesis, err = core.LoadGenesisConf(config.Chain.Genesis)
	if err != nil {
		return nil, err
	}
	n.accountManager = account.NewManager(n)

	// init random seed.
	rand.Seed(time.Now().UTC().UnixNano())

	return n, nil
}

// Setup setup neblet
func (n *Neblet) Setup() {
	var err error
	logging.CLog().Info("Setuping Neblet...")

	// storage
	n.storage, err = storage.NewMemoryStorage()
	// n.storage, err = storage.NewDiskStorage(n.config.Chain.Datadir)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"dir": n.config.Chain.Datadir,
			"err": err,
		}).Fatal("Failed to open disk storage.")
	}

	// net
	n.netService, err = p2p.NewNetService(n)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Failed to setup net service.")
	}

	// core
	n.eventEmitter = core.NewEventEmitter(1024)
	n.blockChain, err = core.NewBlockChain(n)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Failed to setup blockchain.")
	}
	gasPrice := util.NewUint128FromString(n.config.Chain.GasPrice)
	gasLimit := util.NewUint128FromString(n.config.Chain.GasLimit)
	n.blockChain.TransactionPool().SetGasConfig(gasPrice, gasLimit)
	n.blockChain.BlockPool().RegisterInNetwork(n.netService)
	n.blockChain.TransactionPool().RegisterInNetwork(n.netService)

	// consensus
	n.consensus, err = dpos.NewDpos(n)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Failed to setup consensus.")
	}
	n.blockChain.SetConsensusHandler(n.consensus)

	// sync
	n.syncService = nsync.NewService(n.blockChain, n.netService)
	n.blockChain.SetSyncService(n.syncService)

	// api
	n.apiServer = rpc.NewAPIServer(n)

	logging.CLog().Info("Setuped Neblet.")
}

// Start starts the services of the neblet.
func (n *Neblet) Start() {
	n.lock.Lock()
	defer n.lock.Unlock()

	logging.CLog().Info("Starting Neblet...")

	if n.running {
		logging.CLog().WithFields(logrus.Fields{
			"err": "neblet is already running",
		}).Fatal("Failed to start neblet.")
	}
	n.running = true

	if n.config.Stats.EnableMetrics {
		metrics.Start(n)
	}

	if err := n.netService.Start(); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Failed to start net service.")
	}

	if err := n.apiServer.Start(); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Failed to start api server.")
	}

	if err := n.apiServer.RunGateway(); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Failed to start api gateway.")
	}

	n.blockChain.Start()
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
			logging.CLog().WithFields(logrus.Fields{
				"err": err,
			}).Fatal("Failed to enable mining.")
		}
	}

	// first sync
	if len(n.Config().Network.Seed) > 0 {
		n.blockChain.StartActiveSync()
	} else {
		logging.CLog().Info("This is a seed node.")
		n.Consensus().ResumeMining()
	}

	metricsNebstartGauge.Update(1)

	logging.CLog().Info("Started Neblet.")
}

// Stop stops the services of the neblet.
func (n *Neblet) Stop() {
	n.lock.Lock()
	defer n.lock.Unlock()

	logging.CLog().Info("Stopping Neblet...")

	if n.consensus != nil {
		n.consensus.Stop()
		n.consensus = nil
	}

	if n.syncService != nil {
		n.syncService.Stop()
		n.syncService = nil
	}

	if n.eventEmitter != nil {
		n.eventEmitter.Stop()
		n.eventEmitter = nil
	}

	if n.blockChain != nil {
		n.blockChain.TransactionPool().Stop()
		n.blockChain.BlockPool().Stop()
		n.blockChain.Stop()
		n.blockChain = nil
	}

	if n.apiServer != nil {
		n.apiServer.Stop()
		n.apiServer = nil
	}

	if n.managementServer != nil {
		n.managementServer.Stop()
		n.managementServer = nil
	}

	if n.netService != nil {
		n.netService.Stop()
		n.netService = nil
	}

	if n.config.Stats.EnableMetrics {
		metrics.Stop()
	}

	n.accountManager = nil

	n.running = false

	logging.CLog().Info("Stopped Neblet.")
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
func (n *Neblet) Config() *nebletpb.Config {
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

// SyncService return sync service
func (n *Neblet) SyncService() *nsync.Service {
	return n.syncService
}
