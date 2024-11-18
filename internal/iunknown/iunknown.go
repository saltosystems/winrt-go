//go:build windows

package iunknown

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

// Class instance represents a WinRT delegate class.
type Instance interface {
	GetIID() *ole.GUID
	AddRef() uintptr
	Release() uintptr
}

var mutex = sync.RWMutex{}
var instances = make(map[uintptr]Instance)

// RegisterCallbacks adds the given pointer and the Delegate it points to to our instances.
// This is required to redirect received callbacks to the correct object instance.
// The function returns the callbacks to use when creating a new delegate instance.
func RegisterInstance(ptr unsafe.Pointer, inst Instance) *ole.IUnknownVtbl {
	mutex.Lock()
	defer mutex.Unlock()
	instances[uintptr(ptr)] = inst

	return &ole.IUnknownVtbl{
		QueryInterface: queryInterfaceCallback,
		AddRef:         addRefCallback,
		Release:        releaseCallback,
	}
}

func GetInstance(ptr unsafe.Pointer) (Instance, bool) {
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
	instance, ok := GetInstance(instancePtr)
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

func addRef(instancePtr unsafe.Pointer) uintptr {
	instance, ok := GetInstance(instancePtr)
	if !ok {
		// instance not found
		return ole.E_FAIL
	}

	return instance.AddRef()
}

func release(instancePtr unsafe.Pointer) uintptr {
	instance, ok := GetInstance(instancePtr)
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
