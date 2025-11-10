---
page_title: "truenas_smb_shares Data Source - terraform-provider-truenas"
subcategory: "File Sharing"
description: |-
  Fetches information about all SMB/CIFS shares on TrueNAS system.
---

# truenas_smb_shares (Data Source)

Fetches information about all SMB/CIFS shares on TrueNAS system. This data source returns detailed configuration for all SMB shares, including paths, permissions, and special features like Time Machine support.

## Example Usage

### List All SMB Shares

```terraform
data "truenas_smb_shares" "all" {
  output "smb_share_count" {
    value = length(data.truenas_smb_shares.all.shares)
  }
  
  output "smb_share_names" {
    value = [
      for share in data.truenas_smb_shares.all.shares
      : share.name
    ]
  }
}
```

### Find Specific SMB Share

```terraform
data "truenas_smb_shares" "shares" {
}

locals {
  # Find share by name
  media_share = {
    for share in data.truenas_smb_shares.shares.shares
    : share
    if share.name == "media"
  }
  
  # Find Time Machine enabled shares
  timemachine_shares = [
    for share in data.truenas_smb_shares.shares.shares
    : share
    if share.timemachine
  ]
}

output "media_share_config" {
  value = local.media_share
}

output "timemachine_share_count" {
  value = length(local.timemachine_shares)
}
```

### SMB Share Analysis

```terraform
data "truenas_smb_shares" "shares" {
}

locals {
  share_analysis = {
    total_shares = length(data.truenas_smb_shares.shares.shares)
    enabled_shares = length([
      for share in data.truenas_smb_shares.shares.shares
      : share
      if share.enabled
    ])
    readonly_shares = length([
      for share in data.truenas_smb_shares.shares.shares
      : share
      if share.readonly
    ])
    guest_accessible = length([
      for share in data.truenas_smb_shares.shares.shares
      : share
      if share.guestok
    ])
    timemachine_enabled = length([
      for share in data.truenas_smb_shares.shares.shares
      : share
      if share.timemachine
    ])
    recyclebin_enabled = length([
      for share in data.truenas_smb_shares.shares.shares
      : share
      if share.recyclebin
    ])
  }
}

output "smb_analysis" {
  value = local.share_analysis
}
```

### Export SMB Share Information

```terraform
data "truenas_smb_shares" "shares" {
}

locals {
  export_data = {
    for share in data.truenas_smb_shares.shares.shares
    : share.name => {
      id          = share.id
      path        = share.path
      enabled     = share.enabled
      readonly    = share.readonly
      browsable   = share.browsable
      guestok     = share.guestok
      recyclebin  = share.recyclebin
      timemachine = share.timemachine
      purpose     = share.purpose
      home        = share.home
    }
  }
}

output "smb_export" {
  value = local.export_data
}
```

### Validate SMB Configuration

```terraform
data "truenas_smb_shares" "shares" {
}

locals {
  # Check for shares with guest access
  guest_accessible_shares = [
    for share in data.truenas_smb_shares.shares.shares
    : share
    if share.enabled && share.guestok
  ]
  
  # Check for shares without recycle bin
  no_recyclebin_shares = [
    for share in data.truenas_smb_shares.shares.shares
    : share
    if share.enabled && !share.recyclebin && !share.readonly
  ]
  
  # Check for non-browsable shares
  hidden_shares = [
    for share in data.truenas_smb_shares.shares.shares
    : share
    if share.enabled && !share.browsable
  ]
}

output "security_analysis" {
  value = {
    guest_accessible = [
      for share in local.guest_accessible_shares
      : share.name
    ]
    no_recyclebin = [
      for share in local.no_recyclebin_shares
      : share.name
    ]
    hidden_shares = [
      for share in local.hidden_shares
      : share.name
    ]
  }
}
```

### Time Machine Share Discovery

