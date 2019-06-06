package logger

import (
	"log"
	"strings"
	"unsafe"
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
		"Ldate":         Ldate,
		"Ltime":         Ltime,
		"Lmicroseconds": Lmicroseconds,
		"Llongfile":     Llongfile,
		"Lshortfile":    Lshortfile,
		"LUTC":          LUTC,
		"LstdFlags":     LstdFlags,
	}
)

func StrToFlag(flags string) int {
	var Flags = 0
	fields := strings.Split(flags, "|")
	for _, field := range fields {
		if flag, ok := strFlag[strings.TrimSpace(field)]; ok {
			Flags |= flag
		}
	}
	return Flags
}

func StrToLvl(level string) Level {
	if lvl, ok := strLevel[level]; ok {
		return lvl
	}
	return DEBUG
}
