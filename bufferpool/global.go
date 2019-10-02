package bufferpool

import (
	"log"
	"time"
	"unsafe"
)

var gbp *bufferpool
var locker uint32

func init() {
	gbp = New()
	go static()
}

func Alloc(length int) []byte {
	b, e := gbp.Alloc(length)
	if e != nil {
		log.Println(e)
	}
	return b
}

func Realloc(src []byte, length int) []byte {
	dst := Alloc(length)
	copy(dst, src)
	Free(src)
	return dst
}

func Free(buf []byte) {
	gbp.Free(buf)
}

func AllocPointer(length int) unsafe.Pointer {
	ptr, e := gbp.AllocPointer(length)
	if e != nil {
		log.Println(e)
	}
	return ptr

}

func FreePointer(ptr unsafe.Pointer) {
	gbp.FreePointer(ptr)
}

func static() {
	tick := time.NewTicker(time.Second)
	for {
		select {
		case <-tick.C:
			log.Println(3, gbp.memCap[3])
		}
	}
}
