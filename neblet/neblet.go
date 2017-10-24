package neblet

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/multiformats/go-multiaddr"
	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/consensus/pow"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net/p2p"
	"github.com/nebulasio/go-nebulas/rpc"
	nsync "github.com/nebulasio/go-nebulas/sync"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrNebletAlreadyRunning throws when the neblet is already running.
	ErrNebletAlreadyRunning = errors.New("neblet is already running")
)

// Neblet manages life cycle of blockchain services.
type Neblet struct {
	config nebletpb.Config

	accountManager *account.Manager

	// p2pManager *p2p.Manager
	netService *p2p.NetService

	consensus consensus.Consensus

	blockChain *core.BlockChain

	snycManager *nsync.Manager

	rpcServer *rpc.Server

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
	n.lock.Lock()
	defer n.lock.Unlock()
	log.Info("Starting neblet...")

	if n.running {
		return ErrNebletAlreadyRunning
	}
	n.running = true

	// TODO: use new proto config.
	p2pConfig := n.getP2PConfig()
	// TODO: handle err
	// n.p2pManager = p2p.NewManager(p2pConfig)
	n.netService, _ = p2p.NewNetService(p2pConfig)

	n.blockChain = core.NewBlockChain(core.TestNetID)
	fmt.Printf("chainID is %d\n", n.blockChain.ChainID())
	n.blockChain.BlockPool().RegisterInNetwork(n.netService)
	n.blockChain.TransactionPool().RegisterInNetwork(n.netService)

	n.consensus = pow.NewPow(n)
	n.blockChain.SetConsensusHandler(n.consensus)

	// start sync service
	n.snycManager = nsync.NewManager(n.blockChain, n.consensus, n.netService)

	n.rpcServer = rpc.NewServer(n)

	// start.
	n.netService.Start()
	n.blockChain.BlockPool().Start()
	n.blockChain.TransactionPool().Start()
	n.consensus.Start()
	n.snycManager.Start()
	go n.rpcServer.Start()

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

	if n.rpcServer != nil {
		n.rpcServer.Stop()
		n.rpcServer = nil
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

// TODO: move this to p2p package.
func (n *Neblet) getP2PConfig() *p2p.Config {
	config := p2p.DefautConfig()
	config.IP = localHost()
	seed := n.config.P2P.Seed
	if len(seed) > 0 {
		seed, err := multiaddr.NewMultiaddr(seed)
		if err != nil {
			log.Error("param seed error, creating seed node fail", err)
			return nil
		}
		config.BootNodes = []multiaddr.Multiaddr{seed}
	}
	if port := n.config.P2P.Port; port > 0 {
		config.Port = uint(port)
	}
	// P2P network randseed, in this release we use port as randseed
	// config.Randseed = time.Now().Unix()
	config.Randseed = int64(config.Port)
	return config
}

// TODO: move this to p2p package.
func localHost() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}
