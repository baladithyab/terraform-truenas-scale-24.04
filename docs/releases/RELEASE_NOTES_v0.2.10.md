# Release Notes - v0.2.10

**Release Date**: October 30, 2025  
**Provider**: TrueNAS Terraform Provider  
**Repository**: https://github.com/baladithyab/terraform-provider-truenas  
**Compatibility**: TrueNAS Scale 24.04

---

## 🔧 Critical Bug Fix Release

v0.2.10 is a **critical bug fix release** that resolves VM state tracking issues discovered in v0.2.9.

---

## 🐛 What Was Fixed

### Issue: VM Read() Function Returns Unknown Values

**Problem in v0.2.9:**
```
Error: Provider produced inconsistent result after apply

When applying changes to truenas_vm.talos_worker_muninn_01, provider produced an unexpected new value:
.arch_type: was unknown, but now known.
.cpu_mode: was unknown, but now known.
.cpu_model: was unknown, but now known.
.machine_type: was unknown, but now known.
.min_memory: was unknown, but now known.
.time: was unknown, but now known.

This is a bug in the provider, which should be reported in the provider's own issue tracker.
```

**Root Cause:**
- v0.2.9 Read() function only read **9 properties**: id, name, description, vcpus, cores, threads, memory, autostart, status
- Missing properties: min_memory, bootloader, cpu_mode, cpu_model, machine_type, arch_type, time
- Terraform expected all values to be **known** after apply
- Provider returned **unknown values** for missing properties

**Fix in v0.2.10:**
- Completely rewrote Read() function to parse **ALL 16 VM properties**
- Handle null values for optional properties
- Convert API response types correctly

---

## ✅ Properties Now Read Correctly

### Before v0.2.9 (Only 9 Properties)
```go
// v0.2.9 Read() - INCOMPLETE
if name, ok := result["name"].(string); ok {
    data.Name = types.StringValue(name)
}
if description, ok := result["description"].(string); ok {
    data.Description = types.StringValue(description)
}
if vcpus, ok := result["vcpus"].(float64); ok {
    data.VCPUs = types.Int64Value(int64(vcpus))
}
// ... cores, threads, memory, autostart, status
// Missing: min_memory, bootloader, cpu_mode, cpu_model, machine_type, arch_type, time
```

### After v0.2.10 (All 16 Properties)
```go
// v0.2.10 Read() - COMPLETE
✅ id - VM identifier
✅ name - VM name
✅ description - VM description (optional)
✅ vcpus - Number of virtual CPUs
✅ cores - Number of cores per socket
✅ threads - Number of threads per core
✅ memory - Memory in MB
✅ min_memory - Minimum memory (optional)
✅ autostart - Autostart on boot
✅ bootloader - Bootloader type (optional)
✅ cpu_mode - CPU mode (optional)
✅ cpu_model - CPU model (optional)
✅ machine_type - Machine type (optional)
✅ arch_type - Architecture type (optional)
✅ time - Time configuration (optional)
✅ status - VM status
```

---

## 🔍 Pattern Recognition

### This is the Third Occurrence of the Same Bug!

**v0.2.6 - Dataset Read() Fix:**
- Problem: Read() only read 4 properties
- Solution: Rewrote to read ALL dataset properties

**v0.2.9 - NFS Share Read() Fix:**
- Problem: Read() only read 4 properties
- Solution: Rewrote to read ALL NFS share properties

**v0.2.10 - VM Read() Fix:**
- Problem: Read() only read 9 properties
- Solution: Rewrote to read ALL VM properties

**Lesson Learned**: When implementing a new resource, the Read() function MUST parse ALL properties from the API response, not just the required ones!

---

## 🔄 Upgrading from v0.2.9

### No Breaking Changes

**Step 1**: Update version:
```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.10"  # Changed from 0.2.9
    }
  }
}
```

**Step 2**: Upgrade:
```bash
terraform init -upgrade
terraform plan
terraform apply
```

**All v0.2.9 configurations will work in v0.2.10!**

---

## ✅ Verified Working Behavior

### VM Creation and State Tracking
```hcl
resource "truenas_vm" "example" {
  name   = "my_vm"
  memory = 4096
  vcpus  = 2
  cores  = 1
  threads = 1
  # Optional properties like cpu_mode, machine_type, etc. are now tracked correctly
}
```

**Before v0.2.10**:
```
✅ VM created successfully
❌ Terraform reports "unknown values" for 6 optional properties
❌ State tracking incomplete
```

**After v0.2.10**:
```
✅ VM created successfully
✅ All 16 properties tracked correctly in state
✅ No "unknown values" errors
✅ State tracking complete
```

---

## 📊 Release Statistics

| Metric | Value |
|--------|-------|
| **Version** | 0.2.10 |
| **Release Type** | Critical Bug Fix |
| **Resources Fixed** | 1 (VMs) |
| **Breaking Changes** | 0 |
| **Files Changed** | 2 |
| **Lines Changed** | +110, -1 |
| **Platforms** | 5 |

---

## ⚠️ Important Notice

**All v0.2.9 users should upgrade to v0.2.10 immediately.**

v0.2.9 has a critical bug that prevents proper state tracking for VMs. v0.2.10 fixes this issue.

---

## 🔄 Version History

### v0.2.6 → v0.2.10 Read() Function Fixes
- **v0.2.6**: Fixed dataset Read() to read all properties
- **v0.2.7**: Fixed integer property parsing for datasets
- **v0.2.8**: Fixed NFS shares, snapshot tasks, and VM creation
- **v0.2.9**: Fixed NFS share Read() to read all properties
- **v0.2.10**: Fixed VM Read() to read all properties

**All core resources now have complete Read() implementations!**

---

## 🎉 Complete State Tracking for All Resources

### v0.2.10 Achievement: Full CRUD Support

| Resource | Create | Read | Update | Delete | State Tracking |
|----------|--------|------|--------|--------|----------------|
| **Datasets** | ✅ | ✅ | ✅ | ✅ | ✅ Complete |
| **NFS Shares** | ✅ | ✅ | ✅ | ✅ | ✅ Complete |
| **Snapshot Tasks** | ✅ | ✅ | ✅ | ✅ | ✅ Complete |
| **VMs** | ✅ | ✅ | ✅ | ✅ | ✅ Complete |

**All core resources now have full CRUD support with complete state tracking!**

---

## 🚀 What's Next

### Planned for v0.3.0
- Replication task management
- Cloud sync task management
- Service management
- Certificate management
- Cron job management

---

## 📞 Support

- **GitHub Issues**: https://github.com/baladithyab/terraform-provider-truenas/issues
- **Repository**: https://github.com/baladithyab/terraform-provider-truenas
- **TrueNAS Version**: Scale 24.04 (REST API)

---

**🎉 v0.2.10 completes the Read() function fixes for all core resources!** 🎉

**All datasets, NFS shares, snapshot tasks, and VMs now have complete state tracking!**

**Recommendation**: All users should upgrade to v0.2.10 for proper VM state tracking and complete infrastructure management.

