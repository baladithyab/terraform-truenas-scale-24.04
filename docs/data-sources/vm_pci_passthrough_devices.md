---
page_title: "truenas_vm_pci_passthrough_devices Data Source - terraform-provider-truenas"
subcategory: "Virtual Machines"
description: |-
  Fetches available PCI passthrough devices for VM attachment.
---

# truenas_vm_pci_passthrough_devices (Data Source)

Fetches available PCI passthrough devices for VM attachment. This data source provides comprehensive information about PCI devices that can be passed through to virtual machines, including availability, IOMMU groups, and device capabilities.

## Example Usage

### Basic PCI Device Discovery

```terraform
data "truenas_vm_pci_passthrough_devices" "all" {
  available_only = false
  
  output "pci_devices" {
    value = data.truenas_vm_pci_passthrough_devices.all.devices
  }
}

output "device_count" {
  value = length(keys(data.truenas_vm_pci_passthrough_devices.all.devices))
}
```

### Available Devices Only

```terraform
data "truenas_vm_pci_passthrough_devices" "available" {
  available_only = true
  
  output "available_devices" {
    value = data.truenas_vm_pci_passthrough_devices.available.devices
  }
}

output "available_count" {
  value = length(keys(data.truenas_vm_pci_passthrough_devices.available.devices))
}
```

### GPU Device Selection

```terraform
data "truenas_vm_pci_passthrough_devices" "pci_devices" {
  available_only = true
}

locals {
  # Find GPU devices
  gpu_devices = {
    for id, device in data.truenas_vm_pci_passthrough_devices.pci_devices.devices
    : device
    if contains(lower(device.description), "nvidia") || 
       contains(lower(device.description), "amd") || 
       contains(lower(device.description), "radeon") ||
       contains(lower(device.description), "geforce")
  }
  
  # Get first available GPU
  first_gpu = length(local.gpu_devices) > 0 ? values(local.gpu_devices)[0] : null
}

output "gpu_selection" {
  value = {
    available_gpus = local.gpu_devices
    selected_gpu = local.first_gpu
  }
}
```

### Network Device Selection

```terraform
data "truenas_vm_pci_passthrough_devices" "pci_devices" {
  available_only = true
}

locals {
  # Find network devices
  network_devices = {
    for id, device in data.truenas_vm_pci_passthrough_devices.pci_devices.devices
    : device
    if contains(lower(device.controller_type), "ethernet") ||
       contains(lower(device.controller_type), "network")
  }
  
  # Find USB controllers
  usb_devices = {
    for id, device in data.truenas_vm_pci_passthrough_devices.pci_devices.devices
    : device
    if contains(lower(device.controller_type), "usb")
  }
}

output "device_selection" {
  value = {
    network_devices = local.network_devices
    usb_devices = local.usb_devices
  }
}
```

### IOMMU Group Analysis

```terraform
data "truenas_vm_pci_passthrough_devices" "pci_devices" {
  available_only = false
}

locals {
  # Group devices by IOMMU group
  iommu_groups = {
    for id, device in data.truenas_vm_pci_passthrough_devices.pci_devices.devices
    : device.iommu_group => {
      devices = []
    }
  }
  
  # Populate groups
  grouped_devices = merge([
    for id, device in data.truenas_vm_pci_passthrough_devices.pci_devices.devices
    : {
      "${device.iommu_group}" = {
        devices = concat(local.iommu_groups["${device.iommu_group}"].devices, [device])
      }
    }
  ])
}

output "iommu_analysis" {
  value = {
    groups = local.grouped_devices
    group_count = length(keys(local.grouped_devices))
  }
}
```

### Critical Device Identification

```terraform
data "truenas_vm_pci_passthrough_devices" "pci_devices" {
  available_only = true
}

locals {
  # Find critical devices (should not passthrough)
  critical_devices = {
    for id, device in data.truenas_vm_pci_passthrough_devices.pci_devices.devices
    : device
    if device.critical
  }
  
  # Find safe devices for passthrough
  safe_devices = {
    for id, device in data.truenas_vm_pci_passthrough_devices.pci_devices.devices
    : device
    if device.available && !device.critical
  }
}

output "safety_analysis" {
  value = {
    critical_devices = local.critical_devices
    safe_devices = local.safe_devices
    safe_count = length(keys(local.safe_devices))
  }
}
```

### Complete PCI Passthrough Setup

