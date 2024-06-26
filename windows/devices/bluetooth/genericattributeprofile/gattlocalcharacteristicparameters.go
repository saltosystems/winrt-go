// Code generated by winrt-go-gen. DO NOT EDIT.

//go:build windows

//nolint:all
package genericattributeprofile

import (
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go/windows/foundation/collections"
	"github.com/saltosystems/winrt-go/windows/storage/streams"
)

const SignatureGattLocalCharacteristicParameters string = "rc(Windows.Devices.Bluetooth.GenericAttributeProfile.GattLocalCharacteristicParameters;{faf73db4-4cff-44c7-8445-040e6ead0063})"

type GattLocalCharacteristicParameters struct {
	ole.IUnknown
}

func NewGattLocalCharacteristicParameters() (*GattLocalCharacteristicParameters, error) {
	inspectable, err := ole.RoActivateInstance("Windows.Devices.Bluetooth.GenericAttributeProfile.GattLocalCharacteristicParameters")
	if err != nil {
		return nil, err
	}
	return (*GattLocalCharacteristicParameters)(unsafe.Pointer(inspectable)), nil
}

func (impl *GattLocalCharacteristicParameters) SetStaticValue(value *streams.IBuffer) error {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattLocalCharacteristicParameters))
	defer itf.Release()
	v := (*iGattLocalCharacteristicParameters)(unsafe.Pointer(itf))
	return v.SetStaticValue(value)
}

func (impl *GattLocalCharacteristicParameters) GetStaticValue() (*streams.IBuffer, error) {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattLocalCharacteristicParameters))
	defer itf.Release()
	v := (*iGattLocalCharacteristicParameters)(unsafe.Pointer(itf))
	return v.GetStaticValue()
}

func (impl *GattLocalCharacteristicParameters) SetCharacteristicProperties(value GattCharacteristicProperties) error {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattLocalCharacteristicParameters))
	defer itf.Release()
	v := (*iGattLocalCharacteristicParameters)(unsafe.Pointer(itf))
	return v.SetCharacteristicProperties(value)
}

func (impl *GattLocalCharacteristicParameters) GetCharacteristicProperties() (GattCharacteristicProperties, error) {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattLocalCharacteristicParameters))
	defer itf.Release()
	v := (*iGattLocalCharacteristicParameters)(unsafe.Pointer(itf))
	return v.GetCharacteristicProperties()
}

func (impl *GattLocalCharacteristicParameters) SetReadProtectionLevel(value GattProtectionLevel) error {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattLocalCharacteristicParameters))
	defer itf.Release()
	v := (*iGattLocalCharacteristicParameters)(unsafe.Pointer(itf))
	return v.SetReadProtectionLevel(value)
}

func (impl *GattLocalCharacteristicParameters) GetReadProtectionLevel() (GattProtectionLevel, error) {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattLocalCharacteristicParameters))
	defer itf.Release()
	v := (*iGattLocalCharacteristicParameters)(unsafe.Pointer(itf))
	return v.GetReadProtectionLevel()
}

func (impl *GattLocalCharacteristicParameters) SetWriteProtectionLevel(value GattProtectionLevel) error {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattLocalCharacteristicParameters))
	defer itf.Release()
	v := (*iGattLocalCharacteristicParameters)(unsafe.Pointer(itf))
	return v.SetWriteProtectionLevel(value)
}

func (impl *GattLocalCharacteristicParameters) GetWriteProtectionLevel() (GattProtectionLevel, error) {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattLocalCharacteristicParameters))
	defer itf.Release()
	v := (*iGattLocalCharacteristicParameters)(unsafe.Pointer(itf))
	return v.GetWriteProtectionLevel()
}

func (impl *GattLocalCharacteristicParameters) SetUserDescription(value string) error {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattLocalCharacteristicParameters))
	defer itf.Release()
	v := (*iGattLocalCharacteristicParameters)(unsafe.Pointer(itf))
	return v.SetUserDescription(value)
}

func (impl *GattLocalCharacteristicParameters) GetUserDescription() (string, error) {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattLocalCharacteristicParameters))
	defer itf.Release()
	v := (*iGattLocalCharacteristicParameters)(unsafe.Pointer(itf))
	return v.GetUserDescription()
}

func (impl *GattLocalCharacteristicParameters) GetPresentationFormats() (*collections.IVector, error) {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattLocalCharacteristicParameters))
	defer itf.Release()
	v := (*iGattLocalCharacteristicParameters)(unsafe.Pointer(itf))
	return v.GetPresentationFormats()
}

const GUIDiGattLocalCharacteristicParameters string = "faf73db4-4cff-44c7-8445-040e6ead0063"
const SignatureiGattLocalCharacteristicParameters string = "{faf73db4-4cff-44c7-8445-040e6ead0063}"

type iGattLocalCharacteristicParameters struct {
	ole.IInspectable
}

type iGattLocalCharacteristicParametersVtbl struct {
	ole.IInspectableVtbl

	SetStaticValue              uintptr
	GetStaticValue              uintptr
	SetCharacteristicProperties uintptr
	GetCharacteristicProperties uintptr
	SetReadProtectionLevel      uintptr
	GetReadProtectionLevel      uintptr
	SetWriteProtectionLevel     uintptr
	GetWriteProtectionLevel     uintptr
	SetUserDescription          uintptr
	GetUserDescription          uintptr
	GetPresentationFormats      uintptr
}

