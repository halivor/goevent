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
	PANIC
)

const (
	Ldate         = log.Ldate
	Ltime         = log.Ltime
	Lmicroseconds = log.Lmicroseconds
	Llongfile     = log.Llongfile
	Lshortfile    = log.Lshortfile
	LUTC          = log.LUTC
	LstdFlags     = Ldate | Ltime
)

type Logger interface {
	SetFlags(flags int)
	SetLevel(level Level)
	SetPrefix(prefix string)
	Trace(v ...interface{})
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Panic(v ...interface{})
}

var glog *logger

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
	l.depth = 2
	l.Logger = log.New(l, prefix, flag)
	l.Writer = w
	log.Println(len(mFile), len(mWriter), len(mLogs))

	return l, nil
}

func NewStdOut(prefix string, flag int, level Level) *logger {
	locker.Lock()
	l := freeList[0]
	freeList = freeList[1:]
	locker.Unlock()

	l.level = level
	l.depth = 2
	l.Logger = log.New(l, prefix, flag)
	l.Writer = os.Stdout
	return l
}

func Release(l Logger) {
}

func SetPrefix(prefix string) {
	glog.SetPrefix(prefix)
}

func SetFlags(flags int) {
	glog.SetFlags(flags)
}

func SetLevel(level Level) {
	glog.SetLevel(level)
}

func Trace(v ...interface{}) {
	glog.Trace(v)
}

func Debug(v ...interface{}) {
	glog.Debug(v)
}

func Info(v ...interface{}) {
	glog.Info(v)
}

func Warn(v ...interface{}) {
	glog.Warn(v)
}

func Panic(v ...interface{}) {
	glog.Panic(v)
}
