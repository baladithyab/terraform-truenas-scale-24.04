# Changelog

All notable changes to the TrueNAS Terraform Provider will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned for v0.3.0
- Replication task management
- Cloud sync task management
- Service management (start/stop/configure)
- Certificate management
- Cron job management

## [0.2.0] - 2025-10-26

### Fixed
- **Data Sources Now Functional** 🎉
  - `data.truenas_pool` - Query pool information (status, health, capacity)
  - `data.truenas_dataset` - Query dataset information
  - Fixed schema registration issues that caused "no schema available" errors

- **Import Functionality Verified** ✅
  - All 14 resources support import
  - NFS share import working correctly
  - SMB share import working correctly
  - Snapshot import with custom format (`dataset@snapshotname`)

- **Snapshot Resources Fully Operational** 📸
  - `truenas_snapshot` - Manual snapshot creation
  - `truenas_periodic_snapshot_task` - Automated snapshot scheduling
  - Fixed schema validation errors

### Verified
- ✅ All 14 resources compile and register correctly
- ✅ All 2 data sources compile and register correctly
- ✅ Build process produces working 25MB binary
- ✅ Tested against TrueNAS Scale 24.04
- ✅ Import functionality tested with real infrastructure

### Documentation
- Added `GAPS_ANALYSIS_RESPONSE.md` - Response to community testing
- Added `RELEASE_v0.2.0_PLAN.md` - Release planning and testing guide
- Updated `API_COVERAGE.md` - Added version information and warnings
- Updated `CHANGELOG.md` - This file

### Known Issues
- None identified in v0.2.0 testing

### Breaking Changes
- None - fully backward compatible with v0.1.0

### Migration from v0.1.0
No changes required. Simply update your provider version:

```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.0"  # Update from 0.1.0
    }
  }
}
```

Then run:
```bash
terraform init -upgrade
```

## [0.1.0] - 2025-10-15

### Added - Initial Release

#### Resources (14)
**Storage & File Sharing (3)**
- `truenas_dataset` - ZFS dataset management
- `truenas_nfs_share` - NFS share management
- `truenas_smb_share` - SMB/CIFS share management

**User Management (2)**
- `truenas_user` - User account management
- `truenas_group` - Group management

**Virtual Machines (1)**
- `truenas_vm` - Virtual machine management

**iSCSI (3)**
- `truenas_iscsi_target` - iSCSI target management
- `truenas_iscsi_extent` - iSCSI extent (storage) management
- `truenas_iscsi_portal` - iSCSI portal (network listener) management

**Network (2)**
- `truenas_interface` - Network interface management (PHYSICAL, VLAN, BRIDGE, LAG)
- `truenas_static_route` - Static route management

**Kubernetes/Apps (1)**
- `truenas_chart_release` - Kubernetes application deployment

**Snapshots (2)**
- `truenas_snapshot` - ZFS snapshot management
- `truenas_periodic_snapshot_task` - Automated snapshot scheduling

#### Data Sources (2)
- `data.truenas_dataset` - Query dataset information
- `data.truenas_pool` - Query pool information

#### Features
- ✅ Full CRUD operations for all resources
- ✅ Import support for all resources
- ✅ Comprehensive examples for each resource
- ✅ Complete documentation (10 guides)
- ✅ Kubernetes migration capabilities
- ✅ Multi-tier snapshot strategies

#### Documentation
- `README.md` - Main overview and quick start
- `QUICKSTART.md` - Getting started guide
- `API_COVERAGE.md` - Complete API status tracking
- `API_ENDPOINTS.md` - Full endpoint reference
- `PROJECT_SUMMARY.md` - Technical implementation details
- `IMPORT_GUIDE.md` - Import documentation for all resources
- `KUBERNETES_MIGRATION.md` - Complete migration guide (5 workflows)
- `TESTING.md` - Testing guide
- `CONTRIBUTING.md` - Contribution guidelines

#### Examples
- Complete examples for all 14 resources
- Production-ready Kubernetes deployment examples
- Multi-tier snapshot configuration examples
- Network configuration examples (VLAN, Bridge, LAG)
- iSCSI complete setup examples

### Known Issues in v0.1.0
- ⚠️ Data sources may not work correctly (fixed in v0.2.0)
- ⚠️ Some import functionality may be missing (fixed in v0.2.0)
- ⚠️ Snapshot resources may have schema errors (fixed in v0.2.0)

**Recommendation**: Upgrade to v0.2.0 when available

## Version History Summary

| Version | Date | Resources | Data Sources | Key Features |
|---------|------|-----------|--------------|--------------|
| v0.2.0 | 2025-10-27 | 14 | 2 | Data sources fixed, import verified |
| v0.1.0 | 2025-10-15 | 14 | 2 | Initial release |

## Upgrade Guide

### From v0.1.0 to v0.2.0

**No breaking changes** - This is a bug fix release.

1. Update your `terraform` block:
   ```hcl
   terraform {
     required_providers {
       truenas = {
         source  = "registry.terraform.io/baladithyab/truenas"
         version = "~> 0.2.0"
       }
     }
   }
   ```

2. Upgrade the provider:
   ```bash
   terraform init -upgrade
   ```

3. Verify everything works:
   ```bash
   terraform plan
   ```

4. If you were using workarounds for data sources, you can now remove them:
   ```hcl
   # OLD (v0.1.0 workaround with HTTP provider)
   data "http" "pool_info" {
     url = "${var.truenas_base_url}/api/v2.0/pool/id/Loki"
     # ...
   }
   
   # NEW (v0.2.0 native data source)
   data "truenas_pool" "loki" {
     id = "Loki"
   }
   ```

## Support

- **GitHub Issues**: https://github.com/baladithyab/terraform-truenas-scale-24.04/issues
- **Documentation**: https://github.com/baladithyab/terraform-truenas-scale-24.04
- **TrueNAS Version**: Scale 24.04 (REST API)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:
- Reporting bugs
- Suggesting features
- Submitting pull requests
- Testing new resources

## License

Mozilla Public License 2.0

---

**Note**: This provider is specific to TrueNAS Scale 24.04 which uses REST API. 
TrueNAS Scale 25.04+ uses WebSocket/JSON-RPC and is not compatible with this provider.

