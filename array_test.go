package winrt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetEmpty(t *testing.T) {
	a := NewArrayIterable([]any{}, SignatureInt32)
	it, err := a.First()
	require.NoError(t, err)

	ok, err := it.GetHasCurrent()
	require.NoError(t, err)
	require.False(t, ok)
}

func Test_GetCurrent(t *testing.T) {
	a := NewArrayIterable([]any{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, SignatureInt32)

	it, err := a.First()
	require.NoError(t, err)

	i := 1
	for {
		hasMore, err := it.GetHasCurrent()
		require.NoError(t, err)
		if !hasMore {
			break
		}

		ptr, err := it.GetCurrent()
		require.NoError(t, err)
		require.Equal(t, i, int(uintptr(ptr)))
		_, err = it.MoveNext()
		require.NoError(t, err)
		i++
	}
	require.Equal(t, 11, i)
}

func Test_GetMany(t *testing.T) {
	a := NewArrayIterable([]any{101, 202, 303}, SignatureInt32)

	it, err := a.First()
	require.NoError(t, err)

	r3, err := it.GetCurrent()
	require.NoError(t, err)
	require.True(t, int(uintptr(r3)) == 101)

	resp, n, err := it.GetMany(12)
	require.NoError(t, err)
	require.Equal(t, uint32(3), n)

	var i uint32
	var j int = 101
	for i = 0; i < n; i++ {
		val := int(uintptr(resp[i]))
		require.Equal(t, j, val)
		j += 101
	}

	// no more items
	hasMore, err := it.GetHasCurrent()
	require.NoError(t, err)
	require.False(t, hasMore)

	_, err = it.GetCurrent()
	require.Error(t, err)

	ok, err := it.MoveNext()
	require.NoError(t, err)
	require.False(t, ok)
}
