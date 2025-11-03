# GPU Passthrough Example

This example demonstrates how to create a TrueNAS VM with GPU passthrough using the TrueNAS Terraform provider.

## Prerequisites

1. **IOMMU Enabled**: IOMMU (Intel VT-d or AMD-Vi) must be enabled in your system BIOS/UEFI
2. **GPU Isolation**: The GPU must be isolated from the host OS in TrueNAS (System > Advanced > Isolated GPU PCI IDs)
3. **TrueNAS Scale 24.04+**: This feature requires TrueNAS Scale 24.04 or later
4. **Provider Version**: terraform-provider-truenas v0.2.14 or later

## Features Demonstrated

- **Data Sources**:
  - `truenas_vm_iommu_enabled`: Check if IOMMU is enabled
  - `truenas_gpu_pci_choices`: Discover available GPUs
  - `truenas_vm_pci_passthrough_devices`: List all available PCI passthrough devices

- **VM Resource Enhancements**:
  - `pci_devices`: Attach PCI passthrough devices (GPUs, network cards, etc.)
  - `hide_from_msr`: Hide KVM hypervisor from MSR discovery (useful for NVIDIA GPUs)
  - `ensure_display_device`: Control virtual display device (set to false for GPU passthrough)
  - `cpu_mode`: Set to HOST-PASSTHROUGH for best GPU performance

## Usage

1. **Discover Available Devices**:
   ```bash
   # Check if IOMMU is enabled
   terraform apply -target=data.truenas_vm_iommu_enabled.check
   
   # List available GPUs
   terraform apply -target=data.truenas_gpu_pci_choices.gpus
   
   # List all available PCI passthrough devices
   terraform apply -target=data.truenas_vm_pci_passthrough_devices.all
   ```

2. **Find Your GPU's PCI ID**:
   - Look at the output from `truenas_gpu_pci_choices.gpus`
   - Note the PCI address (e.g., `0000:3b:00.0`)
   - Convert to pptdev format: `pci_0000_3b_00_0` (replace `:` and `.` with `_`, prefix with `pci_`)

3. **Update the Configuration**:
   - Edit `main.tf` and update the `pptdev` value in the `pci_devices` block
   - Adjust CPU, memory, and other settings as needed

4. **Apply the Configuration**:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## Important Notes

### IOMMU Groups
When passing through a GPU, all devices in the same IOMMU group must be isolated from the host. Use the `truenas_vm_pci_passthrough_devices` data source to check IOMMU group information.

### GPU Isolation
Before creating the VM, ensure the GPU is isolated in TrueNAS:
1. Go to System > Advanced
2. Add the GPU's PCI ID to "Isolated GPU PCI IDs"
3. Reboot the TrueNAS system

### NVIDIA GPUs
For NVIDIA GPUs, you typically need:
- `cpu_mode = "HOST-PASSTHROUGH"`
- `hide_from_msr = true`
- `ensure_display_device = false`

### Guest OS Drivers
After creating the VM, you'll need to:
1. Install the guest operating system
2. Install appropriate GPU drivers in the guest OS
3. For NVIDIA GPUs on Linux, you may need to blacklist the nouveau driver

## Example Output

```hcl
iommu_enabled = true

available_gpus = {
  "ASPEED Technology, Inc. ASPEED Graphics Family" = "0000:04:00.0"
  "NVIDIA Corporation Device 2584" = "0000:3b:00.0"
}

vm_id = "7"
vm_status = "STOPPED"
```

## Troubleshooting

### IOMMU Not Enabled
If `iommu_enabled = false`, enable IOMMU in your system BIOS/UEFI:
- **Intel**: Enable VT-d
- **AMD**: Enable AMD-Vi or IOMMU

### GPU Not Available
If the GPU doesn't appear in `available_pci_devices`:
1. Check if it's already in use by the host
2. Verify it's isolated in System > Advanced > Isolated GPU PCI IDs
3. Reboot TrueNAS after changing isolation settings

### VM Fails to Start
Common issues:
- GPU not properly isolated from host
- IOMMU group conflicts
- Missing or incorrect pptdev ID
- Insufficient memory or CPU allocation

## References

- [TrueNAS Scale VM Documentation](https://www.truenas.com/docs/scale/scaletutorials/virtualization/)
- [PCI Passthrough Guide](https://www.truenas.com/docs/scale/scaletutorials/virtualization/pcipassthrough/)

