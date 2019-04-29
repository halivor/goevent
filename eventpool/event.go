package eventpool

import (
	"syscall"
)

type EP_EVENT int32

const (
	EV_READ EP_EVENT = 1 << iota
	EV_WRITE
	EV_ERROR
	EV_EDGE
)

var evStr map[EP_EVENT]string = map[EP_EVENT]string{
	EV_READ:  "ev in",
	EV_WRITE: "ev out",
	EV_ERROR: "ev err",
	EV_EDGE:  "ev et",
}
var seMap map[int32]EP_EVENT = map[int32]EP_EVENT{
	syscall.EPOLLIN:  EV_READ,
	syscall.EPOLLOUT: EV_WRITE,
	syscall.EPOLLERR: EV_ERROR,
	syscall.EPOLLET:  EV_EDGE,
}

// TODO: 改用数组处理
var esMap map[EP_EVENT]int32 = map[EP_EVENT]int32{
	EV_READ:  syscall.EPOLLIN,
	EV_WRITE: syscall.EPOLLOUT,
	EV_ERROR: syscall.EPOLLERR,
	EV_EDGE:  syscall.EPOLLET,
}

type Event interface {
	Fd() int
	CallBack(ev EP_EVENT)
	Event() EP_EVENT
	Release()
}

func (t EP_EVENT) String() string {
	if s, ok := evStr[t]; ok {
		return s
	}
	return "no such event"
}

func EpsToSys(ev EP_EVENT) uint32 {
	var ses uint32 = 0
	for k, v := range esMap {
		if ev&k != 0 {
			ses |= uint32(v)
		}
	}
	return ses
}

func SysToEps(ev uint32) EP_EVENT {
	var es EP_EVENT
	for k, v := range seMap {
		if ev&uint32(k) != 0 {
			es |= v
		}
	}
	return es
}
