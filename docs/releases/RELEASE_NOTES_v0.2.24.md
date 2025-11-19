# Release Notes v0.2.24 - Configurable Cloud-Init Device Order

**Release Date:** 2025-11-19
**Type:** Enhancement Release

## 🔧 Enhancement: Configurable Cloud-Init Device Order

This release adds the ability to customize the boot order of Cloud-Init ISO devices, removing the previous hardcoded limitation.

### What's New

#### New `device_order` Parameter
The `cloud_init` block now supports an optional `device_order` parameter that allows you to control when the Cloud-Init ISO boots in relation to other devices.

**Key Benefits:**
- **Flexibility**: Override the default boot order (10000) to suit your needs
- **Control**: Ensure Cloud-Init ISO boots at the desired point in the boot sequence
- **Backward Compatible**: Defaults to 10000 if not specified, maintaining existing behavior

### Configuration Options

The `cloud_init.device_order` parameter:
- Type: `Number` (Optional)
- Default: `10000`
- Purpose: Determines boot priority (lower values boot first)
- Use Case: Customize when Cloud-Init configuration is applied during boot

### Example Usage

#### Default Behavior (Unchanged)
```terraform
resource "truenas_vm" "ubuntu_cloud" {
  name   = "ubuntu-cloud-init"
  vcpus  = 2
  memory = 4096

  cloud_init {
    user_data = <<EOF
#cloud-config
hostname: ubuntu-server
# ... configuration ...
EOF
    # device_order defaults to 10000
  }
}
```

#### Custom Device Order
```terraform
resource "truenas_vm" "ubuntu_cloud_custom" {
  name   = "ubuntu-cloud-custom"
  vcpus  = 2
  memory = 4096

  cloud_init {
    user_data = <<EOF
#cloud-config
hostname: ubuntu-server
# ... configuration ...
EOF
    
    # Boot Cloud-Init earlier in the sequence
    device_order = 5000
  }
}
```

### Implementation Details

- **Schema Update**: Added `device_order` as an optional computed Int64 attribute to the `cloud_init` block
- **Logic Update**: Modified [`handleCloudInitCreate()`](../../internal/provider/resource_vm.go:774) to use the configurable value instead of hardcoded `10000`
- **Documentation**: Updated resource documentation with parameter details and examples

### 📚 Documentation Updates

- Enhanced `truenas_vm` resource documentation
- Added example showing `device_order` usage
- Updated Cloud-Init configuration reference

### 🔄 Migration Notes

This release is fully backward compatible:
- Existing configurations without `device_order` will continue to work unchanged
- The default value of `10000` ensures Cloud-Init ISO boots after regular devices
- No action required for existing deployments

---

**Download:** Available from the Terraform Registry and GitHub Releases
**Documentation:** [Complete Documentation](../../docs/)
**Support:** [GitHub Issues](https://github.com/baladithyab/terraform-provider-truenas/issues)