```terraform
data "truenas_smb_shares" "shares" {
}

locals {
  timemachine_shares = [
    for share in data.truenas_smb_shares.shares.shares
    : share
    if share.timemachine && share.enabled
  ]
}

output "timemachine_shares" {
  value = [
    for share in local.timemachine_shares
    : {
      name = share.name
      path = share.path
      purpose = share.purpose
    }
  ]
}

# Create backup for Time Machine shares
resource "truenas_periodic_snapshot_task" "timemachine_backup" {
  for_each = {
    for share in local.timemachine_shares
    : share.name => share
  }
  
  dataset        = replace(each.value.path, "/mnt/", "")
  recursive      = true
  enabled        = true
  naming_schema  = "tm-backup-%Y-%m-%d_%H-%M"
  schedule       = "0 */6 * * *"
  lifetime_value = 14
  lifetime_unit  = "DAY"
  
  depends_on = [data.truenas_smb_shares.shares]
}
```

### Home Share Detection

```terraform
data "truenas_smb_shares" "shares" {
}

locals {
  home_shares = [
    for share in data.truenas_smb_shares.shares.shares
    : share
    if share.home && share.enabled
  ]
}

output "home_shares" {
  value = [
    for share in local.home_shares
    : {
      name = share.name
      path = share.path
      browsable = share.browsable
    }
  ]
}
```

## Schema

### Read-Only

