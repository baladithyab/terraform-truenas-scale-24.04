# VM Boot Order Configuration

This example demonstrates how to control the boot order of devices (NICs, disks, and CDROMs) in TrueNAS VMs using the `order` attribute.

## Overview

The `order` attribute determines the boot priority of devices:
- **Lower order values boot first**
- **Higher order values boot later**
- If `order` is not specified, devices are assigned auto-incrementing orders starting at 1000

## Common Use Cases

### 1. Install OS from ISO (Boot from CDROM first)

```hcl
resource "truenas_vm" "install_from_iso" {
  name   = "myvm"
  memory = 4096
  vcpus  = 2

  # CDROM boots FIRST (order 1)
  cdrom_devices = [{
    path  = "/mnt/pool/isos/ubuntu-22.04.iso"
    order = 1
  }]

  # Disk boots SECOND (order 2)
  disk_devices = [{
    path  = "/dev/zvol/pool/vms/disk0"
    order = 2
  }]
}
```

**Use this when:**
- Installing a new operating system
- The VM needs to boot from ISO first
- After OS installation, you can change the order or remove the CDROM

### 2. Boot from Disk (Normal Operation)

```hcl
resource "truenas_vm" "boot_from_disk" {
  name   = "myvm"
  memory = 4096
  vcpus  = 2

  # Disk boots FIRST (order 1)
  disk_devices = [{
    path  = "/dev/zvol/pool/vms/disk0"
    order = 1
  }]

  # CDROM boots SECOND (fallback)
  cdrom_devices = [{
    path  = "/mnt/pool/isos/rescue.iso"
    order = 2
  }]
}
```

**Use this when:**
- OS is already installed on the disk
- Normal VM operation
- CDROM is only for rescue/recovery

### 3. Talos Linux Installation

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
    order = 1
  }]

  # Boot from disk SECOND (after Talos is installed)
  disk_devices = [{
    path  = "/dev/zvol/pool/vms/talos-disk0"
    type  = "VIRTIO"
    order = 2
  }]

  bootloader = "UEFI"
}
```

**Talos Installation Process:**
1. VM boots from ISO (order 1)
2. Talos installs to disk
3. After installation, Talos on disk takes over
4. ISO remains attached but boots second

## Boot Order Behavior

### How TrueNAS/libvirt Uses Boot Order

The `order` field in the TrueNAS API maps directly to the libvirt `<boot order='X'/>` attribute in the VM's XML configuration. This controls which device the BIOS/UEFI firmware tries to boot from first.

**Example libvirt XML:**
```xml
<disk type='block' device='disk'>
  <boot order='1'/>  <!-- Boots FIRST -->
  ...
</disk>
<disk type='file' device='cdrom'>
  <boot order='2'/>  <!-- Boots SECOND -->
  ...
</disk>
```

### Default Behavior (No Order Specified)

If you don't specify `order`, devices are created with auto-incrementing orders starting at 1000:

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

**Result:** NICs → Disks → CDROMs (in the order they appear in the config)

## Prerequisites

1. **Create zvols for disks:**
   ```bash
   zfs create -V 32G pool/vms/myvm-disk0
   ```

2. **Upload ISO files:**
   - ISOs should be accessible at `/mnt/pool/path/to/file.iso`
   - Upload via TrueNAS UI: Storage → Pools → Upload

3. **Verify network interfaces:**
   - Check available interfaces in TrueNAS UI: Network → Interfaces
   - Common interfaces: `eno1`, `eno2`, `br0`

## Usage

1. **Copy the example configuration:**
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. **Edit terraform.tfvars:**
   ```hcl
   truenas_base_url = "http://10.0.0.83:81"
   truenas_api_key  = "your-api-key-here"
   ```

3. **Initialize Terraform:**
   ```bash
   terraform init
   ```

4. **Review the plan:**
   ```bash
   terraform plan
   ```

5. **Apply the configuration:**
   ```bash
   terraform apply
   ```

## Changing Boot Order After Creation

To change the boot order of an existing VM:

1. **Update the order values in your Terraform config:**
   ```hcl
   disk_devices = [{
     path  = "/dev/zvol/pool/vms/disk0"
     order = 1  # Changed from 2 to 1
   }]

   cdrom_devices = [{
     path  = "/mnt/pool/isos/ubuntu.iso"
     order = 2  # Changed from 1 to 2
   }]
   ```

2. **Apply the changes:**
   ```bash
   terraform apply
   ```

**Note:** Changing device order may require recreating the devices. The VM may need to be stopped and restarted for changes to take effect.

## Troubleshooting

### VM Boots from Wrong Device

**Problem:** VM boots from disk instead of CDROM (or vice versa)

**Solution:**
1. Check the `order` values in your Terraform config
2. Verify the order in TrueNAS UI: Virtualization → VMs → Edit → Devices
3. Ensure lower order value is on the device you want to boot first
4. Restart the VM after making changes

### Order Not Taking Effect

**Problem:** Changed order in Terraform but VM still boots from old device

**Solution:**
1. Stop the VM: `terraform apply` with `start_on_create = false`
2. Verify devices were updated in TrueNAS UI
3. Start the VM manually or set `start_on_create = true`
4. Check VM console to see boot sequence

### Multiple Devices with Same Order

**Problem:** What happens if multiple devices have the same order?

**Answer:** TrueNAS/libvirt behavior is undefined. Always use unique order values for each device.

## Best Practices

1. **Use explicit order values** for critical boot sequences (OS installation, Talos, etc.)
2. **Leave gaps between order values** (1, 10, 20) to allow inserting devices later
3. **Document your boot order** in the VM description or comments
4. **Test boot order** by checking VM console during startup
5. **Remove or change CDROM order** after OS installation is complete

## Examples in This Directory

- `install_from_iso` - Boot from CDROM first for OS installation
- `boot_from_disk` - Boot from disk first (normal operation)
- `talos_worker` - Talos Linux worker with proper boot order
- `multi_disk` - Multiple disks with specific boot order
- `default_order` - Default ordering (no explicit order values)

## Additional Resources

- [TrueNAS VM Documentation](https://www.truenas.com/docs/scale/scaletutorials/virtualization/)
- [libvirt Domain XML Format](https://libvirt.org/formatdomain.html#boot-order)
- [Talos Linux Installation](https://www.talos.dev/latest/introduction/getting-started/)

