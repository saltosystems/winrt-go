package provider

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go"
	"github.com/stretchr/testify/require"
)

func init() {
	ole.CoInitialize(0)
}

func Test_GetMany_StorageProperyItem(t *testing.T) {
	prop1, err := NewStorageProviderItemProperty()
	require.NoError(t, err)
	prop1.SetId(1)
	prop1.SetValue("Value1")
	prop1.SetIconResource("shell32.dll,-44")

	prop2, err := NewStorageProviderItemProperty()
	require.NoError(t, err)
	prop2.SetId(2)
	prop2.SetValue("Value2")
	prop2.SetIconResource("shell32.dll,-44")

	a := winrt.NewArrayIterable([]any{prop1, prop2}, SignatureStorageProviderItemProperty)

	it, err := a.First()
	require.NoError(t, err)
	resp, n, err := it.GetMany(3) // only 2 is available
	require.NoError(t, err)
	require.Equal(t, uint32(2), n)

	// Extract and print the StorageProviderItemProperty objects
	for i := uint32(0); i < n; i++ {
		itemPtr := unsafe.Pointer(resp[i])
		item := (*StorageProviderItemProperty)(itemPtr)
		// Access properties of the StorageProviderItemProperty
		id, err := item.GetId()
		require.NoError(t, err)
		require.Equal(t, int32(i+1), id)
		value, err := item.GetValue()
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("Value%d", i+1), value)
		iconResource, err := item.GetIconResource()
		require.NoError(t, err)
		require.Equal(t, "shell32.dll,-44", iconResource)
	}
}
