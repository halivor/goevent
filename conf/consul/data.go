package consul

import (
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

type kv struct {
	Watch sync.Map // TODO: 在watch中的key可以直接读取
	KVs   sync.Map
	Keys  sync.Map
	Vals  sync.Map
}

func (c *Consul) GetKVs(prefix string) map[string]string {
	if kvs, ok := c.kv.KVs.Load(prefix); ok {
		return kvs.(map[string]string)
	}
	keys, meta, _ := c.KV().List(prefix, &api.QueryOptions{})
	c.Index.Store(prefix, meta.LastIndex)
	kvs := make(map[string]string, len(keys))
	for _, key := range keys {
		kvs[key.Key] = string(key.Value)
	}
	c.kv.KVs.Store(prefix, kvs)
	return kvs
}

func (c *Consul) GetVal(key string) string {
	if val, ok := c.kv.Vals.Load(key); ok {
		return val.(string)
	}

	kv, meta, _ := c.KV().Get(key, nil)
	c.kv.Vals.Store(key, string(kv.Value))
	c.Index.Store(key, meta.LastIndex)
	return string(kv.Value)
}

func (c *Consul) WatchKVs(prefix string) (<-chan map[string]string, chan struct{}) {
	cb := make(chan map[string]string)
	stop := make(chan struct{}, 1)
	go func(key string, ch chan map[string]string, stop chan struct{}) {
		for {
			select {
			case <-stop:
				close(stop)
				close(cb)
				return
			default:
				dst, e := c.watchKey(key)
				if e != nil {
					close(stop)
					close(cb)
					return
				}

				isrc, iok := c.kv.KVs.Load(key)
				src, ok := isrc.(map[string]string)
				if iok && ok && KVsEqual(src, dst) {
					continue
				}
				c.kv.KVs.Store(prefix, dst)
				cb <- dst
			}
		}
	}(prefix, cb, stop)

	return cb, stop
}

func (c *Consul) watchKey(key string) (map[string]string, error) {
	idx := c.getIdx(key)
	kvs, meta, e := c.KV().List(key, &api.QueryOptions{
		WaitIndex: idx, WaitTime: time.Minute,
	})
	if e != nil {
		return nil, e
	}
	dst := make(map[string]string, len(kvs))
	for _, kv := range kvs {
		dst[kv.Key] = string(kv.Value)
	}
	c.Index.Store(key, meta.LastIndex)
	return dst, nil
}

// 考虑废弃
func (c *Consul) GetKeys(prefix string) map[string]struct{} {
	if mks, ok := c.kv.Keys.Load(prefix); ok {
		return mks.(map[string]struct{})
	}
	keys, meta, _ := c.KV().Keys(prefix, "", nil)
	mks := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		mks[key] = struct{}{}
	}
	c.kv.Keys.Store(prefix, mks)
	c.Index.Store(prefix, meta.LastIndex)
	return mks
}

func KVsEqual(src, dst map[string]string) bool {
	for k, vs := range src {
		if vd, ok := dst[k]; !ok || vd != vs {
			return false
		}
	}

	if len(src) == len(dst) {
		return true
	}

	for k, vd := range dst {
		if vs, ok := src[k]; !ok || vs != vd {
			return false
		}
	}
	return true
}
