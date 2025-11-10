---
page_title: "truenas_gpu_pci_choices Data Source - terraform-provider-truenas"
subcategory: "Virtual Machines"
description: |-
  Fetches available GPU PCI device choices from TrueNAS system.
---

# truenas_gpu_pci_choices (Data Source)

Fetches available GPU PCI device choices from TrueNAS system. This data source returns a map of GPU descriptions to PCI addresses, useful for identifying available GPUs for VM passthrough.

## Example Usage

### Basic GPU Discovery

```terraform
data "truenas_gpu_pci_choices" "available" {
  output "gpu_devices" {
    value = data.truenas_gpu_pci_choices.available.choices
  }
}

output "gpu_count" {
  value = length(keys(data.truenas_gpu_pci_choices.available.choices))
}
```

### Select Specific GPU for Passthrough

```terraform
data "truenas_gpu_pci_choices" "gpus" {
}

locals {
  # Find NVIDIA GPU
  nvidia_gpu = {
    for k, v in data.truenas_gpu_pci_choices.gpus.choices
    : v
    if contains(lower(k), "nvidia")
  }
  
  # Find AMD GPU
  amd_gpu = {
    for k, v in data.truenas_gpu_pci_choices.gpus.choices
    : v
    if contains(lower(k), "amd")
  }
}

output "selected_gpu" {
  value = local.nvidia_gpu != null ? local.nvidia_gpu : local.amd_gpu
}
```

### GPU Passthrough Configuration

```terraform
data "truenas_gpu_pci_choices" "gpus" {
}

# Create VM with GPU passthrough
resource "truenas_vm" "gaming_vm" {
  name   = "gaming-vm"
  vcpus  = 8
  memory = 16384
  
  autostart = true
}

resource "truenas_vm_device" "gpu" {
  vm_id       = truenas_vm.gaming_vm.id
  device_type = "PCI"
  order       = 100
  
  pci_config = [{
    pptdev = values(data.truenas_gpu_pci_choices.gpus.choices)[0]
  }]
  
  depends_on = [data.truenas_gpu_pci_choices.gpus]
}
```

### Multiple GPU Selection

```terraform
data "truenas_gpu_pci_choices" "gpus" {
}

locals {
  gpu_list = [
    for desc, pci in data.truenas_gpu_pci_choices.gpus.choices
    : {
      description = desc
      pci_address = pci
    }
  ]
  
  # Sort by description for consistent selection
  sorted_gpus = sort(local.gpu_list, lambda gpu: gpu.description)
}

output "available_gpus" {
  value = {
    for gpu in local.sorted_gpus
    : gpu.description => gpu.pci_address
  }
}

# Use first available GPU
resource "truenas_vm_device" "first_gpu" {
  vm_id       = truenas_vm.example.id
  device_type = "PCI"
  order       = 100
  
  pci_config = [{
    pptdev = local.sorted_gpus[0].pci_address
  }]
  
  depends_on = [data.truenas_gpu_pci_choices.gpus]
}
```

### GPU Validation

```terraform
data "truenas_gpu_pci_choices" "gpus" {
}

locals {
  has_nvidia = anytrue([
    for k, v in data.truenas_gpu_pci_choices.gpus.choices
    : contains(lower(k), "nvidia")
  ])
  
  has_amd = anytrue([
    for k, v in data.truenas_gpu_pci_choices.gpus.choices
    : contains(lower(k), "amd")
  ])
  
  has_intel = anytrue([
    for k, v in data.truenas_gpu_pci_choices.gpus.choices
    : contains(lower(k), "intel")
  ])
}

output "gpu_analysis" {
  value = {
    total_gpus = length(keys(data.truenas_gpu_pci_choices.gpus.choices))
    has_nvidia  = local.has_nvidia
    has_amd     = local.has_amd
    has_intel   = local.has_intel
    gpu_details = data.truenas_gpu_pci_choices.gpus.choices
  }
}
```

### Conditional GPU Passthrough

```terraform
data "truenas_gpu_pci_choices" "gpus" {
}

locals {
  # Check if any GPUs are available
  gpu_available = length(keys(data.truenas_gpu_pci_choices.gpus.choices)) > 0
  
  # Get first GPU if available
  first_gpu = local.gpu_available ? values(data.truenas_gpu_pci_choices.gpus.choices)[0] : null
}

# Only create GPU device if GPU is available
resource "truenas_vm_device" "conditional_gpu" {
  count     = local.gpu_available ? 1 : 0
  vm_id     = truenas_vm.example.id
  device_type = "PCI"
  order     = 100
  
  pci_config = [{
    pptdev = local.first_gpu
  }]
  
  depends_on = [data.truenas_gpu_pci_choices.gpus]
}
```

## Schema

### Read-Only

- `id` (String) Data source identifier (always 'gpu_pci_choices').
- `choices` (Map of String) Map of GPU descriptions to PCI addresses (e.g., 'NVIDIA Corporation Device 2584' -> '0000:3b:00.0').

## Notes

### GPU Discovery

The data source returns a map where:
- **Key**: GPU description (vendor and device name)
- **Value**: PCI address in format `domain:bus:slot.function`

