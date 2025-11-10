# Import Guide

This guide explains how to import existing TrueNAS resources into Terraform state.

## Overview

All resources in this provider support importing existing infrastructure. This allows you to:
- Migrate existing TrueNAS configurations to Terraform
- Recover from state file loss
- Adopt resources created outside of Terraform

## Import Syntax

```bash
terraform import <resource_type>.<resource_name> <import_id>
```

## Supported Resources

### Storage & File Sharing

#### Dataset
```bash
terraform import truenas_dataset.mydata tank/mydata
```
**Import ID Format**: Dataset name (e.g., `tank/mydata`)

#### NFS Share
```bash
terraform import truenas_nfs_share.share1 1
```
**Import ID Format**: Share ID (numeric)

#### SMB Share
```bash
terraform import truenas_smb_share.share1 1
```
**Import ID Format**: Share ID (numeric)

### User Management

#### User
```bash
terraform import truenas_user.john 1000
```
**Import ID Format**: User ID (numeric)

#### Group
```bash
terraform import truenas_group.developers 1000
```
**Import ID Format**: Group ID (numeric)

### Virtual Machines

#### VM
```bash
terraform import truenas_vm.ubuntu ubuntu-vm
```
**Import ID Format**: VM name

### iSCSI

#### iSCSI Target
```bash
terraform import truenas_iscsi_target.target1 1
```
**Import ID Format**: Target ID (numeric)

#### iSCSI Extent
```bash
terraform import truenas_iscsi_extent.extent1 1
```
**Import ID Format**: Extent ID (numeric)

#### iSCSI Portal
```bash
terraform import truenas_iscsi_portal.portal1 1
```
**Import ID Format**: Portal ID (numeric)

### Network

#### Interface
```bash
terraform import truenas_interface.eth0 eth0
```
**Import ID Format**: Interface name (e.g., `eth0`, `vlan10`, `br0`, `bond0`)

#### Static Route
```bash
terraform import truenas_static_route.route1 1
```
**Import ID Format**: Route ID (numeric)

### Kubernetes/Apps

#### Chart Release
```bash
terraform import truenas_chart_release.plex plex
```
**Import ID Format**: Release name

### Snapshots

#### Snapshot
```bash
terraform import truenas_snapshot.backup tank/mydata@backup-2024-01-15
```
**Import ID Format**: `dataset@snapshotname`

#### Periodic Snapshot Task
```bash
terraform import truenas_periodic_snapshot_task.hourly 1
```
**Import ID Format**: Task ID (numeric)

## Import Workflow

### 1. Create Resource Block

First, create an empty resource block in your Terraform configuration:

```hcl
resource "truenas_dataset" "mydata" {
  # Configuration will be populated after import
}
```

### 2. Run Import Command

```bash
terraform import truenas_dataset.mydata tank/mydata
```

### 3. Generate Configuration

After import, use `terraform show` to see the current state:

```bash
terraform show
```

### 4. Update Configuration

Copy the relevant attributes from `terraform show` output to your `.tf` file:

```hcl
resource "truenas_dataset" "mydata" {
  name        = "tank/mydata"
  compression = "lz4"
  atime       = "off"
  quota       = 1099511627776  # 1TB
}
```

### 5. Verify

Run `terraform plan` to ensure no changes are detected:

```bash
terraform plan
```

If the plan shows changes, adjust your configuration until it matches the imported state.

## Bulk Import Example

To import multiple resources, create a script:

```bash
#!/bin/bash

# Import all datasets
terraform import truenas_dataset.tank tank
terraform import truenas_dataset.media tank/media
terraform import truenas_dataset.backups tank/backups

# Import all shares
terraform import truenas_nfs_share.media_nfs 1
terraform import truenas_smb_share.media_smb 2

# Import all users
terraform import truenas_user.alice 1001
terraform import truenas_user.bob 1002

echo "Import complete!"
```

## Finding Import IDs

### Numeric IDs

For resources that use numeric IDs, you can find them using the TrueNAS API:

```bash
# List all NFS shares
curl -H "Authorization: Bearer $API_KEY" \
  http://truenas-ip/api/v2.0/sharing/nfs

# List all users
curl -H "Authorization: Bearer $API_KEY" \
  http://truenas-ip/api/v2.0/user

# List all iSCSI targets
curl -H "Authorization: Bearer $API_KEY" \
  http://truenas-ip/api/v2.0/iscsi/target
```

### Name-Based IDs

For resources that use names:

```bash
# List all datasets
curl -H "Authorization: Bearer $API_KEY" \
  http://truenas-ip/api/v2.0/pool/dataset

# List all VMs
curl -H "Authorization: Bearer $API_KEY" \
  http://truenas-ip/api/v2.0/vm

# List all interfaces
curl -H "Authorization: Bearer $API_KEY" \
  http://truenas-ip/api/v2.0/interface
```

### Snapshots

List snapshots with their full names:

```bash
curl -H "Authorization: Bearer $API_KEY" \
  http://truenas-ip/api/v2.0/zfs/snapshot
```

The `name` field will be in the format `dataset@snapshotname`.

## Common Issues

### Issue: "Resource not found"

**Cause**: The import ID is incorrect or the resource doesn't exist.

**Solution**: Verify the resource exists using the TrueNAS API or web UI.

### Issue: "Resource already managed by Terraform"

**Cause**: The resource is already in the Terraform state.

**Solution**: Remove it from state first:
```bash
terraform state rm truenas_dataset.mydata
```

### Issue: Plan shows changes after import

**Cause**: Your Terraform configuration doesn't match the actual resource state.

**Solution**: 
1. Run `terraform show` to see the imported state
2. Update your `.tf` file to match
3. Pay attention to computed vs. required attributes

## Best Practices

1. **Import one resource at a time** - Easier to troubleshoot
2. **Use version control** - Commit after each successful import
3. **Test in non-production first** - Verify the import process
4. **Document import IDs** - Keep a mapping of resource names to IDs
5. **Verify with plan** - Always run `terraform plan` after import
6. **Backup state file** - Before bulk imports

## Advanced: Importing Entire Infrastructure

For importing a complete TrueNAS setup:

1. **Inventory**: List all resources using the API
2. **Generate**: Create Terraform configuration files
3. **Import**: Run import commands for each resource
4. **Validate**: Verify with `terraform plan`
5. **Refactor**: Organize into modules if needed

Example inventory script:

```bash
#!/bin/bash
API_KEY="your-api-key"
BASE_URL="http://truenas-ip/api/v2.0"

echo "=== Datasets ==="
curl -s -H "Authorization: Bearer $API_KEY" $BASE_URL/pool/dataset | jq -r '.[].name'

echo "=== NFS Shares ==="
curl -s -H "Authorization: Bearer $API_KEY" $BASE_URL/sharing/nfs | jq -r '.[] | "\(.id) - \(.path)"'

echo "=== SMB Shares ==="
curl -s -H "Authorization: Bearer $API_KEY" $BASE_URL/sharing/smb | jq -r '.[] | "\(.id) - \(.name)"'

echo "=== Users ==="
curl -s -H "Authorization: Bearer $API_KEY" $BASE_URL/user | jq -r '.[] | "\(.uid) - \(.username)"'

echo "=== VMs ==="
curl -s -H "Authorization: Bearer $API_KEY" $BASE_URL/vm | jq -r '.[] | "\(.id) - \(.name)"'
```

## See Also

- [Terraform Import Documentation](https://www.terraform.io/docs/cli/import/index.html)
- [TrueNAS API Documentation](https://www.truenas.com/docs/api/)
- [CONTRIBUTING.md](CONTRIBUTING.md) - For adding import support to new resources

