package bufferpool

import (
	"log"
	"os"
	"runtime"
	"sync/atomic"
	"unsafe"
)

const (
	BUF_MIN_LEN = 1024
	BUF_MAX_LEN = 4 * 1024 * 1024
	MEM_SIZE    = 6
)

type BufferPool interface {
	Alloc(length int) ([]byte, error)
	Release(buffer []byte)
}

// bufferpool =
//     128 * 8000 * 8 = 8M
//     512 * 8000 * 2 = 8M
//    1024 * 8000 * 1 = 8M
//    2048 * 4000     = 8M
//    4096 * 2000     = 8M
//    8192 * 1000     = 8M
var memSize [MEM_SIZE]int = [MEM_SIZE]int{128, 512, 1024, 2048, 4096, 8192}
var memCnt [MEM_SIZE]int = [MEM_SIZE]int{8000 * 8, 8000 * 2, 8000 * 1, 4000, 2000, 1000}
var memSizeIdx map[int]int = map[int]int{ // TODO: 暂时没用，考虑丢弃
	128: 0, 512: 1, 1024: 2, 2048: 3, 4096: 4, 8192: 5,
}

/*
 * array / map ~= 30 : 1
 * assignment performace
 * num            array    map
 * 1000*1000      385us    11788   us
 * 1000*1000*1000 349899us 11776979us
**/
// TODO: 考虑对每个map使用lock，不再使用全局lock
type bufferpool struct {
	locker        uint32
	memLocker     [MEM_SIZE]*uint32
	memCache      [MEM_SIZE][]unsafe.Pointer
	cFirst, cLast int
	cEnd          int
	memSlice      [MEM_SIZE][]unsafe.Pointer
	memCap        [MEM_SIZE]int
	memRef        map[unsafe.Pointer]int
}

var (
	gpool [][]byte
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	gpool = make([][]byte, 128)
}

func New() *bufferpool {
	bp := &bufferpool{
		locker: 0,
		memRef: make(map[unsafe.Pointer]int, 8192),
	}
	for idx := 0; idx < MEM_SIZE; idx++ {
		bp.memSlice[idx] = make([]unsafe.Pointer, memCnt[idx]*10)
		bp.allocMemory(idx)
	}
	return bp
}

func (bp *bufferpool) allocMemory(idx int) {
	size := memSize[idx]
	num := memCnt[idx]

	pool := make([]byte, size*num)
	gpool = append(gpool, pool)

	var list []unsafe.Pointer
	mc := bp.memCache[idx]
	switch {
	case mc == nil:
		list = bp.memSlice[idx][:num]
		bp.memCap[idx] = num
	case mc != nil && cap(bp.memSlice[idx]) >= (bp.memCap[idx]+num)*2:
		// TODO: 更好的处理slice大小
		copy(bp.memSlice[idx], mc)
		list = bp.memSlice[idx][len(mc) : len(mc)+num]
		bp.memCap[idx] += num
	case mc != nil && cap(bp.memSlice[idx]) < (bp.memCap[idx]+num)*2:
		bp.memSlice[idx] = make([]unsafe.Pointer, (bp.memCap[idx]+num)*2)
		copy(bp.memSlice[idx], mc)
		list = bp.memSlice[idx][len(mc) : len(mc)+num]
		bp.memCap[idx] += num
	default:
	}
	bp.memCache[idx] = bp.memSlice[idx][:len(mc)+len(list)]

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
	for idx := 0; idx < MEM_SIZE; idx++ {
		if length <= memSize[idx] {
			// TODO: 手工处理位置信息与使用slice处理流程性能对比
			switch mc := bp.memCache[idx]; {
			case len(mc) > 1:
				p = mc[0]
				bp.memCache[idx] = mc[1:]
			case len(mc) == 0:
				return nil, os.ErrInvalid
			case len(mc) == 1:
				p = mc[0]
				bp.memCache[idx] = mc[1:]
				bp.allocMemory(idx)
			default:
				return nil, os.ErrInvalid
			}
			//log.Println("alloc  ", p)
			bp.memRef[p] = idx
			return p, nil
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
	if idx, ok := bp.memRef[ptr]; ok {
		if cap(bp.memCache[idx])-len(bp.memCache[idx]) == 0 {
			src := bp.memCache[idx]
			//log.Println()
			dst := bp.memSlice[idx][:len(src)]
			copy(dst, src)
			//log.Println(idx, cap(bp.memSlice[idx]),
			//	"src", len(src), cap(src),
			//	"dst", len(dst), cap(dst))
			bp.memCache[idx] = dst
		}
		bp.memCache[idx] = append(bp.memCache[idx], ptr)
		delete(bp.memRef, ptr)
		//log.Println("release", ptr)
	}
}
