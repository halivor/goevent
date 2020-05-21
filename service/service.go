package service

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

type Service interface {
	Init(interface{})
	Get(key string) map[string]Value
	Put(key, val string)
	Watch(key string) <-chan map[string]Value
	Lock(key string)
	Unlock(key string)
}

var (
	msvc = map[string]Service{}
	mtx  sync.Mutex
)

func Add(key string, api map[string]string) {
}

func Register(name string, svc Service) {
	mtx.Lock()
	defer mtx.Unlock()
	msvc[name] = svc
}

func Get(name string) Service {
	mtx.Lock()
	defer mtx.Unlock()
	return msvc[name]
}
