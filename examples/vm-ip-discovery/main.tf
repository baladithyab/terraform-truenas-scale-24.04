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

variable "truenas_base_url" {
  description = "TrueNAS Base URL (e.g., http://10.0.0.83:81)"
  type        = string
}

variable "truenas_api_key" {
  description = "TrueNAS API Key"
  type        = string
  sensitive   = true
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

# ============================================================================
# Example 1: Get MAC Addresses from VMs
# ============================================================================

# Create a VM
resource "truenas_vm" "talos_worker" {
  name        = "talos-worker-01"
  description = "Talos Kubernetes Worker"
  memory      = 4096
  vcpus       = 2
  cores       = 2
  threads     = 1
  autostart   = true
}

# Output MAC addresses for DHCP lookup
output "talos_worker_mac_addresses" {
  description = "MAC addresses of Talos worker NICs - use these to look up DHCP leases"
  value       = truenas_vm.talos_worker.mac_addresses
}

# ============================================================================
# Example 2: Query Guest Agent for IP Addresses (VMs with guest agent only)
# ============================================================================

# Query Ubuntu VM (has guest agent installed)
data "truenas_vm_guest_info" "ubuntu" {
  vm_name      = "ubuntu-vm"
  truenas_host = var.truenas_host
  ssh_user     = var.ssh_user
  ssh_key_path = var.ssh_key_path
}

output "ubuntu_vm_info" {
  description = "Ubuntu VM information from guest agent"
  value = {
    ip_addresses = data.truenas_vm_guest_info.ubuntu.ip_addresses
    hostname     = data.truenas_vm_guest_info.ubuntu.hostname
    os_name      = data.truenas_vm_guest_info.ubuntu.os_name
    os_version   = data.truenas_vm_guest_info.ubuntu.os_version
  }
}

# ============================================================================
# Example 3: Use Guest Agent IPs to Configure Static IPs for Talos
# ============================================================================

# Get IPs from existing VMs
data "truenas_vm_guest_info" "plex" {
  vm_name      = "plex-server"
  truenas_host = var.truenas_host
  ssh_user     = var.ssh_user
  ssh_key_path = var.ssh_key_path
}

data "truenas_vm_guest_info" "nextcloud" {
  vm_name      = "nextcloud"
  truenas_host = var.truenas_host
  ssh_user     = var.ssh_user
  ssh_key_path = var.ssh_key_path
}

# Collect all existing IPs
locals {
  existing_vm_ips = concat(
    data.truenas_vm_guest_info.ubuntu.ip_addresses,
    data.truenas_vm_guest_info.plex.ip_addresses,
    data.truenas_vm_guest_info.nextcloud.ip_addresses,
  )

  # Define static IPs for Talos (avoiding existing IPs)
  # Make sure these don't conflict with existing_vm_ips
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

output "ip_allocation_plan" {
  description = "IP allocation plan for Talos cluster"
  value = {
    existing_ips        = local.existing_vm_ips
    talos_control_plane = local.talos_control_plane_ips
    talos_workers       = local.talos_worker_ips
  }
}

# ============================================================================
# Example 4: MAC Address Lookup Script
# ============================================================================

# Output a script to look up DHCP leases by MAC address
output "dhcp_lookup_script" {
  description = "Script to look up DHCP leases for Talos VMs"
  value       = <<-EOT
    #!/bin/bash
    # Look up DHCP leases for Talos worker VMs
    
    echo "Talos Worker MAC Addresses:"
    echo "${join("\n", truenas_vm.talos_worker.mac_addresses)}"
    
    echo ""
    echo "To find IP addresses, check your DHCP server:"
    echo "  - pfSense: Status > DHCP Leases"
    echo "  - Router: Check DHCP client list"
    echo "  - Linux DHCP server: cat /var/lib/dhcp/dhcpd.leases"
    
    echo ""
    echo "Or use nmap to scan for these MAC addresses:"
    %{for mac in truenas_vm.talos_worker.mac_addresses~}
    echo "nmap -sn 10.0.0.0/24 | grep -B 2 '${mac}'"
    %{endfor~}
  EOT
}

# ============================================================================
# Example 5: Complete Talos Configuration with Static IPs
# ============================================================================

# This shows how you would use the discovered IPs to configure Talos
# (Talos configuration is done outside Terraform, but you can use these values)

output "talos_machine_config_snippet" {
  description = "Talos machine config snippet with static IP"
  value       = <<-EOT
    # Add this to your Talos machine configuration
    machine:
      network:
        interfaces:
          - interface: eth0
            addresses:
              - ${local.talos_worker_ips[0]}/24
            routes:
              - network: 0.0.0.0/0
                gateway: 10.0.0.1
        nameservers:
          - 10.0.0.1
          - 8.8.8.8
  EOT
}

# ============================================================================
# Example 6: Query All VMs and Their MAC Addresses
# ============================================================================

# If you have multiple VMs, you can create a map
locals {
  vm_mac_map = {
    talos_worker = truenas_vm.talos_worker.mac_addresses
    # Add more VMs here as needed
  }
}

output "all_vm_macs" {
  description = "All VM MAC addresses for DHCP lookup"
  value       = local.vm_mac_map
}

