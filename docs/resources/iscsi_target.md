---
page_title: "truenas_iscsi_target Resource - terraform-provider-truenas"
subcategory: "iSCSI"
description: |-
  Manages an iSCSI target on TrueNAS.
---

# truenas_iscsi_target (Resource)

Manages an iSCSI target on TrueNAS. An iSCSI target represents a storage endpoint that iSCSI initiators can connect to, combining extents and portals to provide access to storage resources.

## Example Usage

### Basic Target

```terraform
resource "truenas_iscsi_target" "storage_target" {
  name = "storage-target"
  
  comment = "Primary storage target"
}
```

### Target with Alias

```terraform
resource "truenas_iscsi_target" "backup_target" {
  name  = "backup-target"
  alias = "backup-storage"
  
  comment = "Backup storage target"
}
```

### Target with Portal Groups

```terraform
resource "truenas_iscsi_target" "production_target" {
  name = "production-target"
  
  groups = [1, 2]  # Portal group IDs
  
  comment = "Production storage with multiple portals"
}
```

### Target with Network Restrictions

```terraform
resource "truenas_iscsi_target" "restricted_target" {
  name = "restricted-target"
  
  auth_networks = [
    "192.168.10.0/24",
    "10.0.0.0/8"
  ]
  
  comment = "Target with network access restrictions"
}
```

### Target with Fibre Channel

```terraform
resource "truenas_iscsi_target" "fc_target" {
  name = "fc-target"
  mode = "FC"
  
  comment = "Fibre Channel target"
}
```

### Target with Both iSCSI and FC

```terraform
resource "truenas_iscsi_target" "dual_target" {
  name = "dual-target"
  mode = "BOTH"
  
  groups = [1, 2]
  auth_networks = ["192.168.10.0/24"]
  
  comment = "Dual-mode iSCSI and FC target"
}
```

### Complete Target Configuration

```terraform
resource "truenas_iscsi_target" "complete_target" {
  name  = "complete-target"
  alias = "enterprise-storage"
  mode  = "ISCSI"
  
  groups = [1, 2, 3]
  auth_networks = [
    "192.168.10.0/24",
    "10.0.0.0/8",
    "172.16.0.0/12"
  ]
  
  comment = "Complete enterprise storage target"
}
```

## Schema

### Required

- `name` (String) Target name (IQN will be generated).

### Optional

- `alias` (String) Target alias.
- `mode` (String) Target mode. Options: `ISCSI`, `FC`, `BOTH`. Default: `ISCSI`.
- `groups` (List of Number) List of portal group IDs.
- `auth_networks` (List of String) List of authorized networks (CIDR notation).

### Read-Only

- `id` (String) Target identifier.

## Import

iSCSI targets can be imported using target ID:

```shell
terraform import truenas_iscsi_target.existing 1
```

## Notes

### Target Configuration

#### Target Name and IQN
- Target name is used to generate the IQN (iSCSI Qualified Name)
- IQN format: `iqn.yyyy-mm.com.domain:name`
- TrueNAS automatically generates IQN based on system configuration
- Target name should be descriptive and unique

#### Target Modes

##### ISCSI (Default)
- Standard iSCSI target mode
- Supports iSCSI initiators only
- Most common deployment scenario

```terraform
mode = "ISCSI"
```

##### FC (Fibre Channel)
- Fibre Channel target mode
- Supports FC initiators only
- Requires FC hardware and configuration

```terraform
mode = "FC"
```

##### BOTH
- Supports both iSCSI and Fibre Channel
- Provides flexibility for mixed environments
- Requires both iSCSI and FC configuration

```terraform
mode = "BOTH"
```

### Portal Groups

Portal groups define which portals the target uses:

```terraform
groups = [1, 2, 3]  # Use portal groups 1, 2, and 3
```

- Portal groups must be created separately
- Multiple groups provide redundancy and load balancing
- Portal group IDs are numeric identifiers

### Network Access Control

Restrict access to specific networks:

```terraform
auth_networks = [
  "192.168.10.0/24",    # Specific subnet
  "10.0.0.0/8",          # Large network
  "172.16.0.0/12",        # Supernet
  "192.168.100.100/32"    # Single host
]
```

