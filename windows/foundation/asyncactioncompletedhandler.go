// Code generated by winrt-go-gen. DO NOT EDIT.

//go:build windows

//nolint:all
package foundation

import (
	"sync"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go/internal/delegate"
	"github.com/saltosystems/winrt-go/internal/kernel32"
)

const GUIDAsyncActionCompletedHandler string = "a4ed5c81-76c9-40bd-8be6-b1d90fb20ae7"
const SignatureAsyncActionCompletedHandler string = "delegate({a4ed5c81-76c9-40bd-8be6-b1d90fb20ae7})"

type AsyncActionCompletedHandler struct {
	ole.IUnknown
	sync.Mutex
	refs uintptr
	IID  ole.GUID
}

type AsyncActionCompletedHandlerVtbl struct {
	ole.IUnknownVtbl
	Invoke uintptr
}

type AsyncActionCompletedHandlerCallback func(instance *AsyncActionCompletedHandler, asyncInfo *IAsyncAction, asyncStatus AsyncStatus)

var callbacksAsyncActionCompletedHandler = &asyncActionCompletedHandlerCallbacks{
	mu:        &sync.Mutex{},
	callbacks: make(map[unsafe.Pointer]AsyncActionCompletedHandlerCallback),
}

var releaseChannelsAsyncActionCompletedHandler = &asyncActionCompletedHandlerReleaseChannels{
	mu:    &sync.Mutex{},
	chans: make(map[unsafe.Pointer]chan struct{}),
}

func NewAsyncActionCompletedHandler(iid *ole.GUID, callback AsyncActionCompletedHandlerCallback) *AsyncActionCompletedHandler {
	// create type instance
	size := unsafe.Sizeof(*(*AsyncActionCompletedHandler)(nil))
	instPtr := kernel32.Malloc(size)
	inst := (*AsyncActionCompletedHandler)(instPtr)

	// get the callbacks for the VTable
	callbacks := delegate.RegisterDelegate(instPtr, inst)

	// the VTable should also be allocated in the heap
	sizeVTable := unsafe.Sizeof(*(*AsyncActionCompletedHandlerVtbl)(nil))
	vTablePtr := kernel32.Malloc(sizeVTable)

	inst.RawVTable = (*interface{})(vTablePtr)

	vTable := (*AsyncActionCompletedHandlerVtbl)(vTablePtr)
	vTable.IUnknownVtbl = ole.IUnknownVtbl{
		QueryInterface: callbacks.QueryInterface,
		AddRef:         callbacks.AddRef,
		Release:        callbacks.Release,
	}
	vTable.Invoke = callbacks.Invoke

	// Initialize all properties: the malloc may contain garbage
	inst.IID = *iid // copy contents
	inst.Mutex = sync.Mutex{}
	inst.refs = 0

	callbacksAsyncActionCompletedHandler.add(unsafe.Pointer(inst), callback)

	// See the docs in the releaseChannelsAsyncActionCompletedHandler struct
	releaseChannelsAsyncActionCompletedHandler.acquire(unsafe.Pointer(inst))

	inst.addRef()
	return inst
}

func (r *AsyncActionCompletedHandler) GetIID() *ole.GUID {
	return &r.IID
}

// addRef increments the reference counter by one
func (r *AsyncActionCompletedHandler) addRef() uintptr {
	r.Lock()
	defer r.Unlock()
	r.refs++
	return r.refs
}

// removeRef decrements the reference counter by one. If it was already zero, it will just return zero.
func (r *AsyncActionCompletedHandler) removeRef() uintptr {
	r.Lock()
	defer r.Unlock()

	if r.refs > 0 {
		r.refs--
	}

	return r.refs
}

func (instance *AsyncActionCompletedHandler) Invoke(instancePtr, rawArgs0, rawArgs1, rawArgs2, rawArgs3, rawArgs4, rawArgs5, rawArgs6, rawArgs7, rawArgs8 unsafe.Pointer) uintptr {
	asyncInfoPtr := rawArgs0
	asyncStatusRaw := (int32)(uintptr(rawArgs1))

	// See the quote above.
	asyncInfo := (*IAsyncAction)(asyncInfoPtr)
	asyncStatus := (AsyncStatus)(asyncStatusRaw)
	if callback, ok := callbacksAsyncActionCompletedHandler.get(instancePtr); ok {
		callback(instance, asyncInfo, asyncStatus)
	}
	return ole.S_OK
}

func (instance *AsyncActionCompletedHandler) AddRef() uintptr {
	return instance.addRef()
}

func (instance *AsyncActionCompletedHandler) Release() uintptr {
	rem := instance.removeRef()
	if rem == 0 {
		// We're done.
		instancePtr := unsafe.Pointer(instance)
		callbacksAsyncActionCompletedHandler.delete(instancePtr)

		// stop release channels used to avoid
		// https://github.com/golang/go/issues/55015
		releaseChannelsAsyncActionCompletedHandler.release(instancePtr)

		kernel32.Free(unsafe.Pointer(instance.RawVTable))
		kernel32.Free(instancePtr)
	}
	return rem
}

type asyncActionCompletedHandlerCallbacks struct {
	mu        *sync.Mutex
	callbacks map[unsafe.Pointer]AsyncActionCompletedHandlerCallback
}

func (m *asyncActionCompletedHandlerCallbacks) add(p unsafe.Pointer, v AsyncActionCompletedHandlerCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callbacks[p] = v
}

func (m *asyncActionCompletedHandlerCallbacks) get(p unsafe.Pointer) (AsyncActionCompletedHandlerCallback, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	v, ok := m.callbacks[p]
	return v, ok
}

func (m *asyncActionCompletedHandlerCallbacks) delete(p unsafe.Pointer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.callbacks, p)
}

// typedEventHandlerReleaseChannels keeps a map with channels
// used to keep a goroutine alive during the lifecycle of this object.
// This is required to avoid causing a deadlock error.
// See this: https://github.com/golang/go/issues/55015
type asyncActionCompletedHandlerReleaseChannels struct {
	mu    *sync.Mutex
	chans map[unsafe.Pointer]chan struct{}
}

func (m *asyncActionCompletedHandlerReleaseChannels) acquire(p unsafe.Pointer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	c := make(chan struct{})
	m.chans[p] = c

	go func() {
		// we need a timer to trick the go runtime into
		// thinking there's still something going on here
		// but we are only really interested in <-c
		t := time.NewTimer(time.Minute)
		for {
			select {
			case <-t.C:
				t.Reset(time.Minute)
			case <-c:
				t.Stop()
				return
			}
		}
	}()
}

func (m *asyncActionCompletedHandlerReleaseChannels) release(p unsafe.Pointer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c, ok := m.chans[p]; ok {
		close(c)
		delete(m.chans, p)
	}
}
