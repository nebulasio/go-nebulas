package net

import (
	kbucket "github.com/libp2p/go-libp2p-kbucket"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/nebulasio/go-nebulas/cache"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

type routeTableCache struct {
	peerStore  peerstore.Peerstore
	routeTable *kbucket.RoutingTable
}

func newRouteTableCache(conf *cache.PersistableCacheConfig, ps peerstore.Peerstore, rt *kbucket.RoutingTable) cache.Cache {
	c := &routeTableCache{
		peerStore:  ps,
		routeTable: rt,
	}

	return cache.NewPersistableCache(c, conf)
}

// Entries ..
func (c *routeTableCache) Entries() interface{} {
	ret := make(map[peer.ID][]ma.Multiaddr)
	for _, pid := range c.peerStore.Peers() {
		ret[pid] = c.peerStore.Addrs(pid)
	}
	return ret
}

// Keys ..
func (c *routeTableCache) Keys() interface{} {
	return c.peerStore.Peers()
}

// Set adds a new element or replaces the old element
func (c *routeTableCache) Set(k interface{}, v interface{}) error {
	pid, ok := k.(peer.ID)
	if !ok || pid == "" {
		return cache.ErrInvalidCacheKey
	}
	addrs, ok := v.([]ma.Multiaddr)
	if !ok || len(addrs) == 0 {
		return cache.ErrInvalidCacheValue
	}

	if c.routeTable.Find(pid) != "" {
		c.peerStore.SetAddrs(pid, addrs, peerstore.PermanentAddrTTL)
	} else {
		c.peerStore.AddAddrs(pid, addrs, peerstore.PermanentAddrTTL)
	}
	c.routeTable.Update(pid)
	return nil
}

// Get returns the target element
func (c *routeTableCache) Get(k interface{}) interface{} {
	if pid, ok := k.(peer.ID); ok {
		return c.peerStore.Addrs(pid)
	}
	return nil
}

// Take removes and returns the target element in cache.
// this method could be used as a delete
func (c *routeTableCache) Take(k interface{}) (interface{}, error) {
	pid, ok := k.(peer.ID)
	if !ok {
		return nil, cache.ErrInvalidCacheKey
	}
	r := c.peerStore.Addrs(pid)
	c.peerStore.ClearAddrs(pid)
	c.routeTable.Remove(pid)
	return r, nil
}

// Size returns the total number of elements in cache
func (c *routeTableCache) Size() int {
	return c.routeTable.Size()
}

//  k is peer.ID
//  v is []ma.Multiaddr
func (c *routeTableCache) EncodeEntry(k, v interface{}) (*cache.ExportableEntry, error) {

	ret := &cache.ExportableEntry{}
	if k != nil {
		if tk, ok := k.(peer.ID); ok {
			ret.K = tk.Pretty()
		}
	}

	if v != nil {
		if addrs, ok := v.([]ma.Multiaddr); ok {
			lines := make([]string, len(addrs))
			for i, addr := range addrs {
				// line := fmt.Sprintf("%s/ipfs/%s\n", addr, pid.Pretty())
				lines[i] = addr.String()
			}
			ret.V = lines
		}
	}

	if ret.K == nil && ret.V == nil {
		return nil, cache.ErrInvalidEncodeArguments
	}
	return ret, nil
}

//  k is peer.ID
//  v is []ma.Multiaddr
func (c *routeTableCache) DecodeEntry(kv *cache.ExportableEntry) (k, v interface{}, err error) {
	if kv.K != nil {
		// TODO: test decode
		k, err = peer.IDB58Decode(kv.K.(string))
		if err != nil {
			return
		}
	}

	if kv.V != nil {
		lines, ok := kv.V.([]string)
		if !ok {
			return nil, nil, cache.ErrInvalidCacheValue
		}

		addrs := make([]ma.Multiaddr, 0)
		for _, line := range lines {
			addr, err := ma.NewMultiaddr(line)
			if err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"err":  err,
					"line": line,
				}).Error("Invalid multiaddr string.")
				continue
			}
			addrs = append(addrs, addr)
		}
		v = addrs
	}
	return
}
