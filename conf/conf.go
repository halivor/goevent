package conf

import (
	"sync"

	s "github.com/halivor/goutil/service"
)

var (
	svc s.Service
	mtx sync.Mutex
)

func Use(name string) {
	if svc = s.Get(name); svc == nil {
		panic("service not exist")
	}
}

func Init(params interface{}) {
	mtx.Lock()
	defer mtx.Unlock()
	svc.Init(params)
}

func Get(key string) map[string]s.Value {
	return svc.Get(key)
}

func Put(key, value string) {
	svc.Put(key, value)
}

func Watch(key string) <-chan map[string]s.Value {
	return svc.Watch(key)
}
