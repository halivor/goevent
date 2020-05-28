package service

import (
	"context"
	"sync"

	"github.com/golang/protobuf/proto"
	cp "github.com/halivor/common/golang/packet"
	ce "github.com/halivor/common/golang/util/errno"
	"google.golang.org/grpc"
)

type EventType int32

//type Method func(context.Context, *cp.Request) (*cp.Response, error)
type Method func(ctx context.Context, in *cp.Request, opts ...grpc.CallOption) (*cp.Response, error)
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
	SetUp(name string, m Method)
	Call(name string, req proto.Message, rsp proto.Message) ce.Errno
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