- `shares` (Block List) List of SMB shares. See [Share Attributes](#share-attributes) below.

### Share Attributes

Each share contains:

- `id` (Number) SMB share ID.
- `path` (String) Path to be shared.
- `name` (String) Share name.
- `comment` (String) Share comment/description.
- `enabled` (Boolean) Whether the share is enabled.
- `readonly` (Boolean) Whether the share is read-only.
- `browsable` (Boolean) Whether the share is browsable.
- `guestok` (Boolean) Whether guest access is allowed.
- `recyclebin` (Boolean) Whether recycle bin is enabled.
- `purpose` (String) Share purpose (e.g., DEFAULT_SHARE, ENHANCED_TIMEMACHINE).
- `home` (Boolean) Whether this is a home share.
- `timemachine` (Boolean) Whether Time Machine support is enabled.

## Notes

### SMB Share Structure

The data source returns a list of share objects with comprehensive configuration:

```json
{
  "shares": [
    {
      "id": 1,
      "path": "/mnt/tank/data",
      "name": "data",
      "comment": "Data share",
      "enabled": true,
      "readonly": false,
      "browsable": true,
      "guestok": false,
      "recyclebin": true,
      "purpose": "DEFAULT_SHARE",
      "home": false,
      "timemachine": false
    }
  ]
}
```

### Share Types and Purposes

#### Standard Shares
```terraform
purpose = "DEFAULT_SHARE"  # Standard SMB share
```

#### Time Machine Shares
```terraform
purpose = "ENHANCED_TIMEMACHINE"  # macOS Time Machine support
timemachine = true
```

#### Home Shares
```terraform
home = true  # User home directories
```

### Access Control

#### Guest Access
```terraform
guestok = true   # Allow unauthenticated access
guestok = false  # Require authentication
```

#### Read-only Access
```terraform
readonly = true   # Read-only access
readonly = false  # Read-write access
```

#### Browse Visibility
```terraform
browsable = true   # Visible in network browse list
browsable = false  # Hidden share (prefix with $)
```

### Special Features

#### Recycle Bin
```terraform
recyclebin = true   # Enable deleted file recovery
recyclebin = false  # Permanent deletion
```

#### Time Machine Support
```terraform
timemachine = true   # Enable macOS Time Machine
timemachine = false  # Standard SMB share
```

### Use Cases

#### Share Discovery
```terraform
data "truenas_smb_shares" "shares" {
}

locals {
  # Find all data shares
  data_shares = [
    for share in data.truenas_smb_shares.shares.shares
    : share
    if contains(share.name, "data") && share.enabled
  ]
  
  # Find all public shares
  public_shares = [
    for share in data.truenas_smb_shares.shares.shares
    : share
    if share.enabled && share.guestok
  ]
}

output "share_summary" {
  value = {
    data_shares = length(local.data_shares)
    public_shares = length(local.public_shares)
    total_shares = length(data.truenas_smb_shares.shares.shares)
  }
}
```

#### Security Analysis
```terraform
data "truenas_smb_shares" "shares" {
}

locals {
  security_analysis = {
    # Shares with potential security issues
    guest_accessible = [
      for share in data.truenas_smb_shares.shares.shares
      : share.name
      if share.enabled && share.guestok && !share.readonly
    ]
    
    # Shares without protection
    no_recyclebin = [
      for share in data.truenas_smb_shares.shares.shares
      : share.name
      if share.enabled && !share.recyclebin && !share.readonly
    ]
    
    # Hidden shares (potential security)
    hidden_shares = [
      for share in data.truenas_smb_shares.shares.shares
      : share.name
      if share.enabled && !share.browsable
    ]
  }
}

output "security_report" {
  value = local.security_analysis
}
```

#### Backup Planning
```terraform
data "truenas_smb_shares" "shares" {
}

locals {
  # Find shares that need backup
  backup_candidates = [
    for share in data.truenas_smb_shares.shares.shares
    : share
    if share.enabled && 
       share.readonly == false &&
       !contains(share.name, "temp") &&
       !contains(share.name, "cache")
  ]
}

# Create snapshot tasks for backup candidates
resource "truenas_periodic_snapshot_task" "smb_backups" {
  for_each = {
    for share in local.backup_candidates
    : share.name => share
  }
  
  dataset        = replace(each.value.path, "/mnt/", "")
  recursive      = true
  enabled        = true
  naming_schema  = "smb-backup-%Y-%m-%d_%H-%M"
  schedule       = "0 2 * * *"
  lifetime_value = 7
  lifetime_unit  = "DAY"
  
  depends_on = [data.truenas_smb_shares.shares]
}
```

#### macOS Integration
```terraform
data "truenas_smb_shares" "shares" {
}

locals {
  macos_shares = [
    for share in data.truenas_smb_shares.shares.shares
    : share
    if share.enabled && (share.timemachine || share.home)
  ]
}

output "macos_integration" {
  value = {
    timemachine_shares = [
      for share in local.macos_shares
      : share.name
      if share.timemachine
    ]
    home_shares = [
      for share in local.macos_shares
      : share.name
      if share.home
    ]
  }
}
```

## Best Practices

### Security

1. **Authentication**: Disable guest access for sensitive data
2. **Access Control**: Use appropriate permissions at filesystem level
3. **Network Isolation**: Place shares on appropriate networks
4. **Regular Audits**: Review share configurations regularly

### Performance

1. **Recycle Bin**: Enable for user data, disable for high-performance needs
2. **Browse Lists**: Disable for large numbers of shares
3. **Network Optimization**: Configure appropriate SMB settings
4. **Storage Planning**: Ensure adequate storage performance

### User Experience

1. **Descriptive Names**: Use clear, descriptive share names
2. **Comments**: Add helpful comments for share purposes
3. **Browse Organization**: Group related shares
4. **Home Directories**: Use home shares for user data

### macOS Integration

1. **Time Machine**: Use dedicated shares for Time Machine
2. **Home Shares**: Configure properly for macOS home directories
3. **Permissions**: Ensure proper macOS compatibility
4. **Performance**: Optimize for macOS workloads

## Troubleshooting

### Share Not Accessible

1. Verify share is enabled
2. Check SMB service status
3. Test from client system
4. Review network connectivity

### Permission Issues

1. Check filesystem permissions
2. Verify user authentication
3. Review share access settings
4. Test with different users

### Time Machine Issues

1. Verify Time Machine share configuration
2. Check macOS compatibility settings
3. Review share permissions
4. Test Time Machine connection

### Performance Problems

1. Monitor network bandwidth
2. Check storage performance
3. Review SMB configuration
4. Test with different clients

### Data Source Issues

1. Verify SMB service is running
2. Check provider configuration
3. Test with simple configuration
4. Review TrueNAS API access

## See Also

- [truenas_smb_share](../resources/smb_share) - SMB share management
- [truenas_nfs_shares](nfs_shares) - NFS share discovery
- [TrueNAS SMB Documentation](https://www.truenas.com/docs/scale/smb/) - Official SMB configuration guide
- [SMB Best Practices](https://www.truenas.com/docs/scale/smb/bestpractices/) - SMB optimization and security