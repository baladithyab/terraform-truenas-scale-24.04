---
page_title: "truenas_vm_iommu_enabled Data Source - terraform-provider-truenas"
subcategory: "Virtual Machines"
description: |-
  Checks if IOMMU (Intel VT-d / AMD-Vi) is enabled on TrueNAS system.
---

# truenas_vm_iommu_enabled (Data Source)

Checks if IOMMU (Intel VT-d / AMD-Vi) is enabled on TrueNAS system. IOMMU must be enabled for PCI passthrough to work with virtual machines. This data source helps verify system capability before attempting PCI device passthrough.

## Example Usage

### Basic IOMMU Check

```terraform
data "truenas_vm_iommu_enabled" "iommu_check" {
  output "iommu_status" {
    value = data.truenas_vm_iommu_enabled.iommu_check.enabled
  }
}

output "iommu_enabled" {
  value = data.truenas_vm_iommu_enabled.iommu_check.enabled ? "Yes" : "No"
}
```

### Conditional GPU Passthrough

```terraform
data "truenas_vm_iommu_enabled" "iommu" {
}

# Only create GPU passthrough if IOMMU is enabled
resource "truenas_vm_device" "gpu" {
  count     = data.truenas_vm_iommu_enabled.iommu.enabled ? 1 : 0
  vm_id     = truenas_vm.gaming_vm.id
  device_type = "PCI"
  order     = 100
  
  pci_config = [{
    pptdev = "pci_0000_01_00_0"
  }]
  
  depends_on = [data.truenas_vm_iommu_enabled.iommu]
}

output "gpu_passthrough_status" {
  value = data.truenas_vm_iommu_enabled.iommu.enabled ? 
    "GPU passthrough available" : 
    "IOMMU not enabled - GPU passthrough unavailable"
}
```

### IOMMU Validation

```terraform
data "truenas_vm_iommu_enabled" "iommu" {
}

locals {
  iommu_available = data.truenas_vm_iommu_enabled.iommu.enabled
}

# Create validation resource
resource "null_resource" "iommu_validation" {
  count = local.iommu_available ? 0 : 1
  
  provisioner "local-exec" {
    command = <<-EOT
      echo "ERROR: IOMMU is not enabled on this system"
      echo "PCI passthrough will not work without IOMMU"
      echo "Please enable IOMMU in BIOS and reboot"
      exit 1
    EOT
  }
}

output "validation_result" {
  value = local.iommu_available ? 
    "IOMMU is enabled - PCI passthrough available" : 
    "IOMMU validation failed - see error above"
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
}

# Create VM only if IOMMU is available
resource "truenas_vm" "passthrough_vm" {
  count = local.iommu_enabled ? 1 : 0
  
  name   = "passthrough-vm"
  vcpus  = 4
  memory = 8192
  
  autostart = true
  
  depends_on = [data.truenas_vm_iommu_enabled.iommu]
}

# Add GPU passthrough if available
resource "truenas_vm_device" "gpu" {
  count     = local.iommu_enabled && length(local.available_devices) > 0 ? 1 : 0
  vm_id     = truenas_vm.passthrough_vm[0].id
  device_type = "PCI"
  order     = 100
  
  pci_config = [{
    pptdev = keys(local.available_devices)[0]
  }]
  
  depends_on = [
    data.truenas_vm_iommu_enabled.iommu,
    data.truenas_vm_pci_passthrough_devices.pci_devices
  ]
}

output "setup_status" {
  value = {
    iommu_enabled = local.iommu_enabled
    available_devices = length(local.available_devices)
    vm_created = length(truenas_vm.passthrough_vm) > 0
    gpu_added = length(truenas_vm_device.gpu) > 0
  }
}
```

### IOMMU Status Reporting

