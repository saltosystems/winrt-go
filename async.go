package winrt

import (
	"fmt"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go/windows/foundation"
)

// AwaitAsyncOperation waits for the given IAsyncOperation to complete.
// The genericParamSignature is the WinRT signature of the generic parameter type of the IAsyncOperation.
//
// For example, for `IAsyncOperation<Int32>`, the signature would simply be [SignatureInt32].
//
// To obtain the signature of more complex types, use [ParameterizedInstanceGUID].
//
// For example for `IAsyncOperation<IVectorView<GattClientNotificationResult>>`,
// you would use:
//
//	```go
//	winrt.ParameterizedInstanceGUID(collections.GUIDIVectorView, genericattributeprofile.SignatureGattClientNotificationResult)
//	```
func AwaitAsyncOperation(asyncOperation *foundation.IAsyncOperation, genericParamSignature string) error {
	var status foundation.AsyncStatus

	// We need to obtain the GUID of the AsyncOperationCompletedHandler, but its a generic delegate
	// so we also need the generic parameter type's signature:
	// AsyncOperationCompletedHandler<genericParamSignature>
	iid := ParameterizedInstanceGUID(foundation.GUIDAsyncOperationCompletedHandler, genericParamSignature)

	// Wait until the async operation completes.
	waitChan := make(chan struct{})
	handler := foundation.NewAsyncOperationCompletedHandler(ole.NewGUID(iid), func(instance *foundation.AsyncOperationCompletedHandler, asyncInfo *foundation.IAsyncOperation, asyncStatus foundation.AsyncStatus) {
		status = asyncStatus
		close(waitChan)
	})
	defer handler.Release()

	asyncOperation.SetCompleted(handler)

	// Wait until async operation has stopped, and finish.
	<-waitChan

	if status != foundation.AsyncStatusCompleted {
		return fmt.Errorf("async operation failed with status %d", status)
	}
	return nil
}
