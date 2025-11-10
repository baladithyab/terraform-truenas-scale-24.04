# Release Notes - v0.2.14

**Release Date:** 2025-10-31

## 🎯 Boot Order Control

This release adds the ability to control the boot order of VM devices, fixing a critical bug where devices always booted in a fixed order (NICs → Disks → CDROMs) regardless of user configuration.

---

## 🐛 The Bug That Was Fixed

### Before v0.2.14 (BROKEN)

The provider was creating devices with hardcoded boot order based on device **type**:
- **NICs** always got order 1000, 1001, 1002...
- **Disks** always got order 1003, 1004, 1005...
- **CDROMs** always got order 1006, 1007, 1008...

This meant:
- ❌ **Disks ALWAYS booted before CDROMs** - impossible to install OS from ISO
- ❌ **No way to control boot priority** - users couldn't specify boot order
- ❌ **Talos Linux installations failed** - needed to boot from ISO first

### After v0.2.14 (FIXED)

Users can now specify the `order` attribute for each device:
- ✅ **Lower order boots first** - full control over boot sequence
- ✅ **Install OS from ISO** - set CDROM order=1, Disk order=2
- ✅ **Talos Linux works** - boot from ISO for install, disk for operation
- ✅ **Backward compatible** - if order not specified, uses auto-incrementing (1000+)

---

## ✨ What's New

### Boot Order Attribute

All device types now support the `order` attribute:

```hcl
nic_devices = [{
  type       = "VIRTIO"
  nic_attach = "eno1"
  order      = 1003  # Optional: boot order
}]

disk_devices = [{
  path  = "/dev/zvol/pool/vms/disk0"
  type  = "VIRTIO"
  order = 1  # Lower order = boots FIRST
}]

cdrom_devices = [{
  path  = "/mnt/pool/isos/ubuntu.iso"
  order = 2  # Higher order = boots SECOND
}]
```

### How It Works

The `order` field maps directly to libvirt's `<boot order='X'/>` attribute in the VM's XML configuration. This controls which device the BIOS/UEFI firmware tries to boot from first.

**Lower order value = Higher boot priority**

---

## 📖 Common Use Cases

### 1. Install OS from ISO

Boot from CDROM first to install OS, then from disk:

```hcl
resource "truenas_vm" "install_ubuntu" {
  name   = "ubuntu"
  memory = 4096
  vcpus  = 2

  # CDROM boots FIRST (order 1)
  cdrom_devices = [{
    path  = "/mnt/pool/isos/ubuntu-22.04.iso"
    order = 1
  }]

  # Disk boots SECOND (order 2)
  disk_devices = [{
    path  = "/dev/zvol/pool/vms/ubuntu-disk0"
    type  = "VIRTIO"
    order = 2
  }]

  nic_devices = [{
    type       = "VIRTIO"
    nic_attach = "eno1"
  }]

  bootloader = "UEFI"
}
```

**Boot sequence:**
1. VM boots from Ubuntu ISO
2. Install Ubuntu to disk
3. After installation, disk has bootloader and takes over
4. ISO remains attached but boots second

### 2. Normal Operation (Boot from Disk)

After OS is installed, boot from disk first:

```hcl
resource "truenas_vm" "production" {
  name   = "production"
  memory = 8192
  vcpus  = 4

  # Disk boots FIRST (order 1)
  disk_devices = [{
    path  = "/dev/zvol/pool/vms/prod-disk0"
    type  = "VIRTIO"
    order = 1
  }]

  # CDROM boots SECOND (fallback/rescue)
  cdrom_devices = [{
    path  = "/mnt/pool/isos/rescue.iso"
    order = 2
  }]

  nic_devices = [{
    type       = "VIRTIO"
    nic_attach = "eno1"
  }]

  bootloader = "UEFI"
}
```

### 3. Talos Linux Worker Node

Boot from Talos ISO for installation, then from disk:

```hcl
resource "truenas_vm" "talos_worker" {
  name   = "talosworker01"
  memory = 8192
  vcpus  = 4

  nic_devices = [{
    type       = "VIRTIO"
    nic_attach = "eno1"
  }]

  # Boot from Talos ISO FIRST for installation
  cdrom_devices = [{
    path  = "/mnt/pool/isos/talos-v1.10.6-metal-amd64.iso"
    order = 1  # Boots FIRST
  }]

  # Boot from disk SECOND (after Talos is installed)
  disk_devices = [{
    path   = "/dev/zvol/pool/vms/talos-worker-01-disk0"
    type   = "VIRTIO"
    iotype = "THREADS"
    order  = 2  # Boots SECOND
  }]

  bootloader      = "UEFI"
  start_on_create = true
}
```

**Talos installation process:**
1. VM boots from Talos ISO (order 1)
2. Talos installs to disk
3. After installation, Talos on disk takes over
4. ISO remains attached but boots second (order 2)

### 4. Multi-Disk System

Specify boot priority for multiple disks:

```hcl
resource "truenas_vm" "multi_disk" {
  name   = "multidisk"
  memory = 16384
  vcpus  = 8

  # Primary OS disk boots FIRST
  disk_devices = [
    {
      path  = "/dev/zvol/pool/vms/os-disk"
      type  = "VIRTIO"
      order = 1  # Primary boot disk
    },
    {
      path  = "/dev/zvol/pool/vms/data-disk"
      type  = "VIRTIO"
      order = 3  # Data disk (not bootable)
    }
  ]

  # Rescue ISO as fallback
  cdrom_devices = [{
    path  = "/mnt/pool/isos/rescue.iso"
    order = 2  # Boot if primary disk fails
  }]

  nic_devices = [{
    type       = "VIRTIO"
    nic_attach = "eno1"
  }]

  bootloader = "UEFI"
}
```

