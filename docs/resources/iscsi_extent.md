---
page_title: "truenas_iscsi_extent Resource - terraform-provider-truenas"
subcategory: "iSCSI"
description: |-
  Manages an iSCSI extent (storage) on TrueNAS.
---

# truenas_iscsi_extent (Resource)

Manages an iSCSI extent (storage) on TrueNAS. An extent represents a storage resource that can be exported via iSCSI, either as a disk device or a file-based storage.

## Example Usage

### File-based Extent

```terraform
resource "truenas_iscsi_extent" "data_extent" {
  name     = "data-extent"
  type     = "FILE"
  path     = "/mnt/tank/iscsi/data.img"
  filesize = 107374182400  # 100GB in bytes
  
  comment  = "Data storage extent for VMs"
  enabled  = true
}
```

### Disk-based Extent

```terraform
resource "truenas_iscsi_extent" "disk_extent" {
  name    = "disk-extent"
  type    = "DISK"
  disk    = "ada1"
  
  comment = "Physical disk extent for backup storage"
  enabled = true
}
```

### Extent with Custom Block Size

```terraform
resource "truenas_iscsi_extent" "database_extent" {
  name      = "database-extent"
  type      = "FILE"
  path      = "/mnt/tank/iscsi/database.img"
  filesize  = 53687091200  # 50GB in bytes
  
  blocksize = 4096
  pblocksize = true
  
  comment = "Database storage with 4K blocks"
  enabled = true
}
```

### Read-only Extent

```terraform
resource "truenas_iscsi_extent" "iso_extent" {
  name     = "iso-library"
  type     = "FILE"
  path     = "/mnt/tank/iso/library.img"
  filesize = 21474836480  # 20GB in bytes
  
  readonly = true
  
  comment = "Read-only ISO library extent"
  enabled = true
}
```

### Extent with Availability Threshold

```terraform
resource "truenas_iscsi_extent" "critical_extent" {
  name            = "critical-data"
  type            = "FILE"
  path            = "/mnt/tank/iscsi/critical.img"
  filesize        = 107374182400  # 100GB in bytes
  
  avail_threshold = 10  # Alert when less than 10% available
  
  comment = "Critical data extent with monitoring"
  enabled = true
}
```

### Extent with SSD Optimization

```terraform
resource "truenas_iscsi_extent" "ssd_extent" {
  name     = "ssd-storage"
  type     = "FILE"
  path     = "/mnt/ssd-pool/iscsi/fast.img"
  filesize = 21474836480  # 20GB in bytes
  
  rpm      = "SSD"
  serial   = "SSD-EXTENT-001"
  
  comment = "High-performance SSD extent"
  enabled = true
}
```

### Extent for Xen Compatibility

```terraform
resource "truenas_iscsi_extent" "xen_extent" {
  name     = "xen-storage"
  type     = "FILE"
  path     = "/mnt/tank/iscsi/xen.img"
  filesize = 42949672960  # 40GB in bytes
  
  xen      = true
  
  comment = "Xen-compatible storage extent"
  enabled = true
}
```

### Extent with Insecure TPC

```terraform
resource "truenas_iscsi_extent" "backup_extent" {
  name         = "backup-storage"
  type         = "FILE"
  path         = "/mnt/tank/iscsi/backup.img"
  filesize     = 1073741824000  # 1TB in bytes
  
  insecure_tpc = true  # Allow third-party copy
  
  comment = "Backup storage with TPC support"
  enabled = true
}
```

## Schema

### Required

- `name` (String) Extent name.
- `type` (String) Extent type. Options: `DISK`, `FILE`.

### Optional

- `disk` (String) Disk device (for DISK type).
- `path` (String) File path (for FILE type).
- `filesize` (Number) File size in bytes (for FILE type).
- `comment` (String) Comment.
- `enabled` (Boolean) Enable extent. Default: true.
- `readonly` (Boolean) Read-only extent. Default: false.
- `blocksize` (Number) Block size. Options: 512, 1024, 2048, 4096. Default: 512.
- `pblocksize` (Boolean) Use physical block size. Default: false.
- `avail_threshold` (Number) Available space threshold percentage. Default: 0.
- `serial` (String) Serial number.
- `rpm` (String) RPM. Options: `SSD`, `5400`, `7200`, `10000`, `15000`.
- `xen` (Boolean) Xen compatibility mode. Default: false.
- `insecure_tpc` (Boolean) Allow insecure third-party copy. Default: false.

