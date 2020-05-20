package consul

import (
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

type stat struct {
	Svcs sync.Map
}

func (c *Consul) GetStat(svc string) {
	if _, ok := c.Svcs.Load(svc); ok {
		return
	}
	hc, meta, _ := c.Health().State("passing", &api.QueryOptions{
		Filter: c.Proj + " in ServiceTags and " + svc + " in ServiceTags",
	})
	c.Index.Store(svc, meta.LastIndex)
	c.stat.Svcs.Store(svc, hc)
}

func (c *Consul) WatchStat(svc string) (<-chan map[string]*api.HealthCheck, chan struct{}) {
	cb := make(chan map[string]*api.HealthCheck)
	stop := make(chan struct{}, 1)
	go func(svc string, ch chan map[string]*api.HealthCheck, stop chan struct{}) {
		for {
			select {
			case <-stop:
				close(stop)
				close(cb)
				return
			default:
				dst, e := c.watchStat(svc)
				if e != nil {
					close(stop)
					close(cb)
					return
				}
				isrc, iok := c.stat.Svcs.Load(svc)
				src, ok := isrc.(map[string]*api.HealthCheck)
				if ok && iok && statEqual(src, dst) {
					continue
				}
				c.stat.Svcs.Store(svc, dst)
				cb <- dst
			}
		}
	}(svc, cb, stop)

	return cb, stop
}

func (c *Consul) watchStat(svc string) (map[string]*api.HealthCheck, error) {
	idx := c.getIdx(svc)
	filter := c.Proj + " in ServiceTags and " + svc + " in ServiceTags"
	ss, meta, e := c.Health().State(svc, &api.QueryOptions{
		WaitIndex: idx,
		WaitTime:  time.Minute,
		Filter:    filter,
	})
	if e != nil {
		return nil, e
	}

	mss := make(map[string]*api.HealthCheck, len(ss))
	for _, svc := range ss {
		mss[svc.ServiceID] = svc
	}

	c.Index.Store(svc, meta.LastIndex)
	return mss, nil
}

func statEqual(src, dst map[string]*api.HealthCheck) bool {
	return false
}
