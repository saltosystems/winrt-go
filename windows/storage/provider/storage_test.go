package provider

import (
	"fmt"
	"log"
	"testing"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go"
	"github.com/stretchr/testify/require"
)

func init() {
	ole.CoInitialize(0)
}

func Test_GetManyStorageProperyItem(t *testing.T) {

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

	log.Println("prop1", prop1)
	log.Println("prop2", prop2)

	a := winrt.NewArrayIterable([]any{prop1, prop2}, SignatureStorageProviderItemProperty)

	it, err := a.First()
	require.NoError(t, err)
	resp, n, err := it.GetMany(3) // only 2 is available
	require.NoError(t, err)
	require.Equal(t, uint32(2), n)

	println("RESP", n, resp, len(resp))

	// Extract and print the StorageProviderItemProperty objects
	for i := uint32(0); i < n; i++ {
		itemPtr := unsafe.Pointer(resp[i]) // Convert uintptr to pointer

		item := (*StorageProviderItemProperty)(itemPtr) // Cast pointer to StorageProviderItemProperty

		// Access properties of the StorageProviderItemProperty
		id, _ := item.GetId()
		value, _ := item.GetValue()
		iconResource, _ := item.GetIconResource()

		fmt.Printf("Item %d: Id=%d, Value=%s, IconResource=%s\n", i+1, id, value, iconResource)
	}
	require.Equal(t, 1, 2)
}