#### CIDR Notation
- `/24`: 255.255.255.0 mask (256 hosts)
- `/16`: 255.255.0.0 mask (65,536 hosts)
- `/8`: 255.0.0.0 mask (16,777,216 hosts)
- `/32`: Single host

#### Security Considerations
- Be specific with network restrictions
- Use smallest possible networks
- Regularly review and update access lists
- Consider using VPN for remote access

### Target Aliases

Aliases provide human-readable names:

```terraform
alias = "production-storage"
```

Benefits:
- Easier identification in management tools
- Consistent naming across environments
- Simplifies documentation and troubleshooting

### Complete Target Setup

A complete iSCSI setup typically includes:

1. **Extents**: Storage resources
2. **Portals**: Network endpoints
3. **Targets**: Storage endpoints
4. **Associations**: Link targets to extents and portals

```terraform
# Create extent
resource "truenas_iscsi_extent" "data" {
  name     = "data-extent"
  type     = "FILE"
  path     = "/mnt/tank/iscsi/data.img"
  filesize = 107374182400
}

# Create portal
resource "truenas_iscsi_portal" "portal" {
  listen = [
    {
      ip   = "0.0.0.0"
      port = 3260
    }
  ]
}

# Create target
resource "truenas_iscsi_target" "target" {
  name  = "storage-target"
  alias = "main-storage"
  groups = [truenas_iscsi_portal.portal.id]
  auth_networks = ["192.168.10.0/24"]
}
```

## Best Practices

### Planning

1. **Naming Convention**: Use consistent, descriptive names
2. **Network Design**: Plan network access carefully
3. **Security Strategy**: Implement proper access controls
4. **Documentation**: Document target configurations and purposes

### Security

1. **Network Restrictions**: Use specific CIDR blocks
2. **Access Control**: Regularly review authorized networks
3. **Monitoring**: Monitor target access and usage
4. **Authentication**: Implement proper iSCSI authentication

### Performance

1. **Portal Distribution**: Use multiple portals for load balancing
2. **Network Optimization**: Configure appropriate network settings
3. **Resource Planning**: Ensure adequate network bandwidth
4. **Monitoring**: Track performance metrics

### High Availability

1. **Multiple Portals**: Configure redundant network paths
2. **Network Diversity**: Use different network segments
3. **Failover Testing**: Test failover scenarios
4. **Monitoring**: Monitor portal and target health

### Maintenance

1. **Regular Reviews**: Periodically review configurations
2. **Backup Documentation**: Keep current configuration records
3. **Security Audits**: Regular security assessments
4. **Performance Tuning**: Adjust based on usage patterns

## Troubleshooting

### Target Not Accessible

1. Verify target is enabled
2. Check portal configurations
3. Test network connectivity
4. Review authentication settings

### Network Access Issues

1. Verify CIDR notation in auth_networks
2. Check network routing
3. Test from initiator networks
4. Review firewall rules

### Performance Problems

1. Monitor network bandwidth
2. Check portal utilization
3. Verify network configuration
4. Review initiator settings

### Portal Group Issues

1. Verify portal group IDs exist
2. Check portal group configurations
3. Test individual portal connectivity
4. Review portal group assignments

### Import Issues

1. Use correct target ID
2. Verify target exists in TrueNAS
3. Check target is not in use
4. Ensure proper permissions

### Mode Configuration

1. Verify hardware supports selected mode
2. Check FC configuration for FC/BOTH modes
3. Test connectivity for each protocol
4. Review mode-specific settings

## See Also

- [truenas_iscsi_extent](iscsi_extent) - iSCSI extent management
- [truenas_iscsi_portal](iscsi_portal) - iSCSI portal configuration
- [TrueNAS iSCSI Documentation](https://www.truenas.com/docs/scale/iscsi/) - Official iSCSI configuration guide
- [iSCSI Target Configuration](https://www.truenas.com/docs/scale/iscsi/targets/) - Target setup and management
- [iSCSI Security Best Practices](https://www.truenas.com/docs/scale/iscsi/security/) - Security recommendations