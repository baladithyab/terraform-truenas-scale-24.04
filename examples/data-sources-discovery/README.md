# TrueNAS Data Sources Discovery Example

This example demonstrates how to use the TrueNAS provider's data sources to discover and query existing resources on your TrueNAS system.

## Features Demonstrated

### 1. **NFS Shares Discovery** (`truenas_nfs_shares`)
- List all NFS shares on the system
- Filter enabled shares
- Access share properties (path, networks, permissions)
- Generate Kubernetes StorageClass configurations

### 2. **SMB Shares Discovery** (`truenas_smb_shares`)
- List all SMB/CIFS shares
- Filter by properties (Time Machine, guest access, etc.)
- Access share metadata (name, path, purpose)

### 3. **VMs Discovery** (`truenas_vms`)
- List all virtual machines
- Filter by status (RUNNING, STOPPED)
- Access VM properties (vcpus, memory, autostart)
- Identify VMs needing attention

### 4. **Single VM Query** (`truenas_vm`)
- Query specific VM by name or ID
- Get detailed VM configuration
- Access CPU mode, bootloader, and other settings

## Prerequisites

- TrueNAS Scale 24.04 or later
- TrueNAS API key with read permissions
- Terraform 1.0 or later

## Usage

1. **Set your API key:**
   ```bash
   export TF_VAR_truenas_api_key="your-api-key-here"
   ```

2. **Update the provider configuration:**
   Edit `main.tf` and set the correct `base_url` for your TrueNAS system.

3. **Initialize Terraform:**
   ```bash
   terraform init
   ```

4. **Run the discovery:**
   ```bash
   terraform plan
   ```

5. **View outputs:**
   ```bash
   terraform apply
   ```

## Example Outputs

### NFS Shares
```hcl
nfs_shares = {
  "/mnt/Loki/midgard/media" = {
    id       = 1
    enabled  = true
    comment  = "Media libraries"
    networks = ["10.0.0.0/24"]
  }
}
```

### SMB Shares
```hcl
smb_shares = {
  "media" = {
    id       = 1
    path     = "/mnt/Loki/midgard/media"
    enabled  = true
    readonly = false
    comment  = "Media share"
  }
}
```

### VMs
```hcl
all_vms = {
  "talos-worker-1" = {
    id     = "7"
    status = "RUNNING"
    vcpus  = 4
    memory = 8192
  }
}
```

## Practical Use Cases

### 1. Generate Kubernetes NFS StorageClasses
The example automatically generates configuration for Kubernetes NFS StorageClasses based on enabled NFS shares:

```hcl
nfs_storage_class_config = {
  "Loki-midgard-media" = {
    server = "10.0.0.83"
    path   = "/mnt/Loki/midgard/media"
  }
}
```

### 2. Monitoring and Alerting
Identify VMs that should be running but are stopped:

```hcl
vms_needing_start = ["talos-worker-2"]
```

### 3. Inventory Reporting
Get a summary of all resources:

```hcl
truenas_inventory = {
  nfs_shares = {
    total   = 5
    enabled = 4
  }
  smb_shares = {
    total       = 3
    timemachine = 1
  }
  vms = {
    total   = 8
    running = 6
  }
}
```

### 4. Dynamic Configuration
Use discovered resources to configure other Terraform resources:

```hcl
# Create Kubernetes ConfigMap with NFS mount points
resource "kubernetes_config_map" "nfs_mounts" {
  metadata {
    name = "nfs-mounts"
  }
  
  data = {
    for share in data.truenas_nfs_shares.all.shares :
    replace(share.path, "/", "_") => share.path
    if share.enabled
  }
}
```

## Data Source Reference

### `truenas_nfs_shares`
Returns a list of all NFS shares with properties:
- `id` - Share ID
- `path` - Export path
- `comment` - Description
- `enabled` - Whether share is active
- `readonly` - Read-only flag
- `networks` - Allowed networks (CIDR)
- `hosts` - Allowed hosts
- `maproot_user`, `maproot_group` - Root mapping
- `mapall_user`, `mapall_group` - All users mapping

### `truenas_smb_shares`
Returns a list of all SMB shares with properties:
- `id` - Share ID
- `name` - Share name
- `path` - Share path
- `comment` - Description
- `enabled` - Whether share is active
- `readonly` - Read-only flag
- `browsable` - Browsable flag
- `guestok` - Guest access allowed
- `recyclebin` - Recycle bin enabled
- `purpose` - Share purpose
- `home` - Home share flag
- `timemachine` - Time Machine support

### `truenas_vms`
Returns a list of all VMs with properties:
- `id` - VM ID
- `name` - VM name
- `description` - VM description
- `vcpus` - Virtual CPU count
- `cores` - Cores per socket
- `threads` - Threads per core
- `memory` - Memory in MiB
- `autostart` - Autostart flag
- `status` - Current status (RUNNING, STOPPED, etc.)

### `truenas_vm`
Query a specific VM by name or ID. Returns all properties from `truenas_vms` plus:
- `min_memory` - Minimum memory
- `bootloader` - Bootloader type
- `cpu_mode` - CPU mode
- `cpu_model` - CPU model

## Benefits Over HTTP Data Sources

Previously, you might have used `data "http"` resources to query the TrueNAS API:

```hcl
# OLD WAY - Don't do this anymore
data "http" "nfs_shares" {
  url = "http://10.0.0.83:81/api/v2.0/sharing/nfs"
  request_headers = {
    Authorization = "Bearer ${var.truenas_api_key}"
  }
}

locals {
  nfs_shares = jsondecode(data.http.nfs_shares.body)
}
```

**New way is better because:**
- ✅ **Type-safe** - Proper Terraform schema with validation
- ✅ **Better errors** - Clear error messages instead of JSON parsing failures
- ✅ **Auto-completion** - IDE support for attributes
- ✅ **Consistent** - Same patterns as other Terraform providers
- ✅ **Documented** - Built-in documentation via `terraform-docs`
- ✅ **Maintainable** - No manual JSON parsing or error handling

## Next Steps

1. **Remove HTTP data sources** from your existing Terraform code
2. **Replace with native data sources** from this provider
3. **Simplify your code** by removing JSON parsing logic
4. **Add type safety** by using proper Terraform attributes

## Related Examples

- [VM GPU Passthrough](../vm-gpu-passthrough/) - GPU device discovery and attachment
- [Basic VM](../basic-vm/) - Creating VMs with the provider
- [Dataset Management](../dataset-management/) - Managing ZFS datasets

