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

	EV_TIMEOUT
	EV_PERSIST
	EV_END
)

type Event interface {
	Fd() int
	CallBack(ev EP_EVENT)
	Event() EP_EVENT
	Release()
}

var evStr map[EP_EVENT]string = map[EP_EVENT]string{
	EV_READ:  "in ",
	EV_WRITE: "out ",
	EV_ERROR: "err ",
	EV_EDGE:  "edge ",

	EV_TIMEOUT: "timeout ",
	EV_PERSIST: "persist ",
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

var esArr [32]int32 = [32]int32{
	EV_READ:  syscall.EPOLLIN,
	EV_WRITE: syscall.EPOLLOUT,
	EV_ERROR: syscall.EPOLLERR,
	EV_EDGE:  syscall.EPOLLET,
}

func (t EP_EVENT) String() string {
	s := []byte("ev ")
	for k, v := range evStr {
		if k&t != 0 {
			s = append(s, []byte(v)...)
		}
	}
	if len(s) == 3 {
		return "no such event"
	}
	return string(s)
}

func EpsToSys(ev EP_EVENT) uint32 {
	var ses uint32 = 0
	for i := uint32(0); i < 4; i++ {
		if v := esArr[ev&(1<<i)]; v != 0 {
			ses |= uint32(v)
		}
	}
	/*for k, v := range esMap {*/
	//if ev&k != 0 {
	//ses |= uint32(v)
	//}
	/*}*/
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
