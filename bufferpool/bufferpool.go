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
	memType     = 6
)

type BufferPool interface {
	Alloc(length int) ([]byte, error)
	Release(buffer []byte)
}

// bufferpool =
//    128 * 8000 * 8 = 8M
//    512 * 8000 * 2 = 8M
//   1024 * 8000 * 1 = 8M
//   2048 * 4000     = 8M
//   4096 * 2000     = 8M
//   8192 * 1000     = 8M
var memSize [memType]int = [memType]int{128, 512, 1024, 2048, 4096, 8192}
var memCnt [memType]int = [memType]int{8000 * 8, 8000 * 2, 8000 * 1, 4000, 2000, 1000}

// 考虑array和map性能差距
// 内存分配应该是一个调用频率极高的模块
type bufferpool struct {
	locker   uint32
	memCache map[int][]unsafe.Pointer
	memSlice map[int][]unsafe.Pointer
	memCap   map[int]int
	memRef   map[unsafe.Pointer]int
}

var (
	gpool [][]byte
)

func init() {
	gpool = make([][]byte, 128)
}

func New() *bufferpool {
	bp := &bufferpool{
		locker:   0,
		memCache: make(map[int][]unsafe.Pointer, 128),
		memSlice: make(map[int][]unsafe.Pointer, 128), // 可以改
		memCap:   make(map[int]int),
		memRef:   make(map[unsafe.Pointer]int, 8192),
	}
	for i, size := range memSize {
		bp.memSlice[size] = make([]unsafe.Pointer, memCnt[i]*10)
		bp.allocMemory(size, memCnt[i])
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
	for i := 0; i < memType; i++ {
		if size := memSize[i]; length <= size {
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
					bp.allocMemory(memCnt[i], size)
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
