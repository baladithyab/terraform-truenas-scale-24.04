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

## [0.2.14] - 2025-10-31

### Added
- **Boot Order Control**: Device boot order can now be explicitly configured
  - **New Attribute**: `order` field for all device types (nic_devices, disk_devices, cdrom_devices)
  - **Features**:
    - Control which device boots first (lower order = boots first)
    - Useful for OS installation (boot from CDROM first, then disk)
    - Useful for Talos Linux (boot from ISO for install, then disk for operation)
    - Default behavior: auto-incrementing order starting at 1000 if not specified
  - **Use Cases**:
    - Install OS from ISO: CDROM order=1, Disk order=2
    - Normal operation: Disk order=1, CDROM order=2
    - Multi-disk systems: Specify boot priority for each disk
  - **Example**:
    ```hcl
    resource "truenas_vm" "talos_worker" {
      name = "talosworker01"

      # Boot from ISO FIRST for installation
      cdrom_devices = [{
        path  = "/mnt/pool/isos/talos.iso"
        order = 1  # Boots FIRST
      }]

      # Boot from disk SECOND (after OS is installed)
      disk_devices = [{
        path  = "/dev/zvol/pool/vms/disk0"
        order = 2  # Boots SECOND
      }]
    }
    ```

### Fixed
- **Boot Order Bug**: Previously, devices were always ordered by type (NICs, then disks, then CDROMs) regardless of user configuration. Now the `order` field is properly respected and passed to TrueNAS API, which maps to libvirt's `<boot order='X'/>` attribute.

### Documentation
- Added `examples/vm-boot-order/` directory with comprehensive boot order examples
- Added README with detailed boot order configuration guide
- Documented common use cases: OS installation, Talos Linux, multi-disk systems

## [0.2.13] - 2025-10-31

### Added
- **VM Device Configuration**: VMs can now be created with network interfaces, disks, and CDROM devices
  - **New Attributes**:
    - `nic_devices` - List of network interface devices (VIRTIO, E1000, etc.)
    - `disk_devices` - List of disk devices (VIRTIO, AHCI, etc.)
    - `cdrom_devices` - List of CDROM devices for ISO mounting
  - **Features**:
    - Auto-generation of MAC addresses (leave `mac` empty)
    - Support for multiple NICs, disks, and CDROMs per VM
    - Full control over device types and attributes
    - Devices are automatically created during VM creation
    - Devices are automatically read and populated in state
  - **Use Cases**:
    - Create fully functional VMs with network connectivity
    - Attach zvol-based disks for VM storage
    - Mount ISO files for OS installation
    - Configure Talos Linux worker nodes with proper networking
  - **Example**:
    ```hcl
    resource "truenas_vm" "example" {
      name   = "myvm"
      memory = 4096
      vcpus  = 2

      nic_devices = [{
        type       = "VIRTIO"
        nic_attach = "eno1"
      }]

      disk_devices = [{
        path = "/dev/zvol/pool/vms/myvm-disk0"
        type = "VIRTIO"
      }]

      cdrom_devices = [{
        path = "/mnt/pool/isos/ubuntu.iso"
      }]
    }
    ```

### Fixed
- VMs created without devices now properly support device configuration
- Device reading now correctly populates all device types in state
- MAC addresses are properly exported even when auto-generated (null in API)

### Documentation
- Added comprehensive `examples/vm-with-devices/` with README
- Documented all device types and their attributes
- Added troubleshooting guide for common device issues
- Included Talos Linux worker node example

## [0.2.12] - 2025-10-31

