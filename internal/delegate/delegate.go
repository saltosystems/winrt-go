//go:build windows

package delegate

import (
	"sync"
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
)

// Only a limited number of callbacks may be created in a single Go process,
// and any memory allocated for these callbacks is never released.
// Between NewCallback and NewCallbackCDecl, at least 1024 callbacks can always be created.
var (
	queryInterfaceCallback = syscall.NewCallback(queryInterface)
	addRefCallback         = syscall.NewCallback(addRef)
	releaseCallback        = syscall.NewCallback(release)
)

// invokeCallbacks holds one Invoke callback per parameter count, indexed by
// it. On 32-bit Windows COM uses stdcall, where the callee pops its own
// arguments, so the callback must declare exactly the parameters WinRT pushes:
// one declaring more corrupts the caller's stack on return.
var invokeCallbacks = [...]uintptr{
	syscall.NewCallback(func(i unsafe.Pointer) uintptr { return invoke(i) }),
	syscall.NewCallback(func(i, a0 unsafe.Pointer) uintptr { return invoke(i, a0) }),
	syscall.NewCallback(func(i, a0, a1 unsafe.Pointer) uintptr { return invoke(i, a0, a1) }),
	syscall.NewCallback(func(i, a0, a1, a2 unsafe.Pointer) uintptr { return invoke(i, a0, a1, a2) }),
	syscall.NewCallback(func(i, a0, a1, a2, a3 unsafe.Pointer) uintptr { return invoke(i, a0, a1, a2, a3) }),
	syscall.NewCallback(func(i, a0, a1, a2, a3, a4 unsafe.Pointer) uintptr { return invoke(i, a0, a1, a2, a3, a4) }),
	syscall.NewCallback(func(i, a0, a1, a2, a3, a4, a5 unsafe.Pointer) uintptr { return invoke(i, a0, a1, a2, a3, a4, a5) }),
	syscall.NewCallback(func(i, a0, a1, a2, a3, a4, a5, a6 unsafe.Pointer) uintptr {
		return invoke(i, a0, a1, a2, a3, a4, a5, a6)
	}),
	syscall.NewCallback(func(i, a0, a1, a2, a3, a4, a5, a6, a7 unsafe.Pointer) uintptr {
		return invoke(i, a0, a1, a2, a3, a4, a5, a6, a7)
	}),
	syscall.NewCallback(func(i, a0, a1, a2, a3, a4, a5, a6, a7, a8 unsafe.Pointer) uintptr {
		return invoke(i, a0, a1, a2, a3, a4, a5, a6, a7, a8)
	}),
}

// Delegate represents a WinRT delegate class.
type Delegate interface {
	GetIID() *ole.GUID
	Invoke(instancePtr, rawArgs0, rawArgs1, rawArgs2, rawArgs3, rawArgs4, rawArgs5, rawArgs6, rawArgs7, rawArgs8 unsafe.Pointer) uintptr
	AddRef() uintptr
	Release() uintptr
}

// Callbacks contains the syscalls registered on Windows.
type Callbacks struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	Invoke         uintptr
}

var mutex = sync.RWMutex{}
var instances = make(map[uintptr]Delegate)

// RegisterCallbacks adds the given pointer and the Delegate it points to to our instances.
// This is required to redirect received callbacks to the correct object instance.
// The function returns the callbacks to use when creating a new delegate instance.
// argCount is the number of parameters the delegate's Invoke method takes.
func RegisterCallbacks(ptr unsafe.Pointer, inst Delegate, argCount int) *Callbacks {
	mutex.Lock()
	defer mutex.Unlock()
	instances[uintptr(ptr)] = inst

	return &Callbacks{
		QueryInterface: queryInterfaceCallback,
		AddRef:         addRefCallback,
		Release:        releaseCallback,
		Invoke:         invokeCallbacks[argCount],
	}
}

func getInstance(ptr unsafe.Pointer) (Delegate, bool) {
	mutex.RLock() // locks writing, allows concurrent read
	defer mutex.RUnlock()

	i, ok := instances[uintptr(ptr)]
	return i, ok
}

func removeInstance(ptr unsafe.Pointer) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(instances, uintptr(ptr))
}

func queryInterface(instancePtr unsafe.Pointer, iidPtr unsafe.Pointer, ppvObject *unsafe.Pointer) uintptr {
	instance, ok := getInstance(instancePtr)
	if !ok {
		// instance not found
		return ole.E_POINTER
	}

	// Checkout these sources for more information about the QueryInterface method.
	//   - https://docs.microsoft.com/en-us/cpp/atl/queryinterface
	//   - https://docs.microsoft.com/en-us/windows/win32/api/unknwn/nf-unknwn-iunknown-queryinterface(refiid_void)

	if ppvObject == nil {
		// If ppvObject (the address) is nullptr, then this method returns E_POINTER.
		return ole.E_POINTER
	}

	// This function must adhere to the QueryInterface defined here:
	// https://docs.microsoft.com/en-us/windows/win32/api/unknwn/nn-unknwn-iunknown
	if iid := (*ole.GUID)(iidPtr); ole.IsEqualGUID(iid, instance.GetIID()) || ole.IsEqualGUID(iid, ole.IID_IUnknown) || ole.IsEqualGUID(iid, ole.IID_IInspectable) {
		*ppvObject = instancePtr
	} else {
		*ppvObject = nil
		// Return E_NOINTERFACE if the interface is not supported
		return ole.E_NOINTERFACE
	}

	// If the COM object implements the interface, then it returns
	// a pointer to that interface after calling IUnknown::AddRef on it.
	(*ole.IUnknown)(*ppvObject).AddRef()

	// Return S_OK if the interface is supported
	return ole.S_OK
}

func invoke(instancePtr unsafe.Pointer, rawArgs ...unsafe.Pointer) uintptr {
	instance, ok := getInstance(instancePtr)
	if !ok {
		// instance not found
		return ole.E_FAIL
	}

	var a [9]unsafe.Pointer
	copy(a[:], rawArgs)
	return instance.Invoke(instancePtr, a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8])
}

func addRef(instancePtr unsafe.Pointer) uintptr {
	instance, ok := getInstance(instancePtr)
	if !ok {
		// instance not found
		return ole.E_FAIL
	}

	return instance.AddRef()
}

func release(instancePtr unsafe.Pointer) uintptr {
	instance, ok := getInstance(instancePtr)
	if !ok {
		// instance not found
		return ole.E_FAIL
	}

	rem := instance.Release()
	if rem == 0 {
		// remove this delegate
		removeInstance(instancePtr)
	}
	return rem
}
