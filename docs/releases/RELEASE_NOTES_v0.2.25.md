# Release Notes - v0.2.25

**Release Date:** November 19, 2024

## Overview

This release improves the cloud-init device ordering logic by implementing automatic order calculation, removing the hardcoded default value of 10000.

## Features

### Cloud-Init Device Order Auto-Increment

**Enhanced cloud-init device ordering with intelligent auto-increment logic**

Previously, cloud-init ISO devices were created with a hardcoded default `device_order` of 10000 when not explicitly specified. This release implements smart auto-incrementing:

- **Auto-calculation**: When `device_order` is not provided in configuration, the system automatically calculates the next available order number as `max(existing_device_orders) + 1`
- **Default fallback**: If no devices exist on the VM, defaults to 1000 (instead of the previous 10000)
- **User control preserved**: Users can still explicitly specify `device_order` to override the auto-increment behavior

**Benefits:**
- More predictable device ordering
- Eliminates unnecessarily high order values
- Maintains backward compatibility with explicit ordering

**Example:**
```hcl
resource "truenas_vm" "example" {
  name   = "my-vm"
  memory = 2048
  
  cloud_init = {
    user_data = file("user-data.yaml")
    meta_data = file("meta-data.yaml")
    # device_order is now optional and auto-calculated
  }
}
```

## Technical Changes

### Modified Files

- **[`internal/provider/resource_vm.go`](../../internal/provider/resource_vm.go)**
  - Added [`getMaxDeviceOrder()`](../../internal/provider/resource_vm.go:784) helper function to retrieve maximum device order from existing VM devices
  - Updated [`handleCloudInitCreate()`](../../internal/provider/resource_vm.go:813) to implement auto-increment logic
  - Updated schema description for `device_order` attribute to reflect new behavior
  
- **[`main.go`](../../main.go:25)**
  - Version bumped from `0.2.24` to `0.2.25`

### API Verification

Confirmed via TrueNAS API specification ([`docs/api/openapi.json`](../../docs/api/openapi.json)) that the `order` field is **optional** for both `vm_create` and `vm_device_create` operations, validating this implementation approach.

## Breaking Changes

None. This is a backward-compatible enhancement.

## Upgrade Notes

No special upgrade steps required. Existing configurations will continue to work:
- Configurations with explicit `device_order` values will use those values unchanged
- Configurations without `device_order` will now benefit from intelligent auto-increment instead of hardcoded 10000

## Known Issues

None specific to this release.

---

**Full Changelog:** [v0.2.24...v0.2.25](https://github.com/baladithyab/terraform-provider-truenas/compare/v0.2.24...v0.2.25)