```terraform
data "truenas_vm_iommu_enabled" "iommu" {
}

data "truenas_vm_pci_passthrough_devices" "pci_devices" {
  available_only = true
  
  depends_on = [data.truenas_vm_iommu_enabled.iommu]
}

locals {
  iommu_enabled = data.truenas_vm_iommu_enabled.iommu.enabled
  available_devices = data.truenas_vm_pci_passthrough_devices.pci_devices.devices
  
  # Select GPU for passthrough
  selected_gpu = {
    for id, device in local.available_devices
    : device
    if contains(lower(device.description), "nvidia") && device.available
  }
  
  # Select network card for passthrough
  selected_nic = {
    for id, device in local.available_devices
    : device
    if contains(lower(device.controller_type), "ethernet") && device.available
  }
}

# Create VM only if IOMMU is enabled
resource "truenas_vm" "passthrough_vm" {
  count = local.iommu_enabled ? 1 : 0
  
  name   = "passthrough-vm"
  vcpus  = 8
  memory = 16384
  
  autostart = true
  
  depends_on = [data.truenas_vm_iommu_enabled.iommu]
}

# Add GPU passthrough
resource "truenas_vm_device" "gpu" {
  count     = local.iommu_enabled && length(local.selected_gpu) > 0 ? 1 : 0
  vm_id     = truenas_vm.passthrough_vm[0].id
  device_type = "PCI"
  order     = 100
  
  pci_config = [{
    pptdev = keys(local.selected_gpu)[0]
  }]
  
  depends_on = [
    data.truenas_vm_iommu_enabled.iommu,
    data.truenas_vm_pci_passthrough_devices.pci_devices
  ]
}

# Add network passthrough
resource "truenas_vm_device" "nic" {
  count     = local.iommu_enabled && length(local.selected_nic) > 0 ? 1 : 0
  vm_id     = truenas_vm.passthrough_vm[0].id
  device_type = "PCI"
  order     = 200
  
  pci_config = [{
    pptdev = keys(local.selected_nic)[0]
  }]
  
  depends_on = [
    data.truenas_vm_iommu_enabled.iommu,
    data.truenas_vm_pci_passthrough_devices.pci_devices
  ]
}
```

## Schema

### Optional

- `available_only` (Boolean) If true, only return devices where available=true. Default: true.

### Read-Only

