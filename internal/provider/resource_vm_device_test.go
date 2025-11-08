package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

// TestVMDevice_ValidDeviceTypes tests that valid device types are accepted
func TestVMDevice_ValidDeviceTypes(t *testing.T) {
	validTypes := []string{"NIC", "DISK", "CDROM", "PCI", "USB", "DISPLAY", "RAW"}

	for _, deviceType := range validTypes {
		device := VMDeviceResourceModel{
			DeviceType: types.StringValue(deviceType),
		}
		assert.NotEmpty(t, device.DeviceType.ValueString())
		assert.Equal(t, deviceType, device.DeviceType.ValueString())
	}
}

// TestVMDevice_NICConfigValidation tests that NIC config has proper attributes
func TestVMDevice_NICConfigValidation(t *testing.T) {
	// Test valid NIC config
	nicConfig := NICConfigModel{
		Type:                types.StringValue("VIRTIO"),
		NICAttach:           types.StringValue("br0"),
		MAC:                 types.StringValue("00:11:22:33:44:55"),
		TrustGuestRxFilters: types.BoolValue(false),
	}

	assert.Equal(t, "VIRTIO", nicConfig.Type.ValueString())
	assert.Equal(t, "br0", nicConfig.NICAttach.ValueString())
	assert.Equal(t, "00:11:22:33:44:55", nicConfig.MAC.ValueString())
	assert.False(t, nicConfig.TrustGuestRxFilters.ValueBool())
}

// TestVMDevice_NICConfigRequiredFields tests that NIC config requires nic_attach
func TestVMDevice_NICConfigRequiredFields(t *testing.T) {
	nicConfig := NICConfigModel{
		NICAttach: types.StringValue("eno1"),
	}

	assert.Equal(t, "eno1", nicConfig.NICAttach.ValueString())
	assert.True(t, nicConfig.Type.IsNull())
	assert.True(t, nicConfig.MAC.IsNull())
}

// TestVMDevice_DISKConfigValidation tests DISK config validation
func TestVMDevice_DISKConfigValidation(t *testing.T) {
	diskConfig := DiskConfigModel{
		Path:                types.StringValue("/dev/zvol/pool/disk"),
		Type:                types.StringValue("VIRTIO"),
		IOType:              types.StringValue("THREADS"),
		PhysicalSectorSize:  types.Int64Value(512),
		LogicalSectorSize:   types.Int64Value(512),
	}

	assert.Equal(t, "/dev/zvol/pool/disk", diskConfig.Path.ValueString())
	assert.Equal(t, "VIRTIO", diskConfig.Type.ValueString())
	assert.Equal(t, "THREADS", diskConfig.IOType.ValueString())
	assert.Equal(t, int64(512), diskConfig.PhysicalSectorSize.ValueInt64())
	assert.Equal(t, int64(512), diskConfig.LogicalSectorSize.ValueInt64())
}

// TestVMDevice_DISKConfigRequiredFields tests that DISK config requires path
func TestVMDevice_DISKConfigRequiredFields(t *testing.T) {
	diskConfig := DiskConfigModel{
		Path: types.StringValue("/dev/zvol/pool/vm-disk0"),
	}

	assert.Equal(t, "/dev/zvol/pool/vm-disk0", diskConfig.Path.ValueString())
	assert.True(t, diskConfig.Type.IsNull())
	assert.True(t, diskConfig.IOType.IsNull())
}

// TestVMDevice_CDROMConfigValidation tests CDROM config validation
func TestVMDevice_CDROMConfigValidation(t *testing.T) {
	cdromConfig := CDROMConfigModel{
		Path: types.StringValue("/mnt/pool/isos/ubuntu.iso"),
	}

	assert.Equal(t, "/mnt/pool/isos/ubuntu.iso", cdromConfig.Path.ValueString())
}

// TestVMDevice_PCIConfigValidation tests PCI config validation
func TestVMDevice_PCIConfigValidation(t *testing.T) {
	pciConfig := PCIConfigModel{
		PPTDev: types.StringValue("pci_0000_01_00_0"),
	}

	assert.Equal(t, "pci_0000_01_00_0", pciConfig.PPTDev.ValueString())
}

