---
page_title: "truenas_nfs_shares Data Source - terraform-provider-truenas"
subcategory: "File Sharing"
description: |-
  Fetches information about all NFS shares on TrueNAS system.
---

# truenas_nfs_shares (Data Source)

Fetches information about all NFS shares on TrueNAS system. This data source returns detailed configuration for all NFS shares, including paths, permissions, and network access settings.

## Example Usage

### List All NFS Shares

```terraform
data "truenas_nfs_shares" "all" {
  output "nfs_share_count" {
    value = length(data.truenas_nfs_shares.all.shares)
  }
  
  output "nfs_share_paths" {
    value = [
      for share in data.truenas_nfs_shares.all.shares
      : share.path
    ]
  }
}
```

### Find Specific NFS Share

```terraform
data "truenas_nfs_shares" "shares" {
}

locals {
  # Find share by path
  data_share = {
    for share in data.truenas_nfs_shares.shares.shares
    : share
    if share.path == "/mnt/tank/data"
  }
  
  # Find enabled shares
  enabled_shares = [
    for share in data.truenas_nfs_shares.shares.shares
    : share
    if share.enabled
  ]
}

output "data_share_config" {
  value = local.data_share
}

output "enabled_share_count" {
  value = length(local.enabled_shares)
}
```

### NFS Share Analysis

```terraform
data "truenas_nfs_shares" "shares" {
}

locals {
  share_analysis = {
    total_shares = length(data.truenas_nfs_shares.shares.shares)
    enabled_shares = length([
      for share in data.truenas_nfs_shares.shares.shares
      : share
      if share.enabled
    ])
    readonly_shares = length([
      for share in data.truenas_nfs_shares.shares.shares
      : share
      if share.readonly
    ])
    shares_with_networks = length([
      for share in data.truenas_nfs_shares.shares.shares
      : share
      if length(share.networks) > 0
    ])
  }
}

output "nfs_analysis" {
  value = local.share_analysis
}
```

### Export NFS Share Information

```terraform
data "truenas_nfs_shares" "shares" {
}

locals {
  export_data = {
    for share in data.truenas_nfs_shares.shares.shares
    : share.path => {
      id           = share.id
      enabled      = share.enabled
      readonly     = share.readonly
      comment      = share.comment
      networks     = share.networks
      hosts        = share.hosts
      maproot_user = share.maproot_user
      mapall_user  = share.mapall_user
    }
  }
}

output "nfs_export" {
  value = local.export_data
}
```

### Validate NFS Configuration

```terraform
data "truenas_nfs_shares" "shares" {
}

locals {
  # Check for shares without network restrictions
  unrestricted_shares = [
    for share in data.truenas_nfs_shares.shares.shares
    : share
    if share.enabled && length(share.networks) == 0
  ]
  
  # Check for shares with root mapping
  root_mapped_shares = [
    for share in data.truenas_nfs_shares.shares.shares
    : share
    if share.enabled && share.maproot_user != null
  ]
}

output "security_issues" {
  value = {
    unrestricted_shares = [
      for share in local.unrestricted_shares
      : share.path
    ]
    root_mapped_shares = [
      for share in local.root_mapped_shares
      : {
        path = share.path
        maproot_user = share.maproot_user
      }
    ]
  }
}
```

### Conditional Resource Based on NFS Shares

```terraform
data "truenas_nfs_shares" "shares" {
}

locals {
  # Check if data share exists
  data_share_exists = anytrue([
    for share in data.truenas_nfs_shares.shares.shares
    : share.path == "/mnt/tank/data" && share.enabled
  ])
}

# Create backup job only if data share exists
resource "truenas_periodic_snapshot_task" "data_backup" {
  count = local.data_share_exists ? 1 : 0
  
  dataset        = "tank/data"
  recursive      = true
  enabled        = true
  naming_schema  = "backup-%Y-%m-%d_%H-%M"
  schedule       = "0 2 * * *"
  lifetime_value = 7
  lifetime_unit  = "DAY"
  
  depends_on = [data.truenas_nfs_shares.shares]
}
```

## Schema

### Read-Only

