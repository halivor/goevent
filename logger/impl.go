package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	bp "github.com/halivor/goutility/bufferpool"
)

type logger struct {
	id     int
	level  Level
	prefix string
	flag   int
	depth  int

	sync.Mutex
	os.FileInfo
	io.Writer
}

func (l *logger) SetLevel(level Level) {
	if level < TRACE || level > WARN {
		return
	}
	l.level = level
}

func (l *logger) SetPrefix(prefix string) {
	if len(prefix) > 0 {
		l.prefix = prefix + " "
	}
}

func (l *logger) SetFlags(flag int) {
	l.flag = flag
}

func (l *logger) Trace(v ...interface{}) {
	if TRACE < l.level {
		return
	}
	l.Output("TRACE ", fmt.Sprintln(v...))
}

func (l *logger) Debug(v ...interface{}) {
	if DEBUG < l.level {
		return
	}
	l.Output("DEBUG ", fmt.Sprintln(v...))
}

func (l *logger) Info(v ...interface{}) {
	if INFO < l.level {
		return
	}
	l.Output("INFO ", fmt.Sprintln(v...))
}

func (l *logger) Warn(v ...interface{}) {
	if WARN < l.level {
		return
	}
	l.Output("WARN ", fmt.Sprintln(v...))
	l.Flush()
}

func (l *logger) Panic(v ...interface{}) {
	if WARN < l.level {
		return
	}
	l.Output("PANIC ", fmt.Sprintln(v...))
	l.FlushAll()
	panic(v)
}

func (l *logger) Flush() {
	chFlush <- l.Writer
}

func (l *logger) FlushAll() {
	chFlush <- nil
	chFlush <- os.Stdin
}

func (l *logger) Output(level, s string) {
	now := time.Now() // get this early.
	var file string
	var line int
	if l.flag&(Lshortfile|Llongfile) != 0 {
		var ok bool
		_, file, line, ok = runtime.Caller(l.depth)
		if !ok {
			file = "???"
			line = 0
		}
	}
	length := len(level) + len(l.prefix) + 128 + len(s)
	buf := bp.Alloc(length)[:0]
	buf = append(buf, l.prefix...)
	buf = append(buf, level...)
	l.formatHeader(&buf, now, level, file, line)
	buf = append(buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		buf = append(buf, '\n')
	}
	chNl <- &nlogs{w: l.Writer, data: buf}
}

func (l *logger) formatHeader(buf *[]byte, t time.Time, level, file string, line int) {
	if l.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if l.flag&LUTC != 0 {
			t = t.UTC()
		}
		if l.flag&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '.')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '.')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
		}
		*buf = append(*buf, ' ')
	}
	if l.flag&(Lshortfile|Llongfile) != 0 {
		if l.flag&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ": "...)
	}
}

func itoa(buf *[]byte, i int, wid int) {
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}
