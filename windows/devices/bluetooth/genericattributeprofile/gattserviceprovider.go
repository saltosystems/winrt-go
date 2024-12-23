// Code generated by winrt-go-gen. DO NOT EDIT.

//go:build windows

//nolint:all
package genericattributeprofile

import (
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go/windows/foundation"
)

const SignatureGattServiceProvider string = "rc(Windows.Devices.Bluetooth.GenericAttributeProfile.GattServiceProvider;{7822b3cd-2889-4f86-a051-3f0aed1c2760})"

type GattServiceProvider struct {
	ole.IUnknown
}

func (impl *GattServiceProvider) GetService() (*GattLocalService, error) {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattServiceProvider))
	defer itf.Release()
	v := (*iGattServiceProvider)(unsafe.Pointer(itf))
	return v.GetService()
}

func (impl *GattServiceProvider) GetAdvertisementStatus() (GattServiceProviderAdvertisementStatus, error) {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattServiceProvider))
	defer itf.Release()
	v := (*iGattServiceProvider)(unsafe.Pointer(itf))
	return v.GetAdvertisementStatus()
}

func (impl *GattServiceProvider) AddAdvertisementStatusChanged(handler *foundation.TypedEventHandler) (foundation.EventRegistrationToken, error) {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattServiceProvider))
	defer itf.Release()
	v := (*iGattServiceProvider)(unsafe.Pointer(itf))
	return v.AddAdvertisementStatusChanged(handler)
}

func (impl *GattServiceProvider) RemoveAdvertisementStatusChanged(token foundation.EventRegistrationToken) error {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattServiceProvider))
	defer itf.Release()
	v := (*iGattServiceProvider)(unsafe.Pointer(itf))
	return v.RemoveAdvertisementStatusChanged(token)
}

func (impl *GattServiceProvider) StartAdvertising() error {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattServiceProvider))
	defer itf.Release()
	v := (*iGattServiceProvider)(unsafe.Pointer(itf))
	return v.StartAdvertising()
}

func (impl *GattServiceProvider) StartAdvertisingWithParameters(parameters *GattServiceProviderAdvertisingParameters) error {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattServiceProvider))
	defer itf.Release()
	v := (*iGattServiceProvider)(unsafe.Pointer(itf))
	return v.StartAdvertisingWithParameters(parameters)
}

func (impl *GattServiceProvider) StopAdvertising() error {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiGattServiceProvider))
	defer itf.Release()
	v := (*iGattServiceProvider)(unsafe.Pointer(itf))
	return v.StopAdvertising()
}

const GUIDiGattServiceProvider string = "7822b3cd-2889-4f86-a051-3f0aed1c2760"
const SignatureiGattServiceProvider string = "{7822b3cd-2889-4f86-a051-3f0aed1c2760}"

type iGattServiceProvider struct {
	ole.IInspectable
}

type iGattServiceProviderVtbl struct {
	ole.IInspectableVtbl

	GetService                       uintptr
	GetAdvertisementStatus           uintptr
	AddAdvertisementStatusChanged    uintptr
	RemoveAdvertisementStatusChanged uintptr
	StartAdvertising                 uintptr
	StartAdvertisingWithParameters   uintptr
	StopAdvertising                  uintptr
}

func (v *iGattServiceProvider) VTable() *iGattServiceProviderVtbl {
	return (*iGattServiceProviderVtbl)(unsafe.Pointer(v.RawVTable))
}

func (v *iGattServiceProvider) GetService() (*GattLocalService, error) {
	var out *GattLocalService
	hr, _, _ := syscall.SyscallN(
		v.VTable().GetService,
		uintptr(unsafe.Pointer(v)),    // this
		uintptr(unsafe.Pointer(&out)), // out GattLocalService
	)

	if hr != 0 {
		return nil, ole.NewError(hr)
	}

	return out, nil
}

func (v *iGattServiceProvider) GetAdvertisementStatus() (GattServiceProviderAdvertisementStatus, error) {
	var out GattServiceProviderAdvertisementStatus
	hr, _, _ := syscall.SyscallN(
		v.VTable().GetAdvertisementStatus,
		uintptr(unsafe.Pointer(v)),    // this
		uintptr(unsafe.Pointer(&out)), // out GattServiceProviderAdvertisementStatus
	)

	if hr != 0 {
		return GattServiceProviderAdvertisementStatusCreated, ole.NewError(hr)
	}

	return out, nil
}

