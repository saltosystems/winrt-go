package main

import (
	"fmt"
	"os"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go"
	"github.com/saltosystems/winrt-go/windows/foundation"
	"github.com/saltosystems/winrt-go/windows/storage"
	"github.com/saltosystems/winrt-go/windows/storage/provider"
	"github.com/saltosystems/winrt-go/windows/storage/streams"
)

func main() {
	ole.RoInitialize(0)
	if err := run2(); err != nil {
		panic(err)
	}
}

func run() error {
	// create info
	infoID := "infoID"

	info, err := provider.NewStorageProviderSyncRootInfo()
	if err != nil {
		return err
	}

	err = info.SetId(infoID)
	if err != nil {
		return err
	}

	storageFolderAsync, err := storage.StorageFolderGetFolderFromPathAsync(`C:\Users\jagoba\projects\winrt-go`)
	if err != nil {
		return err
	}

	if err := awaitAsyncOperation(storageFolderAsync, storage.SignatureStorageFolder); err != nil {
		return err
	}

	res, err := storageFolderAsync.GetResults()
	if err != nil {
		return err
	}

	folder := (*storage.StorageFolder)(res)
	itf := folder.MustQueryInterface(ole.NewGUID(storage.GUIDIStorageFolder))
	defer itf.Release()
	f := (*storage.IStorageFolder)(unsafe.Pointer(itf))
	if err := info.SetPath(f); err != nil {
		return err
	}

	// register info
	err = provider.StorageProviderSyncRootManagerRegister(info)
	if err != nil {
		return err
	}

	// unregister info
	err = provider.StorageProviderSyncRootManagerUnregister(infoID)
	if err != nil {
		return err
	}

	// release info
	info.Release()

	return nil
}

func run2() error {
	// tr := initTestResource(t, withTestBrowseDirFn(defaultBrowseDirTestFunc), withConnectSyncRoot())
	// defer tr.cleanUp()

	roots, err := provider.StorageProviderSyncRootManagerGetCurrentSyncRoots()
	if err != nil {
		return err
	}
	numRoots, err := roots.GetSize()
	if err != nil {
		return err
	}
	fmt.Println("Number of roots:", numRoots)

	tempBase, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	syncRootPath, err := os.MkdirTemp(tempBase, "syncRootPath")
	if err != nil {
		return err
	}

	println(syncRootPath)

	writer, err := streams.NewDataWriter()
	if err != nil {
		return err
	}
	syncRootId := []byte("syncRootIdentity")
	err = writer.WriteBytes(uint32(len(syncRootId)), syncRootId)
	if err != nil {
		return err
	}

	bufferContext, err := writer.DetachBuffer()
	if err != nil {
		return err
	}

	syncRootInfo, err := provider.NewStorageProviderSyncRootInfo()
	if err != nil {
		return err
	}

	err = syncRootInfo.SetContext(bufferContext)
	if err != nil {
		return err
	}

	syncRootInfo.SetId("{00000000-0000-0000-0000-000000000001}")
	if err != nil {
		return err
	}
	storageFolderAsync, err := storage.StorageFolderGetFolderFromPathAsync(syncRootPath)
	if err != nil {
		return err
	}

	if err := awaitAsyncOperation(storageFolderAsync, storage.SignatureStorageFolder); err != nil {
		return err
	}

	res, err := storageFolderAsync.GetResults()
	if err != nil {
		return err
	}

	dir := (*storage.StorageFolder)(res)

	itf3 := dir.MustQueryInterface(ole.NewGUID(storage.GUIDIStorageFolder))
	defer itf3.Release()
	iStorageDir := (*storage.IStorageFolder)(unsafe.Pointer(itf3))
	err = syncRootInfo.SetPath(iStorageDir)
	if err != nil {
		return err
	}
	err = syncRootInfo.SetHydrationPolicy(2)
	if err != nil {
		return err
	}
	err = syncRootInfo.SetHydrationPolicyModifier(0)
	if err != nil {
		return err
	}
	err = syncRootInfo.SetPopulationPolicy(2)
	if err != nil {
		return err
	}
	err = syncRootInfo.SetInSyncPolicy(provider.StorageProviderInSyncPolicyPreserveInsyncForSyncEngine)
	if err != nil {
		return err
	}
	err = syncRootInfo.SetHardlinkPolicy(0)
	if err != nil {
		return err
	}
	err = syncRootInfo.SetVersion("1.0")
	if err != nil {
		return err
	}
	syncRootInfo.SetAllowPinning(true)
	syncRootInfo.SetShowSiblingsAsGroup(false)
	syncRootInfo.SetProtectionMode(0)
	syncRootInfo.SetDisplayNameResource("DisplayNameResource")
	//PrintAllFields(syncRootInfo)
	fmt.Println(">>>>>>> sync root info", syncRootInfo)

	err = provider.StorageProviderSyncRootManagerRegister(syncRootInfo)
	if err != nil {
		return err
	}

	roots, err = provider.StorageProviderSyncRootManagerGetCurrentSyncRoots()
	if err != nil {
		return err
	}
	println("done")
	numRoots, err = roots.GetSize()
	if err != nil {
		return err
	}
	fmt.Println("Number of roots:", numRoots)

	return nil
}

func awaitAsyncOperation(asyncOperation *foundation.IAsyncOperation, genericParamSignature string) error {
	var status foundation.AsyncStatus

	// We need to obtain the GUID of the AsyncOperationCompletedHandler, but its a generic delegate
	// so we also need the generic parameter type's signature:
	// AsyncOperationCompletedHandler<genericParamSignature>
	iid := winrt.ParameterizedInstanceGUID(foundation.GUIDAsyncOperationCompletedHandler, genericParamSignature)

	// Wait until the async operation completes.
	waitChan := make(chan struct{})
	handler := foundation.NewAsyncOperationCompletedHandler(ole.NewGUID(iid), func(instance *foundation.AsyncOperationCompletedHandler, asyncInfo *foundation.IAsyncOperation, asyncStatus foundation.AsyncStatus) {
		status = asyncStatus
		close(waitChan)
	})
	defer handler.Release()

	asyncOperation.SetCompleted(handler)

	// Wait until async operation has stopped, and finish.
	asyncWait := true
	for asyncWait {
		select {
		case <-time.After(30 * time.Second):
			itf, err := asyncOperation.QueryInterface(ole.NewGUID(foundation.GUIDIAsyncInfo))
			if err != nil {
				return err
			}
			defer itf.Release()
			v := (*foundation.IAsyncInfo)(unsafe.Pointer(itf))
			if err := v.Cancel(); err != nil {
				return err
			}
			println("Waiting for operation cancel")
		case <-waitChan:
			asyncWait = false
		}
	}

	if status != foundation.AsyncStatusCompleted {
		return fmt.Errorf("async operation failed with status %d", status)
	}
	return nil
}
