package conf

import (
	"sync"
)

type service interface {
	Init(map[string]interface{})
	Get(key string) []byte
	Put(key, val string)
	Watch(key string) <-chan map[string][]byte
	Lock(key string)
	Unlock(key string)
}

var (
	s    service
	mtx  sync.Mutex
	svcs = map[string]service{}
)

func Register(name string, svc service) {
	svcs[name] = svc
}

func Use(name string) {
	mtx.Lock()
	defer mtx.Unlock()
	if _, ok := svcs[name]; !ok {
		panic(name + " not exists")
	}
	s = svcs[name]
}

func Init(params map[string]interface{}) {
	mtx.Lock()
	defer mtx.Unlock()
	if s == nil {
		panic("unused service")
	}
	s.Init(params)
}

func Get(key string) []byte {
	return s.Get(key)
}

func Put(key, value string) {
	s.Put(key, value)
}

func Watch(key string) <-chan map[string][]byte {
	return s.Watch(key)
}

func Lock(key string) {
	s.Lock(key)
}
func Unlock(key string) {
	s.Unlock(key)
}
