package logger

import (
	"io"
	"log"
	"os"
	"sync"
	"time"

	bp "github.com/halivor/goutility/bufferpool"
)

const (
	INIT_LOGGERS = 4096
	MAX_LOGGERS  = 1024 * 1024
	MAX_LOGS     = 4096
)

type logList struct {
	list [MAX_LOGS]*nlogs
	n    int
}

// 常规日志
type nlogs struct {
	w    io.Writer
	data []byte
}

var (
	locker    sync.RWMutex
	stdout    bool                = false
	arrLogger [MAX_LOGGERS]logger // 所有logger信息
	freeList  []*logger

	chNl    chan *nlogs
	chFlush chan io.Writer
	chReLog chan struct{}

	mFnFI    map[string]map[os.FileInfo]io.Writer // 不同位置的同名文件的文件信息
	mLoggers map[io.Writer]map[*logger]struct{}
	mFile    map[io.Writer]string

	mLogs map[io.Writer]*logList // writer到实际日志的映射
	wSet  map[io.Writer]struct{} // 待写入write
)

func init() {
	chNl = make(chan *nlogs, 1024*1024)
	chFlush = make(chan io.Writer)
	chReLog = make(chan struct{})

	mFnFI = make(map[string]map[os.FileInfo]io.Writer, 1024) // file name -> file info
	mFile = make(map[io.Writer]string, 1024)                 // io.Writer -> file name
	mLoggers = make(map[io.Writer]map[*logger]struct{}, 1024)

	mLogs = make(map[io.Writer]*logList, 256) // write -> log list
	wSet = make(map[io.Writer]struct{}, 128)  // wait to write

	freeList = make([]*logger, 0, MAX_LOGGERS) // free loggers
	for i := 0; i < MAX_LOGGERS; i++ {
		arrLogger[i].id = i
		freeList = append(freeList, &arrLogger[i])
	}
	mLogs[os.Stdout] = &logList{}

	glogInit()
	go write()
}

func glogInit() {
	glog = NewStdOut("", log.Lshortfile|log.Ltime, DEBUG)
	glog.depth = 3
}

func flushAll() {
	for w, _ := range wSet {
		flush(w)
		// syncWrite(w) TODO: 确认实际效果
	}
	wSet = make(map[io.Writer]struct{}, 128)
}

func flush(w io.Writer) {
	if l, ok := mLogs[w]; ok && l.n > 0 {
		for i := 0; i < l.n; i++ {
			if w != nil {
				w.Write(l.list[i].data)
				if stdout && w != os.Stdout {
					os.Stdout.Write(l.list[i].data)
				}
			}
			bp.Release(l.list[i].data)
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

func record(nl *nlogs) {
	wSet[nl.w] = struct{}{}
	switch l, ok := mLogs[nl.w]; {
	case ok && l.n+2 < MAX_LOGS:
		l.list[l.n] = nl
		l.n++
	case ok && l.n+2 == MAX_LOGS:
		l.list[l.n] = nl
		l.n++
		flush(nl.w)
	default:
		l := &logList{}
		mLogs[nl.w] = l
		l.list[l.n] = nl
		l.n++
	}
}

func write() {
	t := time.NewTicker(time.Millisecond * 200)
	for {
		select {
		case w, ok := <-chFlush:
			if !ok {
				chFlush = nil
			}
			switch {
			case w == nil:
				flushAll()
			default:
				flush(w)
			}
		default:
			select {
			case <-t.C:
				flushAll()
			case nl, ok := <-chNl:
				if !ok {
					chNl = nil
				}
				record(nl)
			case w, ok := <-chFlush:
				if !ok {
					chFlush = nil
				}
				switch {
				case w == nil:
					flushAll()
				default:
					flush(w)
				}
			case _, ok := <-chReLog:
				if !ok {
					chReLog = nil
				}
				flushAll()
				reLog() // 切换日志
			}
		}
	}
}

func newLogger(file string) (lg *logger, e error) {
	locker.Lock()
	defer locker.Unlock()

	lg = freeList[0]
	freeList = freeList[1:]

	// TODO: 创建目录
	switch lg.FileInfo, e = os.Stat(file); {
	case e != nil:
		// 文件不存在，直接创建
		var fp *os.File
		if fp, e = os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644); e == nil {
			if lg.FileInfo, e = os.Stat(file); e != nil {
				freeList = append(freeList, lg)
				return nil, e
			}
			lg.Writer = fp
			if _, ok := mFnFI[lg.FileInfo.Name()]; !ok {
				mFnFI[lg.FileInfo.Name()] = make(map[os.FileInfo]io.Writer, 8)
			}
			mFnFI[lg.FileInfo.Name()][lg.FileInfo] = fp

			mFile[fp] = file
			if _, ok := mLoggers[lg]; !ok {
				mLoggers[fp] = make(map[*logger]struct{}, 8)
			}
			mLoggers[fp][lg] = struct{}{}
		}
	case e == nil:
		// 文件已存在，根据打开情况处理
		if efis, ok := mFnFI[lg.FileInfo.Name()]; ok && len(efis) > 0 {
			for efi, w := range efis {
				if os.SameFile(efi, lg.FileInfo) {
					lg.Writer = w
					lg.FileInfo = efi
					if _, ok := mLoggers[w]; !ok {
						mLoggers[w] = make(map[*logger]struct{}, 8)
					}
					mLoggers[w][lg] = struct{}{}
					return lg, nil
				}
			}
		}
		// 若日志文件未打开，直接创建
		if fp, e := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644); e == nil {
			lg.Writer = fp
			lg.FileInfo, e = os.Stat(file)
			if _, ok := mFnFI[lg.FileInfo.Name()]; !ok {
				mFnFI[lg.FileInfo.Name()] = make(map[os.FileInfo]io.Writer, 8)
			}
			mFnFI[lg.FileInfo.Name()][lg.FileInfo] = fp

			mFile[fp] = file
			if _, ok := mLoggers[lg]; !ok {
				mLoggers[fp] = make(map[*logger]struct{}, 8)
			}
			mLoggers[fp][lg] = struct{}{}
		}
	default:
		freeList = append(freeList, lg)
		return nil, os.ErrInvalid
	}

	return lg, nil
}

func ReLog() {
	chReLog <- struct{}{}
}

func reLog() {
	locker.Lock()
	for w, fn := range mFile {
		if fp, ok := w.(*os.File); ok {
			fp.Close()
			now := time.Now().Format("20060102.150405")
			if fn, ok := mFile[fp]; ok {
				if fi, e := os.Stat(fn); e == nil && fi.Size() > 1024*1024 {
					os.Rename(fn, fn+"."+now)
				}
			}
			nfp, _ := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			*fp = *nfp
		}
	}
	locker.Unlock()
}
