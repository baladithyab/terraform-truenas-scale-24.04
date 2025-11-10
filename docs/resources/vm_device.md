---
page_title: "truenas_vm_device Resource - terraform-provider-truenas"
subcategory: "Virtual Machines"
description: |-
  Manages a VM device independently on TrueNAS.
---

# truenas_vm_device (Resource)

Manages a VM device independently on TrueNAS. This resource allows you to manage VM devices (NICs, Disks, CDROMs, PCI devices, USB devices, Displays, and RAW devices) separately from the VM itself, providing fine-grained control over virtual hardware.

## Example Usage

### Network Interface Device

```terraform
resource "truenas_vm_device" "vm_nic" {
  vm_id       = "1"
  device_type = "NIC"
  order       = 100
  
  nic_config = [{
    nic_attach           = "br0"
    type                 = "VIRTIO"
    mac                  = "52:54:00:ab:cd:ef"
    trust_guest_rx_filters = false
  }]
}
```

### Disk Device

```terraform
resource "truenas_vm_device" "vm_disk" {
  vm_id       = "1"
  device_type = "DISK"
  order       = 200
  
  disk_config = [{
    path                = "/dev/zvol/tank/vm1-disk0"
    type                = "VIRTIO"
    iotype              = "THREADS"
    physical_sectorsize  = 512
    logical_sectorsize   = 512
  }]
}
```

### CDROM Device

```terraform
resource "truenas_vm_device" "vm_cdrom" {
  vm_id       = "1"
  device_type = "CDROM"
  order       = 300
  
  cdrom_config = [{
    path = "/mnt/tank/iso/ubuntu-22.04.iso"
  }]
}
```

### PCI Passthrough Device

```terraform
resource "truenas_vm_device" "vm_pci" {
  vm_id       = "1"
  device_type = "PCI"
  order       = 400
  
  pci_config = [{
    pptdev = "pci_0000_3b_00_0"
  }]
}
```

### USB Device

```terraform
resource "truenas_vm_device" "vm_usb" {
  vm_id       = "1"
  device_type = "USB"
  order       = 500
  
  usb_config = [{
    controller = "ehci"
    device     = "usb_1_1_2"
  }]
}
```

### Display Device

```terraform
resource "truenas_vm_device" "vm_display" {
  vm_id       = "1"
  device_type = "DISPLAY"
  order       = 600
  
  display_config = [{
    type       = "VNC"
    port       = 5900
    bind       = "0.0.0.0"
    password   = "secure-password"
    web        = true
    resolution = "1920x1080"
    web_port   = 5901
    wait       = false
  }]
}
```

### RAW Device

```terraform
resource "truenas_vm_device" "vm_raw" {
  vm_id       = "1"
  device_type = "RAW"
  order       = 700
  
  raw_config = [{
    path = "/mnt/tank/vm1-raw.img"
    size = 10737418240  # 10GB
    boot = false
  }]
}
```

### Complete VM with Multiple Devices

```terraform
# Boot disk
resource "truenas_vm_device" "boot_disk" {
  vm_id       = "1"
  device_type = "DISK"
  order       = 100
  
  disk_config = [{
    path   = "/dev/zvol/tank/vm1-boot"
    type   = "VIRTIO"
    boot   = true
  }]
}

# Data disk
resource "truenas_vm_device" "data_disk" {
  vm_id       = "1"
  device_type = "DISK"
  order       = 200
  
  disk_config = [{
    path   = "/dev/zvol/tank/vm1-data"
    type   = "VIRTIO"
  }]
}

# Network interface
resource "truenas_vm_device" "network" {
  vm_id       = "1"
  device_type = "NIC"
  order       = 300
  
  nic_config = [{
    nic_attach = "br0"
    type       = "VIRTIO"
  }]
}

# GPU passthrough
resource "truenas_vm_device" "gpu" {
  vm_id       = "1"
  device_type = "PCI"
  order       = 400
  
  pci_config = [{
    pptdev = "pci_0000_01_00_0"
  }]
}
```

## Schema

### Required

- `vm_id` (String) ID of the VM this device belongs to.
- `device_type` (String) Type of device. Options: `NIC`, `DISK`, `CDROM`, `PCI`, `USB`, `DISPLAY`, `RAW`.

