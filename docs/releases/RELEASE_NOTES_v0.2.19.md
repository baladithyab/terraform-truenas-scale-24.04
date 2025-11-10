# Release v0.2.19 - Critical Hotfix

**Release Date**: January 8, 2025  
**Type**: Critical Hotfix  
**Status**: ⚠️ URGENT - Upgrade from v0.2.18 immediately

---

## ⚠️ Critical Fix

This is a **critical hotfix** for v0.2.18 that fixes a severe bug causing VM corruption.

### What Was Broken in v0.2.18

The VM Update() function was sending zero values for computed fields (cores, threads, vcpus) to the TrueNAS API, causing:
- Libvirt XML validation errors: "Zero is not permitted"
- VM corruption when updating any VM attribute
- Complete failure of `desired_state` transitions
- VMs becoming unqueryable and unmanageable

### What's Fixed in v0.2.19

- ✅ VM Update() now only sends changed fields to TrueNAS API
- ✅ Computed fields (cores, threads, vcpus, bootloader, cpu_mode) are preserved from state
- ✅ VM state transitions (STOPPED ↔ RUNNING) work correctly
- ✅ No more "Zero cores" errors
- ✅ VM updates no longer cause corruption

---

## 🚀 Upgrading from v0.2.18

**If you're using v0.2.18, upgrade immediately:**

```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/dariusbakunas/truenas"
      version = "~> 0.2.19"  # Update to 0.2.19
    }
  }
}
```

Then run:
```bash
terraform init -upgrade
terraform plan  # Review changes
terraform apply # Fix any corrupted VMs
```

### If You Have Corrupted VMs from v0.2.18

If v0.2.18 corrupted your VMs (showing "Zero cores" errors):

1. Upgrade to v0.2.19
2. Run `terraform apply` - the fix will restore valid CPU topology
3. If VMs are still corrupted, you may need to delete them manually via TrueNAS UI

---

## 📋 Technical Details

### Root Cause

The v0.2.18 Update() function built API payloads with all VM attributes, including computed fields that had zero/empty values in the plan. TrueNAS rejected these updates, leaving VMs in corrupted state.

### The Fix

Modified [`internal/provider/resource_vm.go:534`](internal/provider/resource_vm.go:534) to:
- Compare plan vs state to detect actual changes
- Only send fields that have changed
- Preserve computed values when not explicitly set
- Validate field values before sending to API

### Verification

Comprehensive testing performed (see [`TEST_REPORT_v0.2.18_FIX_VERIFICATION.md`](TEST_REPORT_v0.2.18_FIX_VERIFICATION.md)):
- ✅ VM creation: PASS
- ✅ VM update (STOPPED → RUNNING): PASS
- ✅ VM update (RUNNING → STOPPED): PASS  
- ✅ CPU topology preservation: PASS
- ✅ No libvirt XML errors: PASS

---

## 📦 What's in This Release

### Bug Fixes
- **CRITICAL**: Fixed VM Update() sending zero values for computed fields ([#issue](link))
- Fixed libvirt XML "Zero is not permitted" errors
- Fixed VM corruption during state transitions
- Fixed `desired_state` attribute causing VM updates to fail

### Changed Files
- [`internal/provider/resource_vm.go`](internal/provider/resource_vm.go) - Update() function fixed
- [`CHANGELOG.md`](CHANGELOG.md) - v0.2.19 entry added

### No New Features
This is a pure bugfix release. All v0.2.18 features remain:
- `truenas_vm_device` resource ✅
- VM lifecycle with `desired_state` ✅ (now actually works!)
- Enhanced `truenas_vm_guest_info` ✅
- All documentation ✅

---

## ⚙️ Compatibility

- **TrueNAS Scale**: 24.04 (24.04.x)
- **TrueNAS Scale 25.x**: ❌ Not compatible (uses JSON-RPC instead of REST)
- **Terraform**: >= 1.0
- **Go**: 1.21+

---

## 📚 Documentation

All v0.2.18 documentation remains valid:
- [README.md](README.md) - Getting started
- [../guides/VM_IP_DISCOVERY.md](../guides/VM_IP_DISCOVERY.md) - IP discovery methods
- [KNOWN_LIMITATIONS.md](KNOWN_LIMITATIONS.md) - API limitations and workarounds
- [API_COVERAGE.md](API_COVERAGE.md) - Implementation status

---

## 🐛 Known Issues

None. The critical v0.2.18 bug is resolved.

If you encounter any issues with v0.2.19, please [file an issue](https://github.com/baladithyab/terraform-provider-truenas/issues).

---

## 💬 Support

- **Issues**: [GitHub Issues](https://github.com/baladithyab/terraform-provider-truenas/issues)
- **Discussions**: [GitHub Discussions](https://github.com/baladithyab/terraform-provider-truenas/discussions)
- **Changelog**: [CHANGELOG.md](CHANGELOG.md)

---

## 🙏 Thank You

Thank you for your patience with v0.2.18. This hotfix ensures the provider works reliably for VM management on TrueNAS Scale 24.04.