func (v *iGattServiceProvider) AddAdvertisementStatusChanged(handler *foundation.TypedEventHandler) (foundation.EventRegistrationToken, error) {
	var out foundation.EventRegistrationToken
	hr, _, _ := syscall.SyscallN(
		v.VTable().AddAdvertisementStatusChanged,
		uintptr(unsafe.Pointer(v)),       // this
		uintptr(unsafe.Pointer(handler)), // in foundation.TypedEventHandler
		uintptr(unsafe.Pointer(&out)),    // out foundation.EventRegistrationToken
	)

	if hr != 0 {
		return foundation.EventRegistrationToken{}, ole.NewError(hr)
	}

	return out, nil
}

func (v *iGattServiceProvider) RemoveAdvertisementStatusChanged(token foundation.EventRegistrationToken) error {
	hr, _, _ := syscall.SyscallN(
		v.VTable().RemoveAdvertisementStatusChanged,
		uintptr(unsafe.Pointer(v)),      // this
		uintptr(unsafe.Pointer(&token)), // in foundation.EventRegistrationToken
	)

	if hr != 0 {
		return ole.NewError(hr)
	}

	return nil
}

func (v *iGattServiceProvider) StartAdvertising() error {
	hr, _, _ := syscall.SyscallN(
		v.VTable().StartAdvertising,
		uintptr(unsafe.Pointer(v)), // this
	)

	if hr != 0 {
		return ole.NewError(hr)
	}

	return nil
}

func (v *iGattServiceProvider) StartAdvertisingWithParameters(parameters *GattServiceProviderAdvertisingParameters) error {
	hr, _, _ := syscall.SyscallN(
		v.VTable().StartAdvertisingWithParameters,
		uintptr(unsafe.Pointer(v)),          // this
		uintptr(unsafe.Pointer(parameters)), // in GattServiceProviderAdvertisingParameters
	)

	if hr != 0 {
		return ole.NewError(hr)
	}

	return nil
}

func (v *iGattServiceProvider) StopAdvertising() error {
	hr, _, _ := syscall.SyscallN(
		v.VTable().StopAdvertising,
		uintptr(unsafe.Pointer(v)), // this
	)

	if hr != 0 {
		return ole.NewError(hr)
	}

	return nil
}

const GUIDiGattServiceProviderStatics string = "31794063-5256-4054-a4f4-7bbe7755a57e"
const SignatureiGattServiceProviderStatics string = "{31794063-5256-4054-a4f4-7bbe7755a57e}"

type iGattServiceProviderStatics struct {
	ole.IInspectable
}

type iGattServiceProviderStaticsVtbl struct {
	ole.IInspectableVtbl

	GattServiceProviderCreateAsync uintptr
}

func (v *iGattServiceProviderStatics) VTable() *iGattServiceProviderStaticsVtbl {
	return (*iGattServiceProviderStaticsVtbl)(unsafe.Pointer(v.RawVTable))
}

func GattServiceProviderCreateAsync(serviceUuid syscall.GUID) (*foundation.IAsyncOperation, error) {
	inspectable, err := ole.RoGetActivationFactory("Windows.Devices.Bluetooth.GenericAttributeProfile.GattServiceProvider", ole.NewGUID(GUIDiGattServiceProviderStatics))
	if err != nil {
		return nil, err
	}
	v := (*iGattServiceProviderStatics)(unsafe.Pointer(inspectable))

	var out *foundation.IAsyncOperation
	hr, _, _ := syscall.SyscallN(
		v.VTable().GattServiceProviderCreateAsync,
		uintptr(unsafe.Pointer(v)),            // this
		uintptr(unsafe.Pointer(&serviceUuid)), // in syscall.GUID
		uintptr(unsafe.Pointer(&out)),         // out foundation.IAsyncOperation
	)

	if hr != 0 {
		return nil, ole.NewError(hr)
	}

	return out, nil
}
