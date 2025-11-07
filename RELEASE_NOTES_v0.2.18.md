# TrueNAS Terraform Provider v0.2.18 Release Notes

Release Date: 2025-11-07

## Overview

Version 0.2.18 is a major feature release that introduces standalone VM device management, declarative VM lifecycle control, and enhanced VM IP discovery with improved security and reliability. This release significantly improves the VM management experience with hot-add/remove device capabilities, automatic state management, and better SSH authentication handling.

---

## Major Features

### 1. New Resource: `truenas_vm_device`

Manage VM devices independently of the VM resource, enabling dynamic device management without recreating VMs.

**Key Features:**
- Standalone device management (create, update, delete devices on existing VMs)
- Hot-add/remove capabilities for supported device types
- Full support for all device types: NIC, DISK, CDROM, PCI, USB, DISPLAY, RAW
- Import support for existing devices
- Independent lifecycle from VM resource

**Why This Matters:**
- Add storage to running VMs without downtime
- Dynamically attach/detach network interfaces
- Manage GPU passthrough devices separately
- Simplify complex VM configurations
- Enable modular infrastructure as code

**Example - Adding a Network Interface:**
```hcl
resource "truenas_vm_device" "additional_nic" {
  vm_id = truenas_vm.example.id
  dtype = "NIC"
  attributes = {
    type       = "VIRTIO"
    nic_attach = "br0"
  }
  order = 1001
}
```

**Example - Adding Storage:**
```hcl
resource "truenas_vm_device" "extra_disk" {
  vm_id = truenas_vm.example.id
  dtype = "DISK"
  attributes = {
    path = "/dev/zvol/tank/vms/extra-storage"
    type = "VIRTIO"
  }
  order = 1002
}
```

**Example - GPU Passthrough:**
```hcl
resource "truenas_vm_device" "gpu" {
  vm_id = truenas_vm.ml_workstation.id
  dtype = "PCI"
  attributes = {
    pptdev = "pci_0000_01_00_0"  # GPU PCI address
  }
}
```

**Import Existing Devices:**
```bash
terraform import truenas_vm_device.my_device <vm_id>:<device_id>
```

---

### 2. VM Lifecycle Management with `desired_state`

Declaratively control VM power state with automatic state transitions and drift detection.

**New Attribute:**
- **`desired_state`**: String (Optional) - Target VM state: "RUNNING", "STOPPED", or "SUSPENDED"
- **Default**: Maintains current state (no automatic start/stop)
- **Behavior**: Provider automatically transitions VM to match desired state
- **Drift Detection**: Automatically restarts stopped VMs if `desired_state = "RUNNING"`

**Replaces:** `start_on_create` (now deprecated but still supported for backward compatibility)

**Why This Matters:**
- Declarative state management (Infrastructure as Code best practices)
- Automatic recovery if VM state drifts
- Simplifies VM lifecycle in automation and CI/CD
- No more manual VM start/stop operations
- Eliminates race conditions in complex deployments

**Example - Always Running VM:**
```hcl
resource "truenas_vm" "web_server" {
  name          = "nginx-server"
  memory        = 4096
  vcpus         = 2
  desired_state = "RUNNING"  # VM will always be running
  
  disk_devices = [{
    path = "/dev/zvol/tank/vms/nginx-disk"
    type = "VIRTIO"
  }]
}
```

**Example - Stopped VM (for maintenance):**
```hcl
resource "truenas_vm" "backup_server" {
  name          = "backup-vm"
  memory        = 2048
  vcpus         = 1
  desired_state = "STOPPED"  # VM will be stopped
}
```

**Example - Suspended VM (save RAM state):**
```hcl
resource "truenas_vm" "dev_environment" {
  name          = "dev-vm"
  memory        = 8192
  vcpus         = 4
  desired_state = "SUSPENDED"  # VM will be suspended
}
```

**Migration from `start_on_create`:**
```hcl
# OLD (v0.2.17 and earlier)
resource "truenas_vm" "example" {
  name            = "my-vm"
  memory          = 4096
  vcpus           = 2
  start_on_create = true  # Deprecated but still works
}

# NEW (v0.2.18)
resource "truenas_vm" "example" {
  name          = "my-vm"
  memory        = 4096
  vcpus         = 2
  desired_state = "RUNNING"  # Declarative and drift-aware
}
```

---

## Enhancements

