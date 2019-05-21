package logger

import (
	"io"
	"os"
	_ "path"
	"sync"
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
	che chan *elogs

	mFile   map[string][]os.FileInfo  // 不同位置的同名文件的文件信息
	mWriter map[os.FileInfo]io.Writer // 实际文件到writer的映射
	mLogs   map[io.Writer]*logList    // writer到实际日志的映射
	wSet    map[io.Writer]struct{}    // 待写入write
)

func init() {
	chn = make(chan *nlogs, 1024)
	che = make(chan *elogs, 1024)

	mFile = make(map[string][]os.FileInfo, 1024)    // filename -> file info
	mWriter = make(map[os.FileInfo]io.Writer, 1024) // file info -> writer
	mLogs = make(map[io.Writer]*logList)            // write -> log list
	wSet = make(map[io.Writer]struct{}, 128)        // wait to write

	freeList = make([]*logger, 0, MAX_LOGGERS) // free loggers
	for i := 0; i < MAX_LOGGERS; i++ {
		arrLogger[i].id = i
		freeList = append(freeList, &arrLogger[i])
	}
	mLogs[os.Stdout] = &logList{}

	go write()
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
			for w, _ := range wSet {
				flush(w)
			}
			wSet = make(map[io.Writer]struct{}, 128)
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
		case <-che:
			// 立即清空对应writer的日志
		}
	}
}

// TODO: error
func newFile(file string) (w io.Writer, e error) {
	locker.Lock()
	defer locker.Unlock()

	fi, e := os.Stat(file)
	if e == nil {
		if efis, ok := mFile[fi.Name()]; ok {
			for _, efi := range efis {
				if w, ok := mWriter[efi]; os.SameFile(efi, fi) && ok {
					return w, nil
				}
			}
		}
	}
	if w, e = os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644); e == nil {
		mLogs[w] = &logList{}
		if _, ok := mFile[fi.Name()]; !ok {
			mFile[fi.Name()] = make([]os.FileInfo, 0, 1024)
		}
		mFile[fi.Name()] = append(mFile[fi.Name()], fi)
		mWriter[fi] = w
	}
	return
}
