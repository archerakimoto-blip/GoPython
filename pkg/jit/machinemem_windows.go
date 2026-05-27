// +build windows

package jit

import (
	"fmt"
	"syscall"
	"unsafe"
)

func mmapRWX(size int) ([]byte, error) {
	// Windows doesn't have mmap in the standard library
	// We need to use VirtualAlloc instead
	return nil, fmt.Errorf("mmap not available on Windows - use VirtualAlloc")
}

func virtualAlloc(size int) (uintptr, error) {
	// Windows VirtualAlloc constants
	const (
		MEM_COMMIT     = 0x1000
		MEM_RESERVE    = 0x2000
		PAGE_EXECUTE_READWRITE = 0x40
	)
	
	// Get system page size
	var minAddr uintptr = 0
	pageSize := syscall.Getpagesize()
	alignedSize := (size + pageSize - 1) &^ (pageSize - 1)
	
	// Call VirtualAlloc
	addr, err := syscall.VirtualAlloc(minAddr, alignedSize, MEM_COMMIT|MEM_RESERVE, PAGE_EXECUTE_READWRITE)
	if err != nil {
		return 0, fmt.Errorf("VirtualAlloc failed: %v", err)
	}
	
	return uintptr(addr), nil
}

func allocateRWXMemory(size int) (uintptr, error) {
	// Try using VirtualAlloc on Windows
	addr, err := virtualAlloc(size)
	if err == nil && addr != 0 {
		return addr, nil
	}
	
	// Fallback - this likely won't be executable
	mem := make([]byte, size)
	return uintptr(unsafe.Pointer(&mem[0])), fmt.Errorf("fallback memory allocation - may not be executable")
}
