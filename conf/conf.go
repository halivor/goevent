package conf

import (
	"sync"

	s "github.com/halivor/goutil/service"
)

var (
	sc  s.Service
	mtx sync.Mutex
)

func Use(name string) {
	if sc = s.Get(name); sc == nil {
		panic("service not exist")
	}
}

func Init(params interface{}) {
	mtx.Lock()
	defer mtx.Unlock()
	sc.Init(params)
}

func Get(key string) map[string]s.Value {
	mtx.Lock()
	defer mtx.Unlock()
	return sc.Get(key)
}

func Put(key, value string) {
	sc.Put(key, value)
}

func Watch(key string) <-chan map[string]s.Value {
	return sc.Watch(key)
}
