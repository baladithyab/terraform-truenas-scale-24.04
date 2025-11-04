terraform {
  required_providers {
    truenas = {
      source  = "baladithyab/truenas"
      version = "~> 0.2.16"
    }
  }
}

provider "truenas" {
  base_url = var.truenas_base_url
  api_key  = var.truenas_api_key
}

# Create zvol for VM disk
resource "truenas_dataset" "vm_disk" {
  name    = "Loki/vms/talostest01-disk0"
  type    = "VOLUME"
  volsize = 10 * 1024 * 1024 * 1024 # 10GB

  force_destroy     = true
  recursive_destroy = true
}

# Minimal Talos VM - Just ISO, Disk, and Network
# Mimicking Proxmox setup: boot_order = ["scsi0", "ide2"]
# In TrueNAS terms: disk (order=1), cdrom (order=2)
resource "truenas_vm" "talos_minimal" {
  name        = "talostest01"
  description = "Minimal Talos VM to test boot and IP assignment"
  memory      = 4096
  vcpus       = 1
  cores       = 4
  threads     = 1
  cpu_mode    = "HOST-PASSTHROUGH"

  # Network device (VIRTIO like Proxmox)
  nic_devices = [{
    type       = "VIRTIO"
    nic_attach = var.network_bridge
  }]

  # CDROM device - order=1000 (boot FIRST, like UI-created VM)
  cdrom_devices = [{
    path  = "/mnt/Loki/isos/talos-v1.11.3-metal-amd64-factory.iso"
    order = 1000
  }]

  # Disk device - order=1001 (boot SECOND, after CDROM)
  disk_devices = [{
    path  = "/dev/zvol/${truenas_dataset.vm_disk.id}"
    type  = "AHCI"
    order = 1001
  }]

  # Display device - SPICE for console access
  display_devices = [{
    type       = "SPICE"
    port       = 5904
    bind       = "0.0.0.0"
    password   = "talos123"
    web        = true
    web_port   = 5905
    resolution = "1024x768"
    wait       = false
  }]

  bootloader            = "UEFI"
  machine_type          = null  # Don't set machine_type to avoid arch_type requirement
  ensure_display_device = true  # Ensure display device is created
  start_on_create       = true

  depends_on = [truenas_dataset.vm_disk]
}

output "vm_id" {
  value       = truenas_vm.talos_minimal.id
  description = "VM ID for talos-minimal-test"
}

output "vm_name" {
  value       = truenas_vm.talos_minimal.name
  description = "VM name"
}

output "nic_devices" {
  value       = truenas_vm.talos_minimal.nic_devices
  description = "Network device configuration with order"
}

output "disk_devices" {
  value       = truenas_vm.talos_minimal.disk_devices
  description = "Disk device configuration with order"
}

output "cdrom_devices" {
  value       = truenas_vm.talos_minimal.cdrom_devices
  description = "CDROM device configuration with order"
}