### 3. Enhanced `truenas_vm_guest_info` Data Source

Significant security and reliability improvements for VM IP discovery.

**Authentication Validation:**
- Provider now validates SSH authentication **before** querying guest agent
- Eliminates wasted time on invalid credentials
- Clear, actionable error messages for authentication failures
- Faster feedback on configuration issues

**New Security Attribute:** `ssh_strict_host_key_checking`
- **Type**: Boolean (Optional)
- **Default**: `false` (permissive, accepts any host key)
- **When `true`**: SSH enforces host key verification (production recommended)
- **When `false`**: SSH accepts any host key (convenient for automation)

**Example - Strict Security (Production):**
```hcl
data "truenas_vm_guest_info" "ubuntu" {
  vm_name                      = "ubuntu-vm"
  truenas_host                 = "10.0.0.83"
  ssh_user                     = "root"
  ssh_key_path                 = "~/.ssh/truenas_key"
  ssh_strict_host_key_checking = true  # Enforce host key verification
}
```

**Prepare for strict checking:**
```bash
# Add TrueNAS host key to known_hosts
ssh-keyscan -H 10.0.0.83 >> ~/.ssh/known_hosts
```

**New Timeout Attribute:** `ssh_timeout_seconds`
- **Type**: Integer (Optional)
- **Default**: `10` seconds
- **Purpose**: Configure SSH connection timeout
- **Use Cases**: 
  - Increase for slow networks
  - Increase for busy TrueNAS systems
  - Decrease for faster failure detection

**Example - Extended Timeout:**
```hcl
data "truenas_vm_guest_info" "slow_network" {
  vm_name             = "remote-vm"
  truenas_host        = "remote.example.com"
  ssh_user            = "root"
  ssh_key_path        = "~/.ssh/truenas_key"
  ssh_timeout_seconds = 30  # Wait up to 30 seconds
}
```

**Improved Error Messages:**
- Authentication failures show specific SSH error details
- Timeout errors clearly indicate connection timeout reached
- Guest agent errors distinguish between agent problems vs connection issues
- sshpass availability check with clear installation instructions

---

## Documentation

### 4. Major Documentation Updates

**New Documentation:**

**[`KNOWN_LIMITATIONS.md`](KNOWN_LIMITATIONS.md)** - Comprehensive limitations guide
- TrueNAS version compatibility (24.04 only)
- VM IP discovery limitations with detailed explanations
- Static IP configuration workarounds and alternatives
- Network configuration constraints
- Summary table of all limitations with severity ratings
- Clear workarounds for each limitation

**Updated Documentation:**

**[`README.md`](README.md)**
- Added prominent "Known Limitations" section near the top
- Updated features list with VM lifecycle management
- Updated resources list to include [`truenas_vm_device`](internal/provider/resource_vm_device.go)
- Added link to [`KNOWN_LIMITATIONS.md`](KNOWN_LIMITATIONS.md)
- Enhanced feature highlights with new capabilities
- Improved quick start examples

**[`VM_IP_DISCOVERY.md`](VM_IP_DISCOVERY.md)**
- Documented [`ssh_strict_host_key_checking`](internal/provider/datasource_vm_guest_info.go) attribute
- Documented [`ssh_timeout_seconds`](internal/provider/datasource_vm_guest_info.go) attribute
- Added authentication validation section with examples
- Expanded troubleshooting section with new error messages
- Added common error messages reference table
- Documented host key verification process step-by-step
- Added sshpass installation instructions for all platforms

**[`API_COVERAGE.md`](API_COVERAGE.md)**
- Updated total resources count: 14 → 15 (added [`truenas_vm_device`](internal/provider/resource_vm_device.go))
- Updated total data sources count: 2 → 10 (corrected previous releases)
- Updated API coverage: ~2.2% → ~2.5%
- Added device management status as implemented
- Updated VM section to show lifecycle management as implemented
- Updated metrics and implementation status

---

## Bug Fixes

### 5. Critical Security and Reliability Fixes

**Authentication Security:**
- **Fixed**: Missing authentication validation before guest agent queries
  - **Before**: Provider attempted guest agent query with invalid credentials, leading to confusing timeout errors
  - **After**: Clear authentication error message immediately with specific SSH error details
  - **Impact**: Faster debugging (saves 10+ seconds per misconfiguration), clearer error messages

