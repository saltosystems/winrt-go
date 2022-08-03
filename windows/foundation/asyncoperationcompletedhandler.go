// Code generated by winrt-go-gen. DO NOT EDIT.

//go:build windows

//nolint
package foundation

import (
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go"
)

/*
#include <stdint.h>

// Note: these functions have a different signature but because they are only
// used as function pointers (and never called) and because they use C name
// mangling, the signature doesn't really matter.
void winrt_AsyncOperationCompletedHandler_Invoke(void);
void winrt_AsyncOperationCompletedHandler_QueryInterface(void);
uint64_t winrt_AsyncOperationCompletedHandler_AddRef(void);
uint64_t winrt_AsyncOperationCompletedHandler_Release(void);

// The Vtable structure for WinRT AsyncOperationCompletedHandler interfaces.
typedef struct {
	void *QueryInterface;
	void *AddRef;
	void *Release;
	void *Invoke;
} AsyncOperationCompletedHandlerVtbl_t;

// The Vtable itself. It can be kept constant.
static const AsyncOperationCompletedHandlerVtbl_t winrt_AsyncOperationCompletedHandlerVtbl = {
	(void*)winrt_AsyncOperationCompletedHandler_QueryInterface,
	(void*)winrt_AsyncOperationCompletedHandler_AddRef,
	(void*)winrt_AsyncOperationCompletedHandler_Release,
	(void*)winrt_AsyncOperationCompletedHandler_Invoke,
};

// A small helper function to get the Vtable.
const AsyncOperationCompletedHandlerVtbl_t * winrt_getAsyncOperationCompletedHandlerVtbl(void) {
	return &winrt_AsyncOperationCompletedHandlerVtbl;
}
*/
import "C"

const GUIDAsyncOperationCompletedHandler string = "fcdcf02c-e5d8-4478-915a-4d90b74b83a5"
const SignatureAsyncOperationCompletedHandler string = "delegate({fcdcf02c-e5d8-4478-915a-4d90b74b83a5})"

type AsyncOperationCompletedHandler struct {
	ole.IUnknown
	IID      *ole.GUID
	RefCount *winrt.RefCount
	Callback AsyncOperationCompletedHandlerCallback
}

type AsyncOperationCompletedHandlerCallback func(instance *AsyncOperationCompletedHandler, asyncInfo *IAsyncOperation, asyncStatus AsyncStatus)

func NewAsyncOperationCompletedHandler(iid *ole.GUID, callback AsyncOperationCompletedHandlerCallback) *AsyncOperationCompletedHandler {
	inst := (*AsyncOperationCompletedHandler)(C.malloc(C.size_t(unsafe.Sizeof(AsyncOperationCompletedHandler{}))))
	inst.RawVTable = (*interface{})((unsafe.Pointer)(C.winrt_getAsyncOperationCompletedHandlerVtbl()))
	inst.IID = iid
	inst.RefCount = winrt.NewRefCount()
	inst.Callback = callback

	inst.RefCount.AddRef()
	return inst
}
