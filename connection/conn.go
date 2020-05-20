package connection

import (
	"fmt"
	"syscall"

	log "github.com/halivor/goutil/logger"
)

type Conn interface {
	Fd() int
	Send(message []byte) (int, error)
	Recv(buf []byte) (int, error)
	Close()
}

func NewConn(fd int) Conn {
	SetSndBuf(fd, DEFAULT_BUFFER_SIZE)
	SetRcvBuf(fd, DEFAULT_BUFFER_SIZE)
	prefix := fmt.Sprintf("[sock(%d)] ", fd)
	return &c{
		fd:     fd,
		ss:     ESTAB,
		Logger: log.NewLog("sock.log", prefix, log.LstdFlags, log.WARN),
	}
}

func NewTcpConn() (*c, error) {
	fd, e := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if e != nil {
		return nil, e
	}
	SetSndBuf(fd, DEFAULT_BUFFER_SIZE)
	SetRcvBuf(fd, DEFAULT_BUFFER_SIZE)
	prefix := fmt.Sprintf("[tcp(%d)] ", fd)
	return &c{
		fd:     fd,
		ss:     CREATE,
		Logger: log.NewLog("tcp.log", prefix, log.LstdFlags, log.WARN),
	}, nil
}
