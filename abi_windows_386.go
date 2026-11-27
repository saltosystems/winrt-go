//go:build windows && 386

package winrt

import (
	"syscall"
	"unsafe"
)

// Uint64Arg lays out a uint64 passed by value: on 386 it takes two stack
// words, low word first.
func Uint64Arg(v uint64) []uintptr { return []uintptr{uintptr(v), uintptr(v >> 32)} }

// GUIDArg lays out a GUID passed by value: on 386 its 16 bytes go on the
// stack, as four words.
func GUIDArg(g *syscall.GUID) []uintptr {
	w := (*[4]uint32)(unsafe.Pointer(g))
	return []uintptr{uintptr(w[0]), uintptr(w[1]), uintptr(w[2]), uintptr(w[3])}
}
