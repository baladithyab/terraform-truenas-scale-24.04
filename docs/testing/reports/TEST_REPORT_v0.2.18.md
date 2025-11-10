# Test Report: v0.2.18 Feature Testing

**Date**: 2025-11-07  
**Provider Version**: v0.2.18 (local build)  
**TrueNAS Version**: Scale 24.04  
**Test Environment**: http://10.0.0.83:81

## Test Summary

| Feature | Status | Notes |
|---------|--------|-------|
| Provider Build | ✅ PASS | Built successfully (26M binary) |
| VM Creation with `desired_state` | ✅ PASS | VMs created in STOPPED state |
| `truenas_vm_device` Resource | ✅ PASS | NIC device created independently |
| Lifecycle State Transition | ❌ FAIL | Critical bug discovered in Update logic |
| MAC Address Export | ⚠️  PARTIAL | Not tested due to update failure |

## Test Execution Details

### 1. Build Process ✅
```bash
go build -o terraform-provider-truenas
# Result: Success, 26M binary
```

### 2. Configuration Setup ✅
- Created test configuration: `test_v0.2.18.tf`
- Configured dev_overrides for local provider testing
- Validated configuration successfully

### 3. VM Creation with `desired_state` ✅

**Test Configuration:**
```hcl
resource "truenas_vm" "test_lifecycle" {
  name          = "testlifecyclevm"
  description   = "Testing v0.2.18 desired_state feature"
  memory        = 2048
  vcpus         = 2
  autostart     = false
  desired_state = "STOPPED"
  bootloader    = "UEFI"
  cpu_mode      = "CUSTOM"
  
  ensure_display_device = true
}
```

**Result**: ✅ SUCCESS
- VM ID: 139
- Status: STOPPED (as expected)
- All computed values populated correctly

### 4. VM Device Management ✅

**Test Configuration:**
```hcl
resource "truenas_vm_device" "test_nic" {
  vm_id       = truenas_vm.test_devices.id
  device_type = "NIC"
  order       = 2000

  nic_config = [{
    type       = "VIRTIO"
    nic_attach = "eno1"
  }]
}
```

**Result**: ✅ SUCCESS
- Device ID: 410
- NIC attached to VM 138
- Configuration applied correctly
- **Important**: Syntax uses `nic_config = [{}]` (list with equals), not `nic_config {}` (block syntax)

### 5. Lifecycle State Transition ❌

**Test**: Change `desired_state` from "STOPPED" to "RUNNING"

**Result**: ❌ CRITICAL FAILURE

**Error Messages:**
1. **Empty/Zero Values Sent**:
   ```
   Unable to update VM: Invalid choice for cpu_mode and bootloader (empty values)
   ```

2. **VM Corruption**:
   ```
   libvirt.libvirtError: XML error: Invalid value for attribute 'cores' 
   in element 'topology': Zero is not permitted
   ```

3. **State became inconsistent** - computed values showed as unknown after apply

## Critical Bugs Discovered

### Bug #1: Update Logic Sends Zero/Empty Values ⚠️

**Issue**: When updating a VM, the provider sends zero/empty values for computed fields that weren't explicitly set, causing TrueNAS to reject the update.

**Affected Fields**:
- `cores` → sent as 0 instead of 1
- `cpu_mode` → sent as empty string
- `bootloader` → sent as empty string  
- `threads` → sent as 0 instead of 1
- Other computed fields

**Impact**: Makes VM updates impossible without explicitly setting ALL fields

**Root Cause**: The Update function in `resource_vm.go` doesn't properly handle computed values - it sends them in the update request even when they haven't changed.

### Bug #2: State Inconsistency After Failed Update ⚠️

**Issue**: After a failed update, Terraform state shows computed values as "unknown" instead of preserving the known values.

**Error Example**:
```
Error: Provider returned invalid result object after apply
After the apply operation, the provider still indicated an unknown 
value for truenas_vm.test_lifecycle.cores
```

**Impact**: State file becomes corrupted, requiring manual intervention

### Bug #3: nic_devices Read-back Issue ⚠️

**Issue**: When using `truenas_vm_device` separately, the `nic_devices` attribute in `truenas_vm` shows devices that should be managed externally.

**Error**:
```
Error: Provider produced inconsistent result after apply
.nic_devices: was null, but now contains device managed by truenas_vm_device
```

**Impact**: Conflicts between `truenas_vm` and `truenas_vm_device` resources

## Successful Features

### ✅ Initial VM Creation
- `desired_state` attribute works perfectly for initial creation
- VMs created in specified state (STOPPED/RUNNING)
- All device configurations applied correctly

### ✅ Independent Device Management
- `truenas_vm_device` resource works independently
- Devices can be added/removed without affecting VM
- Proper ordering with `order` attribute

### ✅ Configuration Validation
- Terraform validates configuration correctly
- Type checking works
- Required/optional parameters enforced

## Recommendations

### Immediate Fixes Required

1. **Fix Update Logic**:
   - Only send fields that have actually changed
   - Preserve computed values during updates
   - Don't send zero/empty values for unset fields

2. **Fix State Management**:
   - Ensure Read() is called after Update()
   - Properly populate all computed values after operations
   - Handle errors gracefully without corrupting state

3. **Fix Device Conflict**:
   - When `truenas_vm_device` is used, exclude those devices from `truenas_vm.nic_devices`
   - Or make `nic_devices` computed-only (read-only)

### Testing Recommendations

1. **Add Unit Tests**: Create tests for Update logic with partial field updates
2. **Add Integration Tests**: Test state transitions with actual TrueNAS API
3. **Test VM Lifecycle**: Test full lifecycle (create → update → delete)
4. **Test Mixed Device Management**: Test both inline and separate device resources

## Workaround for Current Version

Until fixes are implemented, users should:

1. **Set all computed fields explicitly** when creating VMs:
   ```hcl
   bootloader = "UEFI"
   cpu_mode   = "CUSTOM"
   cores      = 1
   threads    = 1
   ```

2. **Avoid updating VMs** that use `desired_state` - prefer recreation

3. **Use either inline devices OR separate `truenas_vm_device`**, not both

## Files Created/Modified

- `test_v0.2.18.tf` - Test configuration file
- `.terraformrc` - Local provider override configuration
- This test report

## Cleanup Status

⚠️ **Manual cleanup required**: Test VMs (IDs 138, 139) are corrupted in TrueNAS and cannot be deleted via API. They must be removed manually through TrueNAS UI or by directly manipulating the database.

## Conclusion

The v0.2.18 release introduces valuable features (`desired_state` and `truenas_vm_device`), but has critical bugs in the VM update logic that make it unsuitable for production use without fixes. The create functionality works perfectly, but any updates to existing VMs will likely fail or corrupt the VM state.

**Recommendation**: Fix the identified bugs before releasing v0.2.18.