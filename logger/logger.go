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
	Traceln(v ...interface{})
	Debug(v ...interface{})
	Debugln(v ...interface{})
	Info(v ...interface{})
	Infoln(v ...interface{})
	Warn(v ...interface{})
	Warnln(v ...interface{})
}

var glog *logger

const (
	logLen = int(unsafe.Sizeof(logger{}))
)

func NewFileLog(file string, prefix string, flag int, level Level) (l *logger, e error) {
	switch l, e = newLogger(file); {
	case e != nil:
		log.Panicln(e)
		return nil, e
	default:
		l.level = level
		l.depth = 2
		l.Logger = log.New(l, prefix, flag)
		//log.Println(len(mFnFI), len(mLogs))
	}

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

func Release(l Logger) { // TODO: 放到 for/select 内部
	locker.Lock()
	defer locker.Unlock()
	switch lg, ok := l.(*logger); {
	case ok && lg.Writer == os.Stdout:
		lg.FileInfo = nil
		lg.Writer = nil
		lg.Logger = nil
		freeList = append(freeList, lg)
	case ok:
		if lgs, ok := mLoggers[lg.Writer]; ok {
			delete(lgs, lg)    // io.writer -> []*logger
			if len(lgs) == 0 { // io.writer -> close
				// base name -> []file info -> io.writer
				if fw, ok := mFnFI[lg.FileInfo.Name()]; ok {
					delete(fw, lg.FileInfo)
					if len(fw) == 0 {
						delete(mFnFI, lg.FileInfo.Name())
					}
				}
				delete(mLoggers, lg.Writer) // io.writer -> logger
				delete(mFile, lg.Writer)    // io.writer -> file name

				chFlush <- lg.Writer
				if fp, ok := lg.Writer.(*os.File); ok {
					fp.Close()
				}
			}
		}

		lg.Logger = nil
		lg.FileInfo = nil
		lg.Writer = nil
		freeList = append(freeList, lg)
	}
	return
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
	glog.Traceln(v)
}

func Debug(v ...interface{}) {
	glog.Debug(v)
}

func Info(v ...interface{}) {
	glog.Infoln(v)
}

func Warn(v ...interface{}) {
	glog.Warnln(v)
}
