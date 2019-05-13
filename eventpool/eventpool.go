package eventpool

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"time"
)

type EventPool interface {
	AddEvent(ev Event) error
	ModEvent(ev Event) error
	DelEvent(ev Event) error
	AddTmo(exp Expire)
}

const (
	EP_TIMEOUT = 1000 // 1000ms

	MaxEvents = 128
	maxConns  = 1024 * 1024
)

type eventpool struct {
	fd   int
	ev   []syscall.EpollEvent // 每次被唤醒，最大处理event数
	es   map[int]Event        // pool中的event
	tmo  *minHeap
	stop bool
	*log.Logger
}

// TODO: error => panic
func New() *eventpool {
	return new(nil)
}

func new(epo *eventpool) *eventpool {
	fd, e := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	switch {
	case e == nil && epo == nil:
		epo = &eventpool{
			fd:     fd,
			ev:     make([]syscall.EpollEvent, MaxEvents),
			es:     make(map[int]Event, maxConns),
			tmo:    newHeap(),
			stop:   false,
			Logger: log.New(os.Stderr, fmt.Sprintf("[ep(%d)] ", fd), log.LstdFlags),
		}
	case e == nil:
		epo.fd = fd
		epo.Logger = log.New(os.Stderr, fmt.Sprintf("[ep(%d)] ", fd), log.LstdFlags)
	default:
		// EINVAL (epoll_create1()) Invalid value specified in flags.
		// EMFILE The per-user limit on the number of epoll instances imposed by
		//        /proc/sys/fs/epoll/max_user_instances was encountered.  See
		//        epoll(7) for further details.
		// EMFILE The per-process limit on the number of open file descriptors
		//        has been reached.
		// ENFILE The system-wide limit on the total number of open files has
		//        been reached.
		// ENOMEM There was insufficient memory to create the kernel object.
		panic(e)
	}
	return epo
}

func (ep *eventpool) AddEvent(ev Event) error {
	ep.es[ev.Fd()] = ev
	ep.Println("add event", ev.Fd(), ev.Event())
	switch e := syscall.EpollCtl(ep.fd,
		syscall.EPOLL_CTL_ADD,
		ev.Fd(),
		&syscall.EpollEvent{
			Events: EpsToSys(ev.Event()),
			Fd:     int32(ev.Fd()),
		},
	); e {
	case nil:
	case syscall.EBADF, syscall.EEXIST, syscall.EINVAL, syscall.ELOOP,
		syscall.ENOMEM, syscall.ENOSPC, syscall.EPERM:
		return e
	default:
		// 不应该有其他错误信息
		ep.Println("event add", ev, e)
		return e
	}
	return nil
}

func (ep *eventpool) ModEvent(ev Event) error {
	ep.Println("mod event", ev.Fd(), ev.Event())
	switch e := syscall.EpollCtl(ep.fd,
		syscall.EPOLL_CTL_MOD,
		ev.Fd(),
		&syscall.EpollEvent{
			Events: EpsToSys(ev.Event()),
			Fd:     int32(ev.Fd()),
		},
	); e {
	case nil:
	case syscall.EBADF, syscall.EINVAL, syscall.ELOOP, syscall.ENOENT,
		syscall.ENOSPC, syscall.ENOMEM, syscall.EPERM:
		return e
	default:
		// 理论上不存在其他错误信息
		ep.Println("event mod", ev, e)
		return e
	}
	return nil
}

func (ep *eventpool) DelEvent(ev Event) error {
	ep.Println("del event", ev.Fd(), ev.Event())
	delete(ep.es, ev.Fd())
	switch e := syscall.EpollCtl(ep.fd,
		syscall.EPOLL_CTL_DEL,
		ev.Fd(),
		&syscall.EpollEvent{
			Events: EpsToSys(ev.Event()),
			Fd:     int32(ev.Fd()),
		},
	); e {
	case nil:
	case syscall.EBADF, syscall.EINVAL, syscall.ELOOP, syscall.ENOENT,
		syscall.ENOSPC, syscall.ENOMEM, syscall.EPERM:
		return e
	default:
		// 理论上不存在其他错误信息
		ep.Println("event del", ev, e)
		return e
	}
	return nil
}

func (ep *eventpool) Run() {
	for !ep.stop {
		switch n, e := syscall.EpollWait(ep.fd, ep.ev, ep.tmo.ExpireAfter()); e {
		case syscall.EINTR:
		case nil:
			if n == 0 {
				now := time.Now().UnixNano() / (1000 * 1000)
				for ep.tmo.Top() <= now {
					ep.Println("timeout event")
					exp := ep.tmo.Pop()
					if ev, ok := exp.(Event); ok {
						ev.CallBack(EV_TIMEOUT)
					}
				}
			}
			for i := 0; i < n; i++ {
				fd := int(ep.ev[i].Fd)
				ep.es[fd].CallBack(SysToEps(ep.ev[i].Events))
			}
		default:
			// 理论上不存在，若存在则直接重建
			// EBADF  epfd is not a valid file descriptor.
			// EFAULT The memory area pointed to by events is not accessible with
			//        write permissions.
			// EINVAL epfd is not an epoll file descriptor, or maxevents is less
			//        than or equal to zero.

			ep.Println("epoll wait error", e)
			if e := ep.rebuild(); e != nil {
				ep.Println("ep run failed,", e)
				return
			}
		}
	}
	ep.Println("event pool stop")
}

func (ep *eventpool) Release() {
	for _, ev := range ep.es {
		ev.Release()
	}
}

func (ep *eventpool) rebuild() (e error) {
	defer func() {
		if r := recover(); r != nil {
			ep.Println("rebuild panic", r)
			e = os.ErrInvalid
		}
	}()
	syscall.Close(ep.fd)
	new(ep)
	for _, ev := range ep.es {
		if e := ep.AddEvent(ev); e != nil {
			ev.Release()
		}
	}
	return nil
}

func (ep *eventpool) Stop() {
	ep.stop = true
}

func (ep *eventpool) AddTmo(exp Expire) {
	ep.tmo.Push(exp)
}
