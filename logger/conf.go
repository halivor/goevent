package logger

import (
	"log"
	"os"
	"strings"
	"time"
	"unsafe"
)

const (
	INIT_LOGGERS = 4096
	MAX_LOGGERS  = 4096
	MAX_LOGS     = 4096
)

type Level uint32

const (
	TRACE Level = 1 + iota
	DEBUG
	INFO
	WARN
)

const (
	Ldate         = log.Ldate
	Ltime         = log.Ltime
	Lmicroseconds = log.Lmicroseconds
	Llongfile     = log.Llongfile
	Lshortfile    = log.Lshortfile
	LUTC          = log.LUTC
	LstdFlags     = Ltime | Lshortfile

	logLen = int(unsafe.Sizeof(logger{}))
)

var (
	strLevel map[string]Level = map[string]Level{
		"TRACE": TRACE,
		"DEBUG": DEBUG,
		"INFO":  INFO,
		"WARN":  WARN,
	}
	strFlag map[string]int = map[string]int{
		"LDATE":         Ldate,
		"LTIME":         Ltime,
		"LMICROSECONDS": Lmicroseconds,
		"LLONGFILE":     Llongfile,
		"LSHORTFILE":    Lshortfile,
		"LUTC":          LUTC,
		"LSTDFLAGS":     LstdFlags,
	}
	gPath      string
	flushMsecs time.Duration = 500 * time.Millisecond
)

func StrToFlag(flags string) int {
	var Flags = 0
	fields := strings.Split(flags, "|")
	for _, field := range fields {
		if flag, ok := strFlag[strings.ToUpper(strings.TrimSpace(field))]; ok {
			Flags |= flag
		}
	}
	return Flags
}

func StrToLvl(level string) Level {
	if lvl, ok := strLevel[strings.ToUpper(strings.TrimSpace(level))]; ok {
		return lvl
	}
	return DEBUG
}

func SetGlobalPath(path string) {
	gPath = path
	if e := os.MkdirAll(gPath, 0755); e != nil {
		panic(e)
	}
}

func SetFlushMsecs(msec int) {
	flushMsecs = time.Duration(msec) * time.Millisecond
	chItvl <- struct{}{}
}

type Conf struct {
	Name   string `json:"name,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Level  string `json:"level,omitempty"`
	Flag   string `json:"flag,omitempty"`
}