- `id` (String) Data source identifier (always 'vm_pci_passthrough_devices').
- `devices` (Map of Object) Map of PCI device IDs to device information objects. See [Device Attributes](#device-attributes) below.

### Device Attributes

Each device contains:

- `pci_address` (String) PCI address in format domain:bus:slot.function.
- `description` (String) Device description.
- `controller_type` (String) Controller type (e.g., Ethernet, USB, SATA).
- `available` (Boolean) Whether device is available for passthrough.
- `critical` (Boolean) Whether device is critical to system operation.
- `iommu_group` (Number) IOMMU group number.
- `vendor` (String) Device vendor name.
- `product` (String) Device product name.

## Notes

### PCI Device Structure

The data source returns a map of device objects:

```json
{
  "pci_0000_01_00_0": {
    "pci_address": "0000:01:00.0",
    "description": "NVIDIA Corporation RTX 3080",
    "controller_type": "VGA",
    "available": true,
    "critical": false,
    "iommu_group": 1,
    "vendor": "NVIDIA Corporation",
    "product": "RTX 3080"
  }
}
```

### PCI Address Format

PCI addresses follow the format: `domain:bus:slot.function`

- **domain**: Typically 0000
- **bus**: Bus number (00-ff)
- **slot**: Slot number (00-1f)
- **function**: Function number (0-7)

Example: `0000:01:00.0`
- Domain: 0000
- Bus: 01
- Slot: 00
- Function: 0

### Device Availability

#### Available Devices
```terraform
available = true  # Device can be safely passed through
```

#### Unavailable Devices
```terraform
available = false  # Device is in use or cannot be passed through
```

#### Critical Devices
```terraform
critical = true   # Device is essential for system operation
critical = false  # Device can be safely passed through
```

### IOMMU Groups

Devices in the same IOMMU group share memory management:

```terraform
iommu_group = 1  # All devices in group 1 share IOMMU
```

#### Group Considerations
- Devices in same group must be passed through together
- Mixing available and unavailable devices in same group
- Critical devices in same group may prevent passthrough

### Controller Types

Common controller types include:
- **VGA**: Graphics controllers
- **Ethernet**: Network controllers
- **USB**: USB controllers
- **SATA**: Storage controllers
- **PCI**: PCI bridges
- **Audio**: Audio devices

### Use Cases

#### GPU Passthrough
```terraform
data "truenas_vm_pci_passthrough_devices" "pci_devices" {
  available_only = true
}

locals {
  gpu_devices = {
    for id, device in data.truenas_vm_pci_passthrough_devices.pci_devices.devices
    : device
    if contains(lower(device.controller_type), "vga") && device.available
  }
}

resource "truenas_vm_device" "gpu" {
  for_each = local.gpu_devices
  
  vm_id     = truenas_vm.gaming_vm.id
  device_type = "PCI"
  order     = 100
  
  pci_config = [{
    pptdev = each.key
  }]
}
```

#### Network Card Passthrough
```terraform
data "truenas_vm_pci_passthrough_devices" "pci_devices" {
  available_only = true
}

locals {
  network_devices = {
    for id, device in data.truenas_vm_pci_passthrough_devices.pci_devices.devices
    : device
    if contains(lower(device.controller_type), "ethernet") && device.available
  }
}

resource "truenas_vm_device" "network" {
  for_each = local.network_devices
  
  vm_id     = truenas_vm.network_vm.id
  device_type = "PCI"
  order     = 200
  
  pci_config = [{
    pptdev = each.key
  }]
}
```

#### USB Controller Passthrough
```terraform
data "truenas_vm_pci_passthrough_devices" "pci_devices" {
  available_only = true
}

locals {
  usb_controllers = {
    for id, device in data.truenas_vm_pci_passthrough_devices.pci_devices.devices
    : device
    if contains(lower(device.controller_type), "usb") && device.available
  }
}

resource "truenas_vm_device" "usb" {
  for_each = local.usb_controllers
  
  vm_id     = truenas_vm.usb_vm.id
  device_type = "PCI"
  order     = 300
  
  pci_config = [{
    pptdev = each.key
  }]
}
```

#### Device Selection Logic
```terraform
data "truenas_vm_pci_passthrough_devices" "pci_devices" {
  available_only = true
}

locals {
  # Prefer GPU over network over USB
  device_priority = {
    for id, device in data.truenas_vm_pci_passthrough_devices.pci_devices.devices
    : id => {
      device = device
      priority = (
        contains(lower(device.controller_type), "vga") ? 1 :
        contains(lower(device.controller_type), "ethernet") ? 2 :
        contains(lower(device.controller_type), "usb") ? 3 : 4
      )
    }
  }
  
  # Sort by priority and select first available
  sorted_devices = sort(values(local.device_priority), lambda d: d.priority)
  selected_device = local.sorted_devices[0].device
}

output "best_device" {
  value = local.selected_device
}
```

## Best Practices

### Device Selection

1. **Availability Check**: Always filter for available devices
2. **Critical Device Avoidance**: Don't pass through critical system devices
3. **IOMMU Group Awareness**: Consider IOMMU group constraints
4. **Performance Needs**: Select appropriate devices for workload

### Safety Considerations

1. **Critical Devices**: Avoid passing through essential system components
2. **IOMMU Groups**: Understand group dependencies
3. **System Stability**: Test passthrough configurations thoroughly
4. **Backup Plans**: Have fallback configurations ready

### Performance Optimization

1. **Device Isolation**: Ensure proper device isolation
2. **Driver Configuration**: Use appropriate guest drivers
3. **Resource Allocation**: Balance device performance with host needs
4. **Monitoring**: Monitor passthrough device performance

### Security

1. **Access Control**: Limit passthrough to authorized VMs
2. **Device Isolation**: Ensure proper device isolation
3. **Monitoring**: Monitor device usage and access
4. **Compliance**: Follow licensing and security requirements

## Troubleshooting

### No Available Devices

1. Verify IOMMU is enabled
2. Check device availability in TrueNAS
3. Review system logs for errors
4. Test with available_only = false

### Device Not Working

1. Verify device is available
2. Check IOMMU group configuration
3. Review VM device configuration
4. Test with different devices

### IOMMU Group Issues

1. Check all devices in group
2. Verify group dependencies
3. Test with different device combinations
4. Review system documentation

### Performance Problems

1. Monitor device utilization
2. Check driver compatibility
3. Review VM configuration
4. Test with different settings

### Data Source Issues

1. Verify IOMMU is enabled
2. Check provider configuration
3. Test with simple configuration
4. Review TrueNAS API access

## See Also

- [truenas_vm_iommu_enabled](vm_iommu_enabled) - IOMMU capability checking
- [truenas_gpu_pci_choices](gpu_pci_choices) - GPU device discovery
- [truenas_vm_device](../resources/vm_device) - VM device management
- [TrueNAS PCI Passthrough Guide](https://www.truenas.com/docs/scale/virtualmachines/pci-passthrough/) - Official PCI passthrough documentation
- [VFIO Configuration](https://wiki.archlinux.org/title/PCI_passthrough_via_OVMF) - Linux VFIO setup guide