```terraform
data "truenas_vm_iommu_enabled" "iommu" {
}

data "truenas_vm_pci_passthrough_devices" "devices" {
  depends_on = [data.truenas_vm_iommu_enabled.iommu]
}

locals {
  iommu_status = {
    enabled = data.truenas_vm_iommu_enabled.iommu.enabled
    message = data.truenas_vm_iommu_enabled.iommu.enabled ? 
      "IOMMU is enabled - PCI passthrough available" :
      "IOMMU is disabled - PCI passthrough unavailable"
  }
  
  passthrough_capability = {
    iommu_enabled = local.iommu_status.enabled
    available_devices = length(data.truenas_vm_pci_passthrough_devices.devices.devices)
    ready_for_passthrough = local.iommu_status.enabled && 
      length(data.truenas_vm_pci_passthrough_devices.devices.devices) > 0
  }
}

output "iommu_report" {
  value = {
    status = local.iommu_status
    capability = local.passthrough_capability
    recommendations = local.iommu_status.enabled ? [
      "PCI passthrough is ready to use",
      "Available devices can be passed through to VMs"
    ] : [
      "Enable IOMMU in BIOS settings",
      "Look for 'Intel VT-d' or 'AMD-Vi' options",
      "Enable virtualization extensions",
      "Reboot system after changes"
    ]
  }
}
```

### Multi-VM IOMMU Check

```terraform
data "truenas_vm_iommu_enabled" "iommu" {
}

locals {
  iommu_available = data.truenas_vm_iommu_enabled.iommu.enabled
}

# Create multiple VMs with different passthrough needs
resource "truenas_vm" "gaming_vm" {
  count = local.iommu_available ? 1 : 0
  
  name   = "gaming-vm"
  vcpus  = 8
  memory = 16384
  
  autostart = true
}

resource "truenas_vm" "workstation_vm" {
  count = local.iommu_available ? 1 : 0
  
  name   = "workstation-vm"
  vcpus  = 6
  memory = 12288
  
  autostart = false
}

output "vm_deployment" {
  value = {
    iommu_available = local.iommu_available
    gaming_vm_created = length(truenas_vm.gaming_vm) > 0
    workstation_vm_created = length(truenas_vm.workstation_vm) > 0
    total_vms = length(truenas_vm.gaming_vm) + length(truenas_vm.workstation_vm)
  }
}
```

## Schema

### Read-Only

- `id` (String) Data source identifier (always 'vm_iommu_enabled').
- `enabled` (Boolean) Whether IOMMU is enabled on the system.

## Notes

### IOMMU Overview

IOMMU (Input/Output Memory Management Unit) is a hardware feature that enables:
- **PCI Device Passthrough**: Direct device access for VMs
- **Memory Protection**: Isolates device memory access
- **Performance**: Near-native performance for passthrough devices
- **Security**: Prevents unauthorized memory access

#### Intel VT-d
- Intel's implementation of IOMMU
- Required for Intel-based systems
- Must be enabled in BIOS

#### AMD-Vi
- AMD's implementation of IOMMU
- Required for AMD-based systems
- Must be enabled in BIOS

### IOMMU Requirements

#### Hardware Requirements
- CPU with IOMMU support (Intel VT-d or AMD-Vi)
- Motherboard with IOMMU support
- Compatible chipset

#### BIOS/UEFI Configuration
```
Intel Systems:
- Intel VT-d: Enabled
- Virtualization Technology: Enabled
- SR-IOV Support: Optional but recommended

AMD Systems:
- AMD-Vi: Enabled
- Virtualization Technology: Enabled
- IOMMU: Enabled
```

#### Software Requirements
- TrueNAS Scale with virtualization support
- Appropriate kernel modules loaded
- Proper system configuration

### IOMMU Detection

The data source checks for:
- Kernel IOMMU support
- Hardware IOMMU capability
- System configuration status

#### Return Values
```terraform
enabled = true   # IOMMU is available and functional
enabled = false  # IOMMU is not available or disabled
```

### Use Cases

