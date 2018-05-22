package core

import (
	"github.com/nebulasio/go-nebulas/cache"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

type transactionPoolCache map[byteutils.HexHash]*Transaction

func newTransactionPoolCache(conf *cache.PersistableCacheConfig) cache.Cache {
	c := make(transactionPoolCache)
	return cache.NewPersistableCache(c, conf)
}

// All ..
func (p transactionPoolCache) Entries() interface{} {
	ret := make([]interface{}, len(p))
	i := 0
	for _, v := range p {
		ret[i] = v
		i++
	}
	return ret
}

// Keys ..
func (p transactionPoolCache) Keys() interface{} {
	ret := make([]byteutils.HexHash, len(p))
	i := 0
	for k := range p {
		ret[i] = k
		i++
	}
	return ret
}

// Set ...
func (p transactionPoolCache) Set(k, v interface{}) error {
	h, ok := k.(byteutils.HexHash)
	if !ok || h == "" {
		return cache.ErrInvalidCacheKey
	}
	tx, ok := v.(*Transaction)
	if !ok || tx == nil {
		return cache.ErrInvalidCacheValue
	}
	p[h] = tx
	return nil
}

// Get ...
func (p transactionPoolCache) Get(k interface{}) interface{} {
	if h, ok := k.(byteutils.HexHash); ok {
		return p[h]
	}
	return nil
}

// Take ..
func (p transactionPoolCache) Take(k interface{}) (interface{}, error) {
	h, ok := k.(byteutils.HexHash)
	if !ok {
		return nil, cache.ErrInvalidCacheKey
	}
	ret := p[h]
	delete(p, h)
	return ret, nil
}

// Size ..
func (p transactionPoolCache) Size() int {
	return len(p)
}

func (p transactionPoolCache) EncodeEntry(k, v interface{}) (*cache.ExportableEntry, error) {

	ret := &cache.ExportableEntry{}
	if k != nil {
		if tk, ok := k.(byteutils.HexHash); ok {
			ret.K = string(tk)
		}
	}

	if v != nil {
		if tx, ok := v.(*Transaction); ok {
			tp, err := tx.ToProto()
			if err != nil {
				return nil, err
			}
			ret.V = tp.(*corepb.Transaction)
		}
	}

	if ret.K == nil && ret.V == nil {
		return nil, cache.ErrInvalidEncodeArguments
	}
	return ret, nil
}

func (p transactionPoolCache) DecodeEntry(e *cache.ExportableEntry) (k, v interface{}, err error) {

	if e.K != nil {
		if s, ok := e.K.(string); ok {
			k = byteutils.HexHash(s)
		} else {
			return nil, nil, cache.ErrInvalidCacheKey
		}
	}

	if e.V != nil {
		if tp, ok := e.V.(*corepb.Transaction); ok {
			tx := new(Transaction)
			err := tx.FromProto(tp)
			if err != nil {
				return nil, nil, err
			}
			v = tx
		} else {
			return nil, nil, cache.ErrInvalidCacheValue
		}
	}
	return
}
