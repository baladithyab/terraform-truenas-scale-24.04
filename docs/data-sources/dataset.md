---
page_title: "truenas_dataset Data Source - terraform-provider-truenas"
subcategory: "Storage & File Sharing"
description: |-
  Fetches information about a ZFS dataset on TrueNAS Scale.
---

# truenas_dataset (Data Source)

Fetches information about a ZFS dataset on TrueNAS Scale. This data source can be used to query dataset properties including type, compression, and usage statistics.

## Example Usage

### Query Dataset by ID

```terraform
data "truenas_dataset" "data" {
  id = "tank/data"
}

output "dataset_info" {
  value = {
    name        = data.truenas_dataset.data.name
    type        = data.truenas_dataset.data.type
    pool        = data.truenas_dataset.data.pool
    compression = data.truenas_dataset.data.compression
  }
}
```

### Check Dataset Usage

```terraform
data "truenas_dataset" "usage" {
  id = "tank/backups"
}

locals {
  used_gb = data.truenas_dataset.usage.used / 1024 / 1024 / 1024
  available_gb = data.truenas_dataset.usage.available / 1024 / 1024 / 1024
}

output "dataset_usage" {
  value = {
    used_gb      = local.used_gb
    available_gb = local.available_gb
  }
}
```

### Conditional Resource Based on Dataset Properties

```terraform
data "truenas_dataset" "source" {
  id = "tank/source-data"
}

# Only create backup if source dataset uses compression
resource "truenas_dataset" "backup" {
  count = data.truenas_dataset.source.compression != "OFF" ? 1 : 0
  
  name = "tank/backup-data"
  type = "FILESYSTEM"
  
  compression = data.truenas_dataset.source.compression
}
```

### Monitor Multiple Datasets

```terraform
data "truenas_dataset" "datasets" {
  for_each = toset([
    "tank/data",
    "tank/backups",
    "tank/media"
  ])
  id = each.key
}

output "dataset_summary" {
  value = {
    for name, dataset in data.truenas_dataset.datasets :
    name => {
      type        = dataset.type
      compression = dataset.compression
      used_gb     = dataset.used / 1024 / 1024 / 1024
      available_gb = dataset.available / 1024 / 1024 / 1024
    }
  }
}
```

### Create Share Based on Dataset Type

```terraform
data "truenas_dataset" "shared_data" {
  id = "tank/shared"
}

# Create NFS share for filesystem datasets
resource "truenas_nfs_share" "nfs_share" {
  count = data.truenas_dataset.shared_data.type == "FILESYSTEM" ? 1 : 0
  
  path = "/mnt/${data.truenas_dataset.shared_data.name}"
  comment = "NFS share for ${data.truenas_dataset.shared_data.name}"
}

# Create iSCSI extent for volume datasets
resource "truenas_iscsi_extent" "iscsi_extent" {
  count = data.truenas_dataset.shared_data.type == "VOLUME" ? 1 : 0
  
  name = "iscsi-${data.truenas_dataset.shared_data.name}"
  path = "/dev/zvol/${data.truenas_dataset.shared_data.name}"
}
```

### Validate Dataset Configuration

```terraform
data "truenas_dataset" "production" {
  id = "tank/production"
}

locals {
  # Validate production dataset has required settings
  is_valid = (
    data.truenas_dataset.production.compression == "LZ4" &&
    data.truenas_dataset.production.type == "FILESYSTEM"
  )
}

resource "null_resource" "validation" {
  count = local.is_valid ? 0 : 1
  
  provisioner "local-exec" {
    command = <<-EOT
      echo "Error: Production dataset validation failed"
      echo "Expected compression: LZ4, Got: ${data.truenas_dataset.production.compression}"
      echo "Expected type: FILESYSTEM, Got: ${data.truenas_dataset.production.type}"
      exit 1
    EOT
  }
}
```

## Schema

### Required

- `id` (String) Dataset identifier (full path including pool name, e.g., "tank/data").

### Read-Only

- `name` (String) Full path of the dataset.
- `type` (String) Dataset type (FILESYSTEM or VOLUME).
- `pool` (String) Pool name.
- `compression` (String) Compression algorithm (OFF, LZ4, GZIP, ZSTD, etc.).
- `available` (Number) Available space in bytes.
- `used` (Number) Used space in bytes.

## Notes

### Dataset Identification

The dataset ID must be the full path including the pool name:
- Correct: `"tank/data"`, `"tank/home/user"`, `"pool/volumes/vol1"`
- Incorrect: `"data"`, `"user"`, `"vol1"`

### Dataset Types

- **FILESYSTEM**: Standard file storage with directories and files
- **VOLUME**: Block device storage (zvol) for iSCSI, VM disks, etc.

### Compression Algorithms

Common compression settings:
- **OFF**: No compression
- **LZ4**: Fast compression, good ratio (recommended)
- **ZSTD**: Better compression, higher CPU usage
- **GZIP**: Various levels (GZIP-1 through GZIP-9)
- **LZJB**: Legacy compression algorithm

