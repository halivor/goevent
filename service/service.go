package service

import (
	"context"
	"sync"

	cp "co.mplat.com/packet"
	ce "co.mplat.com/util/errno"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
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
	Call(string, string, int64, proto.Message, proto.Message) ce.Errno
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