func (v *iGattLocalCharacteristicParameters) VTable() *iGattLocalCharacteristicParametersVtbl {
	return (*iGattLocalCharacteristicParametersVtbl)(unsafe.Pointer(v.RawVTable))
}

func (v *iGattLocalCharacteristicParameters) SetStaticValue(value *streams.IBuffer) error {
	hr, _, _ := syscall.SyscallN(
		v.VTable().SetStaticValue,
		uintptr(unsafe.Pointer(v)),     // this
		uintptr(unsafe.Pointer(value)), // in streams.IBuffer
	)

	if hr != 0 {
		return ole.NewError(hr)
	}

	return nil
}

func (v *iGattLocalCharacteristicParameters) GetStaticValue() (*streams.IBuffer, error) {
	var out *streams.IBuffer
	hr, _, _ := syscall.SyscallN(
		v.VTable().GetStaticValue,
		uintptr(unsafe.Pointer(v)),    // this
		uintptr(unsafe.Pointer(&out)), // out streams.IBuffer
	)

	if hr != 0 {
		return nil, ole.NewError(hr)
	}

	return out, nil
}

func (v *iGattLocalCharacteristicParameters) SetCharacteristicProperties(value GattCharacteristicProperties) error {
	hr, _, _ := syscall.SyscallN(
		v.VTable().SetCharacteristicProperties,
		uintptr(unsafe.Pointer(v)), // this
		uintptr(value),             // in GattCharacteristicProperties
	)

	if hr != 0 {
		return ole.NewError(hr)
	}

	return nil
}

func (v *iGattLocalCharacteristicParameters) GetCharacteristicProperties() (GattCharacteristicProperties, error) {
	var out GattCharacteristicProperties
	hr, _, _ := syscall.SyscallN(
		v.VTable().GetCharacteristicProperties,
		uintptr(unsafe.Pointer(v)),    // this
		uintptr(unsafe.Pointer(&out)), // out GattCharacteristicProperties
	)

	if hr != 0 {
		return GattCharacteristicPropertiesNone, ole.NewError(hr)
	}

	return out, nil
}

func (v *iGattLocalCharacteristicParameters) SetReadProtectionLevel(value GattProtectionLevel) error {
	hr, _, _ := syscall.SyscallN(
		v.VTable().SetReadProtectionLevel,
		uintptr(unsafe.Pointer(v)), // this
		uintptr(value),             // in GattProtectionLevel
	)

	if hr != 0 {
		return ole.NewError(hr)
	}

	return nil
}

func (v *iGattLocalCharacteristicParameters) GetReadProtectionLevel() (GattProtectionLevel, error) {
	var out GattProtectionLevel
	hr, _, _ := syscall.SyscallN(
		v.VTable().GetReadProtectionLevel,
		uintptr(unsafe.Pointer(v)),    // this
		uintptr(unsafe.Pointer(&out)), // out GattProtectionLevel
	)

	if hr != 0 {
		return GattProtectionLevelPlain, ole.NewError(hr)
	}

	return out, nil
}

func (v *iGattLocalCharacteristicParameters) SetWriteProtectionLevel(value GattProtectionLevel) error {
	hr, _, _ := syscall.SyscallN(
		v.VTable().SetWriteProtectionLevel,
		uintptr(unsafe.Pointer(v)), // this
		uintptr(value),             // in GattProtectionLevel
	)

	if hr != 0 {
		return ole.NewError(hr)
	}

	return nil
}

func (v *iGattLocalCharacteristicParameters) GetWriteProtectionLevel() (GattProtectionLevel, error) {
	var out GattProtectionLevel
	hr, _, _ := syscall.SyscallN(
		v.VTable().GetWriteProtectionLevel,
		uintptr(unsafe.Pointer(v)),    // this
		uintptr(unsafe.Pointer(&out)), // out GattProtectionLevel
	)

	if hr != 0 {
		return GattProtectionLevelPlain, ole.NewError(hr)
	}

	return out, nil
}

func (v *iGattLocalCharacteristicParameters) SetUserDescription(value string) error {
	valueHStr, err := ole.NewHString(value)
	if err != nil {
		return err
	}
	hr, _, _ := syscall.SyscallN(
		v.VTable().SetUserDescription,
		uintptr(unsafe.Pointer(v)), // this
		uintptr(valueHStr),         // in string
	)

	if hr != 0 {
		return ole.NewError(hr)
	}

	return nil
}

func (v *iGattLocalCharacteristicParameters) GetUserDescription() (string, error) {
	var outHStr ole.HString
	hr, _, _ := syscall.SyscallN(
		v.VTable().GetUserDescription,
		uintptr(unsafe.Pointer(v)),        // this
		uintptr(unsafe.Pointer(&outHStr)), // out string
	)

	if hr != 0 {
		return "", ole.NewError(hr)
	}

	out := outHStr.String()
	ole.DeleteHString(outHStr)
	return out, nil
}

func (v *iGattLocalCharacteristicParameters) GetPresentationFormats() (*collections.IVector, error) {
	var out *collections.IVector
	hr, _, _ := syscall.SyscallN(
		v.VTable().GetPresentationFormats,
		uintptr(unsafe.Pointer(v)),    // this
		uintptr(unsafe.Pointer(&out)), // out collections.IVector
	)

	if hr != 0 {
		return nil, ole.NewError(hr)
	}

	return out, nil
}
