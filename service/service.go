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

// TODO: method to interface, when call to conver http/grpc call
type Method func(ctx context.Context, in *cp.Request, opts ...grpc.CallOption) (*cp.Response, error)

type Service interface {
	Init(interface{})
	Get(key string) map[string]Value
	Put(key, val string)
	Watch(key string) <-chan map[string]Value
	Lock(key string)
	Unlock(key string)
	SignUp(Server)
	NewClnt(Server, interface{})
	Call(string, string, proto.Message, proto.Message) ce.Errno
	InCall(string, string, *cp.Request, proto.Message) ce.Errno
}

var (
	msvc = map[string]func() Service{}
	mtx  sync.Mutex
)

func Register(name string, svc func() Service) {
	mtx.Lock()
	defer mtx.Unlock()
	msvc[name] = svc
}

func Get(name string) Service {
	mtx.Lock()
	defer mtx.Unlock()
	return msvc[name]()
}

func New(name string) Service {
	mtx.Lock()
	defer mtx.Unlock()
	return msvc[name]()
}
