//go:build windows

package delegate

import (
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go/internal/iunknown"
)

// Only a limited number of callbacks may be created in a single Go process,
// and any memory allocated for these callbacks is never released.
// Between NewCallback and NewCallbackCDecl, at least 1024 callbacks can always be created.
var (
	invokeCallback = syscall.NewCallback(invoke)
)

// Delegate represents a WinRT delegate class.
type Delegate interface {
	iunknown.Instance
	Invoke(instancePtr, rawArgs0, rawArgs1, rawArgs2, rawArgs3, rawArgs4, rawArgs5, rawArgs6, rawArgs7, rawArgs8 unsafe.Pointer) uintptr
}

// Callbacks contains the syscalls registered on Windows.
type Callbacks struct {
	ole.IUnknownVtbl
	Invoke uintptr
}

// RegisterDelegate adds the given pointer and the Delegate it points to to our instances.
// This is required to redirect received callbacks to the correct object instance.
// The function returns the callbacks to use when creating a new delegate instance.
func RegisterDelegate(ptr unsafe.Pointer, inst Delegate) *Callbacks {
	vtbl := iunknown.RegisterInstance(ptr, inst)

	return &Callbacks{
		IUnknownVtbl: *vtbl,
		Invoke:       invokeCallback,
	}
}

func invoke(instancePtr, rawArgs0, rawArgs1, rawArgs2, rawArgs3, rawArgs4, rawArgs5, rawArgs6, rawArgs7, rawArgs8 unsafe.Pointer) uintptr {
	instance, ok := iunknown.GetInstance(instancePtr)
	if !ok {
		// instance not found
		return ole.E_FAIL
	}

	d, ok := instance.(Delegate)
	if !ok {
		// instance not found
		return ole.E_FAIL
	}

	return d.Invoke(instancePtr, rawArgs0, rawArgs1, rawArgs2, rawArgs3, rawArgs4, rawArgs5, rawArgs6, rawArgs7, rawArgs8)
}
