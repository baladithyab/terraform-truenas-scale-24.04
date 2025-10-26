terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/YOUR_USERNAME/truenas"
      version = "~> 1.0"
    }
  }
}

provider "truenas" {
  base_url = var.truenas_base_url
  api_key  = var.truenas_api_key
}

# Create a dataset for iSCSI storage
resource "truenas_dataset" "iscsi_storage" {
  name        = "tank/iscsi"
  type        = "FILESYSTEM"
  compression = "LZ4"
  comments    = "iSCSI storage dataset"
}

# Create iSCSI portal
resource "truenas_iscsi_portal" "main" {
  comment = "Main iSCSI Portal"
  
  listen {
    ip   = "0.0.0.0"
    port = 3260
  }
}

# Create file-based extent
resource "truenas_iscsi_extent" "vm_disk1" {
  name     = "vm-disk-1"
  type     = "FILE"
  path     = "/mnt/${truenas_dataset.iscsi_storage.name}/vm-disk-1.img"
  filesize = 107374182400  # 100GB
  comment  = "VM Disk 1"
  enabled  = true
  readonly = false
  
  depends_on = [truenas_dataset.iscsi_storage]
}

# Create another extent
resource "truenas_iscsi_extent" "vm_disk2" {
  name     = "vm-disk-2"
  type     = "FILE"
  path     = "/mnt/${truenas_dataset.iscsi_storage.name}/vm-disk-2.img"
  filesize = 53687091200  # 50GB
  comment  = "VM Disk 2"
  enabled  = true
  readonly = false
  
  depends_on = [truenas_dataset.iscsi_storage]
}

# Create iSCSI target
resource "truenas_iscsi_target" "vm_target" {
  name          = "vm-target-1"
  alias         = "VM Storage Target"
  mode          = "ISCSI"
  groups        = [truenas_iscsi_portal.main.id]
  auth_networks = ["192.168.1.0/24"]
}

# Output the target IQN
output "iscsi_target_id" {
  value       = truenas_iscsi_target.vm_target.id
  description = "iSCSI Target ID"
}

output "iscsi_portal_id" {
  value       = truenas_iscsi_portal.main.id
  description = "iSCSI Portal ID"
}

output "extent_ids" {
  value = {
    disk1 = truenas_iscsi_extent.vm_disk1.id
    disk2 = truenas_iscsi_extent.vm_disk2.id
  }
  description = "iSCSI Extent IDs"
}