**SSH Connection Reliability:**
- **Fixed**: No timeout configuration options
  - **Before**: Hard-coded 10-second timeout could cause failures on slow networks or busy systems
  - **After**: Configurable via `ssh_timeout_seconds` attribute (default: 10)
  - **Impact**: Works reliably across different network conditions and system loads

- **Fixed**: No host key verification options
  - **Before**: Always accepted any host key (potential MITM vulnerability)
  - **After**: Optional strict host key checking via `ssh_strict_host_key_checking` attribute
  - **Impact**: Better security posture for production environments

**Error Handling:**
- **Improved**: Better diagnostics for common issues
  - Authentication failures now show specific SSH error (e.g., "Permission denied", "Host key verification failed")
  - Timeout errors clearly indicate connection timeout vs guest agent timeout
  - Guest agent errors distinguish between agent not installed vs agent communication issues
  - sshpass availability check with installation instructions if missing

---

## Breaking Changes

**None** - This release is fully backward compatible with v0.2.17.

### Behavior Changes (Non-Breaking)

**`start_on_create` Deprecation:**
- **Status**: Deprecated but still fully functional
- **Timeline**: Will be removed in v0.3.0 (3+ months notice)
- **Migration Path**: Use `desired_state = "RUNNING"` instead
- **Reason**: `desired_state` provides declarative state management and drift detection

**Default SSH Behavior:**
- **New Attributes**: Both `ssh_strict_host_key_checking` and `ssh_timeout_seconds` are optional with safe defaults
- **Existing Configurations**: Continue to work unchanged (default: permissive, 10-second timeout)
- **No State Changes**: Existing resources are unaffected

---

## Technical Details

### Files Changed

**New Files:**
- [`internal/provider/resource_vm_device.go`](internal/provider/resource_vm_device.go) - New resource implementation (500+ lines)
- [`examples/resources/truenas_vm_device/resource.tf`](examples/resources/truenas_vm_device/resource.tf) - Example configurations

**Modified Files:**
- [`internal/provider/resource_vm.go`](internal/provider/resource_vm.go) - Added `desired_state` attribute and state management logic
- [`internal/provider/datasource_vm_guest_info.go`](internal/provider/datasource_vm_guest_info.go) - Enhanced with security options and validation
- [`internal/provider/provider.go`](internal/provider/provider.go) - Registered [`truenas_vm_device`](internal/provider/resource_vm_device.go) resource
- [`README.md`](README.md) - Updated with limitations and new features
- [`VM_IP_DISCOVERY.md`](VM_IP_DISCOVERY.md) - Documented new security features
- [`API_COVERAGE.md`](API_COVERAGE.md) - Updated coverage statistics
- [`KNOWN_LIMITATIONS.md`](KNOWN_LIMITATIONS.md) - New comprehensive limitations guide

### Implementation Details

**VM Device Resource:**
- Full CRUD operations using TrueNAS REST API endpoints
- Device types mapped to API dtype parameter
- Attribute validation based on device type
- Import support using `vm_id:device_id` format
- Order management for boot priority

**VM Lifecycle Management:**
- State tracking using TrueNAS VM status API
- Automatic state transition logic in Create, Update, and Read operations
- API endpoints: `/vm/id/{id}/start`, `/vm/id/{id}/stop`, `/vm/id/{id}/suspend`
- Drift detection on every refresh

**Guest Info Security:**
- Pre-flight authentication validation using SSH test connection
- Configurable timeout using SSH `-o ConnectTimeout` option
- Host key checking using SSH `-o StrictHostKeyChecking` option
- Improved error parsing from SSH command output

### API Endpoints Used

**New Endpoints:**
- `POST /vm/id/{id}/device` - Create VM device
- `GET /vm/id/{id}/device` - List VM devices
- `GET /vm/device/id/{device_id}` - Get device details
- `PUT /vm/device/id/{device_id}` - Update device
- `DELETE /vm/device/id/{device_id}` - Delete device

**Enhanced Usage:**
- `POST /vm/id/{id}/start` - Start VM (used by desired_state)
- `POST /vm/id/{id}/stop` - Stop VM (used by desired_state)
- `POST /vm/id/{id}/suspend` - Suspend VM (used by desired_state)
- `GET /vm/id/{id}` - Get VM status (enhanced state tracking)

---

## Compatibility

