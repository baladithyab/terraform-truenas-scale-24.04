terraform {
  required_providers {
    truenas = {
      source = "terraform-providers/truenas"
    }
  }
}

provider "truenas" {
  base_url = "http://10.0.0.213:81"
  api_key  = var.truenas_api_key
}

variable "truenas_api_key" {
  description = "TrueNAS API Key"
  type        = string
  sensitive   = true
}

# Get information about the pool
data "truenas_pool" "tank" {
  id = "tank"
}

# Create a parent dataset for the project
resource "truenas_dataset" "project" {
  name        = "tank/myproject"
  type        = "FILESYSTEM"
  compression = "LZ4"
  atime       = "OFF"
  comments    = "Project root dataset"
}

# Create a dataset for shared data
resource "truenas_dataset" "shared" {
  name        = "tank/myproject/shared"
  type        = "FILESYSTEM"
  compression = "LZ4"
  atime       = "OFF"
  quota       = 53687091200 # 50GB
  comments    = "Shared data for the project"

  depends_on = [truenas_dataset.project]
}

# Create a dataset for user homes
resource "truenas_dataset" "homes" {
  name        = "tank/myproject/homes"
  type        = "FILESYSTEM"
  compression = "LZ4"
  atime       = "OFF"
  comments    = "User home directories"

  depends_on = [truenas_dataset.project]
}

# Create a group for project members
resource "truenas_group" "project_users" {
  name = "project_users"
  sudo = false
  smb  = true
}

# Create users
resource "truenas_user" "alice" {
  username  = "alice"
  full_name = "Alice Smith"
  email     = "alice@example.com"
  password  = var.alice_password
  group     = truenas_group.project_users.gid
  home      = "/mnt/tank/myproject/homes/alice"
  shell     = "/bin/bash"
  sudo      = false
  smb       = true
}

resource "truenas_user" "bob" {
  username  = "bob"
  full_name = "Bob Johnson"
  email     = "bob@example.com"
  password  = var.bob_password
  group     = truenas_group.project_users.gid
  home      = "/mnt/tank/myproject/homes/bob"
  shell     = "/bin/bash"
  sudo      = false
  smb       = true
}

variable "alice_password" {
  description = "Password for Alice"
  type        = string
  sensitive   = true
}

variable "bob_password" {
  description = "Password for Bob"
  type        = string
  sensitive   = true
}

# Create an NFS share for the shared dataset
resource "truenas_nfs_share" "shared" {
  path     = "/mnt/${truenas_dataset.shared.name}"
  comment  = "NFS share for project shared data"
  networks = ["192.168.1.0/24"]
  readonly = false
  enabled  = true

  depends_on = [truenas_dataset.shared]
}

# Create an SMB share for the shared dataset
resource "truenas_smb_share" "shared" {
  name       = "project_shared"
  path       = "/mnt/${truenas_dataset.shared.name}"
  comment    = "SMB share for project shared data"
  enabled    = true
  browsable  = true
  guestok    = false
  readonly   = false
  recyclebin = true
  shadowcopy = true

  depends_on = [truenas_dataset.shared]
}

# Create an SMB share for user homes
resource "truenas_smb_share" "homes" {
  name       = "homes"
  path       = "/mnt/${truenas_dataset.homes.name}"
  comment    = "User home directories"
  enabled    = true
  browsable  = true
  guestok    = false
  readonly   = false
  recyclebin = true
  shadowcopy = true

  depends_on = [truenas_dataset.homes]
}

# Outputs
output "pool_info" {
  description = "Information about the pool"
  value = {
    name      = data.truenas_pool.tank.name
    status    = data.truenas_pool.tank.status
    healthy   = data.truenas_pool.tank.healthy
    available = data.truenas_pool.tank.available
    size      = data.truenas_pool.tank.size
  }
}

output "dataset_info" {
  description = "Information about created datasets"
  value = {
    project = truenas_dataset.project.name
    shared  = truenas_dataset.shared.name
    homes   = truenas_dataset.homes.name
  }
}

output "nfs_share_path" {
  description = "NFS share path"
  value       = truenas_nfs_share.shared.path
}

output "smb_shares" {
  description = "SMB share names"
  value = {
    shared = truenas_smb_share.shared.name
    homes  = truenas_smb_share.homes.name
  }
}

output "users_created" {
  description = "Users created"
  value = {
    alice = truenas_user.alice.username
    bob   = truenas_user.bob.username
  }
}

