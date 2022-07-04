package winrt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test that we can create a `GUID` for a "generic" WinRT type.
const (
	guidTypedEventHandler                                    = "9de1c534-6ae1-11e0-84e1-18a905bcc53f"
	signatureBluetoothLEAdvertisementWatcher                 = "rc(Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementWatcher;{a6ac336f-f3d3-4297-8d6c-c81ea6623f40})"
	signatureBluetoothLEAdvertisementReceivedEventArgs       = "rc(Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementReceivedEventArgs;{27987ddf-e596-41be-8d43-9e6731d4a913})"
	signatureBluetoothLEAdvertisementWatcherStoppedEventArgs = "rc(Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementWatcherStoppedEventArgs;{dd40f84d-e7b9-43e3-9c04-0685d085fd8c})"
)

// TypedEventHandler<IBluetoothLEAdvertisementWatcher, IBluetoothLEAdvertisementReceivedEventArgs>
func TestParametrizedReceivedEvent(t *testing.T) {
	expected := "{90EB4ECA-D465-5EA0-A61C-033C8C5ECEF2}"
	guid := ParameterizedInstanceGUID(guidTypedEventHandler, signatureBluetoothLEAdvertisementWatcher, signatureBluetoothLEAdvertisementReceivedEventArgs)

	assert.Equal(t, expected, guid)
}

// TypedEventHandler<IBluetoothLEAdvertisementWatcher, IBluetoothLEAdvertisementReceivedEventArgs>
func TestParametrizedStoppedEvent(t *testing.T) {
	expected := "{9936A4DB-DC99-55C3-9E9B-BF4854BD9EAB}"
	guid := ParameterizedInstanceGUID(guidTypedEventHandler, signatureBluetoothLEAdvertisementWatcher, signatureBluetoothLEAdvertisementWatcherStoppedEventArgs)

	assert.Equal(t, expected, guid)
}
