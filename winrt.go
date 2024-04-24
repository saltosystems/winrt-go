package winrt

// common
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.IClosable
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.IAsyncOperation`1
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.AsyncOperationCompletedHandler`1
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.AsyncStatus
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.DateTime
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Deferral
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.DeferralCompletedHandler
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.IReference`1

// advertisement
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementWatcherStatus
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementWatcher -method-filter add_Received -method-filter remove_Received -method-filter add_Stopped -method-filter remove_Stopped -method-filter Start -method-filter Stop -method-filter get_Status -method-filter get_AllowExtendedAdvertisements -method-filter put_AllowExtendedAdvertisements -method-filter get_ScanningMode -method-filter put_ScanningMode -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementReceivedEventArgs -method-filter get_RawSignalStrengthInDBm -method-filter get_BluetoothAddress -method-filter get_Advertisement -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementWatcherStoppedEventArgs
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEManufacturerData
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisement -method-filter get_LocalName -method-filter put_LocalName -method-filter get_ServiceUuids -method-filter get_ManufacturerData -method-filter get_DataSections -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementDataSection -method-filter get_DataType -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementPublisher -method-filter get_Advertisement -method-filter Start -method-filter Stop -method-filter get_Status -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementPublisherStatus
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEScanningMode

// bluetooth
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.BluetoothLEDevice -method-filter FromBluetoothAddressAsync -method-filter FromBluetoothAddressWithBluetoothAddressTypeAsync -method-filter Close -method-filter get_ConnectionStatus -method-filter add_ConnectionStatusChanged -method-filter remove_ConnectionStatusChanged -method-filter get_BluetoothDeviceId -method-filter GetGattServicesWithCacheModeAsync -method-filter GetGattServicesAsync -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.BluetoothConnectionStatus
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.BluetoothAddressType
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.BluetoothDeviceId -method-filter !FromId
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.BluetoothCacheMode
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.BluetoothError

//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattSession -method-filter FromDeviceIdAsync -method-filter get_MaintainConnection -method-filter put_MaintainConnection -method-filter get_CanMaintainConnection -method-filter Close -method-filter get_MaxPduSize -method-filter add_MaxPduSizeChanged -method-filter remove_MaxPduSizeChanged -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattDeviceServicesResult -method-filter !get_ProtocolError
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattCommunicationStatus
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattDeviceService -method-filter get_Uuid -method-filter Close -method-filter GetCharacteristicsAsync -method-filter GetCharacteristicsWithCacheModeAsync -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattCharacteristicsResult -method-filter !get_ProtocolError
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattCharacteristic -method-filter get_Uuid -method-filter get_CharacteristicProperties -method-filter WriteValueWithOptionAsync -method-filter WriteValueAsync -method-filter ReadValueWithCacheModeAsync -method-filter ReadValueAsync -method-filter WriteClientCharacteristicConfigurationDescriptorAsync -method-filter add_ValueChanged -method-filter remove_ValueChanged -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattCharacteristicProperties
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattWriteOption
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattReadResult -method-filter !get_ProtocolError
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattClientCharacteristicConfigurationDescriptorValue
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattValueChangedEventArgs
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattClientNotificationResult
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattLocalCharacteristic
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattLocalCharacteristicParameters
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattLocalCharacteristicResult
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattLocalDescriptorParameters
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattLocalService
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattProtectionLevel
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattReadRequest
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattReadRequestedEventArgs
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattRequestState
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattServiceProvider
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattServiceProviderAdvertisementStatus
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattServiceProviderAdvertisingParameters
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattServiceProviderResult
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattSubscribedClient
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattWriteRequest
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattWriteRequestedEventArgs

// event
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.TypedEventHandler`2
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.EventRegistrationToken

// buffer
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.IBuffer
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.Buffer -method-filter !CreateCopyFromMemoryBuffer -method-filter !CreateMemoryBufferOverIBuffer
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.IDataReader -method-filter ReadBytes -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.DataReader -method-filter FromBuffer -method-filter ReadBytes -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.IDataWriter -method-filter WriteBytes -method-filter DetachBuffer -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.DataWriter -method-filter WriteBytes -method-filter DetachBuffer -method-filter DataWriter -method-filter Close -method-filter !*

// vector
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Collections.IVector`1
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Collections.IVectorView`1
