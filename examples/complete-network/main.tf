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

# Configure primary interface with static IP
resource "truenas_interface" "primary" {
  name        = "eth0"
  type        = "PHYSICAL"
  description = "Primary Network Interface"
  ipv4_dhcp   = false
  
  aliases {
    address = "192.168.1.10"
    netmask = 24
  }
}

# Create VLAN for management
resource "truenas_interface" "mgmt_vlan" {
  name                  = "vlan100"
  type                  = "VLAN"
  description           = "Management VLAN"
  vlan_parent_interface = truenas_interface.primary.name
  vlan_tag              = 100
  
  aliases {
    address = "10.0.100.10"
    netmask = 24
  }
}

# Create VLAN for storage
resource "truenas_interface" "storage_vlan" {
  name                  = "vlan200"
  type                  = "VLAN"
  description           = "Storage VLAN"
  vlan_parent_interface = truenas_interface.primary.name
  vlan_tag              = 200
  
  aliases {
    address = "10.0.200.10"
    netmask = 24
  }
}

# Create bridge for VMs
resource "truenas_interface" "vm_bridge" {
  name           = "br0"
  type           = "BRIDGE"
  description    = "VM Bridge"
  bridge_members = ["eth1"]
  
  aliases {
    address = "192.168.100.1"
    netmask = 24
  }
}

# Add default route
resource "truenas_static_route" "default" {
  destination = "default"
  gateway     = "192.168.1.1"
  description = "Default gateway"
  
  depends_on = [truenas_interface.primary]
}

# Add route to remote network
resource "truenas_static_route" "remote_network" {
  destination = "10.10.0.0/16"
  gateway     = "192.168.1.254"
  description = "Route to remote datacenter"
  
  depends_on = [truenas_interface.primary]
}

# Outputs
output "primary_interface" {
  value = {
    name = truenas_interface.primary.name
    id   = truenas_interface.primary.id
  }
  description = "Primary interface details"
}

output "vlans" {
  value = {
    management = truenas_interface.mgmt_vlan.name
    storage    = truenas_interface.storage_vlan.name
  }
  description = "VLAN interface names"
}

output "routes" {
  value = {
    default = truenas_static_route.default.id
    remote  = truenas_static_route.remote_network.id
  }
  description = "Static route IDs"
}

