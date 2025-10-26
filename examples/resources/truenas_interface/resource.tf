# Physical interface with static IP
resource "truenas_interface" "eth0" {
  name        = "eth0"
  type        = "PHYSICAL"
  description = "Primary network interface"
  ipv4_dhcp   = false
  
  aliases {
    address = "192.168.1.10"
    netmask = 24
  }
}

# VLAN interface
resource "truenas_interface" "vlan10" {
  name                    = "vlan10"
  type                    = "VLAN"
  description             = "VLAN 10 - Management"
  vlan_parent_interface   = "eth0"
  vlan_tag                = 10
  vlan_pcp                = 0
  
  aliases {
    address = "10.0.10.1"
    netmask = 24
  }
}

# Bridge interface
resource "truenas_interface" "br0" {
  name            = "br0"
  type            = "BRIDGE"
  description     = "VM Bridge"
  bridge_members  = ["eth1", "eth2"]
  
  aliases {
    address = "192.168.100.1"
    netmask = 24
  }
}

# Link Aggregation (LACP)
resource "truenas_interface" "bond0" {
  name         = "bond0"
  type         = "LINK_AGGREGATION"
  description  = "LACP Bond"
  lag_ports    = ["eth2", "eth3"]
  lag_protocol = "LACP"
  
  aliases {
    address = "192.168.2.10"
    netmask = 24
  }
}

# Interface with DHCP
resource "truenas_interface" "eth1_dhcp" {
  name        = "eth1"
  type        = "PHYSICAL"
  description = "Secondary interface with DHCP"
  ipv4_dhcp   = true
  ipv6_auto   = true
}

# Interface with multiple IPs
resource "truenas_interface" "eth2_multi" {
  name        = "eth2"
  type        = "PHYSICAL"
  description = "Interface with multiple IPs"
  
  aliases {
    address = "192.168.3.10"
    netmask = 24
  }
  
  aliases {
    address = "192.168.3.11"
    netmask = 24
  }
  
  aliases {
    address = "10.0.0.10"
    netmask = 8
  }
}

# Import an existing interface
# terraform import truenas_interface.existing eth0

