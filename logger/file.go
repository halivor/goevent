package logger

import (
	"io"
	"os"
	"time"
)

var (
	freeList []*logger
)

func init() {
	freeList = make([]*logger, 0, MAX_LOGGERS*10) // free loggers
	for i := 0; i < MAX_LOGGERS; i++ {
		arrLogger[i].id = i
		freeList = append(freeList, &arrLogger[i])
	}
	slice["freeList"] = freeList
}

func createFile(file string) (*logger, error) {
	fp, e := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if e != nil {
		return nil, e
	}
	fi, e := os.Stat(file)
	if e != nil {
		return nil, e
	}

	lg := freeList[0]
	freeList = freeList[1:]
	lg.FileInfo = fi
	lg.Writer = fp

	if _, ok := mFnFI[fi.Name()]; !ok {
		mFnFI[fi.Name()] = make(map[os.FileInfo]io.Writer, 8)
	}
	mFnFI[fi.Name()][fi] = fp

	if _, ok := mLoggers[lg]; !ok {
		mLoggers[fp] = make(map[*logger]struct{}, 8)
	}
	mLoggers[fp][lg] = struct{}{}

	mFiles[fp] = file
	return lg, e
}

func useFile(file string, fi os.FileInfo) (*logger, error) {
	// 文件已存在，根据打开情况处理
	if efis, ok := mFnFI[fi.Name()]; ok && len(efis) > 0 {
		for efi, w := range efis {
			if os.SameFile(efi, fi) {
				lg := freeList[0]
				freeList = freeList[1:]
				lg.Writer = w
				lg.FileInfo = efi
				mLoggers[w][lg] = struct{}{}
				return lg, nil
			}
		}
	}
	// 若日志文件未打开，直接创建
	return createFile(file)
}

func freeFile() {
}

func reLog() {
	locker.Lock()
	for w, fn := range mFiles {
		if fp, ok := w.(*os.File); ok {
			fp.Close()
			now := time.Now().Format("20060102.150405")
			if fi, e := os.Stat(fn); e == nil && fi.Size() > 1024*1024 {
				os.Rename(fn, fn+"."+now)
			}
			nfp, _ := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			*fp = *nfp
		}
	}
	locker.Unlock()
}

func reOpen() {
	locker.Lock()
	for w, fn := range mFiles {
		if fp, ok := w.(*os.File); ok {
			fp.Close()
			nfp, _ := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			*fp = *nfp
		}
	}
	locker.Unlock()
}