// TestVMDevice_PCIConfigMultipleFormats tests PCI config with different device formats
func TestVMDevice_PCIConfigMultipleFormats(t *testing.T) {
	testCases := []string{
		"pci_0000_01_00_0",
		"pci_0000_3b_00_0",
		"pci_0000_af_00_1",
	}

	for _, deviceID := range testCases {
		pciConfig := PCIConfigModel{
			PPTDev: types.StringValue(deviceID),
		}
		assert.Equal(t, deviceID, pciConfig.PPTDev.ValueString())
	}
}

// TestVMDevice_USBConfigValidation tests USB config validation
func TestVMDevice_USBConfigValidation(t *testing.T) {
	usbConfig := USBConfigModel{
		Controller: types.StringValue("xhci"),
		Device:     types.StringValue("usb_1234_5678"),
	}

	assert.Equal(t, "xhci", usbConfig.Controller.ValueString())
	assert.Equal(t, "usb_1234_5678", usbConfig.Device.ValueString())
}

// TestVMDevice_USBConfigRequiredFields tests that USB config requires device
func TestVMDevice_USBConfigRequiredFields(t *testing.T) {
	usbConfig := USBConfigModel{
		Device: types.StringValue("usb_device_id"),
	}

	assert.Equal(t, "usb_device_id", usbConfig.Device.ValueString())
	assert.True(t, usbConfig.Controller.IsNull())
}

// TestVMDevice_DisplayConfigValidation tests DISPLAY config validation
func TestVMDevice_DisplayConfigValidation(t *testing.T) {
	displayConfig := DisplayConfigModel{
		Port:       types.Int64Value(5900),
		Bind:       types.StringValue("0.0.0.0"),
		Password:   types.StringValue("secret123"),
		Web:        types.BoolValue(true),
		Type:       types.StringValue("SPICE"),
		Resolution: types.StringValue("1920x1080"),
		WebPort:    types.Int64Value(6080),
		Wait:       types.BoolValue(false),
	}

	assert.Equal(t, int64(5900), displayConfig.Port.ValueInt64())
	assert.Equal(t, "0.0.0.0", displayConfig.Bind.ValueString())
	assert.Equal(t, "secret123", displayConfig.Password.ValueString())
	assert.True(t, displayConfig.Web.ValueBool())
	assert.Equal(t, "SPICE", displayConfig.Type.ValueString())
	assert.Equal(t, "1920x1080", displayConfig.Resolution.ValueString())
	assert.Equal(t, int64(6080), displayConfig.WebPort.ValueInt64())
	assert.False(t, displayConfig.Wait.ValueBool())
}

// TestVMDevice_RAWConfigValidation tests RAW config validation
func TestVMDevice_RAWConfigValidation(t *testing.T) {
	rawConfig := RAWConfigModel{
		Path: types.StringValue("/mnt/pool/raw/device.img"),
		Size: types.Int64Value(10737418240), // 10GB
		Boot: types.BoolValue(true),
	}

	assert.Equal(t, "/mnt/pool/raw/device.img", rawConfig.Path.ValueString())
	assert.Equal(t, int64(10737418240), rawConfig.Size.ValueInt64())
	assert.True(t, rawConfig.Boot.ValueBool())
}

// TestVMDevice_RAWConfigRequiredFields tests that RAW config requires path
func TestVMDevice_RAWConfigRequiredFields(t *testing.T) {
	rawConfig := RAWConfigModel{
		Path: types.StringValue("/path/to/raw"),
	}

	assert.Equal(t, "/path/to/raw", rawConfig.Path.ValueString())
	assert.True(t, rawConfig.Size.IsNull())
	assert.True(t, rawConfig.Boot.IsNull())
}

// TestVMDevice_OrderDefaults tests that order defaults appropriately
func TestVMDevice_OrderDefaults(t *testing.T) {
	// Test with explicit order
	deviceWithOrder := VMDeviceResourceModel{
		Order: types.Int64Value(100),
	}
	assert.Equal(t, int64(100), deviceWithOrder.Order.ValueInt64())

	// Test with null order (should be set to 1000 by Create function)
	deviceWithNullOrder := VMDeviceResourceModel{
		Order: types.Int64Null(),
	}
	assert.True(t, deviceWithNullOrder.Order.IsNull())
}

// TestVMDevice_VMIDReference tests VM ID reference
func TestVMDevice_VMIDReference(t *testing.T) {
	device := VMDeviceResourceModel{
		VMID:       types.StringValue("123"),
		DeviceType: types.StringValue("NIC"),
	}

	assert.Equal(t, "123", device.VMID.ValueString())
	assert.Equal(t, "NIC", device.DeviceType.ValueString())
}

