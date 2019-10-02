package logger

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	ml := make([]Logger, 1024)
	for i := 0; i < len(ml); i++ {
		ml[i], _ = New("/data/logs/logger.test.log",
			fmt.Sprintf("%04d", i), LstdFlags, TRACE)
	}
	tick := time.NewTicker(time.Microsecond * 50)
	data := ""
	for i := 0; i < 1024; i++ {
		data += "0"
	}
	for {
		select {
		case <-tick.C:
			n := rand.Intn(len(ml))
			ml[n].Debug(data)
		}
	}
}
