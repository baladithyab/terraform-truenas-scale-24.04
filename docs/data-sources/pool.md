---
page_title: "truenas_pool Data Source - terraform-provider-truenas"
subcategory: "Storage & File Sharing"
description: |-
  Fetches information about a ZFS pool on TrueNAS Scale.
---

# truenas_pool (Data Source)

Fetches information about a ZFS pool on TrueNAS Scale. This data source can be used to query pool details including health status, capacity, and mount path.

## Example Usage

### Query Pool by Name

```terraform
data "truenas_pool" "main" {
  id = "tank"
}

output "pool_status" {
  value = data.truenas_pool.main.status
}

output "pool_healthy" {
  value = data.truenas_pool.main.healthy
}

output "available_space" {
  value = data.truenas_pool.main.available
}
```

### Query Pool by ID

```terraform
data "truenas_pool" "pool_by_id" {
  id = "1"
}

output "pool_info" {
  value = {
    name      = data.truenas_pool.pool_by_id.name
    size      = data.truenas_pool.pool_by_id.size
    available = data.truenas_pool.pool_by_id.available
    path      = data.truenas_pool.pool_by_id.path
  }
}
```

### Use Pool Information for Dataset Creation

```terraform
data "truenas_pool" "storage" {
  id = "tank"
}

# Only create dataset if pool is healthy
resource "truenas_dataset" "data" {
  count = data.truenas_pool.storage.healthy ? 1 : 0
  
  name = "${data.truenas_pool.storage.name}/data"
  type = "FILESYSTEM"
  
  comments = "Data storage on ${data.truenas_pool.storage.name} pool"
}
```

### Monitor Pool Capacity

```terraform
data "truenas_pool" "monitor" {
  id = "tank"
}

locals {
  pool_size_gb = data.truenas_pool.monitor.size / 1024 / 1024 / 1024
  available_gb = data.truenas_pool.monitor.available / 1024 / 1024 / 1024
  used_percent = (data.truenas_pool.monitor.size - data.truenas_pool.monitor.available) / data.truenas_pool.monitor.size * 100
}

output "pool_capacity" {
  value = {
    total_size_gb = local.pool_size_gb
    available_gb  = local.available_gb
    used_percent  = local.used_percent
  }
}

# Alert if pool is over 80% full
resource "null_resource" "capacity_alert" {
  count = local.used_percent > 80 ? 1 : 0
  
  provisioner "local-exec" {
    command = "echo 'Warning: Pool ${data.truenas_pool.monitor.name} is ${local.used_percent}% full!'"
  }
}
```

### Conditional Resource Creation Based on Pool

```terraform
data "truenas_pool" "primary" {
  id = "tank"
}

data "truenas_pool" "backup" {
  id = "backup"
}

# Create datasets on healthy pools only
resource "truenas_dataset" "primary_data" {
  count = data.truenas_pool.primary.healthy ? 1 : 0
  
  name = "${data.truenas_pool.primary.name}/primary-data"
  type = "FILESYSTEM"
}

resource "truenas_dataset" "backup_data" {
  count = data.truenas_pool.backup.healthy ? 1 : 0
  
  name = "${data.truenas_pool.backup.name}/backup-data"
  type = "FILESYSTEM"
}
```

## Schema

### Required

- `id` (String) Pool identifier. Can be either the pool name (e.g., "tank") or numeric pool ID.

### Read-Only

- `name` (String) Pool name.
- `status` (String) Pool status (e.g., "ONLINE", "DEGRADED", "FAULTED").
- `healthy` (Boolean) Whether the pool is healthy.
- `path` (String) Pool mount path.
- `available` (Number) Available space in bytes.
- `size` (Number) Total pool size in bytes.

## Notes

### Pool Identification

The data source accepts either:
- Pool name (e.g., "tank", "backup")
- Numeric pool ID (e.g., "1", "2")

Using the pool name is recommended as it's more readable and stable.

### Pool Status Values

Common pool statuses:
- **ONLINE**: Pool is healthy and operational
- **DEGRADED**: Pool has redundancy issues but is still functional
- **FAULTED**: Pool has serious issues and may be inaccessible
- **OFFLINE**: Pool is not currently mounted