### Added
- **VM MAC Address Export**: VM resource now exports MAC addresses from all NIC devices
  - **New Attribute**: `mac_addresses` (computed list of strings)
  - **Use Case**: Look up DHCP leases, configure static IPs for Talos, network inventory
  - **Works For**: ALL VMs (including Talos which doesn't support guest agent)
  - **Example**: `output "macs" { value = truenas_vm.example.mac_addresses }`

- **VM Guest Agent Data Source**: New `truenas_vm_guest_info` data source for querying QEMU guest agent
  - **Attributes**: `ip_addresses`, `hostname`, `os_name`, `os_version`
  - **Requirements**: QEMU guest agent installed in VM, SSH access to TrueNAS host
  - **Use Case**: Automatic IP discovery for VMs with guest agent (Ubuntu, Debian, etc.)
  - **Example**: `data "truenas_vm_guest_info" "ubuntu" { vm_name = "ubuntu-vm" ... }`

### Technical Details
- **Files Changed**:
  - `internal/provider/resource_vm.go` - Added `mac_addresses` computed attribute
  - `internal/provider/datasource_vm_guest_info.go` - New data source for guest agent queries
  - `internal/provider/provider.go` - Registered new data source
- **Implementation**:
  - MAC addresses: Read from VM devices array, filter NIC devices, extract MAC attribute
  - Guest agent: SSH to TrueNAS host, run `virsh qemu-agent-command`, parse JSON response
- **API Limitation**: TrueNAS API doesn't expose IP addresses or guest agent data, hence SSH approach

### Use Case: Talos Kubernetes
```hcl
# Query existing VMs to see what IPs are in use
data "truenas_vm_guest_info" "ubuntu" {
  vm_name      = "ubuntu-vm"
  truenas_host = "10.0.0.83"
  ssh_user     = "root"
  ssh_key_path = "~/.ssh/id_rsa"
}

# Create Talos VMs and get MAC addresses
resource "truenas_vm" "talos_worker" {
  name   = "talos-worker-01"
  memory = 4096
  vcpus  = 2
}

output "talos_macs" {
  value = truenas_vm.talos_worker.mac_addresses
}

# Use existing IPs to avoid conflicts when configuring Talos static IPs
locals {
  existing_ips = data.truenas_vm_guest_info.ubuntu.ip_addresses
  talos_ips    = ["10.0.0.111", "10.0.0.112", "10.0.0.113"]
}
```

### Examples
- New `examples/vm-ip-discovery/` directory with complete examples
- New `examples/data-sources/truenas_vm_guest_info/` with data source examples

### Backward Compatibility
- ✅ No breaking changes
- ✅ All v0.2.11 configurations will work in v0.2.12
- ✅ New attributes are computed (read-only)
- ✅ New data source is optional

## [0.2.11] - 2025-10-30

### Added
- **VM Start on Create**: New `start_on_create` attribute for VM resource
  - **Feature**: Automatically start VM after creation
  - **Use Case**: Eliminates manual step of starting VMs after Terraform creates them
  - **Default**: `false` (VMs are created but not started by default)
  - **Behavior**: When set to `true`, provider calls `/vm/id/{id}/start` API endpoint after VM creation
  - **Error Handling**: If start fails, VM is still created successfully and a warning is shown

### Technical Details
- **Files Changed**:
  - `internal/provider/resource_vm.go` - Added `start_on_create` attribute and start logic
- **API Endpoints Used**:
  - `POST /vm/id/{id}/start` - Start a VM
- **Key Changes**:
  - Added `StartOnCreate` field to `VMResourceModel`
  - Added `start_on_create` schema attribute (optional, default: false)
  - Added start logic in Create() function after VM creation
  - Start failures generate warnings (not errors) to avoid blocking VM creation

### Usage Example
```hcl
resource "truenas_vm" "example" {
  name            = "my_vm"
  memory          = 4096
  vcpus           = 2
  cores           = 1
  threads         = 1
  start_on_create = true  # VM will be started automatically after creation
}
```

### Backward Compatibility
- ✅ No breaking changes
- ✅ All v0.2.10 configurations will work in v0.2.11
- ✅ `start_on_create` is optional and defaults to `false`

## [0.2.10] - 2025-10-30

### Fixed
- **Critical Fix**: VM Read() function now properly reads all properties from TrueNAS API
  - **Root Cause**: v0.2.9 Read() only read 9 properties, missing 6 optional properties, causing "unknown values" errors
  - **Impact**: VMs created successfully but Terraform reported inconsistent state after apply
  - **Solution**: Completely rewrote Read() function to parse ALL VM properties:
    - `id` - VM identifier
    - `name` - VM name
    - `description` - VM description (optional)
    - `vcpus` - Number of virtual CPUs
    - `cores` - Number of cores per socket
    - `threads` - Number of threads per core
    - `memory` - Memory in MB
    - `min_memory` - Minimum memory (optional)
    - `autostart` - Autostart on boot
    - `bootloader` - Bootloader type (optional)
    - `cpu_mode` - CPU mode (optional)
    - `cpu_model` - CPU model (optional)
    - `machine_type` - Machine type (optional)
    - `arch_type` - Architecture type (optional)
    - `time` - Time configuration (optional)
    - `status` - VM status

### Technical Details
- **Files Changed**:
  - `internal/provider/resource_vm.go` - Completely rewrote readVM() function
- **Key Changes**:
  - Parse all 16 VM properties from API response
  - Handle null values for optional properties (description, min_memory, bootloader, cpu_mode, cpu_model, machine_type, arch_type, time)
  - Convert API response types correctly (float64 for integers, string for text)

### Pattern Recognition
- **Same issue as v0.2.9 NFS shares and v0.2.6 datasets**: Read() function must read ALL properties to avoid "unknown values" errors
- This is the **third time** we've fixed this pattern:
  - v0.2.6: Dataset Read() only read 4 properties → Fixed
  - v0.2.9: NFS share Read() only read 4 properties → Fixed
  - v0.2.10: VM Read() only read 9 properties → Fixed

### Backward Compatibility
- ✅ No breaking changes
- ✅ All v0.2.9 configurations will work in v0.2.10
- ✅ Fixes critical state tracking issue for VMs

## [0.2.9] - 2025-10-30

### Fixed
- **Critical Fix**: NFS share Read() function now properly reads all properties from TrueNAS API
  - **Root Cause**: v0.2.8 Read() only read 4 properties (path, comment, enabled, ro), causing "unknown values" errors
  - **Impact**: NFS shares created successfully but Terraform reported inconsistent state after apply
  - **Solution**: Completely rewrote Read() function to parse ALL NFS share properties:
    - `id` - Share identifier
    - `path` - Share path
    - `comment` - Share comment
    - `enabled` - Share enabled status
    - `ro` (readonly) - Read-only status
    - `networks` - Authorized networks list
    - `hosts` - Authorized hosts list (defaults to empty array)
    - `security` - Security mechanisms list (defaults to empty array)
    - `maproot_user` - Root user mapping
    - `mapall_user` - All users mapping

### Technical Details
- **Files Changed**:
  - `internal/provider/resource_nfs_share.go` - Completely rewrote readNFSShare() function
- **Key Changes**:
  - Parse all list properties (networks, hosts, security) from API response
  - Default `hosts` and `security` to empty lists (not null) to match Create behavior
  - Handle null values for optional string properties (comment, maproot_user, mapall_user)
  - Convert API response types correctly (float64 for ID, []interface{} for lists)

### Backward Compatibility
- ✅ No breaking changes
- ✅ All v0.2.8 configurations will work in v0.2.9
- ✅ Fixes critical state tracking issue for NFS shares

## [0.2.8] - 2025-10-30

### Fixed
- **Critical Fix**: NFS share creation now works with default values for required fields
  - **Root Cause**: TrueNAS API requires `hosts` and `security` fields, but provider sent null when not specified
  - **Impact**: NFS share creation failed with "null not allowed" errors
  - **Solution**: Default `hosts` and `security` to empty arrays `[]` when not specified

- **Critical Fix**: Snapshot task creation now accepts cron-style schedule format
  - **Root Cause**: Provider expected JSON schedule format, but users provide cron strings like "0 2 * * *"
  - **Impact**: Snapshot task creation failed with JSON parse errors
  - **Solution**:
    - Added `parseCronSchedule()` helper to convert cron format to JSON
    - Added `scheduleToCron()` helper to convert JSON back to cron format
    - Create/Update functions now parse cron strings to JSON
    - Read function now converts JSON back to cron strings

- **Critical Fix**: VM creation now works with optional string fields
  - **Root Cause**: Provider sent empty strings for optional fields like `cpu_mode` and `time`
  - **Impact**: VM creation failed with "Invalid choice: " errors
  - **Solution**: Only send optional string fields if they have non-empty values

### Technical Details
- **Files Changed**:
  - `internal/provider/resource_nfs_share.go` - Default hosts and security to empty arrays
  - `internal/provider/resource_periodic_snapshot_task.go` - Parse cron schedule format
  - `internal/provider/resource_vm.go` - Only send non-empty string values
- **Affected Resources**: NFS shares, snapshot tasks, VMs

### Backward Compatibility
- ✅ No breaking changes
- ✅ All v0.2.7 configurations will work in v0.2.8
- ✅ Fixes critical issues that prevented NFS shares, snapshot tasks, and VMs from being created

## [0.2.7] - 2025-10-30

### Fixed
- **Critical Fix**: Provider now correctly parses integer properties from TrueNAS API response
  - **Root Cause**: v0.2.6 tried to parse `copies` as float64, but API returns it as a string ("1")
  - **Root Cause**: v0.2.6 didn't handle null values for integer properties (quota, refquota, reservation, refreservation)
  - **Impact**: Terraform reported "unknown values" for integer properties after apply
  - **Solution**:
    - Parse `copies` as string and convert to int64 using `strconv.ParseInt()`
    - Handle null values for all integer properties using type switch
    - Set to `types.Int64Null()` when value is null or missing
  - Integer properties are now correctly parsed and stored in Terraform state

### Technical Details
- **File**: `internal/provider/resource_dataset.go`
- **Changes**:
  - Lines 3-18: Added `strconv` import for string-to-int conversion
  - Lines 480-527: Fixed parsing for copies, reservation, refreservation
  - Lines 573-605: Fixed parsing for quota, refquota
  - All integer properties now handle both null and numeric values correctly
- **Affected Properties**: copies, reservation, refreservation, quota, refquota

### Backward Compatibility
- ✅ No breaking changes
- ✅ All v0.2.6 configurations will work in v0.2.7
- ✅ Fixes critical issue that prevented Terraform from tracking integer property values

## [0.2.6] - 2025-10-30

### Fixed
- **Critical Fix**: Provider now correctly reads all property values from TrueNAS API response
  - **Root Cause**: v0.2.5 `Read()` function only read a few properties (comments, compression, atime, volsize) and set others to null
  - **Impact**: Terraform reported "unknown values" after apply, preventing state tracking
  - **Solution**: Updated `Read()` function to parse and populate ALL properties from API response
  - Properties are now correctly read from the API and stored in Terraform state
  - Applied proper null handling for properties that don't exist in the API response

### Technical Details
- **File**: `internal/provider/resource_dataset.go`
- **Changes**:
  - Lines 439-574: Completely rewrote Read() function to read all properties
  - Added proper parsing for all shared properties: sync, deduplication, readonly, copies, reservation, refreservation
  - Added proper parsing for all FILESYSTEM properties: exec, recordsize, quota, refquota, snapdir
  - Added proper null handling when properties don't exist in API response
- **Affected Properties**: All dataset properties now correctly read from API

### Backward Compatibility
- ✅ No breaking changes
- ✅ All v0.2.5 configurations will work in v0.2.6
- ✅ Fixes critical issue that prevented Terraform from tracking dataset state properly

## [0.2.5] - 2025-10-30

### Fixed
- **Critical Fix**: Provider now correctly omits integer properties with zero values
  - **Root Cause**: v0.2.4 checked `!IsNull()` but still sent zero values (`0`) to the API for unset integer properties
  - **Impact**: TrueNAS API rejected requests with "'copies' must be one of '1 | 2 | 3'" errors for zero values
  - **Solution**: Added value validation (`&& data.PropertyName.ValueInt64() > 0`) for all integer properties
  - Integer properties are now only included in API requests if they have positive (non-zero) values
  - Applied to both `Create()` and `Update()` functions

### Technical Details
- **File**: `internal/provider/resource_dataset.go`
- **Changes**:
  - Lines 231-268: Updated Create() to validate integer values before sending
  - Lines 339-376: Updated Update() to validate integer values before sending
- **Affected Properties**: copies, reservation, refreservation, volsize, quota, refquota
- **String Properties**: No change (already fixed in v0.2.4)

### Backward Compatibility
- ✅ No breaking changes
- ✅ All v0.2.4 configurations will work in v0.2.5
- ✅ Fixes critical issue that prevented datasets from being created when optional integer properties had zero values

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

