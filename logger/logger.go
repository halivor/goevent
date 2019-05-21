package logger

import (
	"log"
	"os"
	"unsafe"
)

type Level uint32

const (
	TRACE Level = 1 + iota
	DEBUG
	INFO
	WARN
	FATAL
)

type Logger interface {
	SetFlags(flags int)
	SetLevel(level Level)
	SetPrefix(prefix string)
	Trace(v ...interface{})
	Traceln(v ...interface{})
	Debug(v ...interface{})
	Debugln(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Fatal(v ...interface{})
}

const (
	logLen = int(unsafe.Sizeof(logger{}))
)

func NewFileLog(file string, prefix string, flag int, level Level) (*logger, error) {
	locker.Lock()
	l := freeList[0]
	freeList = freeList[1:]
	locker.Unlock()

	w, _ := newFile(file)
	l.level = level
	l.Logger = log.New(l, prefix, flag)
	l.Writer = w
	log.Println(len(mFile), len(mWriter), len(mLogs))

	return l, nil
}

func NewStdOut(prefix string, flag int, level Level) Logger {
	locker.Lock()
	l := freeList[0]
	freeList = freeList[1:]
	locker.Unlock()

	l.level = level
	l.Logger = log.New(l, prefix, flag)
	l.Writer = os.Stdout
	return l
}

func Release(l Logger) {
}
