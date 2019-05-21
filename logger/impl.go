package logger

import (
	"io"
	"log"

	bp "github.com/halivor/goutility/bufferpool"
)

type logger struct {
	id    int
	level Level
	*log.Logger
	io.Writer
}

func (l *logger) Write(pb []byte) (int, error) {
	data := bp.Alloc(len(pb))
	copy(data, pb)
	chn <- &nlogs{w: l.Writer, data: data}
	return len(pb), nil
}

func (l *logger) SetLevel(level Level) {
	if level < TRACE || level > FATAL {
		return
	}
	l.level = level
}

func (l *logger) SetPrefix(prefix string) {
	l.SetPrefix(prefix)
}

func (l *logger) SetFlags(flag int) {
	l.SetFlags(flag)
}

func (l *logger) Traceln(v ...interface{}) {
	if TRACE < l.level {
		return
	}
	l.Println(v...)
}

func (l *logger) Trace(v ...interface{}) {
	if TRACE < l.level {
		return
	}
	l.Print(v...)
}
func (l *logger) Debugln(v ...interface{}) {
	if DEBUG < l.level {
		return
	}
	l.Println(v)
}
func (l *logger) Debug(v ...interface{}) {
	if DEBUG < l.level {
		return
	}
	l.Print(v)
}

func (l *logger) Infoln(v ...interface{}) {
	if INFO < l.level {
		return
	}
	l.Println(v...)
}

func (l *logger) Info(v ...interface{}) {
	if INFO < l.level {
		return
	}
	l.Print(v...)
}

func (l *logger) Warnln(v ...interface{}) {
	if WARN < l.level {
		return
	}
	l.Println(v)
}

func (l *logger) Warn(v ...interface{}) {
	if WARN < l.level {
		return
	}
	l.Print(v)
}

func (l *logger) Fatalln(v ...interface{}) {
	l.Println(v)
}

func (l *logger) Fatal(v ...interface{}) {
	l.Print(v)
}
