package logger

import (
	"os"
)

type Logger interface {
	SetFlags(flags int)
	SetLevel(level Level)
	SetPrefix(prefix string)
	Raw(v ...interface{})
	Println(v ...interface{})
	Trace(v ...interface{})
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Panic(v ...interface{})
	Flush()
	FlushAll()
}

func NewLogWithConf(conf *Conf) Logger {
	return NewLog(conf.Name, conf.Prefix,
		StrToFlag(conf.Flag), StrToLvl(conf.Level))
}

func NewLog(file string, prefix string, flag int, level Level) Logger {
	switch l, e := newLogger(gPath + file); {
	case e != nil:
		panic(e)
	default:
		if len(prefix) > 0 {
			l.prefix = prefix + " "
		}
		l.level = level
		l.flag = flag
		l.depth = 2
		return l
	}
}

func New(file string, prefix string, flag int, level Level) (Logger, error) {
	l, e := newLogger(gPath + file)
	switch {
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
	if lg, ok := l.(*logger); ok {
		locker.Lock()
		release(lg)
		locker.Unlock()
	}
	return
}

func StdOutDebug() {
	stdout = true
}

func ReLog() {
	chReLog <- struct{}{}
}
