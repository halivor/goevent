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
	svc service
	mtx sync.Mutex
)

func Register(name string, s service) {
	svc = s
}

func Init(params map[string]interface{}) {
	mtx.Lock()
	defer mtx.Unlock()
	if svc == nil {
		panic("unused service")
	}
	svc.Init(params)
}

func Get(key string) map[string]Value {
	return svc.Get(key)
}

func Put(key, value string) {
	svc.Put(key, value)
}

func Watch(key string) <-chan map[string]Value {
	return svc.Watch(key)
}

func Lock(key string) {
	svc.Lock(key)
}
func Unlock(key string) {
	svc.Unlock(key)
}
