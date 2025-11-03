terraform {
  required_providers {
    truenas = {
      source  = "local/truenas/truenas"
      version = "0.2.14"
    }
  }
}

provider "truenas" {
  base_url = var.truenas_base_url
  api_key  = var.truenas_api_key
}

resource "truenas_vm" "boot_order_test" {
  name        = "bootordertest"
  description = "Test VM for boot order verification"
  memory      = 2048
  vcpus       = 2

  # NIC device - no order specified (should get auto order)
  nic_devices = [{
    type       = "VIRTIO"
    nic_attach = "eno1"
  }]

  # Disk device - order=2 (should boot SECOND)
  disk_devices = [{
    path  = "/dev/zvol/Loki/vms/talos_worker_muninn_01-disk0"
    type  = "VIRTIO"
    order = 2
  }]

  # CDROM device - order=1 (should boot FIRST)
  cdrom_devices = [{
    path  = "/mnt/Loki/isos/talos-v1.10.6-metal-amd64.iso"
    order = 1
  }]

  bootloader      = "UEFI"
  start_on_create = false
}

output "vm_id" {
  value = truenas_vm.boot_order_test.id
}

output "nic_devices" {
  value = truenas_vm.boot_order_test.nic_devices
}

output "disk_devices" {
  value = truenas_vm.boot_order_test.disk_devices
}

output "cdrom_devices" {
  value = truenas_vm.boot_order_test.cdrom_devices
}