---

## 🔄 Default Behavior (Backward Compatible)

If you **don't specify** the `order` attribute, devices are assigned auto-incrementing orders starting at 1000:

```hcl
resource "truenas_vm" "default" {
  name = "myvm"

  nic_devices = [{
    nic_attach = "eno1"
    # order = 1000 (auto-assigned)
  }]

  disk_devices = [{
    path = "/dev/zvol/pool/vms/disk0"
    # order = 1001 (auto-assigned)
  }]

  cdrom_devices = [{
    path = "/mnt/pool/isos/ubuntu.iso"
    # order = 1002 (auto-assigned)
  }]
}
```

**Result:** Devices are ordered in the sequence they appear in your config.

---

## 📊 Device Attributes Reference

### NIC Device

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `type` | string | No | VIRTIO | NIC type (VIRTIO, E1000, etc.) |
| `mac` | string | No | auto-generated | MAC address |
| `nic_attach` | string | Yes | - | Physical interface to attach to |
| `trust_guest_rx_filters` | bool | No | false | Trust guest RX filters |
| **`order`** | **number** | **No** | **auto** | **Boot order (lower = boots first)** |

### Disk Device

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `path` | string | Yes | - | Path to zvol or file |
| `type` | string | No | VIRTIO | Disk type (VIRTIO, AHCI, SCSI) |
| `iotype` | string | No | - | IO type (THREADS, NATIVE) |
| `physical_sectorsize` | number | No | - | Physical sector size in bytes |
| `logical_sectorsize` | number | No | - | Logical sector size in bytes |
| **`order`** | **number** | **No** | **auto** | **Boot order (lower = boots first)** |

### CDROM Device

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `path` | string | Yes | - | Path to ISO file |
| **`order`** | **number** | **No** | **auto** | **Boot order (lower = boots first)** |

---

## 🔧 Migration Guide

### Existing VMs (No Changes Needed)

If you have existing VMs created with v0.2.13 or earlier:
- ✅ **No action required** - they will continue to work
- ✅ **Order is auto-assigned** - devices keep their current order
- ✅ **Backward compatible** - no breaking changes

### Adding Boot Order to Existing VMs

To add explicit boot order to existing VMs:

1. **Add `order` attribute to your Terraform config:**
   ```hcl
   disk_devices = [{
     path  = "/dev/zvol/pool/vms/disk0"
     type  = "VIRTIO"
     order = 1  # Add this line
   }]

   cdrom_devices = [{
     path  = "/mnt/pool/isos/ubuntu.iso"
     order = 2  # Add this line
   }]
   ```

2. **Apply the changes:**
   ```bash
   terraform apply
   ```

3. **Restart the VM** (if needed):
   - Changes may require VM restart to take effect
   - Check TrueNAS UI to verify boot order

---

## 🐛 Bug Fixes

### Fixed: Boot Order Not Respected

**Issue:** Devices were always created with hardcoded order based on type (NICs → Disks → CDROMs), ignoring any user-specified order.

**Root Cause:** The `createDevices()` function used a hardcoded `deviceOrder` variable that incremented for each device, regardless of user configuration.

**Fix:** 
- Added `order` field to all device models (NICDeviceModel, DiskDeviceModel, CDROMDeviceModel)
- Updated schema to include `order` attribute for all device types
- Modified `createDevices()` to use user-specified order if provided
- Updated `readVM()` to read order from API and populate in state

**Impact:** Users can now control boot order, enabling OS installation from ISO and proper Talos Linux deployments.

---

## 📚 Documentation

### New Examples

Added `examples/vm-boot-order/` directory with:
- **main.tf** - 5 complete examples showing different boot order scenarios
- **README.md** - Comprehensive guide to boot order configuration
- **variables.tf** - Variable definitions
- **terraform.tfvars.example** - Template for credentials

### Examples Included

1. **install_from_iso** - Boot from CDROM first for OS installation
2. **boot_from_disk** - Boot from disk first (normal operation)
3. **talos_worker** - Talos Linux worker with proper boot order
4. **multi_disk** - Multiple disks with specific boot order
5. **default_order** - Default ordering (no explicit order values)

---

## ⚠️ Known Limitations

1. **VM Names:** Must be alphanumeric only (no hyphens, underscores, or special characters)
2. **Device Updates:** Changing boot order may require VM restart
3. **Unique Orders:** Each device should have a unique order value (undefined behavior if duplicates)

---

## 🧪 Testing

This release has been tested with:
- ✅ Boot from CDROM first (OS installation)
- ✅ Boot from disk first (normal operation)
- ✅ Talos Linux worker nodes
- ✅ Multi-disk systems
- ✅ Default order (no explicit order specified)
- ✅ Backward compatibility with v0.2.13 configs

---

## 🚀 What's Next

Planned for v0.3.0:
- Device update support (add/remove devices from existing VMs)
- Display device configuration
- USB device passthrough
- PCI device passthrough
- Network bridge management

---

## 🙏 Feedback

If you encounter any issues or have suggestions, please open an issue on GitHub:
https://github.com/baladithyab/terraform-provider-truenas/issues

---

## 👥 Contributors

- @baladithyab - Boot order implementation and bug fix

---

**Full Changelog:** https://github.com/baladithyab/terraform-provider-truenas/compare/v0.2.13...v0.2.14

