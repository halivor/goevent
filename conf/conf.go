package conf

import (
	"sync"

	us "github.com/halivor/goutil/service"
)

var (
	svc us.Service
	mtx sync.Mutex
)

func Use(name string) {
	if svc = us.Get(name); svc == nil {
		panic("service not exist")
	}
}

func Init(params interface{}) {
	mtx.Lock()
	defer mtx.Unlock()
	svc.Init(params)
}

func Get(key string) map[string]us.Value {
	mtx.Lock()
	defer mtx.Unlock()
	return svc.Get(key)
}

func Put(key, value string) {
	svc.Put(key, value)
}

func Watch(key string) <-chan map[string]us.Value {
	return svc.Watch(key)
}
