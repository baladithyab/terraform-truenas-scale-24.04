# TrueNAS Terraform Provider v0.2.16 Release Notes

Release Date: 2025-11-04

## Overview

Version 0.2.16 introduces display device support for VMs, fixes a critical device ordering bug, and improves project organization.

---

## Features

### Display Device Support for VMs

Added full support for SPICE and VNC display devices, enabling console access to virtual machines:

- **SPICE and VNC Support**: Configure display devices with full control over ports, bindings, and passwords
- **Web Access**: Enable web-based console access with configurable ports
- **Resolution Control**: Set custom display resolutions (e.g., 1024x768, 1920x1080)
- **Boot Order Integration**: Display devices can be included in boot order configuration

**Example Usage:**
```hcl
resource "truenas_vm" "example" {
  name   = "my-vm"
  memory = 4096

  display_devices = [{
    type       = "SPICE"
    port       = 5904
    bind       = "0.0.0.0"
    password   = "secure123"
    web        = true
    web_port   = 5905
    resolution = "1920x1080"
    wait       = false
  }]
}
```

**Attributes:**
- `type` - Display type: SPICE or VNC (default: SPICE)
- `port` - Port number for the display server (e.g., 5900 for VNC, 5902 for SPICE)
- `bind` - IP address to bind to (default: 0.0.0.0)
- `password` - Password for display access (optional, sensitive)
- `web` - Enable web access (default: true)
- `web_port` - Port for web access
- `resolution` - Display resolution (e.g., 1024x768, 1920x1080)
- `wait` - Wait for client connection before starting VM
- `order` - Boot order for this device

---

## Bug Fixes

### Fixed Device Order Inconsistency Error

**Issue:** When creating VMs with devices, Terraform would report "Provider produced inconsistent result after apply" errors due to TrueNAS adjusting device order values.

**Example Error:**
```
Error: Provider produced inconsistent result after apply

When applying changes to truenas_vm.talos_minimal, provider
"provider["registry.terraform.io/baladithyab/truenas"]" produced an
unexpected new value: .disk_devices[0].order: was cty.NumberIntVal(1001),
but now cty.NumberIntVal(1002).
```

**Root Cause:** Device order values are server-side assigned and may be adjusted by TrueNAS based on internal logic (e.g., when `ensure_display_device` creates an automatic display device).

**Fix:** Added `UseStateForUnknown()` plan modifier to all device `order` attributes:
- `nic_devices[].order`
- `disk_devices[].order`
- `cdrom_devices[].order`
- `display_devices[].order`
- `pci_devices[].order`

This ensures Terraform accepts server-side order adjustments without raising inconsistency errors.

**Impact:** VM creation and updates with device ordering will now work reliably without manual state manipulation.

---

## Project Organization

### Examples Directory Restructure

Moved test examples to the main examples directory for better discoverability:
- `test-talos-minimal` → `examples/talos-minimal`
- `test-boot-order` → `examples/boot-order`
- `test-volsize` → `examples/volsize`

These examples now join the existing comprehensive examples:
- `examples/complete` - Full-featured VM with all options
- `examples/vm-gpu-passthrough` - GPU passthrough configuration
- `examples/vm-ip-discovery` - IP discovery patterns
- `examples/data-sources` - Data source usage examples

---

## New Example: Talos Minimal VM

Added a minimal Talos Linux VM example demonstrating:
- Boot order configuration (CDROM first, then disk)
- AHCI disk type usage
- SPICE display device setup
- Network bridge configuration
- UEFI bootloader setup

Located at: [examples/talos-minimal/main.tf](examples/talos-minimal/main.tf)

This example is particularly useful for:
- Testing boot order behavior
- Validating minimal VM configurations
- Kubernetes node provisioning with Talos

---

## Technical Details

### Changed Files
- `internal/provider/resource_vm.go` - Display device support + order plan modifiers
- `Makefile` - Version bump to 0.2.16
- Examples reorganization

### Compatibility
- **Terraform**: >= 1.0.0
- **TrueNAS SCALE**: 22.12+ (24.04+ recommended)
- **Go**: 1.21+

---

## Upgrade Notes

### From v0.2.15

No breaking changes. Simply update your provider version:

```hcl
terraform {
  required_providers {
    truenas = {
      source  = "baladithyab/truenas"
      version = "~> 0.2.16"
    }
  }
}
```

Then run:
```bash
terraform init -upgrade
```

### Order Attribute Behavior

If you were experiencing the order inconsistency error, you should:
1. Upgrade to v0.2.16
2. Remove any manual state fixes
3. Run `terraform apply` - the issue will be automatically resolved

---

## What's Next?

### Upcoming in v0.2.17+
- Enhanced VM lifecycle management
- Improved error handling and validation
- Additional device type support
- Performance optimizations

---

## Contributors

- [@baladithyab](https://github.com/baladithyab)

## Resources

- [GitHub Repository](https://github.com/baladithyab/terraform-provider-truenas)
- [Terraform Registry](https://registry.terraform.io/providers/baladithyab/truenas)
- [Issue Tracker](https://github.com/baladithyab/terraform-provider-truenas/issues)

---

**Full Changelog**: [v0.2.15...v0.2.16](https://github.com/baladithyab/terraform-provider-truenas/compare/v0.2.15...v0.2.16)
