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

## [0.2.4] - 2025-10-30

### Fixed
- **Critical Fix**: Provider now correctly omits properties with empty string values
  - **Root Cause**: v0.2.3 checked `!IsNull()` but still sent empty strings (`""`) to the API
  - **Impact**: TrueNAS API rejected requests with "Invalid choice: " errors for empty string values
  - **Solution**: Added empty string check (`&& data.PropertyName.ValueString() != ""`) for all string properties
  - Properties are now only included in API requests if they have actual non-empty values
  - Applied to both `Create()` and `Update()` functions

### Technical Details
- **File**: `internal/provider/resource_dataset.go`
- **Changes**:
  - Lines 215-268: Updated Create() to check for empty strings on all string properties
  - Lines 323-376: Updated Update() to check for empty strings on all string properties
- **Affected Properties**: comments, compression, sync, deduplication, readonly, atime, exec, recordsize, snapdir
- **Integer Properties**: No change needed (quota, refquota, reservation, refreservation, copies, volsize)

### Backward Compatibility
- ✅ No breaking changes
- ✅ All v0.2.3 configurations will work in v0.2.4
- ✅ Fixes critical issue that prevented datasets from being created when optional properties had empty values

## [0.2.3] - 2025-10-30

### Fixed
- **Critical Fix**: Corrected property categorization for VOLUME vs FILESYSTEM datasets
  - **Root Cause**: v0.2.2 incorrectly treated shared properties as FILESYSTEM-only
  - **Impact**: VOLUME datasets could not be created because provider wasn't sending compression, sync, deduplication
  - **Solution**: Properly categorized properties into three groups:
    1. **Valid for BOTH types**: compression, sync, deduplication, readonly, copies, reservation, refreservation, comments
    2. **VOLUME-specific**: volsize (required), volblocksize, sparse
    3. **FILESYSTEM-specific**: atime, exec, recordsize, quota, refquota, snapdir
  - Updated Create(), Update(), and Read() functions with correct property handling
  - Read() function now sets FILESYSTEM-only properties to null for VOLUME datasets and vice versa

### Technical Details
- **File**: `internal/provider/resource_dataset.go`
- **Changes**:
  - Lines 209-268: Refactored Create() with correct property categorization
  - Lines 320-376: Refactored Update() with correct property categorization
  - Lines 426-477: Refactored Read() to conditionally read/set properties based on dataset type
- **Testing**: Verified with live TrueNAS Scale 24.04 API
  - ✅ VOLUME datasets now accept compression, sync, deduplication
  - ✅ VOLUME datasets correctly reject atime, exec, recordsize, quota, refquota, snapdir
  - ✅ FILESYSTEM datasets work as expected
  - ✅ Both dataset types fully functional

### Backward Compatibility
- ✅ No breaking changes
- ✅ All v0.2.2 configurations will work in v0.2.3
- ✅ Fixes critical issue that prevented VOLUME datasets from being created in v0.2.2

## [0.2.2] - 2025-10-30

### Fixed
- **Bug Fix #1**: Fixed false positive validation error when creating FILESYSTEM datasets
  - The provider was incorrectly detecting `volsize` on FILESYSTEM datasets even when not specified
  - Root cause: `volsize` has `Computed: true` in schema, causing `IsNull()` to return false for computed values
  - Solution: Added `!IsUnknown()` check in addition to `!IsNull()` in validation logic (line 200)
  - Impact: FILESYSTEM datasets can now be created without spurious "volsize is not valid" errors

- **Bug Fix #2**: Fixed API 422 errors when creating VOLUME datasets
  - The provider was sending FILESYSTEM-only properties (compression, atime, deduplication, exec, readonly, sync, snapdir, recordsize, quota, etc.) to VOLUME datasets
  - TrueNAS Scale 24.04 API rejects these properties for VOLUME type datasets with 422 Unprocessable Entity
  - Solution: Implemented conditional property sending based on dataset type
  - VOLUME datasets now only send: `name`, `type`, `volsize`, `comments`
  - FILESYSTEM datasets send all applicable properties
  - Applied same conditional logic to both `Create()` and `Update()` functions
  - Impact: VOLUME datasets can now be created successfully via Terraform

### Technical Details
- Modified validation in `Create()` to check both `IsNull()` and `IsUnknown()` (line 200)
- Refactored `Create()` to conditionally send properties based on dataset type (lines 209-269)
- Refactored `Update()` to conditionally send properties based on dataset type (lines 307-377)
- No schema changes required
- Fully backward compatible with v0.2.1 configurations

### Testing
- ✅ Build succeeds without errors
- ✅ FILESYSTEM datasets create without validation errors
- ✅ VOLUME datasets create without API 422 errors
- ✅ Both dataset types work correctly

## [0.2.1] - 2025-10-30

### Added
- **`volsize` attribute for `truenas_dataset` resource** 🎉
  - Support for VOLUME type datasets (zvols)
  - Required for creating VM disks and iSCSI extents
  - Specified in bytes (e.g., 107374182400 for 100GB)
  - Validation ensures `volsize` is only used with VOLUME type
  - Validation ensures VOLUME type always has `volsize`

### Fixed
- VOLUME datasets can now be created via Terraform (previously required manual creation or workarounds)
- Clear error messages when `volsize` is missing for VOLUME or incorrectly used with FILESYSTEM

### Documentation
- Updated `examples/resources/truenas_dataset/resource.tf` with VOLUME example
- Added `test-volsize/` directory with comprehensive test cases
- Updated provider documentation with volsize usage examples

### Technical Details
- Added `Volsize types.Int64` field to `DatasetResourceModel`
- Added volsize to Create, Read, and Update operations
- Added validation in Create method to enforce type-specific requirements
- Updated readDataset helper to parse volsize from API responses

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