### Optional

- `order` (Number) Boot order for this device. Lower values boot first. Default: 1000.
- `nic_config` (Block List) Configuration for NIC device (required when device_type is NIC).
- `disk_config` (Block List) Configuration for DISK device (required when device_type is DISK).
- `cdrom_config` (Block List) Configuration for CDROM device (required when device_type is CDROM).
- `pci_config` (Block List) Configuration for PCI device (required when device_type is PCI).
- `usb_config` (Block List) Configuration for USB device (required when device_type is USB).
- `display_config` (Block List) Configuration for DISPLAY device (required when device_type is DISPLAY).
- `raw_config` (Block List) Configuration for RAW device (required when device_type is RAW).

### NIC Configuration

The `nic_config` block supports:

- `type` (String) NIC type. Options: `VIRTIO`, `E1000`, etc. Default: `VIRTIO`.
- `mac` (String) MAC address (leave empty for auto-generation).
- `nic_attach` (String, Required) Physical network interface to attach to (e.g., eno1, br0).
- `trust_guest_rx_filters` (Boolean) Trust guest RX filters. Default: false.

### Disk Configuration

The `disk_config` block supports:

- `path` (String, Required) Path to disk (e.g., /dev/zvol/pool/vm-disk0).
- `type` (String) Disk type. Options: `VIRTIO`, `AHCI`, etc. Default: `VIRTIO`.
- `iotype` (String) IO type. Options: `THREADS`, `NATIVE`. Default: `THREADS`.
- `physical_sectorsize` (Number) Physical sector size in bytes.
- `logical_sectorsize` (Number) Logical sector size in bytes.

### CDROM Configuration

The `cdrom_config` block supports:

- `path` (String, Required) Path to ISO file (e.g., /mnt/pool/isos/ubuntu.iso).

### PCI Configuration

The `pci_config` block supports:

- `pptdev` (String, Required) PCI device ID to pass through (e.g., 'pci_0000_3b_00_0').

### USB Configuration

The `usb_config` block supports:

- `controller` (String) USB controller type.
- `device` (String, Required) USB device identifier.

### Display Configuration

The `display_config` block supports:

- `port` (Number) Port number for display server.
- `bind` (String) IP address to bind to.
- `password` (String, Sensitive) Password for display access.
- `web` (Boolean) Enable web access.
- `type` (String) Display type: SPICE or VNC.
- `resolution` (String) Display resolution.
- `web_port` (Number) Port for web access.
- `wait` (Boolean) Wait for client connection.

### RAW Configuration

The `raw_config` block supports:

- `path` (String, Required) Path to raw file.
- `size` (Number) Size in bytes.
- `boot` (Boolean) Whether this is a boot device.

### Read-Only

- `id` (String) Device identifier.

## Import

VM devices can be imported using device ID:

```shell
terraform import truenas_vm_device.existing 1
```

## Notes

### Device Types

#### NIC (Network Interface)
- Provides network connectivity to VMs
- Supports various virtual NIC types
- Can attach to bridges or physical interfaces

#### DISK
- Provides storage devices to VMs
- Supports various disk controllers
- Can use ZFS volumes or raw files

#### CDROM
- Provides virtual CD/DVD drive
- Used for OS installation and media
- Supports ISO file mounting

#### PCI
- Enables PCI device passthrough
- Used for GPU passthrough and specialized hardware
- Requires IOMMU support

#### USB
- Provides USB device access
- Can pass through physical USB devices
- Useful for dongles and peripherals

#### DISPLAY
- Provides graphics and input access
- Supports VNC and SPICE protocols
- Enables remote console access

#### RAW
- Provides raw device access
- Can create custom device types
- Flexible device configuration

### Boot Order

Control device boot priority:

```terraform
order = 100  # First boot device
order = 200  # Second boot device
order = 300  # Third boot device
```

Lower numbers boot first. Typical order:
1. CDROM/DVD (for installation)
2. Boot disk
3. Data disks
4. Network devices

### Network Interface Configuration

