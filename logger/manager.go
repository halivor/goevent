package logger

import (
	"io"
	"os"
	"sync"
	"time"
	"unsafe"

	bp "github.com/halivor/goutility/bufferpool"
)

type logList struct {
	list [MAX_LOGS][]byte
	n    int
}

// 常规日志
type nlogs struct {
	w    io.Writer
	data []byte
}

var (
	locker    sync.RWMutex
	stdout    bool                   = false
	arrLogger [MAX_LOGGERS]logger    // 所有logger信息
	slice     map[string]interface{} = map[string]interface{}{}

	chNl    chan []byte    = make(chan []byte, 1024*1024)
	chFlush chan io.Writer = make(chan io.Writer)
	chReLog chan struct{}  = make(chan struct{})
	chItvl  chan struct{}  = make(chan struct{})

	mFnFI    map[string]map[os.FileInfo]io.Writer // 不同位置的同名文件的文件信息
	mLoggers map[io.Writer]map[*logger]struct{}   //
	mFiles   map[io.Writer]string                 // 日志切分时使用

	mLogs map[io.Writer]*logList = make(map[io.Writer]*logList, 32) // writer到实际日志的映射
	wSet  map[io.Writer]struct{} = make(map[io.Writer]struct{}, 32) // 待写入write
)

func init() {
	mFnFI = make(map[string]map[os.FileInfo]io.Writer, 1024) // file name -> file info
	mFiles = make(map[io.Writer]string, 1024)                // io.Writer -> file name
	mLoggers = make(map[io.Writer]map[*logger]struct{}, 1024)

	mLogs[os.Stdout] = &logList{}
	go logging()
}

func logging() {
	t := time.NewTicker(flushMsecs)
	for {
		select {
		case w := <-chFlush:
			switch {
			case w == nil:
				flushAll()
			default:
				flush(w)
			}
		case <-t.C:
			flushAll()
		case <-chReLog:
			flushAll()
			reLog()
		default:
			for i := 0; i < 10; i++ {
				select {
				case nl := <-chNl:
					note(nl)
				case <-t.C:
					flushAll()
				case w := <-chFlush:
					switch {
					case w == nil:
						flushAll()
					default:
						flush(w)
					}
				case <-chReLog:
					flushAll()
					reLog()
				case <-chItvl:
					t.Stop()
					t = time.NewTicker(flushMsecs)
				}
			}
		}
	}
}

func note(data []byte) {
	lg := (*logger)((unsafe.Pointer)((*((*uintptr)(unsafe.Pointer(&data[0]))))))
	wSet[lg.Writer] = struct{}{}
	switch ls, ok := mLogs[lg.Writer]; {
	case ok && ls.n+2 < MAX_LOGS:
		ls.list[ls.n] = data
		ls.n++
	case ok && ls.n+2 == MAX_LOGS:
		ls.list[ls.n] = data
		ls.n++
		flush(lg.Writer)
	default:
		l := &logList{}
		l.list[l.n] = data
		l.n++
		mLogs[lg.Writer] = l
	}
}

func flush(w io.Writer) {
	if l, ok := mLogs[w]; ok && l.n > 0 {
		for i := 0; i < l.n; i++ {
			if w != nil {
				w.Write(l.list[i][ID_LEN:])
				if stdout && w != os.Stdout {
					os.Stdout.Write(l.list[i][ID_LEN:])
				}
			}
			bp.Free(l.list[i])
			l.list[i] = nil
		}
		l.n = 0
	}
}

func syncWrite(w io.Writer) {
	if fp, ok := w.(*os.File); ok {
		fp.Sync()
	}
}

func flushAll() {
	for w, _ := range wSet {
		flush(w)
	}
	wSet = make(map[io.Writer]struct{}, 32)
}

func newLogger(file string) (*logger, error) {
	locker.Lock()
	defer locker.Unlock()
	// TODO: 创建目录
	switch fi, e := os.Stat(file); {
	case os.IsNotExist(e):
		return createFile(file)
	case e == nil:
		return useFile(file, fi)
	}
	return nil, os.ErrInvalid
}

func release(lg *logger) {
	chFlush <- lg.Writer
	if freeList = append(freeList, lg); len(freeList) == cap(freeList) {
		if si, ok := slice["freeList"]; ok {
			if list, ok := si.([]*logger); ok {
				copy(list, freeList)
				freeList = list
			}
		}
	}

	if lg.Writer != os.Stdout {
		if lgs, ok := mLoggers[lg.Writer]; ok {
			delete(lgs, lg)    // io.writer -> []*logger
			if len(lgs) == 0 { // io.writer -> close
				delete(mLoggers, lg.Writer) // io.writer -> logger
				delete(mFiles, lg.Writer)   // io.writer -> file name
				// base name -> []file info -> io.writer
				if fw, ok := mFnFI[lg.FileInfo.Name()]; ok {
					delete(fw, lg.FileInfo)
				}
				if fp, ok := lg.Writer.(*os.File); ok {
					fp.Close()
				}
			}
		}
	}
	lg.FileInfo = nil
	lg.Writer = nil
}
