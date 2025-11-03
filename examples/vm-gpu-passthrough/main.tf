terraform {
  required_providers {
    truenas = {
      source  = "terraform-providers/truenas"
      version = "~> 0.2.14"
    }
  }
}

provider "truenas" {
  base_url = "http://10.0.0.83:81"
  api_key  = var.truenas_api_key
}

variable "truenas_api_key" {
  description = "TrueNAS API key"
  type        = string
  sensitive   = true
}

# Check if IOMMU is enabled (required for PCI passthrough)
data "truenas_vm_iommu_enabled" "check" {}

# Get available GPU PCI IDs
data "truenas_gpu_pci_choices" "gpus" {}

# Get all available PCI passthrough devices
data "truenas_vm_pci_passthrough_devices" "all" {
  available_only = true
}

# Output IOMMU status
output "iommu_enabled" {
  value       = data.truenas_vm_iommu_enabled.check.enabled
  description = "Whether IOMMU is enabled on the TrueNAS system"
}

# Output available GPUs
output "available_gpus" {
  value       = data.truenas_gpu_pci_choices.gpus.choices
  description = "Map of GPU descriptions to PCI addresses"
}

# Output available PCI passthrough devices
output "available_pci_devices" {
  value       = data.truenas_vm_pci_passthrough_devices.all.devices
  description = "Available PCI passthrough devices"
  sensitive   = false
}

# Create a dataset for the VM disk
resource "truenas_dataset" "vm_disk" {
  name = "Loki/VMs/gpu-vm-disk"
  type = "VOLUME"
  volsize = 107374182400 # 100GB in bytes
  volblocksize = 16384
  sparse = false
  force_size = false
  comments = "GPU-enabled VM disk"
}

# Create a VM with GPU passthrough
# This example assumes you have an NVIDIA GPU at pci_0000_3b_00_0
# Adjust the pptdev value based on your system's available devices
resource "truenas_vm" "gpu_vm" {
  name        = "gpu-enabled-vm"
  description = "VM with NVIDIA GPU passthrough"
  
  # CPU and Memory
  vcpus  = 4
  cores  = 4
  threads = 1
  memory = 16384  # 16GB
  
  # GPU passthrough settings
  cpu_mode            = "HOST-PASSTHROUGH"  # Required for best GPU performance
  hide_from_msr       = true                # Hide hypervisor from MSR (helps with GPU drivers)
  ensure_display_device = false             # Disable virtual display when using GPU passthrough
  
  # Boot settings
  bootloader = "UEFI"
  autostart  = false
  start_on_create = false
  
  # Network interface
  nic_devices = [{
    type       = "VIRTIO"
    nic_attach = "eno1"
    trust_guest_rx_filters = false
    order      = 1000
  }]
  
  # Boot disk (VIRTIO for best performance)
  disk_devices = [{
    path   = "/dev/zvol/${truenas_dataset.vm_disk.id}"
    type   = "VIRTIO"
    iotype = "THREADS"
    order  = 1001
  }]
  
  # PCI passthrough device (GPU)
  # Use the data source to find the correct pptdev ID for your GPU
  # Example: pci_0000_3b_00_0 for NVIDIA GPU at 0000:3b:00.0
  pci_devices = [{
    pptdev = "pci_0000_3b_00_0"  # Replace with your GPU's PCI ID
    order  = 1002
  }]
  
  depends_on = [truenas_dataset.vm_disk]
}

# Output VM information
output "vm_id" {
  value       = truenas_vm.gpu_vm.id
  description = "VM ID"
}

output "vm_status" {
  value       = truenas_vm.gpu_vm.status
  description = "VM status"
}

# Example: How to find the correct PCI device ID for your GPU
# 1. Run: terraform apply -target=data.truenas_gpu_pci_choices.gpus
# 2. Check the output for your GPU's PCI address (e.g., "0000:3b:00.0")
# 3. Convert the address to the pptdev format: replace colons and dots with underscores
#    and prefix with "pci_" (e.g., "0000:3b:00.0" becomes "pci_0000_3b_00_0")
# 4. Verify the device is available in the passthrough devices data source:
#    terraform apply -target=data.truenas_vm_pci_passthrough_devices.all
# 5. Update the pci_devices block with the correct pptdev value

# Important notes:
# - IOMMU must be enabled in your system BIOS/UEFI
# - The GPU must be isolated from the host OS (check TrueNAS System > Advanced > Isolated GPU PCI IDs)
# - All devices in the same IOMMU group must be passed through together
# - The VM guest OS needs appropriate GPU drivers installed
# - For NVIDIA GPUs, you may need to hide the hypervisor signature (hide_from_msr = true)