#### NIC Types
- **VIRTIO**: Paravirtualized, best performance
- **E1000**: Intel E1000 emulation, good compatibility
- **RTL8139**: Realtek 8139 emulation, maximum compatibility

#### MAC Address
```terraform
mac = "52:54:00:ab:cd:ef"  # Specific MAC
mac = ""                      # Auto-generate MAC
```

#### Network Attachment
```terraform
nic_attach = "br0"     # Bridge interface
nic_attach = "eno1"    # Physical interface
nic_attach = "vlan10"   # VLAN interface
```

### Disk Configuration

#### Disk Types
- **VIRTIO**: Paravirtualized, best performance
- **AHCI**: SATA controller emulation, good compatibility
- **IDE**: Legacy IDE emulation, maximum compatibility

#### IO Types
- **THREADS**: Threaded I/O (default)
- **NATIVE**: Native I/O, potentially better performance

#### Sector Sizes
```terraform
physical_sectorsize = 512   # Standard sector size
physical_sectorsize = 4096  # Advanced format
logical_sectorsize  = 512   # Common logical size
```

### PCI Passthrough

#### Requirements
- IOMMU must be enabled in BIOS
- VT-d/AMD-Vi support required
- Proper kernel configuration

#### Device Identification
```terraform
pptdev = "pci_0000_3b_00_0"  # Format: pci_domain_bus_slot_function
```

Use `truenas_vm_pci_passthrough_devices` data source to discover available devices.

### Display Configuration

#### Display Types
- **VNC**: VNC protocol, widely supported
- **SPICE**: SPICE protocol, better performance and features

#### VNC Configuration
```terraform
display_config = [{
  type       = "VNC"
  port       = 5900
  bind       = "0.0.0.0"
  password   = "secure-password"
  web        = true
  resolution = "1920x1080"
}]
```

#### SPICE Configuration
```terraform
display_config = [{
  type     = "SPICE"
  port     = 5900
  web      = true
  web_port = 5901
}]
```

## Best Practices

### Device Management

1. **Consistent Ordering**: Use logical boot order numbers
2. **Device Naming**: Use descriptive device names
3. **Documentation**: Document device purposes and configurations
4. **Testing**: Test device configurations before production use

### Performance

1. **VirtIO Devices**: Use VirtIO for best performance
2. **Appropriate Types**: Choose device types based on guest OS
3. **Resource Allocation**: Balance device performance with host resources
4. **Monitoring**: Monitor device performance and utilization

### Security

1. **PCI Passthrough**: Secure passthrough devices properly
2. **Network Isolation**: Use appropriate network attachments
3. **Access Control**: Secure display and USB access
4. **Regular Audits**: Review device configurations regularly

### Maintenance

1. **Device Cleanup**: Remove unused devices
2. **Configuration Backup**: Document device configurations
3. **Performance Tuning**: Adjust settings based on usage
4. **Compatibility Testing**: Test with different guest OS versions

## Troubleshooting

### Device Not Recognized

1. Verify device type and configuration
2. Check VM compatibility
3. Review system logs
4. Test with different device types

### Boot Order Issues

1. Verify order values are correct
2. Check device boot capabilities
3. Test different order sequences
4. Ensure boot device is properly configured

### Network Problems

1. Verify NIC attachment exists
2. Check network interface status
3. Test with different NIC types
4. Review bridge configuration

### Disk Issues

1. Verify disk path exists
2. Check disk permissions
3. Test with different disk types
4. Review storage pool status

### PCI Passthrough Failures

1. Verify IOMMU is enabled
2. Check device availability
3. Review kernel configuration
4. Test with different devices

### Display Connection Problems

1. Verify display configuration
2. Check port availability
3. Test network connectivity
4. Review firewall settings

## See Also

- [truenas_vm](vm) - VM management
- [truenas_vm_pci_passthrough_devices](../data-sources/vm_pci_passthrough_devices) - PCI device discovery
- [truenas_vm_iommu_enabled](../data-sources/vm_iommu_enabled) - IOMMU capability checking
- [TrueNAS VM Documentation](https://www.truenas.com/docs/scale/virtualmachines/) - Official VM documentation
- [VM Device Management](https://www.truenas.com/docs/scale/virtualmachines/devices/) - Device configuration guide