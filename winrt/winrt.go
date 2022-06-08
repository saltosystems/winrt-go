package winrt

import "github.com/go-ole/go-ole"

//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.Buffer -skip-statics

func makeError(hr uintptr) error {
	if hr != 0 {
		return ole.NewError(hr)
	}
	return nil
}
