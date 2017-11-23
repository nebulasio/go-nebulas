package neblet

import (
	"errors"
	"sync"

	"github.com/nebulasio/go-nebulas/util/byteutils"

	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/consensus/pod"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net/p2p"
	"github.com/nebulasio/go-nebulas/rpc"
	"github.com/nebulasio/go-nebulas/storage"
	nsync "github.com/nebulasio/go-nebulas/sync"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrNebletAlreadyRunning throws when the neblet is already running.
	ErrNebletAlreadyRunning = errors.New("neblet is already running")
)

var (
	storageSchemeVersionKey = []byte("scheme")
	storageSchemeVersionVal = []byte("0.1.1")
)

// Neblet manages ldife cycle of blockchain services.
type Neblet struct {
	config nebletpb.Config

	accountManager *account.Manager

	// p2pManager *p2p.Manager
	netService *p2p.NetService

	consensus consensus.Consensus

	blockChain *core.BlockChain

	snycManager *nsync.Manager

	apiServer rpc.Server

	managementServer rpc.Server

	lock sync.RWMutex

	running bool
}

// New returns a new neblet.
func New(config nebletpb.Config) *Neblet {
	n := &Neblet{config: config}
	n.accountManager = account.NewManager(n)
	return n
}

// Start starts the services of the neblet.
func (n *Neblet) Start() error {
	var err error
	n.lock.Lock()
	defer n.lock.Unlock()
	log.Info("Starting neblet...")

	if n.running {
		return ErrNebletAlreadyRunning
	}
	n.running = true

	//n.accountManager = account.NewManager(n)

	n.netService, err = p2p.NewNetService(n)
	if err != nil {
		log.Error("new NetService occurs error ", err)
		return err
	}

	storage, err := storage.NewDiskStorage(n.config.GetStorage().Location)
	// storage, err := storage.NewMemoryStorage()
	if err != nil {
		return err
	}
	if err := n.CheckSchemeVersion(storage); err != nil {
		return err
	}

	n.blockChain, err = core.NewBlockChain(core.TestNetID, storage)
	if err != nil {
		return err
	}
	n.blockChain.BlockPool().RegisterInNetwork(n.netService)
	n.blockChain.TransactionPool().RegisterInNetwork(n.netService)

	n.consensus = pod.NewPoD(n)
	n.blockChain.SetConsensusHandler(n.consensus)

	// start sync service
	n.snycManager = nsync.NewManager(n.blockChain, n.consensus, n.netService)

	n.apiServer = rpc.NewAPIServer(n)
	n.managementServer = rpc.NewManagementServer(n)

	// start.
	err = n.netService.Start()
	if err != nil {
		return err
	}
	n.blockChain.BlockPool().Start()
	n.blockChain.TransactionPool().Start()
	n.consensus.Start()
	n.snycManager.Start()
	go n.apiServer.Start()
	go n.managementServer.Start()
	go n.apiServer.RunGateway()

	// TODO: error handling
	return nil
}

// Stop stops the services of the neblet.
func (n *Neblet) Stop() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	log.Info("Stopping neblet...")

	if n.consensus != nil {
		n.consensus.Stop()
		n.consensus = nil
	}

	if n.blockChain != nil {
		n.blockChain.BlockPool().Stop()
		n.blockChain = nil
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

	n.accountManager = nil

	n.running = false

	return nil
}

// Config returns neblet configuration.
func (n *Neblet) Config() nebletpb.Config {
	return n.config
}

// BlockChain returns block chain reference.
func (n *Neblet) BlockChain() *core.BlockChain {
	return n.blockChain
}

// AccountManager returns account manager reference.
func (n *Neblet) AccountManager() *account.Manager {
	return n.accountManager
}

// NetService returns p2p manager reference.
func (n *Neblet) NetService() *p2p.NetService {
	return n.netService
}

// CheckSchemeVersion checks if the storage scheme version is compatiable
func (n *Neblet) CheckSchemeVersion(storage storage.Storage) error {
	version, err := storage.Get(storageSchemeVersionKey)
	if err != nil {
		storage.Put(storageSchemeVersionKey, storageSchemeVersionVal)
		return nil
	}
	if !byteutils.Equal(version, storageSchemeVersionVal) {
		return errors.New("incompatible storage schema version, pls migrate your storage")
	}
	return nil
}
