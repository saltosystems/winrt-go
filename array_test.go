package winrt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetCurrent(t *testing.T) {
	a := NewArrayIterable([]any{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, SignatureInt32)

	it, err := a.First()
	require.NoError(t, err)

	var ok bool = false
	i := 1
	for ok, err = it.MoveNext(); err == nil && ok; ok, err = it.MoveNext() {
		b, err := it.GetHasCurrent()
		require.NoError(t, err)
		require.True(t, b)
		ptr, err := it.GetCurrent()
		require.NoError(t, err)
		require.Equal(t, i, int(uintptr(ptr)))
		i++
	}
}

func Test_GetMany(t *testing.T) {
	a := NewArrayIterable([]any{101, 202, 303}, SignatureInt32)

	it, err := a.First()
	require.NoError(t, err)
	resp, n, err := it.GetMany(12)
	require.NoError(t, err)
	require.Equal(t, uint32(3), n)

	println("RESP", n, resp, len(resp))

	var i uint32
	var j int = 101
	for i = 0; i < n; i++ {
		val := int(uintptr(resp[i]))
		require.Equal(t, j, val)
		j += 101
	}
}
