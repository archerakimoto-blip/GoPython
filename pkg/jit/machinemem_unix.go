// +build !windows

package jit

import (
	"fmt"
	"syscall"
	"unsafe"
)

func mmapRWX(size int) ([]byte, error) {
	// Use syscall.Mmap to allocate executable memory
	// PROT_READ | PROT_WRITE | PROT_EXEC = 0x7
	prot := syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC
	
	// MAP_ANONYMOUS - not backed by a file
	// MAP_PRIVATE - changes are private
	mapFlags := syscall.MAP_ANONYMOUS | syscall.MAP_PRIVATE
	
	// Align size to page boundary
	pageSize := syscall.Getpagesize()
	alignedSize := (size + pageSize - 1) &^ (pageSize - 1)
	
	// Call mmap: mmap(fd, offset, length, prot, flags)
	// fd = -1 for anonymous mapping
	fd := -1
	offset := int64(0)
	
	addr, err := syscall.Mmap(fd, offset, alignedSize, prot, mapFlags)
	if err != nil {
		return nil, fmt.Errorf("mmap failed: %v", err)
	}
	
	// Make the memory executable
	// On some systems, we might need to set the protection explicitly
	// Most Unix systems will honor the PROT_EXEC flag in mmap
	if err := syscall.Mprotect(addr, prot); err != nil {
		syscall.Munmap(addr)
		return nil, fmt.Errorf("mprotect failed: %v", err)
	}
	
	return addr, nil
}

func virtualAlloc(size int) (uintptr, error) {
	// Windows-specific - not implemented in this Unix build
	return 0, fmt.Errorf("VirtualAlloc not available on Unix")
}

func allocateRWXMemory(size int) (uintptr, error) {
	// Try using mmap on Unix systems
	mem, err := mmapRWX(size)
	if err == nil && len(mem) > 0 {
		return uintptr(unsafe.Pointer(&mem[0])), nil
	}
	
	// Fallback - this likely won't be executable
	mem = make([]byte, size)
	return uintptr(unsafe.Pointer(&mem[0])), fmt.Errorf("fallback memory allocation - may not be executable")
}
