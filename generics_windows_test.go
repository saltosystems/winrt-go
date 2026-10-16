package winrt

import (
	"syscall"
	"testing"
	"unsafe"

	"github.com/saltosystems/winrt-go/windows/foundation/collections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const eBounds = 0x8000000B

var testGUIDs = []syscall.GUID{
	{Data1: 0x0000180f, Data2: 0x0000, Data3: 0x1000, Data4: [8]byte{0x80, 0x00, 0x00, 0x80, 0x5f, 0x9b, 0x34, 0xfb}},
	{Data1: 0x0000cdab, Data2: 0x0000, Data3: 0x1000, Data4: [8]byte{0x80, 0x00, 0x00, 0x80, 0x5f, 0x9b, 0x34, 0xfb}},
	{Data1: 0x6e400001, Data2: 0xb5a3, Data3: 0xf393, Data4: [8]byte{0xe0, 0xa9, 0xe5, 0x0e, 0x24, 0xdc, 0xca, 0x9e}},
}

// guidVector returns an IVector<GUID> whose GetAt and GetMany write whole GUIDs like WinRT does.
func guidVector(guids []syscall.GUID) *collections.IVector {
	vtbl := &collections.IVectorVtbl{
		GetAt: syscall.NewCallback(func(this, index uintptr, out *syscall.GUID) uintptr {
			if index >= uintptr(len(guids)) {
				return eBounds
			}
			*out = guids[index]
			return 0
		}),
		GetMany: syscall.NewCallback(func(this, startIndex, capacity uintptr, items *syscall.GUID, actual *uint32) uintptr {
			if startIndex > uintptr(len(guids)) {
				return eBounds
			}
			n := copy(unsafe.Slice(items, capacity), guids[startIndex:])
			*actual = uint32(n)
			return 0
		}),
	}
	v := &collections.IVector{}
	v.RawVTable = (*interface{})(unsafe.Pointer(vtbl))
	return v
}

func TestIVectorGetAtValueType(t *testing.T) {
	v := guidVector(testGUIDs)

	for i, want := range testGUIDs {
		var got syscall.GUID
		require.NoError(t, v.GetAt(uint32(i), unsafe.Pointer(&got)))
		assert.Equal(t, want, got)
	}

	var got syscall.GUID
	assert.Error(t, v.GetAt(uint32(len(testGUIDs)), unsafe.Pointer(&got)))
}

func TestIVectorGetManyValueType(t *testing.T) {
	v := guidVector(testGUIDs)

	got := make([]syscall.GUID, 2)
	n, err := v.GetMany(1, uint32(len(got)), unsafe.Pointer(&got[0]))
	require.NoError(t, err)
	assert.Equal(t, uint32(2), n)
	assert.Equal(t, testGUIDs[1:], got)
}