- **TrueNAS Version**: Scale 24.04 (REST API)
- **Terraform Version**: >= 1.0
- **Go Version**: >= 1.21 (for building from source)

**Important**: TrueNAS Scale 25.04+ uses WebSocket/JSON-RPC and is **NOT** compatible with this provider.

---

## Installation

### Using from Terraform Registry

```hcl
terraform {
  required_providers {
    truenas = {
      source  = "baladithyab/truenas"
      version = "~> 0.2.18"
    }
  }
}

provider "truenas" {
  base_url = "http://your-truenas-server:81"
  api_key  = var.truenas_api_key
}
```

### Building from Source

```bash
git clone https://github.com/baladithyab/terraform-truenas-scale-24.04.git
cd terraform-truenas-scale-24.04
git checkout v0.2.18
make build
make install
```

---

## Upgrade Guide

### From v0.2.17 to v0.2.18

**No Breaking Changes** - This is a feature and enhancement release.

**Steps:**

1. **Update Provider Version:**
   ```hcl
   terraform {
     required_providers {
       truenas = {
         source  = "baladithyab/truenas"
         version = "~> 0.2.18"
       }
     }
   }
   ```

2. **Upgrade Provider:**
   ```bash
   terraform init -upgrade
   ```

3. **Verify Configuration:**
   ```bash
   terraform plan
   ```

4. **Apply Changes:**
   ```bash
   terraform apply
   ```

### Migration: `start_on_create` → `desired_state`

**Recommended** (but not required - `start_on_create` still works):

```hcl
# Before (v0.2.17)
resource "truenas_vm" "example" {
  name            = "my-vm"
  memory          = 4096
  vcpus           = 2
  start_on_create = true
}

# After (v0.2.18)
resource "truenas_vm" "example" {
  name          = "my-vm"
  memory        = 4096
  vcpus         = 2
  desired_state = "RUNNING"  # More explicit and supports drift detection
}
```

**Benefits of Migration:**
- Explicit state declaration (clearer intent)
- Automatic drift detection and correction
- Support for STOPPED and SUSPENDED states
- Better integration with infrastructure automation

**Migration Steps:**
1. Replace `start_on_create = true` with `desired_state = "RUNNING"`
2. Remove `start_on_create = false` (default behavior is maintained)
3. Run `terraform plan` to verify no changes
4. Run `terraform apply` to update state

### Optional: Enable Strict Host Key Checking

For production environments, enable strict host key checking:

```hcl
data "truenas_vm_guest_info" "example" {
  vm_name                      = "my-vm"
  truenas_host                 = "10.0.0.83"
  ssh_user                     = "root"
  ssh_key_path                 = "~/.ssh/truenas_key"
  ssh_strict_host_key_checking = true  # Add this line for production
  ssh_timeout_seconds          = 30    # Optional: increase if needed
}
```

**Preparation:**
```bash
# Add TrueNAS host key to known_hosts before enabling strict checking
ssh-keyscan -H 10.0.0.83 >> ~/.ssh/known_hosts

# Verify host key is added
ssh-keygen -F 10.0.0.83
```

### What to Expect After Upgrade

**New Capabilities:**
- Create/manage VM devices independently using [`truenas_vm_device`](internal/provider/resource_vm_device.go)
- Control VM power state declaratively using `desired_state`
- Better SSH authentication feedback with validation
- Configurable SSH timeouts for slow networks
- Optional strict host key checking for production

**No State Changes:**
- Existing VMs continue to work unchanged
- Existing guest info data sources work with new defaults
- No resource recreation required
- No manual intervention needed

---

## Testing

### Tested Scenarios

**VM Device Management:**
- ✅ Create NIC devices on existing VMs
- ✅ Create DISK devices on running VMs
- ✅ Create CDROM devices for ISO mounting
- ✅ Create PCI (GPU) passthrough devices
- ✅ Create DISPLAY devices with custom settings
- ✅ Update device attributes (order, type, etc.)
- ✅ Delete devices from running VMs
- ✅ Import existing devices

**VM Lifecycle:**
- ✅ VMs with `desired_state = "RUNNING"` start automatically
- ✅ VMs with `desired_state = "STOPPED"` stop automatically
- ✅ VMs with `desired_state = "SUSPENDED"` suspend automatically
- ✅ State drift detection and correction
- ✅ Backward compatibility with `start_on_create`
- ✅ VMs without `desired_state` maintain current state

