package etcd

import (
	"context"
	_ "fmt"
	"sync"
	"time"

	svc "github.com/halivor/goutil/service"
	api "go.etcd.io/etcd/clientv3"
	apicc "go.etcd.io/etcd/clientv3/concurrency"
)

type conn struct {
	cc  *api.Client
	mtx map[string]*apicc.Mutex
}

var mtx sync.Mutex

func init() {
	svc.Register("etcd", &conn{mtx: map[string]*apicc.Mutex{}})
}

func (c *conn) Init(params interface{}) {
	mtx.Lock()
	defer mtx.Unlock()
	if c.cc != nil {
		return
	}

	eps := params.([]string)
	if len(eps) == 0 {
		eps = []string{"127.0.0.1:2379"}
	}

	var e error
	if c.cc, e = api.New(api.Config{
		Endpoints:   eps,
		DialTimeout: time.Second * 3,
	}); e != nil {
		panic(e)
	}
}

func (c *conn) Get(key string) map[string]svc.Value {
	rc, e := c.cc.Get(context.TODO(), key)
	if e != nil {
		return nil
	}

	mp := map[string]svc.Value{}
	for _, kv := range rc.Kvs {
		mp[string(kv.Key)] = &data{event: svc.EVENT_ADD, kv: kv}
	}
	return mp
}

func (c *conn) Watch(key string) <-chan map[string]svc.Value {
	ch := make(chan map[string]svc.Value, 1)
	go func(key string, ch chan map[string]svc.Value) {
		defer close(ch)
		wc := c.cc.Watch(context.Background(), key)
		for {
			select {
			case rc, ok := <-wc:
				if !ok {
					time.Sleep(time.Second * 1)
					wc = c.cc.Watch(context.Background(), key)
					continue
				}
				md := map[string]svc.Value{}
				for _, ev := range rc.Events {
					evt := svc.EVENT_DEL
					switch { //TODO: 记录一下日志
					case ev.IsCreate():
						evt = svc.EVENT_ADD
					case ev.IsModify():
						evt = svc.EVENT_MOD
					}
					md[string(ev.Kv.Key)] = &data{event: evt, kv: ev.Kv}
				}
				if c.invalide(ch) {
					return
				}
				ch <- md
			}
		}
	}(key, ch)
	return ch
}

func (c *conn) invalide(ch <-chan map[string]svc.Value) bool {
	if len(ch) >= cap(ch) {
		return true
	}
	return false
}

func (c *conn) Put(key, val string) {
	c.cc.Put(context.Background(), key, val)
}

func (c *conn) Lock(key string) {
	ns, _ := apicc.NewSession(c.cc)
	nm := apicc.NewMutex(ns, key)
	c.mtx[key] = nm
	nm.Lock(context.TODO())
}

func (c *conn) Unlock(key string) {
	nm := c.mtx[key]
	nm.Unlock(context.TODO())
	delete(c.mtx, key)
}