### Read-Only

- `id` (String) Extent identifier.

## Import

iSCSI extents can be imported using the extent ID:

```shell
terraform import truenas_iscsi_extent.existing 1
```

## Notes

### Extent Types

#### FILE Type
- Creates a file-based storage extent
- Requires `path` and `filesize` parameters
- File is created and managed by TrueNAS
- Suitable for VM disks and flexible storage

#### DISK Type
- Uses entire physical disk as extent
- Requires `disk` parameter (device name)
- Disk must not be in use by other services
- Suitable for dedicated storage devices

### Block Size Configuration

Block size affects performance and compatibility:

```terraform
blocksize = 512    # Standard block size (default)
blocksize = 1024   # 1K blocks
blocksize = 2048   # 2K blocks  
blocksize = 4096   # 4K blocks (recommended for modern systems)
```

- **512 bytes**: Maximum compatibility
- **4096 bytes**: Better performance for large files
- Use `pblocksize = true` to detect physical block size

### File Size Calculation

File size must be specified in bytes:

```terraform
# Common sizes
filesize = 10737418240   # 10GB
filesize = 107374182400  # 100GB
filesize = 1099511627776 # 1TB

# Calculate: desired_size_gb * 1024^3
# Example: 500GB * 1024^3 = 536870912000 bytes
```

### Performance Optimization

#### SSD Configuration
```terraform
rpm = "SSD"
blocksize = 4096
pblocksize = true
```

#### HDD Configuration
```terraform
rpm = "7200"  # or 5400, 10000, 15000
blocksize = 4096
```

### Availability Monitoring

Set thresholds for space monitoring:

```terraform
avail_threshold = 10  # Alert at 10% free space
avail_threshold = 5   # Alert at 5% free space (critical)
```

### Compatibility Options

#### Xen Compatibility
```terraform
xen = true  # Enable Xen-specific optimizations
```

#### Third-Party Copy
```terraform
insecure_tpc = true  # Allow third-party copy operations
```

### Security Considerations

#### Read-only Extents
```terraform
readonly = true  # Prevent modifications
```

Use for:
- ISO libraries
- Template storage
- Backup archives

## Best Practices

### Planning

1. **Size Planning**: Plan for growth and allocate sufficient space
2. **Performance**: Match block size to workload requirements
3. **Backup Strategy**: Implement regular backups of extent data
4. **Monitoring**: Set appropriate availability thresholds

### Performance

1. **Block Size**: Use 4K blocks for modern workloads
2. **SSD Optimization**: Enable SSD-specific settings
3. **Physical Block Size**: Use `pblocksize = true` for optimal performance
4. **Storage Pool**: Place extents on appropriate storage pools

### Security

1. **Access Control**: Use read-only extents where appropriate
2. **Network Security**: Secure iSCSI network access
3. **Authentication**: Implement proper iSCSI authentication
4. **Monitoring**: Monitor extent access and usage

### Maintenance

1. **Regular Checks**: Monitor extent health and availability
2. **Capacity Planning**: Track usage trends and plan expansion
3. **Backup Verification**: Regularly test backup procedures
4. **Performance Tuning**: Adjust settings based on usage patterns

## Troubleshooting

### Extent Creation Fails

1. Check sufficient disk space
2. Verify path permissions
3. Ensure storage pool is online
4. Check for conflicting extents

### Performance Issues

1. Verify block size configuration
2. Check storage pool performance
3. Monitor network bandwidth
4. Review disk health status

### Access Problems

1. Verify iSCSI service is running
2. Check network connectivity
3. Review authentication settings
4. Ensure extent is enabled

### Space Issues

1. Monitor availability thresholds
2. Check for unused extents
3. Review storage pool capacity
4. Implement cleanup procedures

### Import Issues

1. Use correct extent ID
2. Verify extent exists in TrueNAS
3. Check extent is not in use
4. Ensure proper permissions

## See Also

- [truenas_iscsi_portal](iscsi_portal) - iSCSI portal management
- [truenas_iscsi_target](iscsi_target) - iSCSI target configuration
- [TrueNAS iSCSI Documentation](https://www.truenas.com/docs/scale/iscsi/) - Official iSCSI configuration guide
- [iSCSI Best Practices](https://www.truenas.com/docs/scale/iscsi/bestpractices/) - Performance and security recommendations