- `shares` (Block List) List of NFS shares. See [Share Attributes](#share-attributes) below.

### Share Attributes

Each share contains:

- `id` (Number) NFS share ID.
- `path` (String) Path to be exported.
- `comment` (String) Share comment/description.
- `enabled` (Boolean) Whether the share is enabled.
- `readonly` (Boolean) Whether the share is read-only.
- `maproot_user` (String) Map root user to this user.
- `maproot_group` (String) Map root group to this group.
- `mapall_user` (String) Map all users to this user.
- `mapall_group` (String) Map all groups to this group.
- `networks` (List of String) Allowed networks (CIDR notation).
- `hosts` (List of String) Allowed hosts.

## Notes

### NFS Share Structure

The data source returns a list of share objects with comprehensive configuration:

```json
{
  "shares": [
    {
      "id": 1,
      "path": "/mnt/tank/data",
      "comment": "Data share",
      "enabled": true,
      "readonly": false,
      "maproot_user": "root",
      "maproot_group": "wheel",
      "mapall_user": null,
      "mapall_group": null,
      "networks": ["192.168.1.0/24"],
      "hosts": ["server1.example.com"]
    }
  ]
}
```

### User and Group Mapping

#### Root Mapping
- `maproot_user`: Maps remote root user to specified local user
- `maproot_group`: Maps remote root group to specified local group

#### All Users Mapping
- `mapall_user`: Maps all remote users to specified local user
- `mapall_group`: Maps all remote groups to specified local group

#### Security Considerations
```terraform
# Secure configuration
maproot_user = "nobody"  # Don't map root to privileged user
mapall_user = "nobody"   # Map all users to unprivileged user

# Less secure (for trusted networks)
maproot_user = "root"    # Map root to root
mapall_user = null       # Don't remap users
```

### Network Access Control

#### CIDR Networks
```terraform
networks = [
  "192.168.1.0/24",    # Local network
  "10.0.0.0/8",        # Private networks
  "172.16.0.0/12",     # Private networks
  "203.0.113.5/32"     # Single host
]
```

#### Host-based Access
```terraform
hosts = [
  "server1.example.com",
  "server2.example.com",
  "backup-server.local"
]
```

#### Access Priority
1. Host-based access takes precedence over network-based
2. More specific networks override less specific
3. Order doesn't matter for evaluation

### Share States

#### Enabled/Disabled
```terraform
enabled = true   # Share is active and accessible
enabled = false  # Share is inactive
```

#### Read-only/Read-write
```terraform
readonly = true   # Clients can only read
readonly = false  # Clients can read and write
```

### Use Cases

#### Share Discovery
```terraform
data "truenas_nfs_shares" "shares" {
}

locals {
  # Find all data shares
  data_shares = [
    for share in data.truenas_nfs_shares.shares.shares
    : share
    if contains(share.path, "data") && share.enabled
  ]
}

output "data_shares" {
  value = local.data_shares
}
```

#### Configuration Validation
```terraform
data "truenas_nfs_shares" "shares" {
}

locals {
  # Validate security settings
  secure_shares = [
    for share in data.truenas_nfs_shares.shares.shares
    : share
    if share.enabled && 
       share.readonly == false && 
       length(share.networks) > 0 &&
       share.maproot_user != "root"
  ]
}

output "security_report" {
  value = {
    total_shares = length(data.truenas_nfs_shares.shares.shares)
    secure_shares = length(local.secure_shares)
    insecure_shares = length(data.truenas_nfs_shares.shares.shares) - length(local.secure_shares)
  }
}
```

#### Backup Planning
```terraform
data "truenas_nfs_shares" "shares" {
}

locals {
  # Find shares that need backup
  backup_candidates = [
    for share in data.truenas_nfs_shares.shares.shares
    : share
    if share.enabled && 
       share.readonly == false &&
       !contains(share.path, "temp") &&
       !contains(share.path, "cache")
  ]
}

# Create snapshot tasks for backup candidates
resource "truenas_periodic_snapshot_task" "share_backups" {
  for_each = {
    for share in local.backup_candidates
    : replace(share.path, "/mnt/", "") => share
  }
  
  dataset        = replace(each.value.path, "/mnt/", "")
  recursive      = true
  enabled        = true
  naming_schema  = "backup-%Y-%m-%d_%H-%M"
  schedule       = "0 3 * * *"
  lifetime_value = 7
  lifetime_unit  = "DAY"
  
  depends_on = [data.truenas_nfs_shares.shares]
}
```

## Best Practices

### Security

1. **Network Restrictions**: Always specify allowed networks
2. **User Mapping**: Use unprivileged users for root mapping
3. **Read-only Access**: Use read-only where appropriate
4. **Regular Audits**: Review share configurations regularly

### Performance

1. **Network Optimization**: Place shares on appropriate networks
2. **Access Patterns**: Design for expected usage patterns
3. **Storage Planning**: Ensure adequate storage performance
4. **Monitoring**: Monitor share performance and usage

### Management

1. **Consistent Naming**: Use descriptive share paths and comments
2. **Documentation**: Document share purposes and access
3. **Access Control**: Implement principle of least privilege
4. **Testing**: Test access from client systems

### Maintenance

1. **Regular Reviews**: Periodically review share configurations
2. **Access Audits**: Audit who has access to what
3. **Performance Monitoring**: Monitor share performance
4. **Backup Verification**: Ensure backup of shared data

## Troubleshooting

### Share Not Accessible

1. Verify share is enabled
2. Check network access rules
3. Test from client system
4. Review firewall configuration

### Permission Issues

1. Check user mapping configuration
2. Verify file system permissions
3. Test with different users
4. Review share security settings

### Performance Problems

1. Monitor network bandwidth
2. Check storage performance
3. Review client configuration
4. Test with different mount options

### Network Access Issues

1. Verify CIDR notation
2. Check network routing
3. Test from different networks
4. Review firewall rules

### Data Source Issues

1. Verify NFS service is running
2. Check provider configuration
3. Test with simple configuration
4. Review TrueNAS API access

## See Also

- [truenas_nfs_share](../resources/nfs_share) - NFS share management
- [truenas_smb_shares](smb_shares) - SMB/CIFS share discovery
- [TrueNAS NFS Documentation](https://www.truenas.com/docs/scale/nfs/) - Official NFS configuration guide
- [NFS Best Practices](https://www.truenas.com/docs/scale/nfs/bestpractices/) - NFS optimization and security