//go:build windows

package winrt_test

import (
	"os"
	"testing"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go/winrt"
)

func TestMain(m *testing.M) {
	ole.RoInitialize(1)
	code := m.Run()
	os.Exit(code)
}

func TestNewBuffer(t *testing.T) {
	b, err := winrt.Create(0)
	if err != nil {
		t.Fatal(err)
	}

	if b == nil {
		t.Fatal("b is nil")
	}
}

func TestSetCapacity(t *testing.T) {
	b, err := winrt.Create(0)
	if err != nil {
		t.Fatal(err)
	}

	if b == nil {
		t.Fatal("b is nil")
	}
}
