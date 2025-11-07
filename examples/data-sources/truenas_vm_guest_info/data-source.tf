# Query guest agent information from a VM (using SSH key)
data "truenas_vm_guest_info" "ubuntu" {
  vm_name      = "ubuntu-vm"
  truenas_host = var.truenas_host
  ssh_user     = var.ssh_user
  ssh_key_path = var.ssh_key_path

  # Optional: Configure SSH behavior
  ssh_strict_host_key_checking = true # Default: true (verify host keys)
  ssh_timeout_seconds          = 30   # Default: 30 seconds
}

variable "truenas_host" {
  description = "TrueNAS hostname or IP for SSH access (e.g., 10.0.0.83)"
  type        = string
}

variable "ssh_user" {
  description = "SSH user for TrueNAS host (usually 'root')"
  type        = string
  default     = "root"
}

variable "ssh_key_path" {
  description = "Path to SSH private key for TrueNAS host"
  type        = string
  default     = "~/.ssh/id_rsa"
}

# Output the IP addresses
output "ubuntu_ip_addresses" {
  value = data.truenas_vm_guest_info.ubuntu.ip_addresses
}

output "ubuntu_hostname" {
  value = data.truenas_vm_guest_info.ubuntu.hostname
}

output "ubuntu_os" {
  value = "${data.truenas_vm_guest_info.ubuntu.os_name} ${data.truenas_vm_guest_info.ubuntu.os_version}"
}

# Use the IP addresses in other resources
# For example, configure Talos to avoid these IPs
locals {
  # Get all IPs from VMs with guest agent
  existing_ips = data.truenas_vm_guest_info.ubuntu.ip_addresses

  # Define static IPs for Talos nodes (avoiding existing IPs)
  talos_control_plane_ips = [
    "10.0.0.101",
    "10.0.0.102",
    "10.0.0.103",
  ]

  talos_worker_ips = [
    "10.0.0.111",
    "10.0.0.112",
    "10.0.0.113",
  ]
}

# Example: Query multiple VMs
data "truenas_vm_guest_info" "plex" {
  vm_name                      = "plex-server"
  truenas_host                 = var.truenas_host
  ssh_user                     = var.ssh_user
  ssh_key_path                 = var.ssh_key_path
  ssh_strict_host_key_checking = true
  ssh_timeout_seconds          = 30
}

data "truenas_vm_guest_info" "nextcloud" {
  vm_name                      = "nextcloud"
  truenas_host                 = var.truenas_host
  ssh_user                     = var.ssh_user
  ssh_key_path                 = var.ssh_key_path
  ssh_strict_host_key_checking = true
  ssh_timeout_seconds          = 30
}

# Example: Using password authentication (requires sshpass)
# data "truenas_vm_guest_info" "windows" {
#   vm_name                      = "windows-vm"
#   truenas_host                 = var.truenas_host
#   ssh_user                     = var.ssh_user
#   ssh_password                 = var.truenas_ssh_password
#   ssh_strict_host_key_checking = false # Disable for dynamic hosts
#   ssh_timeout_seconds          = 60    # Longer timeout for slower networks
# }

# Collect all IPs from all VMs
locals {
  all_vm_ips = concat(
    data.truenas_vm_guest_info.ubuntu.ip_addresses,
    data.truenas_vm_guest_info.plex.ip_addresses,
    data.truenas_vm_guest_info.nextcloud.ip_addresses,
  )
}

output "all_vm_ips" {
  description = "All IP addresses from VMs with guest agent"
  value       = local.all_vm_ips
}

