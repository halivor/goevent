package logger

var glog *logger

func init() {
	glog = NewStdOut("", Lshortfile|Ltime, DEBUG)
	glog.depth = 3
}

func SetPrefix(prefix string) {
	glog.SetPrefix(prefix)
}

func SetFlags(flags int) {
	glog.SetFlags(flags)
}

func SetLevel(level Level) {
	glog.SetLevel(level)
}

func Trace(v ...interface{}) {
	glog.Trace(v...)
}

func Debug(v ...interface{}) {
	glog.Debug(v...)
}

func Info(v ...interface{}) {
	glog.Info(v...)
}

func Warn(v ...interface{}) {
	glog.Warn(v...)
	glog.Flush()
}

// TODO: 优雅使用panic
func Panic(v ...interface{}) {
	glog.Warn(v...)
	glog.FlushAll()
	panic(v)
}

func Flush() {
	glog.Flush()
}

func FlushAll() {
	glog.FlushAll()
}
