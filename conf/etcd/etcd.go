package etcd

import (
	"context"
	_ "fmt"
	"sync"
	"time"

	"github.com/halivor/goutil/conf"
	api "go.etcd.io/etcd/clientv3"
	apicc "go.etcd.io/etcd/clientv3/concurrency"
)

type conn struct {
	cc  *api.Client
	mtx map[string]*apicc.Mutex
}

var mtx sync.Mutex

func init() {
	conf.Register("etcd", &conn{mtx: map[string]*apicc.Mutex{}})
}

func (c *conn) Init(params map[string]interface{}) {
	mtx.Lock()
	defer mtx.Unlock()
	if c.cc != nil {
		return
	}

	eps := params["hosts"].([]string)
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

func (c *conn) Get(key string) []byte {
	rc, e := c.cc.Get(context.TODO(), key)
	if e != nil {
		return nil
	}

	var data []byte
	for _, kv := range rc.Kvs {
		data = kv.Value
	}
	return data
}

func (c *conn) Watch(key string) <-chan map[string][]byte {
	ch := make(chan map[string][]byte, 1)
	go func(key string, ch chan map[string][]byte) {
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
				data := map[string][]byte{}
				for _, ev := range rc.Events {
					switch { //TODO: 记录一下日志
					case ev.IsCreate():
					case ev.IsModify():
					default: // delete
					}
					data[string(ev.Kv.Key)] = ev.Kv.Value
				}
				if c.invalide(ch) {
					return
				}
				ch <- data
			}
		}
	}(key, ch)
	return ch
}

func (c *conn) invalide(ch <-chan map[string][]byte) bool {
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