#### Pre-deployment Validation
```terraform
data "truenas_vm_iommu_enabled" "iommu" {
}

locals {
  can_use_passthrough = data.truenas_vm_iommu_enabled.iommu.enabled
}

# Conditional resource creation
resource "truenas_vm" "gpu_vm" {
  count = local.can_use_passthrough ? 1 : 0
  
  name   = "gpu-vm"
  vcpus  = 6
  memory = 8192
  
  depends_on = [data.truenas_vm_iommu_enabled.iommu]
}
```

#### System Capability Reporting
```terraform
data "truenas_vm_iommu_enabled" "iommu" {
}

output "system_capabilities" {
  value = {
    iommu_enabled = data.truenas_vm_iommu_enabled.iommu.enabled
    pci_passthrough_available = data.truenas_vm_iommu_enabled.iommu.enabled
    gpu_passthrough_possible = data.truenas_vm_iommu_enabled.iommu.enabled
  }
}
```

#### Troubleshooting Helper
```terraform
data "truenas_vm_iommu_enabled" "iommu" {
}

locals {
  iommu_status = data.truenas_vm_iommu_enabled.iommu.enabled
}

output "troubleshooting" {
  value = {
    iommu_enabled = local.iommu_status
    issue_detected = !local.iommu_status
    recommended_actions = local.iommu_status ? [
      "IOMMU is properly configured"
    ] : [
      "Check BIOS settings for IOMMU/VT-d/AMD-Vi",
      "Verify CPU supports IOMMU",
      "Ensure TrueNAS is properly configured",
      "Check kernel logs for IOMMU errors"
    ]
  }
}
```

## Best Practices

### System Configuration

1. **BIOS Settings**: Enable all virtualization features
2. **Hardware Compatibility**: Verify hardware supports IOMMU
3. **System Updates**: Keep TrueNAS updated
4. **Testing**: Test IOMMU functionality before deployment

### Deployment Planning

1. **Pre-validation**: Always check IOMMU status before deployment
2. **Fallback Planning**: Have alternative configurations ready
3. **Documentation**: Document IOMMU configuration
4. **Monitoring**: Monitor IOMMU functionality

### Performance Optimization

1. **Hardware Selection**: Choose appropriate hardware
2. **Device Selection**: Select suitable devices for passthrough
3. **Configuration**: Optimize VM and device settings
4. **Testing**: Performance test passthrough configurations

### Security

1. **Device Isolation**: Ensure proper device isolation
2. **Access Control**: Limit passthrough to authorized VMs
3. **Monitoring**: Monitor passthrough device usage
4. **Compliance**: Follow security best practices

## Troubleshooting

### IOMMU Not Enabled

#### BIOS Configuration
1. Enter BIOS/UEFI setup
2. Find virtualization settings
3. Enable Intel VT-d or AMD-Vi
4. Save and reboot

#### Hardware Issues
1. Verify CPU supports IOMMU
2. Check motherboard compatibility
3. Update BIOS/UEFI firmware
4. Test with different hardware

#### Software Issues
1. Update TrueNAS to latest version
2. Check kernel configuration
3. Review system logs
4. Reinstall if necessary

### Data Source Issues

1. Verify provider configuration
2. Check TrueNAS API access
3. Test with simple configuration
4. Review system logs

### Passthrough Problems

1. Verify IOMMU is enabled
2. Check device compatibility
3. Review VM configuration
4. Test with different devices

## See Also

- [truenas_vm_pci_passthrough_devices](vm_pci_passthrough_devices) - PCI device discovery
- [truenas_vm_device](../resources/vm_device) - VM device management
- [truenas_gpu_pci_choices](gpu_pci_choices) - GPU device discovery
- [TrueNAS Virtualization Documentation](https://www.truenas.com/docs/scale/virtualmachines/) - Official VM documentation
- [PCI Passthrough Guide](https://www.truenas.com/docs/scale/virtualmachines/pci-passthrough/) - PCI passthrough configuration
- [IOMMU Configuration](https://wiki.archlinux.org/title/PCI_passthrough_via_OVMF#IOMMU_setup) - Linux IOMMU setup guide