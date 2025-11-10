terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
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

# ============================================================================
# DISCOVER NFS SHARES
# ============================================================================

data "truenas_nfs_shares" "all" {}

output "nfs_shares" {
  value = {
    for share in data.truenas_nfs_shares.all.shares :
    share.path => {
      id      = share.id
      enabled = share.enabled
      comment = share.comment
      networks = share.networks
    }
  }
  description = "All NFS shares on the system"
}

# Filter enabled NFS shares
locals {
  enabled_nfs_shares = [
    for share in data.truenas_nfs_shares.all.shares :
    share if share.enabled
  ]
}

output "enabled_nfs_paths" {
  value       = [for share in local.enabled_nfs_shares : share.path]
  description = "Paths of all enabled NFS shares"
}

# ============================================================================
# DISCOVER SMB SHARES
# ============================================================================

data "truenas_smb_shares" "all" {}

output "smb_shares" {
  value = {
    for share in data.truenas_smb_shares.all.shares :
    share.name => {
      id       = share.id
      path     = share.path
      enabled  = share.enabled
      readonly = share.readonly
      comment  = share.comment
    }
  }
  description = "All SMB shares on the system"
}

# Filter Time Machine shares
locals {
  timemachine_shares = [
    for share in data.truenas_smb_shares.all.shares :
    share if share.timemachine
  ]
}

output "timemachine_share_names" {
  value       = [for share in local.timemachine_shares : share.name]
  description = "Names of all Time Machine shares"
}

# ============================================================================
# DISCOVER VMs
# ============================================================================

data "truenas_vms" "all" {}

output "all_vms" {
  value = {
    for vm in data.truenas_vms.all.vms :
    vm.name => {
      id     = vm.id
      status = vm.status
      vcpus  = vm.vcpus
      memory = vm.memory
    }
  }
  description = "All VMs on the system"
}

# Filter running VMs
locals {
  running_vms = [
    for vm in data.truenas_vms.all.vms :
    vm if vm.status == "RUNNING"
  ]
}

output "running_vm_names" {
  value       = [for vm in local.running_vms : vm.name]
  description = "Names of all running VMs"
}

# ============================================================================
# QUERY SPECIFIC VM
# ============================================================================

# Query VM by name
data "truenas_vm" "worker" {
  name = "talos-worker-1"
}

output "worker_vm_info" {
  value = {
    id          = data.truenas_vm.worker.id
    status      = data.truenas_vm.worker.status
    vcpus       = data.truenas_vm.worker.vcpus
    memory      = data.truenas_vm.worker.memory
    cpu_mode    = data.truenas_vm.worker.cpu_mode
    bootloader  = data.truenas_vm.worker.bootloader
  }
  description = "Information about the worker VM"
}

# Query VM by ID
data "truenas_vm" "by_id" {
  id = "7"
}

output "vm_by_id_name" {
  value       = data.truenas_vm.by_id.name
  description = "Name of VM with ID 7"
}

# ============================================================================
# PRACTICAL USE CASES
# ============================================================================

# Use case 1: Generate Kubernetes NFS StorageClass manifests
locals {
  nfs_storage_classes = {
    for share in local.enabled_nfs_shares :
    replace(replace(share.path, "/mnt/", ""), "/", "-") => {
      server = "10.0.0.83"
      path   = share.path
    }
  }
}

output "nfs_storage_class_config" {
  value       = local.nfs_storage_classes
  description = "Configuration for Kubernetes NFS StorageClasses"
}

# Use case 2: Find VMs that need to be started
locals {
  stopped_vms_with_autostart = [
    for vm in data.truenas_vms.all.vms :
    vm if vm.status == "STOPPED" && vm.autostart
  ]
}

output "vms_needing_start" {
  value       = [for vm in local.stopped_vms_with_autostart : vm.name]
  description = "VMs that are stopped but have autostart enabled"
}

# Use case 3: Inventory report
output "truenas_inventory" {
  value = {
    nfs_shares = {
      total   = length(data.truenas_nfs_shares.all.shares)
      enabled = length(local.enabled_nfs_shares)
    }
    smb_shares = {
      total       = length(data.truenas_smb_shares.all.shares)
      timemachine = length(local.timemachine_shares)
    }
    vms = {
      total   = length(data.truenas_vms.all.vms)
      running = length(local.running_vms)
    }
  }
  description = "TrueNAS resource inventory summary"
}

