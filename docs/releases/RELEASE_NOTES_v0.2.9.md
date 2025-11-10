# Release Notes - v0.2.9

**Release Date**: October 30, 2025  
**Provider**: TrueNAS Terraform Provider  
**Repository**: https://github.com/baladithyab/terraform-provider-truenas  
**Compatibility**: TrueNAS Scale 24.04

---

## 🔧 Critical Bug Fix Release

v0.2.9 is a **critical bug fix release** that resolves NFS share state tracking issues discovered in v0.2.8.

---

## 🐛 What Was Fixed

### Issue: NFS Share Read() Function Returns Unknown Values

**Problem in v0.2.8:**
```
Error: Provider produced inconsistent result after apply

When applying changes to truenas_nfs_share.example, provider produced an unexpected new value:
.hosts: was unknown, but now known.
.security: was unknown, but now known.
.networks: was unknown, but now known.

This is a bug in the provider, which should be reported in the provider's own issue tracker.
```

**Root Cause:**
- v0.2.8 Read() function only read **4 properties**: path, comment, enabled, ro
- Missing properties: id, networks, hosts, security, maproot_user, mapall_user
- Terraform expected all values to be **known** after apply
- Provider returned **unknown values** for missing properties

**Fix in v0.2.9:**
- Completely rewrote Read() function to parse **ALL 10 NFS share properties**
- Default `hosts` and `security` to empty arrays (not null) to match Create behavior
- Handle null values for optional string properties
- Convert API response types correctly

---

## ✅ Properties Now Read Correctly

### Before v0.2.8 (Only 4 Properties)
```go
// v0.2.8 Read() - INCOMPLETE
if path, ok := result["path"].(string); ok {
    data.Path = types.StringValue(path)
}
if comment, ok := result["comment"].(string); ok {
    data.Comment = types.StringValue(comment)
}
if enabled, ok := result["enabled"].(bool); ok {
    data.Enabled = types.BoolValue(enabled)
}
if ro, ok := result["ro"].(bool); ok {
    data.ReadOnly = types.BoolValue(ro)
}
// Missing: id, networks, hosts, security, maproot_user, mapall_user
```

### After v0.2.9 (All 10 Properties)
```go
// v0.2.9 Read() - COMPLETE
✅ id - Share identifier
✅ path - Share path
✅ comment - Share comment
✅ enabled - Share enabled status
✅ ro (readonly) - Read-only status
✅ networks - Authorized networks list
✅ hosts - Authorized hosts list (defaults to empty array)
✅ security - Security mechanisms list (defaults to empty array)
✅ maproot_user - Root user mapping
✅ mapall_user - All users mapping
```

---

## 🔍 Technical Details

### Key Changes in Read() Function

**1. Parse ID**
```go
if id, ok := result["id"].(float64); ok {
    data.ID = types.StringValue(strconv.Itoa(int(id)))
}
```

**2. Parse List Properties (networks, hosts, security)**
```go
// Example: hosts list
if hosts, ok := result["hosts"].([]interface{}); ok && len(hosts) > 0 {
    hostList := make([]string, len(hosts))
    for i, host := range hosts {
        if hostStr, ok := host.(string); ok {
            hostList[i] = hostStr
        }
    }
    hostValues := make([]types.String, len(hostList))
    for i, host := range hostList {
        hostValues[i] = types.StringValue(host)
    }
    listValue, _ := types.ListValueFrom(ctx, types.StringType, hostValues)
    data.Hosts = listValue
} else {
    // Default to empty list to match Create behavior
    emptyList, _ := types.ListValueFrom(ctx, types.StringType, []types.String{})
    data.Hosts = emptyList
}
```

**3. Handle Optional String Properties**
```go
// Example: maproot_user
if maproot, ok := result["maproot_user"].(string); ok && maproot != "" {
    data.Maproot = types.StringValue(maproot)
} else {
    data.Maproot = types.StringNull()
}
```

---

## 🔄 Upgrading from v0.2.8

### No Breaking Changes

**Step 1**: Update version:
```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.9"  # Changed from 0.2.8
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

**All v0.2.8 configurations will work in v0.2.9!**

---

## ✅ Verified Working Behavior

### NFS Share Creation and State Tracking
```hcl
resource "truenas_nfs_share" "example" {
  path     = "/mnt/tank/data"
  comment  = "Example NFS share"
  enabled  = true
  readonly = false
  # hosts and security default to [] automatically
}
```

**Before v0.2.9**:
```
✅ NFS share created successfully
❌ Terraform reports "unknown values" for hosts, security, networks
❌ State tracking incomplete
```

**After v0.2.9**:
```
✅ NFS share created successfully
✅ All properties tracked correctly in state
✅ No "unknown values" errors
✅ State tracking complete
```

---

## 📊 Release Statistics

| Metric | Value |
|--------|-------|
| **Version** | 0.2.9 |
| **Release Type** | Critical Bug Fix |
| **Resources Fixed** | 1 (NFS shares) |
| **Breaking Changes** | 0 |
| **Files Changed** | 2 |
| **Lines Changed** | +118 |
| **Platforms** | 5 |

---

## ⚠️ Important Notice

**All v0.2.8 users should upgrade to v0.2.9 immediately.**

v0.2.8 has a critical bug that prevents proper state tracking for NFS shares. v0.2.9 fixes this issue.

---

## 🔄 Version History

### v0.2.6 → v0.2.7 → v0.2.8 → v0.2.9
- **v0.2.6**: Fixed dataset Read() to read all properties
- **v0.2.7**: Fixed integer property parsing for datasets
- **v0.2.8**: Fixed NFS shares, snapshot tasks, and VM creation
- **v0.2.9**: Fixed NFS share Read() to read all properties

**Pattern Recognition**: This is the **same issue** as v0.2.6 dataset fix!
- v0.2.5 datasets: Read() only read 4 properties → Fixed in v0.2.6
- v0.2.8 NFS shares: Read() only read 4 properties → Fixed in v0.2.9

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

**🎉 v0.2.9 completes the NFS share state tracking functionality!** 🎉

**Recommendation**: All users should upgrade to v0.2.9 for proper NFS share state tracking.

