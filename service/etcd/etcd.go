package etcd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	cp "github.com/halivor/common/golang/packet"
	ce "github.com/halivor/common/golang/util/errno"
	svc "github.com/halivor/goutil/service"
	api "go.etcd.io/etcd/clientv3"
	apicc "go.etcd.io/etcd/clientv3/concurrency"
)

type conn struct {
	cc  *api.Client
	rw  sync.RWMutex
	mtx map[string]*apicc.Mutex
	mms map[string]svc.Method
}

func init() {
	svc.Register("etcd", New())
}

func New() *conn {
	return &conn{
		mtx: map[string]*apicc.Mutex{},
		mms: make(map[string]svc.Method, 64),
	}
}

func (c *conn) SetUp(name string, m svc.Method) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.mms[name] = m
	// TODO: 同步服务注册
}

func (c *conn) Call(name string, req proto.Message, rsp proto.Message) ce.Errno {
	c.rw.RLock()
	m, ok := c.mms[name]
	c.rw.RUnlock()
	if !ok {
		return ce.DATA_INVALID
	}
	rst, e := m(context.Background(), cp.NewRequest(req))
	if e != nil {
		fmt.Println(e)
		return ce.SRV_ERR
	}
	proto.Unmarshal(rst.GetBody(), rsp)
	return ce.Errno(rst.Errno)
}

func (c *conn) Init(params interface{}) {
	c.rw.Lock()
	defer c.rw.Unlock()
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
