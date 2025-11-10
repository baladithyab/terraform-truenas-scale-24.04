---
page_title: "truenas_iscsi_portal Resource - terraform-provider-truenas"
subcategory: "iSCSI"
description: |-
  Manages an iSCSI portal (network listener) on TrueNAS.
---

# truenas_iscsi_portal (Resource)

Manages an iSCSI portal (network listener) on TrueNAS. A portal defines the network endpoints that iSCSI initiators can connect to for accessing iSCSI targets.

## Example Usage

### Basic Portal

```terraform
resource "truenas_iscsi_portal" "default" {
  listen = [
    {
      ip   = "0.0.0.0"
      port = 3260
    }
  ]
  
  comment = "Default iSCSI portal"
}
```

### Portal with Specific IP

```terraform
resource "truenas_iscsi_portal" "storage_network" {
  listen = [
    {
      ip   = "192.168.10.10"
      port = 3260
    }
  ]
  
  comment = "Storage network iSCSI portal"
}
```

### Portal with Multiple Listen Addresses

```terraform
resource "truenas_iscsi_portal" "multi_homed" {
  listen = [
    {
      ip   = "192.168.10.10"
      port = 3260
    },
    {
      ip   = "10.0.0.10"
      port = 3260
    }
  ]
  
  comment = "Multi-homed iSCSI portal"
}
```

### Portal with Custom Port

```terraform
resource "truenas_iscsi_portal" "custom_port" {
  listen = [
    {
      ip   = "0.0.0.0"
      port = 3261
    }
  ]
  
  comment = "iSCSI portal on custom port"
}
```

### Portal with CHAP Authentication

```terraform
resource "truenas_iscsi_portal" "chap_portal" {
  listen = [
    {
      ip   = "0.0.0.0"
      port = 3260
    }
  ]
  
  discovery_authmethod = "CHAP"
  discovery_authgroup  = 1
  
  comment = "iSCSI portal with CHAP authentication"
}
```

### Portal with Mutual CHAP

```terraform
resource "truenas_iscsi_portal" "mutual_chap" {
  listen = [
    {
      ip   = "0.0.0.0"
      port = 3260
    }
  ]
  
  discovery_authmethod = "CHAP_MUTUAL"
  discovery_authgroup  = 2
  
  comment = "iSCSI portal with mutual CHAP authentication"
}
```

### Portal for Management Network

```terraform
resource "truenas_iscsi_portal" "management" {
  listen = [
    {
      ip   = "10.10.10.5"
      port = 3260
    }
  ]
  
  discovery_authmethod = "NONE"
  
  comment = "Management network iSCSI portal"
}
```

### Portal with IPv6

```terraform
resource "truenas_iscsi_portal" "ipv6_portal" {
  listen = [
    {
      ip   = "::"
      port = 3260
    }
  ]
  
  comment = "IPv6 iSCSI portal"
}
```

## Schema

### Required

