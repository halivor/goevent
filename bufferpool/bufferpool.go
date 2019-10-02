package bufferpool

import (
	"log"
	"os"
	"runtime"
	"sync/atomic"
	"unsafe"
)

type BufferPool interface {
	Alloc(length int) ([]byte, error)
	Release(buffer []byte)
}

/*
 * array / map ~= 30 : 1
 * integer assignment performace
 * num            array    map
 * 1000*1000      385us    11788   us
 * 1000*1000*1000 349899us 11776979us
**/
// TODO: 由于使用了两个锁，可能导致性能的下降，考虑如何将锁合并和优化代码处理流程
type bufferpool struct {
	memCache [MEM_ARR_SIZE][]unsafe.Pointer
	mtxMc    [MEM_ARR_SIZE]*uint32
	memSlice [MEM_ARR_SIZE][]unsafe.Pointer
	memCap   [MEM_ARR_SIZE]int
	memBig   map[unsafe.Pointer][]byte
	mtxMb    *uint32
}

var (
	gpool [][]byte
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	gpool = make([][]byte, 256)
}

func New() *bufferpool {
	bp := &bufferpool{
		memBig: make(map[unsafe.Pointer][]byte, 8192),
		mtxMb:  new(uint32),
	}
	for idx := 0; idx < MEM_ARR_SIZE; idx++ {
		bp.mtxMc[idx] = new(uint32)
		bp.memSlice[idx] = make([]unsafe.Pointer, memCnt[idx]*5*(idx+1))
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
		pool[pre] = uint8(idx)
		list[cur-1] = unsafe.Pointer(&pool[pre+1])
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
func (bp *bufferpool) Free(buf []byte) {
	bp.FreePointer(unsafe.Pointer(&buf[0]))
}

func (bp *bufferpool) AllocPointer(length int) (unsafe.Pointer, error) {
	switch {
	case length >= BUF_MAX_LEN:
		buffer := make([]byte, length+1)
		buffer[0] = 'B'
		bp.mbLock()
		bp.memBig[unsafe.Pointer(&buffer[0])] = buffer
		bp.mbUnLock()
		return unsafe.Pointer(&buffer[1]), nil
	default:
		return bp.allocPointer(length)
	}
}
func (bp *bufferpool) FreePointer(ptr unsafe.Pointer) {
	switch buffer := *((*[4]byte)(unsafe.Pointer(uintptr(ptr) - 1))); {
	case buffer[0] == 'B':
		bp.mbLock()
		delete(bp.memBig, unsafe.Pointer(&buffer[0]))
		bp.mbUnLock()
	default:
		bp.freePointer(ptr)
	}
}

func (bp *bufferpool) allocPointer(length int) (p unsafe.Pointer, e error) {
	for idx := 0; idx < MEM_ARR_SIZE; idx++ {
		if length <= memSize[idx]-1 {
			bp.mcLock(idx)
			switch mc := bp.memCache[idx]; {
			case len(mc) > 1:
				p = mc[0]
				bp.memCache[idx] = mc[1:]
			case len(mc) == 1:
				p = mc[0]
				bp.memCache[idx] = mc[1:]
				bp.allocMemory(idx)
			/*case len(mc) == 0: return nil, os.ErrInvalid*/
			default:
				bp.mcUnLock(idx)
				return nil, os.ErrInvalid
			}
			bp.mcUnLock(idx)
			//log.Println("alloc  ", p)
			return p, nil
		}
	}
	return p, os.ErrInvalid
}

func (bp *bufferpool) freePointer(ptr unsafe.Pointer) {
	if idx := int(*((*uint8)(unsafe.Pointer((uintptr(ptr) - 1))))); idx < MEM_ARR_SIZE {
		bp.mcLock(idx)
		if cap(bp.memCache[idx])-len(bp.memCache[idx]) == 0 {
			src := bp.memCache[idx]
			dst := bp.memSlice[idx][:len(src)]
			copy(dst, src)
			log.Println(idx, cap(bp.memSlice[idx]),
				"src", len(src), cap(src),
				"dst", len(dst), cap(dst))
			bp.memCache[idx] = dst
		}
		//log.Println("free   ", ptr)
		bp.memCache[idx] = append(bp.memCache[idx], ptr)
		bp.mcUnLock(idx)
	}
}

func (bp *bufferpool) mcLock(idx int) {
	for !atomic.CompareAndSwapUint32(bp.mtxMc[idx], 0, 1) {
		runtime.Gosched()
	}
}
func (bp *bufferpool) mcUnLock(idx int) {
	atomic.StoreUint32(bp.mtxMc[idx], 0)
}

func (bp *bufferpool) mbLock() {
	for !atomic.CompareAndSwapUint32(bp.mtxMb, 0, 1) {
		runtime.Gosched()
	}
}

func (bp *bufferpool) mbUnLock() {
	atomic.StoreUint32(bp.mtxMb, 0)
}