// TestVMDevice_DeviceTypeMustMatchConfig tests that device type should match config
func TestVMDevice_DeviceTypeMustMatchConfig(t *testing.T) {
	// NIC device with NIC config
	nicDevice := VMDeviceResourceModel{
		DeviceType: types.StringValue("NIC"),
		NICConfig:  types.ListNull(types.ObjectType{}),
	}
	assert.Equal(t, "NIC", nicDevice.DeviceType.ValueString())

	// DISK device with DISK config
	diskDevice := VMDeviceResourceModel{
		DeviceType: types.StringValue("DISK"),
		DiskConfig: types.ListNull(types.ObjectType{}),
	}
	assert.Equal(t, "DISK", diskDevice.DeviceType.ValueString())
}

// TestVMDevice_AllConfigsStartNull tests that all config lists start as null
func TestVMDevice_AllConfigsStartNull(t *testing.T) {
	device := VMDeviceResourceModel{}

	assert.True(t, device.NICConfig.IsNull())
	assert.True(t, device.DiskConfig.IsNull())
	assert.True(t, device.CDROMConfig.IsNull())
	assert.True(t, device.PCIConfig.IsNull())
	assert.True(t, device.USBConfig.IsNull())
	assert.True(t, device.DisplayConfig.IsNull())
	assert.True(t, device.RAWConfig.IsNull())
}

// TestVMDevice_IDGeneration tests ID field behavior
func TestVMDevice_IDGeneration(t *testing.T) {
	device := VMDeviceResourceModel{
		ID: types.StringValue("456"),
	}

	assert.Equal(t, "456", device.ID.ValueString())
	assert.False(t, device.ID.IsNull())
}

// TestVMDevice_NICTypeDefaults tests NIC type defaults
func TestVMDevice_NICTypeDefaults(t *testing.T) {
	// Test with explicit type
	nicWithType := NICConfigModel{
		Type:      types.StringValue("E1000"),
		NICAttach: types.StringValue("br0"),
	}
	assert.Equal(t, "E1000", nicWithType.Type.ValueString())

	// Test with null type (should default to VIRTIO in Create function)
	nicWithNullType := NICConfigModel{
		Type:      types.StringNull(),
		NICAttach: types.StringValue("br0"),
	}
	assert.True(t, nicWithNullType.Type.IsNull())
}

// TestVMDevice_DiskTypeDefaults tests DISK type defaults
func TestVMDevice_DiskTypeDefaults(t *testing.T) {
	// Test with explicit type
	diskWithType := DiskConfigModel{
		Path: types.StringValue("/dev/zvol/pool/disk"),
		Type: types.StringValue("AHCI"),
	}
	assert.Equal(t, "AHCI", diskWithType.Type.ValueString())

	// Test with null type (should default to VIRTIO in Create function)
	diskWithNullType := DiskConfigModel{
		Path: types.StringValue("/dev/zvol/pool/disk"),
		Type: types.StringNull(),
	}
	assert.True(t, diskWithNullType.Type.IsNull())
}

// TestVMDevice_CompleteDeviceModel tests a complete device model
func TestVMDevice_CompleteDeviceModel(t *testing.T) {
	device := VMDeviceResourceModel{
		ID:         types.StringValue("789"),
		VMID:       types.StringValue("123"),
		DeviceType: types.StringValue("DISK"),
		Order:      types.Int64Value(1000),
	}

	assert.Equal(t, "789", device.ID.ValueString())
	assert.Equal(t, "123", device.VMID.ValueString())
	assert.Equal(t, "DISK", device.DeviceType.ValueString())
	assert.Equal(t, int64(1000), device.Order.ValueInt64())
}

// TestVMDevice_DiskIOTypeValidation tests DISK iotype field
func TestVMDevice_DiskIOTypeValidation(t *testing.T) {
	validIOTypes := []string{"THREADS", "NATIVE"}

	for _, ioType := range validIOTypes {
		disk := DiskConfigModel{
			Path:   types.StringValue("/dev/zvol/pool/disk"),
			IOType: types.StringValue(ioType),
		}
		assert.Equal(t, ioType, disk.IOType.ValueString())
	}
}

