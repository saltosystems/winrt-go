//go:build windows

package winrt_test

import (
	"log"
	"os"
	"testing"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go/winrt/windows/storage/streams/buffer"
)

func TestMain(m *testing.M) {
	err := ole.RoInitialize(1)
	if err != nil {
		log.Fatal(err)
	}
	code := m.Run()
	os.Exit(code)
}

func TestNewBuffer(t *testing.T) {
	bufFactory, err := buffer.ActivateIBufferFactory()
	if err != nil {
		t.Fatal(err)
	}

	b, err := bufFactory.Create(10)
	if err != nil {
		t.Fatal(err)
	}

	if b == nil {
		t.Fatal("b is nil")
	}
}

func TestSetCapacity(t *testing.T) {
	bufFactory, err := buffer.ActivateIBufferFactory()
	if err != nil {
		t.Fatal(err)
	}

	bufferCapacity := uint32(12)
	b, err := bufFactory.Create(bufferCapacity)
	if err != nil {
		t.Fatal(err)
	}

	if b == nil {
		t.Fatal("b is nil")
	}

	c, err := b.GetCapacity()
	if err != nil {
		t.Fatal(err)
	}

	if c != bufferCapacity {
		t.Fatalf("Buffer Capacity was set to 10 but GetCapacity returned %d.", c)
	}
}