**Guest Info Security:**
- ✅ SSH authentication validation before guest agent query
- ✅ Strict host key checking when enabled
- ✅ Permissive host key checking by default
- ✅ Custom SSH timeouts (5s, 10s, 30s, 60s tested)
- ✅ Clear error messages for authentication failures
- ✅ Clear error messages for timeout failures
- ✅ Clear error messages for guest agent issues

**Backward Compatibility:**
- ✅ All v0.2.17 configurations work in v0.2.18
- ✅ New attributes are optional with sensible defaults
- ✅ Existing VMs and data sources unaffected
- ✅ No unexpected state changes

---

## Documentation

### Updated Documentation

- **[`README.md`](README.md)** - Main overview with limitations and new features
- **[`CHANGELOG.md`](CHANGELOG.md)** - Complete version history
- **[`KNOWN_LIMITATIONS.md`](KNOWN_LIMITATIONS.md)** - Comprehensive limitations guide (NEW)
- **[`VM_IP_DISCOVERY.md`](VM_IP_DISCOVERY.md)** - Enhanced with security features
- **[`API_COVERAGE.md`](API_COVERAGE.md)** - Updated coverage statistics
- **[`examples/resources/truenas_vm_device/resource.tf`](examples/resources/truenas_vm_device/resource.tf)** - Device examples (NEW)
- **[`examples/resources/truenas_vm/resource.tf`](examples/resources/truenas_vm/resource.tf)** - Updated with desired_state
- **[`examples/data-sources/truenas_vm_guest_info/data-source.tf`](examples/data-sources/truenas_vm_guest_info/data-source.tf)** - Updated with security options

---

## What's Next

### Planned for v0.2.19

**Device Management Enhancements:**
- USB device passthrough improvements
- Display device hot-swap capabilities
- Device validation before creation

**VM Enhancements:**
- Cloud-init support for VM provisioning
- VM cloning capabilities
- Snapshot management for VMs

**Network Enhancements:**
- Network bridge management
- VLAN configuration improvements
- Bond/LAG configuration enhancements

### Planned for v0.3.0 (Major Release)

**New Resources:**
- Replication task management
- Cloud sync task management
- Certificate management
- Service management (start/stop/configure)
- Cron job management

**Breaking Changes:**
- Remove `start_on_create` (deprecated in v0.2.18)
- API endpoint refactoring for better consistency
- Schema improvements for better type safety

**Timeline:** Q1 2026 (3+ months from v0.2.18 release)

---

## Known Limitations

See [`KNOWN_LIMITATIONS.md`](KNOWN_LIMITATIONS.md) for the complete guide. Key limitations:

1. **TrueNAS Version Compatibility**
   - Only compatible with TrueNAS Scale 24.04 (REST API)
   - **NOT** compatible with TrueNAS Scale 25.04+ (WebSocket/JSON-RPC)

2. **VM IP Discovery**
   - Requires QEMU guest agent installed in VM
   - Requires SSH access to TrueNAS host
   - Doesn't work with VMs that don't support guest agent (e.g., Talos Linux)

3. **Network Configuration**
   - Cannot configure static IPs directly in VMs
   - Workaround: Use MAC address + DHCP reservation or cloud-init

4. **Device Updates**
   - Some device types require VM stop/start for updates
   - Hot-add/remove not supported for all device types by TrueNAS

See documentation for workarounds and alternatives.

---

## Support

- **GitHub Issues**: https://github.com/baladithyab/terraform-truenas-scale-24.04/issues
- **Documentation**: https://github.com/baladithyab/terraform-truenas-scale-24.04
- **TrueNAS Version**: Scale 24.04 (REST API)

### Reporting Issues

When reporting issues, please include:
1. TrueNAS Scale version (must be 24.04)
2. Terraform version
3. Provider version (v0.2.18)
4. Full error message and logs
5. Minimal reproduction case

---

## Contributors

- @baladithyab - VM device resource, lifecycle management, guest info enhancements, documentation updates

---

## Acknowledgments

Special thanks to the community for feedback and testing that made this release possible:
- Feature requests for standalone device management
- Bug reports for SSH authentication issues
- Documentation feedback for better clarity

---

**Full Changelog**: https://github.com/baladithyab/terraform-truenas-scale-24.04/blob/main/CHANGELOG.md
