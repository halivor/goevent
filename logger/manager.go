package logger

import (
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	bp "github.com/halivor/goutility/bufferpool"
)

const (
	INIT_LOGGERS = 4096
	MAX_LOGGERS  = 4096
	MAX_LOGS     = 4096
)

type logList struct {
	list [4096]*nlogs
	n    int
}

// 常规日志
type nlogs struct {
	w    io.Writer
	data []byte
}

// 致命日志
type elogs struct {
	w    io.Writer
	data []byte
}

var (
	locker    sync.RWMutex
	arrLogger [MAX_LOGGERS]logger // 所有logger信息
	freeList  []*logger

	chn chan *nlogs
	chw chan *elogs
	chs chan os.Signal

	mFnFI    map[string]map[os.FileInfo]io.Writer // 不同位置的同名文件的文件信息
	mFile    map[io.Writer]string
	mLoggers map[io.Writer]map[*logger]struct{}

	mLogs map[io.Writer]*logList // writer到实际日志的映射
	wSet  map[io.Writer]struct{} // 待写入write
)

func init() {
	chn = make(chan *nlogs, 1024)
	chw = make(chan *elogs, 1024)
	chs = make(chan os.Signal, 1024)
	signal.Notify(chs, syscall.SIGUSR1, syscall.SIGUSR2)

	mFnFI = make(map[string]map[os.FileInfo]io.Writer, 1024) // file name -> file info
	mFile = make(map[io.Writer]string, 1024)                 // io.Writer -> file name
	mLoggers = make(map[io.Writer]map[*logger]struct{})

	mLogs = make(map[io.Writer]*logList)     // write -> log list
	wSet = make(map[io.Writer]struct{}, 128) // wait to write

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
	}
	wSet = make(map[io.Writer]struct{}, 128)
}

func flush(w io.Writer) {
	if l, ok := mLogs[w]; ok {
		for i := 0; i < l.n; i++ {
			w.Write(l.list[i].data)
			bp.Release(l.list[i].data)
			l.list[i] = nil
		}
		l.n = 0
	}
}

func write() {
	tn := time.NewTicker(time.Millisecond * 50)
	for {
		select {
		case <-tn.C:
			flushAll()
		case nl, ok := <-chn:
			if !ok {
				chn = nil
			}
			wSet[nl.w] = struct{}{}
			switch l, ok := mLogs[nl.w]; {
			case ok && l.n+2 < MAX_LOGS:
				l.list[l.n] = nl
				l.n++
			case ok && l.n+2 == MAX_LOGS:
				l.list[l.n] = nl
				l.n++
				flush(nl.w)
			}
			// 缓存达到量级后清理
		case <-chw:
			// 立即清空对应writer的日志
		case <-chs:
			// 重写日志
			flushAll()
			reLog()
		}
	}
}

// TODO: error
func newLogger(file string) (lg *logger, e error) {
	locker.Lock()
	defer locker.Unlock()

	lg = freeList[0]
	freeList = freeList[1:]
	lg.fileName = file

	switch lg.FileInfo, e = os.Stat(file); {
	case e != nil:
		// 文件不存在，直接创建
		var fp *os.File
		if fp, e = os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644); e == nil {
			lg.FileInfo, e = os.Stat(file)
			mLogs[fp] = &logList{}
			if _, ok := mFnFI[lg.FileInfo.Name()]; !ok {
				mFnFI[lg.FileInfo.Name()] = make(map[os.FileInfo]io.Writer, 8)
			}
			mFnFI[lg.FileInfo.Name()][lg.FileInfo] = fp
			mFile[fp] = file
			lg.Writer = fp
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
					if _, ok := mLoggers[w]; !ok {
						mLoggers[w] = make(map[*logger]struct{}, 8)
					}
					mLoggers[w][lg] = struct{}{}
					return
				}
			}
		}

		// 若日志文件未打开，直接创建
		if fp, e := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644); e == nil {
			mLogs[fp] = &logList{}
			if _, ok := mFnFI[lg.FileInfo.Name()]; !ok {
				mFnFI[lg.FileInfo.Name()] = make(map[os.FileInfo]io.Writer, 8)
			}
			mFnFI[lg.FileInfo.Name()][lg.FileInfo] = fp
			mFile[fp] = file
			lg.Writer = fp
			if _, ok := mLoggers[lg]; !ok {
				mLoggers[fp] = make(map[*logger]struct{}, 8)
			}
			mLoggers[fp][lg] = struct{}{}
		}
	default:
	}

	return
}

func ReLog() {
}

func reLog() {
	locker.Lock()
	defer locker.Unlock()
	for w, fn := range mFile {
		if fp, ok := w.(*os.File); ok {
			fp.Close()
			now := time.Now().Format("20060102.150405")
			if fn, ok := mFile[fp]; ok {
				if _, e := os.Stat(fn); e == nil {
					os.Rename(fn, fn+"."+now)
				}
			}
			nfp, _ := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			*fp = *nfp
		}
	}
}
