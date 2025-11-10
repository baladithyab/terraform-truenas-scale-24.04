---
page_title: "truenas_interface Resource - terraform-provider-truenas"
subcategory: "Networking"
description: |-
  Manages a network interface on TrueNAS.
---

# truenas_interface (Resource)

Manages a network interface on TrueNAS. This resource supports various interface types including physical interfaces, VLANs, bridges, and link aggregation groups (LAGs).

## Example Usage

### Basic Physical Interface

```terraform
resource "truenas_interface" "eth0" {
  name        = "eth0"
  type        = "PHYSICAL"
  description = "Primary network interface"
  
  ipv4_dhcp = true
  ipv6_auto = true
}
```

### Static IP Configuration

```terraform
resource "truenas_interface" "eth1" {
  name        = "eth1"
  type        = "PHYSICAL"
  description = "Management interface with static IP"
  
  ipv4_dhcp = false
  
  aliases = [
    {
      address = "192.168.1.100"
      netmask = 24
    }
  ]
}
```

### VLAN Interface

```terraform
resource "truenas_interface" "vlan10" {
  name                = "vlan10"
  type                = "VLAN"
  description         = "VLAN 10 for guest network"
  
  vlan_parent_interface = "eth0"
  vlan_tag            = 10
  vlan_pcp            = 0
  
  ipv4_dhcp = true
}
```

### Bridge Interface

```terraform
resource "truenas_interface" "br0" {
  name        = "br0"
  type        = "BRIDGE"
  description = "Bridge for VM network"
  
  bridge_members = ["eth1", "eth2"]
  
  aliases = [
    {
      address = "10.0.0.1"
      netmask = 24
    }
  ]
}
```

### Link Aggregation (LACP)

```terraform
resource "truenas_interface" "bond0" {
  name        = "bond0"
  type        = "LINK_AGGREGATION"
  description = "LACP bond for high availability"
  
  lag_ports    = ["eth0", "eth1"]
  lag_protocol = "LACP"
  
  ipv4_dhcp = true
}
```

### Link Aggregation (Failover)

```terraform
resource "truenas_interface" "bond1" {
  name        = "bond1"
  type        = "LINK_AGGREGATION"
  description = "Failover bond for redundancy"
  
  lag_ports    = ["eth2", "eth3"]
  lag_protocol = "FAILOVER"
  
  aliases = [
    {
      address = "192.168.2.10"
      netmask = 24
    }
  ]
}
```

### Interface with Custom MTU

```terraform
resource "truenas_interface" "jumbo_frame" {
  name        = "eth2"
  type        = "PHYSICAL"
  description = "Interface with jumbo frames"
  
  ipv4_dhcp = false
  mtu        = 9000
  
  aliases = [
    {
      address = "10.10.10.5"
      netmask = 24
    }
  ]
}
```

### Multiple VLANs on Single Interface

```terraform
resource "truenas_interface" "vlan20" {
  name                = "vlan20"
  type                = "VLAN"
  description         = "VLAN 20 for storage network"
  
  vlan_parent_interface = "eth0"
  vlan_tag            = 20
  
  aliases = [
    {
      address = "172.16.20.1"
      netmask = 24
    }
  ]
}

resource "truenas_interface" "vlan30" {
  name                = "vlan30"
  type                = "VLAN"
  description         = "VLAN 30 for backup network"
  
  vlan_parent_interface = "eth0"
  vlan_tag            = 30
  
  aliases = [
    {
      address = "172.16.30.1"
      netmask = 24
    }
  ]
}
```

## Schema

### Required

- `name` (String) Interface name (e.g., eth0, vlan10, br0, bond0).
- `type` (String) Interface type. Options: `PHYSICAL`, `VLAN`, `BRIDGE`, `LINK_AGGREGATION`.

### Optional