### Health Monitoring

The `healthy` attribute provides a simple boolean check:
- `true`: Pool is in good health (typically ONLINE status)
- `false`: Pool has issues (DEGRADED, FAULTED, etc.)

### Capacity Calculations

The data source provides raw byte values. Common conversions:
```terraform
locals {
  # Convert to GB
  size_gb = data.truenas_pool.example.size / 1024 / 1024 / 1024
  available_gb = data.truenas_pool.example.available / 1024 / 1024 / 1024
  
  # Calculate used space
  used_bytes = data.truenas_pool.example.size - data.truenas_pool.example.available
  used_gb = used_bytes / 1024 / 1024 / 1024
  
  # Calculate percentage used
  used_percent = used_bytes / data.truenas_pool.example.size * 100
}
```

### Use Cases

#### Pre-flight Checks
```terraform
data "truenas_pool" "check" {
  id = "tank"
}

# Only proceed if pool is healthy and has sufficient space
resource "truenas_dataset" "large_dataset" {
  count = data.truenas_pool.check.healthy && data.truenas_pool.check.available > 1099511627776 ? 1 : 0
  
  name = "${data.truenas_pool.check.name}/large-dataset"
  type = "FILESYSTEM"
}
```

#### Multi-Pool Selection
```terraform
data "truenas_pool" "pool1" {
  id = "tank"
}

data "truenas_pool" "pool2" {
  id = "backup"
}

locals {
  # Select pool with most available space
  selected_pool = data.truenas_pool.pool1.available > data.truenas_pool.pool2.available ? data.truenas_pool.pool1 : data.truenas_pool.pool2
}

resource "truenas_dataset" "storage" {
  name = "${local.selected_pool.name}/storage"
  type = "FILESYSTEM"
}
```

#### Pool Monitoring Dashboard
```terraform
data "truenas_pool" "all_pools" {
  for_each = toset(["tank", "backup", "archive"])
  id       = each.key
}

output "pool_summary" {
  value = {
    for pool_name, pool_data in data.truenas_pool.all_pools :
    pool_name => {
      name      = pool_data.name
      status    = pool_data.status
      healthy   = pool_data.healthy
      size_gb   = pool_data.size / 1024 / 1024 / 1024
      free_gb   = pool_data.available / 1024 / 1024 / 1024
      used_pct  = (pool_data.size - pool_data.available) / pool_data.size * 100
    }
  }
}
```

### Integration with Other Resources

The pool data source is commonly used with:

```terraform
data "truenas_pool" "storage" {
  id = "tank"
}

# Create datasets on the pool
resource "truenas_dataset" "app_data" {
  name = "${data.truenas_pool.storage.name}/app-data"
  type = "FILESYSTEM"
}

# Create shares for the datasets
resource "truenas_nfs_share" "app_share" {
  path = "/mnt/${truenas_dataset.app_data.name}"
  comment = "Application data share"
}

# Create VMs using pool storage
resource "truenas_vm" "app_vm" {
  name = "application-server"
  vcpus = 2
  memory = 4096
  
  # VM disks will be created on the pool
}
```

### Best Practices

1. **Always check pool health** before creating resources
2. **Monitor capacity** to prevent running out of space
3. **Use pool names** instead of IDs for better readability
4. **Implement alerts** for pool health and capacity issues
5. **Plan for growth** when allocating storage

### Troubleshooting

**Pool Not Found:**
- Verify pool name spelling
- Check if pool exists in TrueNAS web interface
- Try using numeric pool ID instead

**Incorrect Capacity:**
- Pool statistics may have a delay
- Run `zpool list` on TrueNAS to verify
- Check for pending operations

**Status Issues:**
- Check TrueNAS web interface for detailed pool status
- Look for disk errors or warnings
- Consider running a pool scrub

## See Also

- [truenas_dataset](../resources/dataset) - Create datasets on pools
- [truenas_snapshot](../resources/snapshot) - Create pool snapshots
- [truenas_periodic_snapshot_task](../resources/periodic_snapshot_task) - Automate snapshots
- [ZFS Pool Management](https://www.truenas.com/docs/core/storage/pools/) - TrueNAS documentation