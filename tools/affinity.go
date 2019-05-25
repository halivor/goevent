package tools

import (
	"os"
	"syscall"
	"unsafe"
)

func SetCPUAffinity(index uint8) (uintptr, error) {
	var ma [128]byte
	ma[0] = byte(1 << index)
	mask := uintptr(unsafe.Pointer(&ma[0]))
	r, _, e := syscall.RawSyscall(
		syscall.SYS_SCHED_SETAFFINITY,
		uintptr(os.Getpid()),
		uintptr(unsafe.Sizeof(ma)),
		mask,
	)
	return r, e
}