#### Example Output
```json
{
  "NVIDIA Corporation RTX 3080": "0000:01:00.0",
  "Intel Corporation UHD Graphics": "0000:00:02.0",
  "AMD Corporation Radeon RX 6800": "0000:03:00.0"
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

### GPU Identification

#### NVIDIA GPUs
Common descriptions include:
- "NVIDIA Corporation GeForce RTX 3080"
- "NVIDIA Corporation Quadro RTX 6000"
- "NVIDIA Corporation Tesla T4"

#### AMD GPUs
Common descriptions include:
- "AMD Corporation Radeon RX 6800"
- "AMD Corporation Radeon VII"
- "AMD Corporation Instinct MI100"

#### Intel GPUs
Common descriptions include:
- "Intel Corporation UHD Graphics 630"
- "Intel Corporation Iris Xe Graphics"
- "Intel Corporation Arc A770"

### Use Cases

#### GPU Passthrough for VMs
```terraform
data "truenas_gpu_pci_choices" "gpus" {
}

resource "truenas_vm_device" "gpu_passthrough" {
  vm_id       = truenas_vm.gaming_vm.id
  device_type = "PCI"
  order       = 100
  
  pci_config = [{
    pptdev = values(data.truenas_gpu_pci_choices.gpus.choices)[0]
  }]
  
  depends_on = [data.truenas_gpu_pci_choices.gpus]
}
```

#### Multi-GPU Setups
```terraform
data "truenas_gpu_pci_choices" "gpus" {
}

locals {
  gpu_addresses = values(data.truenas_gpu_pci_choices.gpus.choices)
}

# Create VM with multiple GPUs
resource "truenas_vm_device" "gpu1" {
  count       = length(local.gpu_addresses) >= 1 ? 1 : 0
  vm_id       = truenas_vm.multi_gpu_vm.id
  device_type = "PCI"
  order       = 100
  
  pci_config = [{
    pptdev = local.gpu_addresses[0]
  }]
  
  depends_on = [data.truenas_gpu_pci_choices.gpus]
}

resource "truenas_vm_device" "gpu2" {
  count       = length(local.gpu_addresses) >= 2 ? 1 : 0
  vm_id       = truenas_vm.multi_gpu_vm.id
  device_type = "PCI"
  order       = 200
  
  pci_config = [{
    pptdev = local.gpu_addresses[1]
  }]
  
  depends_on = [data.truenas_gpu_pci_choices.gpus]
}
```

#### GPU Selection Logic
```terraform
data "truenas_gpu_pci_choices" "gpus" {
}

locals {
  # Prefer NVIDIA over AMD over Intel
  preferred_gpu = (
    length([
      for k, v in data.truenas_gpu_pci_choices.gpus.choices
      : v if contains(lower(k), "nvidia")
    ]) > 0 ? 
    values([
      for k, v in data.truenas_gpu_pci_choices.gpus.choices
      : v if contains(lower(k), "nvidia")
    ])[0] :
    length([
      for k, v in data.truenas_gpu_pci_choices.gpus.choices
      : v if contains(lower(k), "amd")
    ]) > 0 ?
    values([
      for k, v in data.truenas_gpu_pci_choices.gpus.choices
      : v if contains(lower(k), "amd")
    ])[0] :
    values(data.truenas_gpu_pci_choices.gpus.choices)[0]
  )
}

output "selected_gpu" {
  value = local.preferred_gpu
}
```

## Best Practices

### GPU Selection

1. **Vendor Preference**: Choose based on workload requirements
2. **Performance Needs**: Select appropriate GPU for use case
3. **Compatibility**: Verify guest OS driver support
4. **Power Requirements**: Consider power consumption

### Passthrough Configuration

1. **IOMMU Required**: Ensure IOMMU is enabled in BIOS
2. **VFIO Support**: Verify kernel VFIO support
3. **Driver Blacklist**: Blacklist host GPU drivers
4. **Isolation**: Properly isolate GPU for VM use

### Performance Optimization

1. **Dedicated GPU**: Use entire GPU for best performance
2. **Memory Allocation**: Consider GPU memory requirements
3. **Power Management**: Configure appropriate power settings
4. **Monitoring**: Monitor GPU performance in VM

### Security

1. **Access Control**: Limit GPU passthrough to authorized VMs
2. **Isolation**: Ensure proper device isolation
3. **Monitoring**: Monitor GPU usage and access
4. **Compliance**: Follow licensing requirements

## Troubleshooting

### No GPUs Detected

1. Verify IOMMU is enabled in BIOS
2. Check kernel VFIO support
3. Review system hardware configuration
4. Update TrueNAS to latest version

### GPU Not Available for Passthrough

1. Check if GPU is in use by host
2. Verify VFIO driver binding
3. Review kernel module configuration
4. Check GPU reset capabilities

### PCI Address Issues

1. Verify PCI address format
2. Check device availability
3. Test with different GPUs
4. Review system logs

### Performance Problems

1. Monitor GPU utilization
2. Check driver versions in guest
3. Verify proper GPU isolation
4. Review VM configuration

### Import Issues

1. Verify data source syntax
2. Check provider configuration
3. Test with simple configuration
4. Review TrueNAS API access

## See Also

- [truenas_vm_pci_passthrough_devices](vm_pci_passthrough_devices) - Comprehensive PCI device discovery
- [truenas_vm_iommu_enabled](vm_iommu_enabled) - IOMMU capability checking
- [truenas_vm_device](../resources/vm_device) - VM device management
- [TrueNAS GPU Passthrough Guide](https://www.truenas.com/docs/scale/virtualmachines/gpu-passthrough/) - Official GPU passthrough documentation
- [VFIO Configuration](https://wiki.archlinux.org/title/PCI_passthrough_via_OVMF) - Linux VFIO setup guide