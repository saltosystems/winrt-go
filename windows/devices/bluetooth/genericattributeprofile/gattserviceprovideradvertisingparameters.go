// Code generated by winrt-go-gen. DO NOT EDIT.

//go:build windows

//nolint:all
package genericattributeprofile

import (
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go/windows/storage/streams"
)

const SignatureGattServiceProviderAdvertisingParameters string = "rc(Windows.Devices.Bluetooth.GenericAttributeProfile.GattServiceProviderAdvertisingParameters;{e2ce31ab-6315-4c22-9bd7-781dbc3d8d82})"

type GattServiceProviderAdvertisingParameters struct {
	ole.IUnknown
}

func NewGattServiceProviderAdvertisingParameters() (*GattServiceProviderAdvertisingParameters, error) {
	inspectable, err := ole.RoActivateInstance("Windows.Devices.Bluetooth.GenericAttributeProfile.GattServiceProviderAdvertisingParameters")
	if err != nil {
		return nil, err
	}
	return (*GattServiceProviderAdvertisingParameters)(unsafe.Pointer(inspectable)), nil
}

func (impl *GattServiceProviderAdvertisingParameters) SetIsConnectable(value bool) error {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattServiceProviderAdvertisingParameters))
	defer itf.Release()
	v := (*iGattServiceProviderAdvertisingParameters)(unsafe.Pointer(itf))
	return v.SetIsConnectable(value)
}

func (impl *GattServiceProviderAdvertisingParameters) GetIsConnectable() (bool, error) {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattServiceProviderAdvertisingParameters))
	defer itf.Release()
	v := (*iGattServiceProviderAdvertisingParameters)(unsafe.Pointer(itf))
	return v.GetIsConnectable()
}

func (impl *GattServiceProviderAdvertisingParameters) SetIsDiscoverable(value bool) error {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattServiceProviderAdvertisingParameters))
	defer itf.Release()
	v := (*iGattServiceProviderAdvertisingParameters)(unsafe.Pointer(itf))
	return v.SetIsDiscoverable(value)
}

func (impl *GattServiceProviderAdvertisingParameters) GetIsDiscoverable() (bool, error) {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattServiceProviderAdvertisingParameters))
	defer itf.Release()
	v := (*iGattServiceProviderAdvertisingParameters)(unsafe.Pointer(itf))
	return v.GetIsDiscoverable()
}

func (impl *GattServiceProviderAdvertisingParameters) SetServiceData(value *streams.IBuffer) error {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattServiceProviderAdvertisingParameters2))
	defer itf.Release()
	v := (*iGattServiceProviderAdvertisingParameters2)(unsafe.Pointer(itf))
	return v.SetServiceData(value)
}

func (impl *GattServiceProviderAdvertisingParameters) GetServiceData() (*streams.IBuffer, error) {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattServiceProviderAdvertisingParameters2))
	defer itf.Release()
	v := (*iGattServiceProviderAdvertisingParameters2)(unsafe.Pointer(itf))
	return v.GetServiceData()
}

const GUIDiGattServiceProviderAdvertisingParameters string = "e2ce31ab-6315-4c22-9bd7-781dbc3d8d82"
const SignatureiGattServiceProviderAdvertisingParameters string = "{e2ce31ab-6315-4c22-9bd7-781dbc3d8d82}"

type iGattServiceProviderAdvertisingParameters struct {
	ole.IInspectable
}

type iGattServiceProviderAdvertisingParametersVtbl struct {
	ole.IInspectableVtbl

	SetIsConnectable  uintptr
	GetIsConnectable  uintptr
	SetIsDiscoverable uintptr
	GetIsDiscoverable uintptr
}

func (v *iGattServiceProviderAdvertisingParameters) VTable() *iGattServiceProviderAdvertisingParametersVtbl {
	return (*iGattServiceProviderAdvertisingParametersVtbl)(unsafe.Pointer(v.RawVTable))
}

func (v *iGattServiceProviderAdvertisingParameters) SetIsConnectable(value bool) error {
	hr, _, _ := syscall.SyscallN(
		v.VTable().SetIsConnectable,
		uintptr(unsafe.Pointer(v)),                // this
		uintptr(*(*byte)(unsafe.Pointer(&value))), // in bool
	)

	if hr != 0 {
		return ole.NewError(hr)
	}

	return nil
}

func (v *iGattServiceProviderAdvertisingParameters) GetIsConnectable() (bool, error) {
	var out bool
	hr, _, _ := syscall.SyscallN(
		v.VTable().GetIsConnectable,
		uintptr(unsafe.Pointer(v)),    // this
		uintptr(unsafe.Pointer(&out)), // out bool
	)

	if hr != 0 {
		return false, ole.NewError(hr)
	}

	return out, nil
}

func (v *iGattServiceProviderAdvertisingParameters) SetIsDiscoverable(value bool) error {
	hr, _, _ := syscall.SyscallN(
		v.VTable().SetIsDiscoverable,
		uintptr(unsafe.Pointer(v)),                // this
		uintptr(*(*byte)(unsafe.Pointer(&value))), // in bool
	)

	if hr != 0 {
		return ole.NewError(hr)
	}

	return nil
}

func (v *iGattServiceProviderAdvertisingParameters) GetIsDiscoverable() (bool, error) {
	var out bool
	hr, _, _ := syscall.SyscallN(
		v.VTable().GetIsDiscoverable,
		uintptr(unsafe.Pointer(v)),    // this
		uintptr(unsafe.Pointer(&out)), // out bool
	)

	if hr != 0 {
		return false, ole.NewError(hr)
	}

	return out, nil
}

const GUIDiGattServiceProviderAdvertisingParameters2 string = "ff68468d-ca92-4434-9743-0e90988ad879"
const SignatureiGattServiceProviderAdvertisingParameters2 string = "{ff68468d-ca92-4434-9743-0e90988ad879}"

type iGattServiceProviderAdvertisingParameters2 struct {
	ole.IInspectable
}

type iGattServiceProviderAdvertisingParameters2Vtbl struct {
	ole.IInspectableVtbl

	SetServiceData uintptr
	GetServiceData uintptr
}

func (v *iGattServiceProviderAdvertisingParameters2) VTable() *iGattServiceProviderAdvertisingParameters2Vtbl {
	return (*iGattServiceProviderAdvertisingParameters2Vtbl)(unsafe.Pointer(v.RawVTable))
}

func (v *iGattServiceProviderAdvertisingParameters2) SetServiceData(value *streams.IBuffer) error {
	hr, _, _ := syscall.SyscallN(
		v.VTable().SetServiceData,
		uintptr(unsafe.Pointer(v)),     // this
		uintptr(unsafe.Pointer(value)), // in streams.IBuffer
	)

	if hr != 0 {
		return ole.NewError(hr)
	}

	return nil
}

func (v *iGattServiceProviderAdvertisingParameters2) GetServiceData() (*streams.IBuffer, error) {
	var out *streams.IBuffer
	hr, _, _ := syscall.SyscallN(
		v.VTable().GetServiceData,
		uintptr(unsafe.Pointer(v)),    // this
		uintptr(unsafe.Pointer(&out)), // out streams.IBuffer
	)

	if hr != 0 {
		return nil, ole.NewError(hr)
	}

	return out, nil
}