# Release Notes - v0.2.0

**Release Date**: October 26, 2025  
**Provider**: TrueNAS Terraform Provider  
**Repository**: https://github.com/baladithyab/terraform-provider-truenas  
**Compatibility**: TrueNAS Scale 24.04

## 🎉 What's New in v0.2.0

This release fixes critical issues identified during community testing and ensures all documented features are fully functional.

### 🐛 Critical Fixes

#### Data Sources Now Working ✅
- **`data.truenas_pool`** - Query pool information (status, health, capacity)
- **`data.truenas_dataset`** - Query dataset information
- **Fixed**: "no schema available" errors that prevented data source usage

#### Import Functionality Verified ✅
- All 14 resources now support import
- **NFS shares** - Import by ID works correctly
- **SMB shares** - Import by ID works correctly
- **Snapshots** - Import with custom format (`dataset@snapshotname`)

#### Snapshot Resources Operational ✅
- **`truenas_snapshot`** - Manual snapshot creation
- **`truenas_periodic_snapshot_task`** - Automated snapshot scheduling
- **Fixed**: Schema validation errors

### ✅ Verified Features

All features have been verified to work correctly:

**Resources (14)**
- ✅ `truenas_dataset` - ZFS dataset management
- ✅ `truenas_nfs_share` - NFS share management
- ✅ `truenas_smb_share` - SMB/CIFS share management
- ✅ `truenas_user` - User account management
- ✅ `truenas_group` - Group management
- ✅ `truenas_vm` - Virtual machine management
- ✅ `truenas_iscsi_target` - iSCSI target management
- ✅ `truenas_iscsi_extent` - iSCSI extent management
- ✅ `truenas_iscsi_portal` - iSCSI portal management
- ✅ `truenas_interface` - Network interface management
- ✅ `truenas_static_route` - Static route management
- ✅ `truenas_chart_release` - Kubernetes application deployment
- ✅ `truenas_snapshot` - ZFS snapshot management
- ✅ `truenas_periodic_snapshot_task` - Automated snapshot scheduling

**Data Sources (2)**
- ✅ `data.truenas_pool` - Query pool information
- ✅ `data.truenas_dataset` - Query dataset information

### 📚 Documentation Updates

- Added `GAPS_ANALYSIS_RESPONSE.md` - Response to community testing
- Added `RELEASE_v0.2.0_PLAN.md` - Release planning guide
- Added `CHANGELOG.md` - Complete version history
- Updated `../api/API_COVERAGE.md` - Version information and warnings
- Updated `README.md` - Version references

## 🔄 Upgrading from v0.1.0

### No Breaking Changes

This is a **bug fix release** with no breaking changes. Upgrading is simple:

**Step 1**: Update your `terraform` block:
```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.0"  # Changed from 0.1.0
    }
  }
}
```

**Step 2**: Upgrade the provider:
```bash
terraform init -upgrade
```

**Step 3**: Verify everything works:
```bash
terraform plan
```

### Removing Workarounds

If you were using workarounds for missing features in v0.1.0, you can now remove them:

**Before (v0.1.0 with HTTP provider workaround):**
```hcl
data "http" "pool_info" {
  url = "${var.truenas_base_url}/api/v2.0/pool/id/Loki"
  request_headers = {
    Authorization = "Bearer ${var.truenas_api_key}"
  }
}

locals {
  pool_data = jsondecode(data.http.pool_info.response_body)
}
```

**After (v0.2.0 with native data source):**
```hcl
data "truenas_pool" "loki" {
  id = "Loki"
}

output "pool_status" {
  value = data.truenas_pool.loki.status
}
```

## 📖 Usage Examples

### Using Data Sources

**Query Pool Information:**
```hcl
data "truenas_pool" "main" {
  id = "tank"
}

output "pool_health" {
  value = data.truenas_pool.main.healthy ? "Healthy" : "Degraded"
}

output "pool_free_space" {
  value = "${data.truenas_pool.main.available / 1024 / 1024 / 1024} GB"
}
```

**Query Dataset Information:**
```hcl
data "truenas_dataset" "media" {
  id = "tank/media"
}

output "dataset_compression" {
  value = data.truenas_dataset.media.compression
}
```

### Importing Resources

**Import NFS Share:**
```bash
# Find share ID from TrueNAS UI or API
terraform import truenas_nfs_share.media 6
```

**Import Snapshot:**
```bash
# Use dataset@snapshotname format
terraform import truenas_snapshot.backup "tank/data@backup-2025-10-26"
```

**Import Periodic Snapshot Task:**
```bash
# Use task ID
terraform import truenas_periodic_snapshot_task.daily 1
```

### Creating Snapshots

**Manual Snapshot:**
```hcl
resource "truenas_snapshot" "pre_upgrade" {
  dataset   = "tank/important"
  name      = "pre-upgrade-${formatdate("YYYY-MM-DD", timestamp())}"
  recursive = true
}
```

**Automated Snapshot Schedule:**
```hcl
resource "truenas_periodic_snapshot_task" "hourly_backup" {
  dataset        = "tank/databases"
  recursive      = false
  lifetime_value = 24
  lifetime_unit  = "HOUR"
  naming_schema  = "auto-%Y%m%d-%H%M"
  
  schedule = jsonencode({
    minute = "0"
    hour   = "*"
    dom    = "*"
    month  = "*"
    dow    = "*"
  })
  
  enabled = true
}
```

## 🔍 What Was Fixed

### Root Cause

The issues reported in community testing were caused by a **version mismatch**:
- v0.1.0 was tagged before some features were fully implemented
- The main branch had all features, but they weren't in the published version
- This release (v0.2.0) includes all features from the main branch

### Verification

All features have been verified:
- ✅ Code exists and compiles
- ✅ Resources registered in provider
- ✅ Build produces working binary (25MB)
- ✅ Tested against TrueNAS Scale 24.04

## 📊 Statistics

| Metric | Value |
|--------|-------|
| **Resources** | 14 |
| **Data Sources** | 2 |
| **Import Support** | 100% (all resources) |
| **Documentation Files** | 13 |
| **Example Configurations** | 17+ |
| **API Coverage** | ~2.2% (14 of 643 endpoints) |
| **Binary Size** | 25MB |

## 🐛 Known Issues

None identified in v0.2.0.

If you encounter any issues, please report them at:
https://github.com/baladithyab/terraform-provider-truenas/issues

## 🙏 Acknowledgments

Special thanks to the Yggdrasil Infrastructure Team for comprehensive testing and detailed gap analysis that led to this release.

## 📞 Support

- **GitHub Issues**: https://github.com/baladithyab/terraform-provider-truenas/issues
- **Documentation**: https://github.com/baladithyab/terraform-provider-truenas
- **TrueNAS Version**: Scale 24.04 (REST API)

## 🔗 Links

- **Repository**: https://github.com/baladithyab/terraform-provider-truenas
- **Changelog**: [CHANGELOG.md](CHANGELOG.md)
- **Import Guide**: [IMPORT_GUIDE.md](IMPORT_GUIDE.md)
- **API Coverage**: [API_COVERAGE.md](API_COVERAGE.md)
- **Gaps Analysis Response**: [GAPS_ANALYSIS_RESPONSE.md](GAPS_ANALYSIS_RESPONSE.md)

## 🚀 What's Next

### Planned for v0.3.0
- Replication task management
- Cloud sync task management
- Service management (start/stop/configure)
- Certificate management
- Cron job management

See [API_COVERAGE.md](API_COVERAGE.md) for the complete roadmap.

---

**Full Changelog**: https://github.com/baladithyab/terraform-provider-truenas/blob/main/CHANGELOG.md

