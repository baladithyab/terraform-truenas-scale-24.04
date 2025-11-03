terraform {
  required_providers {
    truenas = {
      source = "terraform-providers/truenas/truenas"
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

# ============================================================================
# CREATE ZVOL FOR VM DISK
# ============================================================================

resource "truenas_dataset" "vm_disk" {
  name = "Loki/terraform-test-vm-disk0"
  type = "VOLUME"

  # 8GB zvol
  volsize = 8589934592

  # Auto-cleanup on destroy
  force_destroy     = true
  recursive_destroy = true
}

# ============================================================================
# TEST VM WITH TALOS 1.11.3 ISO
# ============================================================================

resource "truenas_vm" "test" {
  name        = "terraformtestvm"
  description = "Test VM for provider validation"

  # CPU configuration
  vcpus   = 2
  cores   = 1
  threads = 1

  # Memory configuration (2GB)
  memory = 2048

  # Boot configuration
  bootloader = "UEFI"
  cpu_mode   = "HOST-PASSTHROUGH"
  autostart  = false

  # Start VM after creation
  start_on_create = true

  # CDROM device with Talos ISO
  cdrom_devices = [
    {
      path  = "/mnt/Loki/isos/talos-v1.11.3-metal-amd64.iso"
      order = 1001
    }
  ]

  # Network device
  nic_devices = [
    {
      type       = "VIRTIO"
      mac        = "00:a0:98:12:34:56"
      nic_attach = "eno1"
      order      = 1000
    }
  ]

  # Disk device (8GB zvol)
  disk_devices = [
    {
      path   = "/dev/zvol/${truenas_dataset.vm_disk.id}"
      type   = "VIRTIO"
      iotype = "THREADS"
      order  = 1002
    }
  ]

  depends_on = [truenas_dataset.vm_disk]
}

# ============================================================================
# TEST DATA SOURCES
# ============================================================================

# Test querying the VM by name
data "truenas_vm" "test_by_name" {
  name = truenas_vm.test.name

  depends_on = [truenas_vm.test]
}

# Test querying the VM by ID
data "truenas_vm" "test_by_id" {
  id = truenas_vm.test.id

  depends_on = [truenas_vm.test]
}

# Test listing all VMs
data "truenas_vms" "all" {
  depends_on = [truenas_vm.test]
}

# Test VM guest info (will work once Talos boots and guest agent is running)
# NOTE: Commented out for initial test - requires SSH key auth to be set up
# data "truenas_vm_guest_info" "test" {
#   vm_name      = truenas_vm.test.name
#   truenas_host = "10.0.0.83"
#   ssh_user     = "root"
#   ssh_key_path = "~/.ssh/id_rsa"
#
#   depends_on = [truenas_vm.test]
# }

# ============================================================================
# OUTPUTS
# ============================================================================

output "vm_id" {
  value       = truenas_vm.test.id
  description = "Created VM ID"
}

output "vm_status" {
  value       = truenas_vm.test.status
  description = "VM status after creation"
}

output "vm_by_name" {
  value = {
    id     = data.truenas_vm.test_by_name.id
    name   = data.truenas_vm.test_by_name.name
    status = data.truenas_vm.test_by_name.status
    vcpus  = data.truenas_vm.test_by_name.vcpus
    memory = data.truenas_vm.test_by_name.memory
  }
  description = "VM queried by name"
}

output "vm_by_id" {
  value = {
    id     = data.truenas_vm.test_by_id.id
    name   = data.truenas_vm.test_by_id.name
    status = data.truenas_vm.test_by_id.status
  }
  description = "VM queried by ID"
}

output "all_vms_count" {
  value       = length(data.truenas_vms.all.vms)
  description = "Total number of VMs on the system"
}

output "test_vm_in_list" {
  value = [
    for vm in data.truenas_vms.all.vms :
    vm if vm.name == truenas_vm.test.name
  ]
  description = "Test VM found in VM list"
}

# output "guest_info" {
#   value = {
#     ip_addresses = try(data.truenas_vm_guest_info.test.ip_addresses, [])
#     hostname     = try(data.truenas_vm_guest_info.test.hostname, "N/A")
#     os_name      = try(data.truenas_vm_guest_info.test.os_name, "N/A")
#     os_version   = try(data.truenas_vm_guest_info.test.os_version, "N/A")
#   }
#   description = "Guest agent information (may be empty if agent not running yet)"
# }

