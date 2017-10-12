package neblet

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/multiformats/go-multiaddr"
	"github.com/nebulasio/go-nebulas/components/net/p2p"
	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/consensus/pow"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/rpc"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrNebletAlreadyRunning throws when the neblet is already running.
	ErrNebletAlreadyRunning = errors.New("neblet is already running")
)

// Neblet manages life cycle of blockchain services.
type Neblet struct {
	config nebletpb.Config

	keyStore *keystore.Keystore

	p2pManager *p2p.Manager

	consensus consensus.Consensus

	blockChain *core.BlockChain

	rpcServer *rpc.Server

	lock sync.RWMutex

	running bool
}

// New returns a new neblet.
func New(config nebletpb.Config) *Neblet {
	return &Neblet{config: config}
}

// Start starts the services of the neblet.
func (n *Neblet) Start() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.running {
		return ErrNebletAlreadyRunning
	}
	n.running = true

	n.keyStore = keystore.DefaultKS

	// TODO: use new proto config.
	p2pConfig := n.getP2PConfig()
	n.p2pManager = p2p.NewManager(p2pConfig)

	n.blockChain = core.NewBlockChain(core.TestNetID)
	fmt.Printf("chainID is %d\n", n.blockChain.ChainID())
	n.blockChain.BlockPool().RegisterInNetwork(n.p2pManager)

	n.consensus = pow.NewPow(n.blockChain, n.p2pManager)
	n.blockChain.SetConsensusHandler(n.consensus)

	n.rpcServer = rpc.NewServer()

	// start.
	n.p2pManager.Start()
	n.blockChain.BlockPool().Start()
	n.consensus.Start()
	go n.rpcServer.Start()

	// TODO: error handling
	return nil
}

// Stop stops the services of the neblet.
func (n *Neblet) Stop() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	n.consensus.Stop()
	n.blockChain.BlockPool().Stop()
	n.p2pManager.Stop()

	n.rpcServer.Stop()
	n.rpcServer = nil

	n.keyStore = nil

	n.running = false

	return nil
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
	config.Randseed = 20170922
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
