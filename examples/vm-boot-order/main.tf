terraform {
  required_providers {
    truenas = {
      source  = "terraform-providers/truenas"
      version = "~> 0.2.14"
    }
  }
}

provider "truenas" {
  base_url = var.truenas_base_url
  api_key  = var.truenas_api_key
}

# Example 1: Boot from CDROM first (for OS installation), then disk
# This is useful when installing an OS - the VM will boot from the ISO first
resource "truenas_vm" "install_from_iso" {
  name        = "installfromiso"
  description = "VM that boots from CDROM first for OS installation"
  vcpus       = 2
  memory      = 4096

  nic_devices = [{
    type       = "VIRTIO"
    nic_attach = "eno1"
    order      = 1003 # NICs don't affect boot order, but we set it for consistency
  }]

  # CDROM with lower order boots FIRST
  cdrom_devices = [{
    path  = "/mnt/pool/isos/ubuntu-22.04.iso"
    order = 1 # Boot from ISO FIRST
  }]

  # Disk with higher order boots SECOND
  disk_devices = [{
    path  = "/dev/zvol/pool/vms/install-disk0"
    type  = "VIRTIO"
    order = 2 # Boot from disk SECOND (after ISO installation completes)
  }]

  bootloader      = "UEFI"
  start_on_create = true
}

# Example 2: Boot from disk first (normal operation after OS is installed)
# This is the typical configuration after OS installation is complete
resource "truenas_vm" "boot_from_disk" {
  name        = "bootfromdisk"
  description = "VM that boots from disk (OS already installed)"
  vcpus       = 2
  memory      = 4096

  nic_devices = [{
    type       = "VIRTIO"
    nic_attach = "eno1"
    order      = 1003
  }]

  # Disk with lower order boots FIRST
  disk_devices = [{
    path  = "/dev/zvol/pool/vms/boot-disk0"
    type  = "VIRTIO"
    order = 1 # Boot from disk FIRST
  }]

  # CDROM with higher order boots SECOND (fallback)
  cdrom_devices = [{
    path  = "/mnt/pool/isos/ubuntu-22.04.iso"
    order = 2 # Boot from ISO only if disk boot fails
  }]

  bootloader      = "UEFI"
  start_on_create = true
}

# Example 3: Talos Linux worker with proper boot order
# Boot from ISO first to install Talos, then from disk for normal operation
resource "truenas_vm" "talos_worker" {
  name        = "talosworker01"
  description = "Talos Linux worker node"
  vcpus       = 4
  memory      = 8192

  nic_devices = [{
    type       = "VIRTIO"
    nic_attach = "eno1"
    order      = 1003
  }]

  # For initial installation: CDROM boots first
  cdrom_devices = [{
    path  = "/mnt/pool/isos/talos-v1.10.6-metal-amd64.iso"
    order = 1 # Boot from Talos ISO FIRST for installation
  }]

  # After installation: disk will have bootloader and boot second
  disk_devices = [{
    path   = "/dev/zvol/pool/vms/talos-worker-01-disk0"
    type   = "VIRTIO"
    iotype = "THREADS"
    order  = 2 # Boot from disk SECOND (after Talos is installed)
  }]

  bootloader      = "UEFI"
  start_on_create = true
}

# Example 4: Multiple disks with specific boot order
resource "truenas_vm" "multi_disk" {
  name        = "multidisk"
  description = "VM with multiple disks and specific boot order"
  vcpus       = 4
  memory      = 8192

  nic_devices = [{
    type       = "VIRTIO"
    nic_attach = "eno1"
    order      = 1004
  }]

  # Boot from first disk (OS disk)
  disk_devices = [
    {
      path  = "/dev/zvol/pool/vms/multi-os-disk"
      type  = "VIRTIO"
      order = 1 # Primary boot disk
    },
    {
      path  = "/dev/zvol/pool/vms/multi-data-disk"
      type  = "VIRTIO"
      order = 3 # Data disk (not bootable, lower priority)
    }
  ]

  # Rescue ISO as fallback
  cdrom_devices = [{
    path  = "/mnt/pool/isos/rescue.iso"
    order = 2 # Boot from rescue ISO if primary disk fails
  }]

  bootloader = "UEFI"
}

# Example 5: Default boot order (no order specified)
# When order is not specified, devices are ordered by type:
# NICs first, then disks, then CDROMs (starting at order 1000)
resource "truenas_vm" "default_order" {
  name        = "defaultorder"
  description = "VM with default device ordering"
  vcpus       = 2
  memory      = 4096

  # No order specified - will get order 1000
  nic_devices = [{
    type       = "VIRTIO"
    nic_attach = "eno1"
  }]

  # No order specified - will get order 1001
  disk_devices = [{
    path = "/dev/zvol/pool/vms/default-disk0"
    type = "VIRTIO"
  }]

  # No order specified - will get order 1002
  cdrom_devices = [{
    path = "/mnt/pool/isos/ubuntu-22.04.iso"
  }]

  bootloader = "UEFI"
}

# Outputs to verify boot order
output "install_from_iso_info" {
  value = {
    name        = truenas_vm.install_from_iso.name
    description = "Boots from CDROM (order 1) first, then disk (order 2)"
  }
}

output "boot_from_disk_info" {
  value = {
    name        = truenas_vm.boot_from_disk.name
    description = "Boots from disk (order 1) first, then CDROM (order 2)"
  }
}

output "talos_worker_info" {
  value = {
    name        = truenas_vm.talos_worker.name
    mac         = truenas_vm.talos_worker.mac_addresses
    description = "Talos worker - boots from ISO (order 1) for install, disk (order 2) after"
  }
}