### Usage Calculations

The data source provides raw byte values. Common conversions:
```terraform
locals {
  # Convert to GB
  used_gb = data.truenas_dataset.example.used / 1024 / 1024 / 1024
  available_gb = data.truenas_dataset.example.available / 1024 / 1024 / 1024
  
  # Calculate total size
  total_bytes = data.truenas_dataset.example.used + data.truenas_dataset.example.available
  total_gb = total_bytes / 1024 / 1024 / 1024
  
  # Calculate usage percentage
  used_percent = data.truenas_dataset.example.used / total_bytes * 100
}
```

### Use Cases

#### Dataset Discovery
```terraform
data "truenas_dataset" "existing" {
  id = "tank/existing-data"
}

# Use existing dataset information for new resources
resource "truenas_nfs_share" "share_existing" {
  path = "/mnt/${data.truenas_dataset.existing.name}"
  comment = "Share for ${data.truenas_dataset.existing.name}"
}
```

#### Capacity Planning
```terraform
data "truenas_dataset" "growth_monitor" {
  id = "tank/application"
}

locals {
  current_usage = data.truenas_dataset.growth_monitor.used
  available_space = data.truenas_dataset.growth_monitor.available
  
  # Alert if less than 20% free
  low_space = available_space < (current_usage * 0.25)
}

resource "null_resource" "capacity_alert" {
  count = local.low_space ? 1 : 0
  
  provisioner "local-exec" {
    command = "echo 'Warning: Dataset ${data.truenas_dataset.growth_monitor.name} running low on space!'"
  }
}
```

#### Configuration Validation
```terraform
data "truenas_dataset" "validate" {
  id = "tank/important-data"
}

locals {
  # Ensure dataset meets requirements
  requirements_met = (
    data.truenas_dataset.validate.compression != "OFF" &&
    data.truenas_dataset.validate.pool == "tank"
  )
}

# Only proceed if requirements are met
resource "truenas_snapshot" "backup" {
  count = local.requirements_met ? 1 : 0
  
  dataset = data.truenas_dataset.validate.name
  name = "auto-backup"
}
```

### Integration with Other Resources

The dataset data source is commonly used with:

```terraform
data "truenas_dataset" "source" {
  id = "tank/source"
}

# Create snapshots
resource "truenas_snapshot" "daily" {
  dataset = data.truenas_dataset.source.name
  name = "daily-${formatdate("YYYY-MM-DD", timestamp())}"
}

# Create shares
resource "truenas_nfs_share" "share" {
  path = "/mnt/${data.truenas_dataset.source.name}"
  comment = "Share for ${data.truenas_dataset.source.name}"
}

# Create periodic snapshot tasks
resource "truenas_periodic_snapshot_task" "auto" {
  dataset = data.truenas_dataset.source.name
  schedule = "0 2 * * *"  # Daily at 2 AM
}
```

### Best Practices

1. **Use full dataset paths** including pool name
2. **Monitor usage** to prevent running out of space
3. **Validate compression** settings match requirements
4. **Check dataset type** before creating shares
5. **Use for discovery** of existing datasets

### Common Patterns

#### Dataset Inventory
```terraform
locals {
  dataset_paths = [
    "tank/data",
    "tank/backups",
    "tank/media",
    "tank/volumes/vm-disks"
  ]
}

data "truenas_dataset" "inventory" {
  for_each = toset(local.dataset_paths)
  id = each.key
}

output "dataset_inventory" {
  value = {
    for path, dataset in data.truenas_dataset.inventory :
    path => {
      type = dataset.type
      pool = dataset.pool
      compression = dataset.compression
      used_gb = dataset.used / 1024 / 1024 / 1024
    }
  }
}
```

#### Compression Audit
```terraform
data "truenas_dataset" "audit" {
  for_each = toset([
    "tank/data",
    "tank/backups",
    "tank/logs"
  ])
  id = each.key
}

locals {
  uncompressed = [
    for path, dataset in data.truenas_dataset.audit :
    path
    if dataset.compression == "OFF"
  ]
}

output "uncompressed_datasets" {
  value = local.uncompressed
}
```

### Troubleshooting

**Dataset Not Found:**
- Verify full path including pool name
- Check dataset exists in TrueNAS web interface
- Ensure correct spelling and case

**Incorrect Usage Data:**
- Dataset statistics may have a delay
- Run `zfs list` on TrueNAS to verify
- Check for ongoing operations

**Type Mismatch:**
- Verify dataset type before creating shares
- Use `truenas_dataset` resource to change type if needed
- Remember VOLUME datasets are for block storage

## See Also

- [truenas_dataset](../resources/dataset) - Manage datasets
- [truenas_snapshot](../resources/snapshot) - Create dataset snapshots
- [truenas_nfs_share](../resources/nfs_share) - Share filesystem datasets
- [truenas_iscsi_extent](../resources/iscsi_extent) - Use volume datasets for iSCSI
- [ZFS Dataset Management](https://www.truenas.com/docs/core/storage/datasets/) - TrueNAS documentation