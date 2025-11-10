terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.14"
    }
  }
}

provider "truenas" {
  base_url = var.truenas_base_url
  api_key  = var.truenas_api_key
}

# Example VM with NIC, Disk, and CDROM devices
resource "truenas_vm" "test_vm" {
  name        = "testvmwithdevices"
  description = "Test VM with NIC, disk, and CDROM devices"
  vcpus       = 2
  cores       = 1
  threads     = 1
  memory      = 4096 # 4GB in MB

  # Network interface devices
  nic_devices = [
    {
      type       = "VIRTIO"
      nic_attach = "eno1"
      # mac is optional - leave empty for auto-generation
      trust_guest_rx_filters = false
    }
  ]

  # Disk devices
  disk_devices = [
    {
      path   = "/dev/zvol/Loki/vms/test-vm-disk0"
      type   = "VIRTIO"
      iotype = "THREADS"
    }
  ]

  # CDROM devices
  cdrom_devices = [
    {
      path = "/mnt/Loki/isos/talos-v1.10.6-metal-amd64.iso"
    }
  ]

  autostart       = false
  start_on_create = false
}

# Example Talos worker VM with proper configuration
resource "truenas_vm" "talos_worker" {
  name        = "talosworkertest"
  description = "Talos Linux worker node"
  vcpus       = 4
  cores       = 2
  threads     = 2
  memory      = 8192 # 8GB in MB

  # Network interface - attach to eno1
  nic_devices = [
    {
      type       = "VIRTIO"
      nic_attach = "eno1"
      # Let TrueNAS auto-generate MAC address
      trust_guest_rx_filters = false
    }
  ]

  # Boot disk
  disk_devices = [
    {
      path   = "/dev/zvol/Loki/vms/talos-worker-test-disk0"
      type   = "VIRTIO"
      iotype = "THREADS"
    }
  ]

  # Talos ISO
  cdrom_devices = [
    {
      path = "/mnt/Loki/isos/talos-v1.10.6-metal-amd64.iso"
    }
  ]

  bootloader      = "UEFI"
  autostart       = false
  start_on_create = true
}

# Output the MAC addresses for network configuration
output "test_vm_mac_addresses" {
  value       = truenas_vm.test_vm.mac_addresses
  description = "MAC addresses of test VM NICs"
}

output "talos_worker_mac_addresses" {
  value       = truenas_vm.talos_worker.mac_addresses
  description = "MAC addresses of Talos worker NICs"
}

output "test_vm_status" {
  value       = truenas_vm.test_vm.status
  description = "Status of test VM"
}

output "talos_worker_status" {
  value       = truenas_vm.talos_worker.status
  description = "Status of Talos worker VM"
}

