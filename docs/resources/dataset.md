---
page_title: "truenas_dataset Resource - terraform-provider-truenas"
subcategory: "Storage & File Sharing"
description: |-
  Manages a ZFS dataset on TrueNAS Scale.
---

# truenas_dataset (Resource)

Manages a ZFS dataset on TrueNAS Scale. Datasets are the fundamental unit of storage organization in ZFS, providing features like compression, quotas, and snapshots.

## Example Usage

### Basic Filesystem Dataset

```terraform
resource "truenas_dataset" "data" {
  name = "tank/data"
  type = "FILESYSTEM"
  
  comments = "General data storage"
}
```

### Dataset with Quota

```terraform
resource "truenas_dataset" "user_home" {
  name = "tank/home/john"
  type = "FILESYSTEM"
  
  quota {
    quota      = 107374182400  # 100GB
    quota_type = "DATASET"
  }
  
  refquota {
    refquota      = 53687091200  # 50GB
    refquota_type = "DATASET"
  }
  
  comments = "John's home directory"
}
```

### Dataset with Compression

```terraform
resource "truenas_dataset" "compressed" {
  name = "tank/backups"
  type = "FILESYSTEM"
  
  compression = "LZ4"
  atime       = "OFF"
  
  comments = "Backup storage with compression"
}
```

### Dataset with ACL Configuration

```terraform
resource "truenas_dataset" "shared" {
  name = "tank/shared"
  type = "FILESYSTEM"
  
  acltype = "NFSV4"
  aclmode = "PASSTHROUGH"
  
  share_type = "SMB"
  
  comments = "Shared dataset for SMB"
}
```

### Volume (Block Device) Dataset

```terraform
resource "truenas_dataset" "iscsi_volume" {
  name = "tank/iscsi/volume1"
  type = "VOLUME"
  
  volsize = 10737418240  # 10GB
  
  comments = "iSCSI block storage"
}
```

### Complete Dataset Configuration

```terraform
resource "truenas_dataset" "production" {
  name = "tank/production"
  type = "FILESYSTEM"
  
  # Quota configuration
  quota {
    quota      = 1099511627776  # 1TB
    quota_type = "DATASET"
  }
  
  refquota {
    refquota      = 536870912000  # 500GB
    refquota_type = "DATASET"
  }
  
  # Performance settings
  compression  = "LZ4"
  atime        = "OFF"
  recordsize   = "128K"
  deduplication = "OFF"
  
  # ACL settings
  acltype = "NFSV4"
  aclmode = "PASSTHROUGH"
  
  # Sharing
  share_type = "GENERIC"
  
  # Metadata
  comments = "Production application data"
}
```

## Schema

### Required

- `name` (String) Full path of the dataset (e.g., `tank/data` or `tank/home/user`). Must include the pool name.
- `type` (String) Dataset type. Options: `FILESYSTEM`, `VOLUME`

### Optional

- `comments` (String) Description or comments about the dataset.
- `sync` (String) Sync behavior. Options: `STANDARD`, `ALWAYS`, `DISABLED`. Default: `STANDARD`
- `compression` (String) Compression algorithm. Options: `OFF`, `LZ4`, `GZIP`, `ZSTD`, `ZLE`, `LZJB`. Default: `LZ4`
- `atime` (String) Access time updates. Options: `ON`, `OFF`. Default: `ON`
- `exec` (String) Execute permissions. Options: `ON`, `OFF`. Default: `ON`
- `readonly` (String) Read-only mode. Options: `ON`, `OFF`. Default: `OFF`
- `recordsize` (String) Record size. Options: `4K` to `1M`. Default: `128K`
- `acltype` (String) ACL type. Options: `NFSV4`, `POSIX`, `OFF`. Default: `INHERIT`
- `aclmode` (String) ACL mode. Options: `PASSTHROUGH`, `RESTRICTED`, `DISCARD`. Default: `DISCARD`
- `casesensitivity` (String) Case sensitivity. Options: `SENSITIVE`, `INSENSITIVE`, `MIXED`. Default: `SENSITIVE`
- `deduplication` (String) Deduplication. Options: `ON`, `OFF`, `VERIFY`. Default: `OFF`
- `share_type` (String) Share type hint. Options: `GENERIC`, `SMB`, `NFS`. Default: `GENERIC`
- `volsize` (Number) Volume size in bytes (required for VOLUME type datasets)
- `volblocksize` (String) Volume block size. Options: `4K` to `128K`. Default: `16K`
- `sparse` (Boolean) Create sparse volume. Default: false
- `quota` (Block) Quota configuration. See [Quota Configuration](#quota-configuration).
- `refquota` (Block) Reference quota configuration. See [Reference Quota Configuration](#reference-quota-configuration).

### Quota Configuration

The `quota` block supports:

- `quota` (Number) Quota size in bytes
- `quota_type` (String) Quota type. Options: `DATASET`, `USEROBJ`, `GROUPOBJ`, `PROJECTOBJ`, `USER`, `GROUP`, `PROJECT`

### Reference Quota Configuration

The `refquota` block supports:

- `refquota` (Number) Reference quota size in bytes  
- `refquota_type` (String) Reference quota type. Options: `DATASET`, `USEROBJ`, `GROUPOBJ`, `PROJECTOBJ`, `USER`, `GROUP`, `PROJECT`

### Read-Only

- `id` (String) The ID of the dataset (same as name).

## Import

Datasets can be imported using their full path:

```shell
terraform import truenas_dataset.example tank/data
```

## Notes

### Dataset Types

- **FILESYSTEM**: Standard file storage with directories and files
- **VOLUME**: Block device storage (zvol) for iSCSI, VM disks, etc.

### Quotas

- **quota**: Limits total space including snapshots
- **refquota**: Limits space used by the dataset itself (excluding snapshots)
- Both are specified in bytes
- Common conversions:
  - 1 GB = 1073741824 bytes
  - 1 TB = 1099511627776 bytes

### Compression

- `LZ4` is recommended for most use cases (fast, good ratio)
- `ZSTD` provides better compression but uses more CPU
- `GZIP` levels 1-9 available (`GZIP-1` through `GZIP-9`)
- Compression is transparent to applications

### Performance Tuning

- **atime=OFF**: Improves performance for read-heavy workloads
- **recordsize**: Match to your workload
  - Small files: 4K-32K
  - Large files/databases: 128K-1M
- **deduplication**: Only use if you have sufficient RAM (5GB per TB of deduplicated data)

### ACL Types

- **NFSV4**: Modern ACLs with rich permissions (recommended for SMB/NFS)
- **POSIX**: Traditional Unix

 permissions
- **OFF**: Disable ACLs

### Share Types

- Provides hints to TrueNAS for optimal configuration
- Does not automatically create shares (use `truenas_nfs_share` or `truenas_smb_share`)

### Nested Datasets

- Create parent datasets before children
- Example: Create `tank/data` before `tank/data/users`
- Use `depends_on` to enforce creation order

### Destroying Datasets

- Datasets with children cannot be destroyed
- Destroy child datasets first
- Snapshots must be deleted before destroying the dataset

## See Also

- [truenas_snapshot](snapshot) - Create dataset snapshots
- [truenas_periodic_snapshot_task](periodic_snapshot_task) - Automate snapshot creation
- [truenas_nfs_share](nfs_share) - Share dataset via NFS
- [truenas_smb_share](smb_share) - Share dataset via SMB
- [truenas_dataset Data Source](../data-sources/dataset) - Query dataset information