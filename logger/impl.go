package logger

import (
	"fmt"
	"io"
	"log"

	bp "github.com/halivor/goutility/bufferpool"
)

type logger struct {
	id    int
	level Level
	depth int
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
	if level < TRACE || level > PANIC {
		return
	}
	l.level = level
}

func (l *logger) SetPrefix(prefix string) {
	l.Logger.SetPrefix(prefix)
}

func (l *logger) SetFlags(flag int) {
	l.Logger.SetFlags(flag)
}

func (l *logger) Trace(v ...interface{}) {
	if TRACE < l.level {
		return
	}
	l.Output(l.depth, fmt.Sprintln(v...))
}

func (l *logger) Debug(v ...interface{}) {
	if DEBUG < l.level {
		return
	}
	l.Output(l.depth, fmt.Sprintln(v...))
}

func (l *logger) Info(v ...interface{}) {
	if INFO < l.level {
		return
	}
	l.Output(l.depth, fmt.Sprintln(v...))
}

func (l *logger) Warn(v ...interface{}) {
	if WARN < l.level {
		return
	}
	l.Output(l.depth, fmt.Sprintln(v...))
}

func (l *logger) Panic(v ...interface{}) {
	// TODO: 利用内存池重写output
	l.Output(l.depth, fmt.Sprintln(v...))
	panic(v)
}
