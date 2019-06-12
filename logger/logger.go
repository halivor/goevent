package logger

import (
	"os"
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
	Flush()
	FlushAll()
}

var glog *logger

func New(file string, prefix string, flag int, level Level) (l *logger, e error) {
	switch l, e = newLogger(file); {
	case e != nil:
		return nil, e
	default:
		if len(prefix) > 0 {
			l.prefix = prefix + " "
		}
		l.level = level
		l.flag = flag
		l.depth = 2
	}

	return l, nil
}

func NewStdOut(prefix string, flag int, level Level) *logger {
	locker.Lock()
	l := freeList[0]
	freeList = freeList[1:]
	locker.Unlock()

	if len(prefix) > 0 {
		l.prefix = prefix + " "
	}
	l.level = level
	l.flag = flag
	l.depth = 2

	l.Writer = os.Stdout
	return l
}

func Release(l Logger) { // TODO: 放到 for/select 内部
	locker.Lock()
	if lg, ok := l.(*logger); ok {
		freeList = append(freeList, lg)
		lg.FileInfo = nil
		lg.Writer = nil
		if lg.Writer != os.Stdout {
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
		}
	}
	locker.Unlock()
	return
}

func StdOutDebug() {
	stdout = true
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
	glog.Flush()
}

// TODO: 优雅使用panic
func Panic(v ...interface{}) {
	glog.Warn(v)
	glog.FlushAll()
	panic(v)
}

func Flush() {
	glog.Flush()
}

func FlushAll() {
	glog.FlushAll()
}
