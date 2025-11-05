# TrueNAS Terraform Provider v0.2.17 Release Notes

Release Date: 2025-11-04

## Overview

Version 0.2.17 enhances the VM resource with intelligent memory ballooning defaults, preventing common memory allocation errors.

---

## Features

### Intelligent min_memory Default

The `min_memory` attribute now automatically defaults to the `memory` value when not explicitly specified, disabling memory ballooning by default.

**Why This Matters:**
- Prevents "virtio_balloon: Out of puff!" errors in VMs
- Simplifies VM configuration - no need to specify min_memory in most cases
- Memory ballooning is now opt-in rather than opt-out
- More predictable VM memory behavior

**Before (v0.2.16 and earlier):**
```hcl
resource "truenas_vm" "example" {
  name       = "my-vm"
  memory     = 8192  # 8GB max
  min_memory = 8192  # Had to explicitly set to disable ballooning
  vcpus      = 2
}
```

**After (v0.2.17):**
```hcl
resource "truenas_vm" "example" {
  name   = "my-vm"
  memory = 8192  # 8GB - min_memory automatically set to 8192
  vcpus  = 2
}
```

**To Enable Memory Ballooning (Optional):**
```hcl
resource "truenas_vm" "flexible" {
  name       = "flexible-vm"
  memory     = 8192   # 8GB max
  min_memory = 2048   # 2GB min - enables ballooning
  vcpus      = 2
}
```

---

## Bug Fixes

None - this is an enhancement release.

---

## Breaking Changes

None - fully backward compatible with v0.2.16.

**Behavior Change (Non-Breaking):**
- VMs created without explicit `min_memory` will now have min_memory = memory
- This is a safer default that prevents memory ballooning errors
- Existing VMs with explicit `min_memory` values are unaffected
- Existing VMs without `min_memory` will see no change in behavior

---

## Technical Details

### Files Changed
- `internal/provider/resource_vm.go` - Enhanced min_memory handling in Create() and Update()
- `Makefile` - Version bump to 0.2.17

### Implementation
- Added plan modifier `UseStateForUnknown()` to min_memory attribute
- Updated Create() function to default min_memory to memory value
- Updated Update() function to default min_memory to memory value
- Enhanced documentation to reflect new default behavior

---

## Upgrade Guide

### From v0.2.16 to v0.2.17

**No Breaking Changes** - This is an enhancement release.

**Steps:**
1. Update version in `terraform` block to `~> 0.2.17`
2. Run `terraform init -upgrade`
3. Run `terraform plan` to verify
4. Run `terraform apply`

**What to Expect:**
- New VMs without `min_memory` will have it automatically set to `memory` value
- Existing VMs will continue to work as before
- No state changes for existing resources

---

## Compatibility

- **TrueNAS Version**: Scale 24.04 (REST API)
- **Terraform Version**: >= 1.0
- **Go Version**: >= 1.21 (for building from source)

---

## Installation

### Using from GitHub

```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.17"
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
git checkout v0.2.17
make build
make install
```

---

## Testing

Tested with:
- ✅ VMs without min_memory specified (defaults to memory value)
- ✅ VMs with explicit min_memory (uses specified value)
- ✅ Memory ballooning enabled (min_memory < memory)
- ✅ Memory ballooning disabled (min_memory = memory)
- ✅ Backward compatibility with v0.2.16 configurations

---

## Documentation

- **README.md** - Main overview and quick start
- **CHANGELOG.md** - Complete version history
- **examples/resources/truenas_vm/resource.tf** - Updated examples
- **RELEASE_NOTES_v0.2.17.md** - This file

---

## What's Next

Planned for future releases:
- Device update support (add/remove devices from existing VMs)
- USB device passthrough
- PCI device passthrough improvements
- Network bridge management
- Cloud-init support

---

## Support

- **GitHub Issues**: https://github.com/baladithyab/terraform-truenas-scale-24.04/issues
- **Documentation**: https://github.com/baladithyab/terraform-truenas-scale-24.04
- **TrueNAS Version**: Scale 24.04 (REST API)

---

## Contributors

- @baladithyab - min_memory default enhancement

---

**Full Changelog**: https://github.com/baladithyab/terraform-truenas-scale-24.04/blob/main/CHANGELOG.md

