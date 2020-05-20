package etcd

import (
	"context"
	_ "fmt"
	"sync"
	"time"

	api "go.etcd.io/etcd/clientv3"
)

type conn struct {
	*api.Client
}

var mtx sync.Mutex
var c *api.Client

// TODO: 重启现有事件
func New(eps []string) {
	mtx.Lock()
	defer mtx.Unlock()
	if c != nil {
		return
	}

	nc, e := api.New(api.Config{
		Endpoints:   eps,
		DialTimeout: time.Second * 3,
	})
	if e != nil {
		panic(e)
	}
	c = nc
}

func Get(key string) []byte {
	rc, e := c.Get(context.TODO(), key)
	if e != nil {
		return nil
	}

	var data []byte
	for _, kv := range rc.Kvs {
		data = kv.Value
	}
	return data
}

func Watch(key string) <-chan []byte {
	ch := make(chan []byte, 1)
	go func(key string, ch chan []byte) {
		defer close(ch)
		wc := c.Watch(context.Background(), key)
		for {
			select {
			case rc, ok := <-wc:
				if !ok {
					time.Sleep(time.Second * 1)
					wc = c.Watch(context.Background(), key)
					continue
				}
				var latest []byte
				for _, ev := range rc.Events {
					switch { //TODO: 记录一下日志
					case ev.IsCreate():
					case ev.IsModify():
					default: // delete
					}
					latest = ev.Kv.Value
				}
				if invalide(ch) {
					return
				}
				ch <- latest
			}
		}
	}(key, ch)
	return ch
}

func invalide(ch <-chan []byte) bool {
	if len(ch) >= cap(ch) {
		return true
	}
	return false
}