- `description` (String) Interface description.
- `ipv4_dhcp` (Boolean) Use DHCP for IPv4. Default: false.
- `ipv6_auto` (Boolean) Use auto-configuration for IPv6. Default: false.
- `aliases` (Block List) Static IP addresses. See [Alias Configuration](#alias-configuration) below.
- `mtu` (Number) Maximum Transmission Unit. Default: 1500.
- `vlan_parent_interface` (String) Parent interface for VLAN (required when type is VLAN).
- `vlan_tag` (Number) VLAN tag (required when type is VLAN).
- `vlan_pcp` (Number) VLAN Priority Code Point (0-7). Default: 0.
- `bridge_members` (List of String) Bridge member interfaces (required when type is BRIDGE).
- `lag_ports` (List of String) LAG member ports (required when type is LINK_AGGREGATION).
- `lag_protocol` (String) LAG protocol. Options: `LACP`, `FAILOVER`, `LOADBALANCE`, `ROUNDROBIN`, `NONE`.

### Alias Configuration

The `aliases` block supports:

- `address` (String, Required) IP address.
- `netmask` (Number, Required) Netmask in CIDR notation (e.g., 24 for /24).

### Read-Only

- `id` (String) Interface identifier (same as name).

## Import

Interfaces can be imported using the interface name:

```shell
terraform import truenas_interface.eth0 eth0
terraform import truenas_interface.vlan10 vlan10
terraform import truenas_interface.br0 br0
```

## Notes

### Interface Types

#### PHYSICAL
- Represents physical network interfaces
- Can be configured with DHCP or static IPs
- Supports MTU configuration
- Can be used as parent for VLANs or member for bridges/LAGs

#### VLAN
- Virtual LAN interface created on physical interface
- Requires `vlan_parent_interface` and `vlan_tag`
- Inherits physical properties from parent
- Supports PCP (Priority Code Point) for QoS

#### BRIDGE
- Layer 2 bridge connecting multiple interfaces
- Requires `bridge_members` list
- Acts as a network switch
- Useful for VM networks and network segmentation

#### LINK_AGGREGATION
- Combines multiple physical interfaces for redundancy/performance
- Requires `lag_ports` list and `lag_protocol`
- Different protocols provide different behaviors:
  - `LACP`: IEEE 802.3ad dynamic aggregation
  - `FAILOVER`: Active/passive redundancy
  - `LOADBALANCE`: Active/active load balancing
  - `ROUNDROBIN`: Packet-level round-robin distribution

### IP Configuration

#### DHCP Configuration
```terraform
ipv4_dhcp = true  # Enable DHCP for IPv4
ipv6_auto = true  # Enable auto-configuration for IPv6
```

#### Static IP Configuration
```terraform
ipv4_dhcp = false

aliases = [
  {
    address = "192.168.1.100"
    netmask = 24
  },
  {
    address = "192.168.1.101"
    netmask = 24
  }
]
```

### MTU Configuration

- Standard Ethernet MTU: 1500 bytes
- Jumbo frames: 9000 bytes (requires network support)
- VLAN overhead: Consider 4 bytes for VLAN tags
- Set consistently across network path

### VLAN Configuration

VLAN interfaces require:
1. Parent physical interface
2. VLAN tag (1-4094)
3. Optional PCP for QoS (0-7)

```terraform
vlan_parent_interface = "eth0"  # Physical interface
vlan_tag            = 100        # VLAN ID
vlan_pcp            = 3          # Priority (optional)
```

### Bridge Configuration

Bridges connect multiple interfaces at layer 2:

```terraform
bridge_members = ["eth1", "eth2", "eth3"]  # Member interfaces
```

Common use cases:
- VM network bridges
- Network segmentation
- Connecting multiple physical networks

### LAG Configuration

Link aggregation provides redundancy and performance:

```terraform
lag_ports    = ["eth0", "eth1"]  # Member interfaces
lag_protocol = "LACP"             # Aggregation protocol
```

Protocol selection:
- `LACP`: Best for switches supporting 802.3ad
- `FAILOVER`: Simple active/passive redundancy
- `LOADBALANCE`: Active/active with load distribution
- `ROUNDROBIN`: Basic packet distribution

## Best Practices

### Interface Naming

- Use descriptive names: `mgmt`, `storage`, `backup`
- Follow consistent naming conventions
- Include VLAN info in VLAN interface names: `eth0.100`, `eth0.200`

### Network Planning

- Plan IP address ranges carefully
- Document VLAN assignments
- Consider future expansion needs
- Test failover scenarios

### Performance Optimization

- Use appropriate MTU sizes
- Enable jumbo frames where supported
- Balance LAG protocols with network capabilities
- Monitor interface utilization

### Security

- Separate management and data networks
- Use VLANs for network segmentation
- Implement proper firewall rules
- Regularly audit interface configurations

## Troubleshooting

### Interface Not Coming Up

1. Check physical cable connections
2. Verify interface type and configuration
3. Check for IP address conflicts
4. Review switch port configuration

### VLAN Issues

1. Verify parent interface is up
2. Check VLAN tag range (1-4094)
3. Ensure switch supports VLAN tagging
4. Verify VLAN configuration on switch

### Bridge Problems

1. Check all member interfaces are up
2. Verify no IP address conflicts
3. Check for spanning tree issues
4. Review bridge member configuration

### LAG Failures

1. Verify all member interfaces are connected
2. Check switch LAG configuration
3. Ensure matching LAG protocol
4. Verify physical interface compatibility

### DHCP Not Working

1. Check DHCP server availability
2. Verify network connectivity
3. Check firewall rules
4. Review interface configuration

### MTU Issues

1. Verify MTU support across network path
2. Check for fragmentation
3. Test with different MTU sizes
4. Ensure consistent MTU configuration

## See Also

- [TrueNAS Network Configuration](https://www.truenas.com/docs/scale/network/) - Official TrueNAS networking documentation
- [VLAN Configuration Guide](https://www.truenas.com/docs/scale/network/vlans/) - VLAN setup and management
- [Link Aggregation](https://www.truenas.com/docs/scale/network/linkaggregation/) - LAG configuration and protocols