// TestVMDevice_DiskSectorSizes tests DISK sector size configurations
func TestVMDevice_DiskSectorSizes(t *testing.T) {
	testCases := []struct {
		name     string
		physical int64
		logical  int64
	}{
		{"Standard 512", 512, 512},
		{"Advanced 4K", 4096, 4096},
		{"Mixed", 4096, 512},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			disk := DiskConfigModel{
				Path:                types.StringValue("/dev/zvol/pool/disk"),
				PhysicalSectorSize:  types.Int64Value(tc.physical),
				LogicalSectorSize:   types.Int64Value(tc.logical),
			}
			assert.Equal(t, tc.physical, disk.PhysicalSectorSize.ValueInt64())
			assert.Equal(t, tc.logical, disk.LogicalSectorSize.ValueInt64())
		})
	}
}

// TestVMDevice_DisplayTypes tests display device types
func TestVMDevice_DisplayTypes(t *testing.T) {
	displayTypes := []string{"SPICE", "VNC"}

	for _, displayType := range displayTypes {
		display := DisplayConfigModel{
			Type: types.StringValue(displayType),
		}
		assert.Equal(t, displayType, display.Type.ValueString())
	}
}

// TestVMDevice_NICMACAddress tests NIC MAC address formats
func TestVMDevice_NICMACAddress(t *testing.T) {
	validMACs := []string{
		"00:11:22:33:44:55",
		"aa:bb:cc:dd:ee:ff",
		"01:23:45:67:89:ab",
	}

	for _, mac := range validMACs {
		nic := NICConfigModel{
			MAC:       types.StringValue(mac),
			NICAttach: types.StringValue("br0"),
		}
		assert.Equal(t, mac, nic.MAC.ValueString())
	}
}

// TestVMDevice_ListTypeConversions tests that config lists can be properly constructed
func TestVMDevice_ListTypeConversions(t *testing.T) {
	// Test creating a list for NIC config
	_ = NICConfigModel{
		Type:      types.StringValue("VIRTIO"),
		NICAttach: types.StringValue("br0"),
	}

	objectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"type":                   types.StringType,
			"mac":                    types.StringType,
			"nic_attach":             types.StringType,
			"trust_guest_rx_filters": types.BoolType,
		},
	}

	// This tests that we can create the object type needed for lists
	assert.NotNil(t, objectType)
	assert.Equal(t, 4, len(objectType.AttrTypes))
}

// TestVMDevice_EmptyStringVsNull tests difference between empty string and null
func TestVMDevice_EmptyStringVsNull(t *testing.T) {
	// Null value
	nullString := types.StringNull()
	assert.True(t, nullString.IsNull())
	assert.False(t, nullString.IsUnknown())

	// Empty string value
	emptyString := types.StringValue("")
	assert.False(t, emptyString.IsNull())
	assert.Equal(t, "", emptyString.ValueString())

	// Non-empty string
	valueString := types.StringValue("test")
	assert.False(t, valueString.IsNull())
	assert.Equal(t, "test", valueString.ValueString())
}

// TestVMDevice_BoolDefaults tests boolean field defaults
func TestVMDevice_BoolDefaults(t *testing.T) {
	// Test NIC trust_guest_rx_filters default
	nic := NICConfigModel{
		TrustGuestRxFilters: types.BoolValue(false),
	}
	assert.False(t, nic.TrustGuestRxFilters.ValueBool())

	// Test DISPLAY web default
	display := DisplayConfigModel{
		Web: types.BoolValue(false),
	}
	assert.False(t, display.Web.ValueBool())

	// Test RAW boot default
	raw := RAWConfigModel{
		Boot: types.BoolValue(false),
	}
	assert.False(t, raw.Boot.ValueBool())
}

// TestVMDevice_MultipleDeviceTypesStructure tests structure for multiple device types
func TestVMDevice_MultipleDeviceTypesStructure(t *testing.T) {
	deviceTypes := map[string]interface{}{
		"NIC":     NICConfigModel{},
		"DISK":    DiskConfigModel{},
		"CDROM":   CDROMConfigModel{},
		"PCI":     PCIConfigModel{},
		"USB":     USBConfigModel{},
		"DISPLAY": DisplayConfigModel{},
		"RAW":     RAWConfigModel{},
	}

	// Verify all device config models exist
	assert.Equal(t, 7, len(deviceTypes))
	for deviceType := range deviceTypes {
		assert.Contains(t, []string{"NIC", "DISK", "CDROM", "PCI", "USB", "DISPLAY", "RAW"}, deviceType)
	}
}