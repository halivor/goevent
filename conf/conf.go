package conf

import (
	"sync"
)

type EventType int32

type Value interface {
	Event() EventType
	Data() []byte
	String() string
}

const (
	EVENT_ADD EventType = 1 << iota
	EVENT_MOD
	EVENT_DEL
)

type service interface {
	Init(map[string]interface{})
	Get(key string) map[string]Value
	Put(key, val string)
	Watch(key string) <-chan map[string]Value
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

func Get(key string) map[string]Value {
	return s.Get(key)
}

func Put(key, value string) {
	s.Put(key, value)
}

func Watch(key string) <-chan map[string]Value {
	return s.Watch(key)
}

func Lock(key string) {
	s.Lock(key)
}
func Unlock(key string) {
	s.Unlock(key)
}