- `listen` (Block List) List of IP addresses and ports to listen on. See [Listen Configuration](#listen-configuration) below.

### Optional

- `comment` (String) Portal comment/description.
- `discovery_authmethod` (String) Discovery authentication method. Options: `NONE`, `CHAP`, `CHAP_MUTUAL`. Default: `NONE`.
- `discovery_authgroup` (Number) Discovery authentication group ID.

### Listen Configuration

The `listen` block supports:

- `ip` (String, Required) IP address to listen on (0.0.0.0 for all IPv4, :: for all IPv6).
- `port` (Number, Optional) Port to listen on. Default: 3260.

### Read-Only

- `id` (String) Portal identifier.

## Import

iSCSI portals can be imported using portal ID:

```shell
terraform import truenas_iscsi_portal.existing 1
```

## Notes

### Portal Configuration

#### IP Address Selection
- **0.0.0.0**: Listen on all IPv4 addresses
- **::**: Listen on all IPv6 addresses
- **Specific IP**: Listen on specific network interface
- **Multiple IPs**: Configure multiple listen addresses for redundancy

#### Port Configuration
- **3260**: Standard iSCSI port (default)
- **Custom ports**: Can use non-standard ports for security
- **Port conflicts**: Ensure ports are not in use by other services

### Authentication Methods

#### NONE (Default)
- No authentication required for discovery
- Suitable for trusted networks
- Not recommended for production environments

```terraform
discovery_authmethod = "NONE"
```

#### CHAP
- One-way CHAP authentication
- Initiator authenticates to target
- More secure than NONE

```terraform
discovery_authmethod = "CHAP"
discovery_authgroup  = 1  # Reference to CHAP group
```

#### CHAP_MUTUAL
- Mutual CHAP authentication
- Both initiator and target authenticate each other
- Most secure option

```terraform
discovery_authmethod = "CHAP_MUTUAL"
discovery_authgroup  = 2  # Reference to CHAP group
```

### Network Considerations

#### Multi-homed Configuration
Configure portals on multiple networks:

```terraform
listen = [
  {
    ip   = "192.168.10.10"  # Storage network
    port = 3260
  },
  {
    ip   = "10.0.0.10"       # Management network
    port = 3260
  }
]
```

#### VLAN Considerations
- Ensure portal IPs are accessible on required VLANs
- Configure firewall rules for iSCSI traffic
- Test connectivity from initiator networks

#### IPv6 Support
- Use `::` for all IPv6 addresses
- Configure specific IPv6 addresses as needed
- Test IPv6 connectivity from initiators

### Security Best Practices

#### Network Isolation
- Use dedicated storage networks
- Implement proper VLAN segmentation
- Configure firewall rules appropriately

#### Authentication
- Always use authentication in production
- Implement mutual CHAP for high security
- Regularly rotate authentication credentials

#### Port Security
- Use standard port 3260 when possible
- Document any custom port usage
- Ensure port accessibility from initiator networks

### Performance Optimization

#### Network Configuration
- Use dedicated network interfaces
- Configure appropriate MTU settings
- Monitor network bandwidth utilization

#### Multi-path Configuration
- Configure multiple portals for redundancy
- Use different network paths
- Test failover scenarios

## Best Practices

### Planning

1. **Network Design**: Plan IP addressing and network topology
2. **Security Strategy**: Implement appropriate authentication methods
3. **Redundancy**: Configure multiple listen addresses where needed
4. **Documentation**: Document portal configurations and purposes

### Security

1. **Authentication**: Use CHAP or mutual CHAP in production
2. **Network Isolation**: Separate iSCSI traffic from other networks
3. **Firewall Rules**: Implement proper firewall configurations
4. **Monitoring**: Monitor portal access and authentication attempts

### Performance

1. **Dedicated Networks**: Use dedicated storage networks
2. **Network Optimization**: Configure appropriate MTU and QoS
3. **Load Distribution**: Distribute initiators across multiple portals
4. **Monitoring**: Monitor network performance and utilization

### Maintenance

1. **Regular Testing**: Test connectivity from initiator systems
2. **Configuration Backup**: Document and backup portal configurations
3. **Security Audits**: Regularly review authentication settings
4. **Performance Monitoring**: Track portal performance metrics

## Troubleshooting

### Portal Not Accessible

1. Verify IP address configuration
2. Check network interface status
3. Test network connectivity
4. Review firewall rules

### Authentication Failures

1. Verify authentication method configuration
2. Check CHAP group settings
3. Validate initiator credentials
4. Review authentication logs

### Port Conflicts

1. Check for port conflicts with other services
2. Verify port is not in use
3. Test with alternative ports
4. Review network service configurations

### Performance Issues

1. Monitor network bandwidth
2. Check for network congestion
3. Verify interface configuration
4. Test with different network paths

### Multi-path Issues

1. Verify all portal IPs are accessible
2. Check initiator multi-path configuration
3. Test failover scenarios
4. Review network routing

### Import Issues

1. Use correct portal ID
2. Verify portal exists in TrueNAS
3. Check portal is not in use
4. Ensure proper permissions

## See Also

- [truenas_iscsi_extent](iscsi_extent) - iSCSI extent management
- [truenas_iscsi_target](iscsi_target) - iSCSI target configuration
- [TrueNAS iSCSI Documentation](https://www.truenas.com/docs/scale/iscsi/) - Official iSCSI configuration guide
- [iSCSI Security Best Practices](https://www.truenas.com/docs/scale/iscsi/security/) - Security recommendations for iSCSI deployments