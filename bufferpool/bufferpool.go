package bufferpool

import (
	_ "log"
	"os"
	"runtime"
	"sync/atomic"
	"unsafe"
)

const (
	BUF_MIN_LEN = 1024
	BUF_MAX_LEN = 4 * 1024 * 1024
)

type BufferPool interface {
	Alloc(length int) ([]byte, error)
	Release(buffer []byte)
}

var arrSize [5]int = [5]int{512, 1024, 2048, 4096, 8192}
var memCnt map[int]int = map[int]int{
	512:  8000 * 4,
	1024: 8000 * 2,
	2048: 8000,
	4096: 4000,
	8192: 2000,
}

// bufferpool =
//    512 * 8000 * 4 = 16M
//   1024 * 8000 * 2 = 16M
//   2048 * 8000     = 16M
//   4096 * 4000     = 16M
//   8192 * 2000     = 16M
type bufferpool struct {
	locker   uint32
	memCache map[int][]unsafe.Pointer
	memSlice map[int][]unsafe.Pointer
	memCap   map[int]int
	memRef   map[unsafe.Pointer]int
}

var (
	gpool  [][]byte
	gslice [][]unsafe.Pointer
)

func init() {
	gpool = make([][]byte, 128)
	gslice = make([][]unsafe.Pointer, 0, 128)
}

func New() *bufferpool {
	bp := &bufferpool{
		locker:   0,
		memCache: make(map[int][]unsafe.Pointer, 128),
		memSlice: make(map[int][]unsafe.Pointer, 128),
		memCap:   make(map[int]int),
		memRef:   make(map[unsafe.Pointer]int, 8192),
	}
	for size, num := range memCnt {
		bp.memSlice[size] = make([]unsafe.Pointer, num*10)
		bp.allocMemory(size, num)
	}
	return bp
}

func (bp *bufferpool) allocMemory(size, num int) {
	pool := make([]byte, size*num)
	gpool = append(gpool, pool)

	var list []unsafe.Pointer
	mc, ok := bp.memCache[size]
	switch {
	case !ok:
		list = bp.memSlice[size][:num]
		bp.memCap[size] = num
	case ok && cap(bp.memSlice[size])-bp.memCap[size] >= num*2:
		// 理论上mc就为0
		copy(bp.memSlice[size], mc)
		list = bp.memSlice[size][len(mc) : len(mc)+num]
		bp.memCap[size] += num
	case ok && cap(bp.memSlice[size])-bp.memCap[size] < num*2:
		// 理论上mc就为0
		bp.memSlice[size] = make([]unsafe.Pointer, (bp.memCap[size]+num)*2)
		copy(bp.memSlice[size], mc)
		list = bp.memSlice[size][len(mc) : len(mc)+num]
		bp.memCap[size] += num
	default:
	}
	bp.memCache[size] = bp.memSlice[size][:len(mc)+len(list)]

	for pre, cur := 0, 1; cur-1 < num; pre, cur = cur*size, cur+1 {
		list[cur-1] = unsafe.Pointer(&pool[pre])
	}
}

func (bp *bufferpool) Alloc(length int) (buf []byte, e error) {
	ptr, e := bp.AllocPointer(length)
	if e != nil {
		return nil, e
	}
	buf = (*((*[BUF_MAX_LEN]byte)(unsafe.Pointer(ptr))))[:length:length]
	return buf, nil
}

func (bp *bufferpool) AllocPointer(length int) (p unsafe.Pointer, e error) {
	defer atomic.StoreUint32(&bp.locker, 0)
	for !atomic.CompareAndSwapUint32(&bp.locker, 0, 1) {
		runtime.Gosched()
	}
	for i := 0; i < len(arrSize); i++ {
		size := arrSize[i]
		if length <= size {
			if mc, ok := bp.memCache[size]; ok {
				switch {
				case len(mc) > 1:
					p = mc[0]
					bp.memCache[size] = mc[1:]
				case len(mc) == 0:
					return nil, os.ErrInvalid
				case len(mc) == 1:
					p = mc[0]
					bp.memCache[size] = mc[1:]
					num := memCnt[size]
					bp.allocMemory(num, size)
				default:
					return nil, os.ErrInvalid
				}
				//log.Println("alloc  ", p)
				bp.memRef[p] = size
				return p, nil
			}
		}
	}
	return nil, os.ErrInvalid
}

func (bp *bufferpool) Release(buf []byte) {
	bp.ReleasePointer(unsafe.Pointer(&buf[0]))
}

func (bp *bufferpool) ReleasePointer(ptr unsafe.Pointer) {
	defer atomic.StoreUint32(&bp.locker, 0)
	for !atomic.CompareAndSwapUint32(&bp.locker, 0, 1) {
		runtime.Gosched()
	}
	if size, ok := bp.memRef[ptr]; ok {
		if cap(bp.memCache[size])-len(bp.memCache[size]) == 0 {
			src := bp.memCache[size]
			dst := bp.memSlice[size][:len(src)]
			copy(dst, src)
			bp.memCache[size] = dst
		}
		bp.memCache[size] = append(bp.memCache[size], ptr)
		delete(bp.memRef, ptr)
		//log.Println("release", ptr)
	}
}
