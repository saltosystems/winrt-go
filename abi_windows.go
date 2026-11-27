//go:build windows && !386

package winrt

import (
	"syscall"
	"unsafe"
)

// Uint64Arg lays out a uint64 passed by value: one word.
func Uint64Arg(v uint64) []uintptr { return []uintptr{uintptr(v)} }

// GUIDArg lays out a GUID passed by value. The x64 and ARM64 conventions
// pass a struct this size as a pointer to a copy.
func GUIDArg(g *syscall.GUID) []uintptr { return []uintptr{uintptr(unsafe.Pointer(g))} }
