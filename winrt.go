package winrt

// advertisement
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementWatcherStatus
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementWatcher -method-filter add_Received -method-filter add_Stopped -method-filter Start -method-filter Stop -method-filter get_Status -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementReceivedEventArgs -method-filter get_RawSignalStrengthInDBm -method-filter get_BluetoothAddress -method-filter get_Advertisement -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementWatcherStoppedEventArgs -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEManufacturerData
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisement -method-filter get_LocalName -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementDataSection -method-filter get_DataType -method-filter !*

// event
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.TypedEventHandler`2
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.EventRegistrationToken

// buffer
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.IBuffer
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.Buffer -method-filter !CreateCopyFromMemoryBuffer -method-filter !CreateMemoryBufferOverIBuffer
