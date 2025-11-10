# Release Notes - v0.2.8

**Release Date**: October 30, 2025  
**Provider**: TrueNAS Terraform Provider  
**Repository**: https://github.com/baladithyab/terraform-provider-truenas  
**Compatibility**: TrueNAS Scale 24.04

---

## 🔧 Critical Bug Fix Release

v0.2.8 is a **critical bug fix release** that resolves issues with NFS shares, snapshot tasks, and VM creation discovered in v0.2.7.

---

## 🐛 What Was Fixed

### Issue 1: NFS Share Creation Fails with Null Values

**Problem in v0.2.7:**
```
Error: Client Error

Unable to create NFS share, got error: API request failed with status 422:
{
  "sharingnfs_create.hosts": [{"message": "null not allowed", "errno": 22}],
  "sharingnfs_create.security": [{"message": "null not allowed", "errno": 22}]
}
```

**Root Cause:**
- TrueNAS API requires `hosts` and `security` fields
- Provider sent `null` when these fields weren't specified
- API rejected null values

**Fix in v0.2.8:**
- Default `hosts` to `[]` (allow all hosts) when not specified
- Default `security` to `[]` (no security restrictions) when not specified

---

### Issue 2: Snapshot Task Creation Fails with JSON Parse Error

**Problem in v0.2.7:**
```
Error: Parse Error

Unable to parse schedule JSON: invalid character '2' after top-level value
```

**Root Cause:**
- Provider expected JSON schedule format
- Users provide cron-style strings like `"0 2 * * *"`
- Provider tried to parse cron string as JSON

**Fix in v0.2.8:**
- Added `parseCronSchedule()` helper to convert cron to JSON
- Added `scheduleToCron()` helper to convert JSON back to cron
- Create/Update functions parse cron strings to JSON
- Read function converts JSON back to cron strings

---

### Issue 3: VM Creation Fails with Empty String Values

**Problem in v0.2.7:**
```
Error: Client Error

Unable to create VM, got error: API request failed with status 422:
{
  "vm_create.cpu_mode": [{"message": "Invalid choice: ", "errno": 22}],
  "vm_create.time": [{"message": "Invalid choice: ", "errno": 22}]
}
```

**Root Cause:**
- Optional fields like `cpu_mode` and `time` were sent as empty strings
- TrueNAS API rejects empty strings for these fields

**Fix in v0.2.8:**
- Only send optional string fields if they have non-empty values
- Same pattern as dataset fixes (v0.2.4)

---

## ✅ Verified Working Behavior

### NFS Shares
```hcl
resource "truenas_nfs_share" "example" {
  path = "/mnt/tank/data"
  # hosts and security default to [] automatically
}
```
**Result**: ✅ NFS share created successfully

### Snapshot Tasks
```hcl
resource "truenas_periodic_snapshot_task" "daily" {
  dataset        = "tank/data"
  schedule       = "0 2 * * *"  # Cron format
  naming_schema  = "auto-%Y%m%d-%H%M"
  lifetime_value = 7
  lifetime_unit  = "DAY"
}
```
**Result**: ✅ Snapshot task created successfully

### VMs
```hcl
resource "truenas_vm" "example" {
  name   = "my-vm"
  memory = 4096
  # cpu_mode and time are optional - omitted if not specified
}
```
**Result**: ✅ VM created successfully

---

## 🔄 Upgrading from v0.2.7

### No Breaking Changes

**Step 1**: Update version:
```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.8"  # Changed from 0.2.7
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

**All v0.2.7 configurations will work in v0.2.8!**

---

## 🔍 Technical Details

### Fix 1: NFS Share Resource

**File**: `internal/provider/resource_nfs_share.go`

**Before (v0.2.7)**:
```go
if !data.Hosts.IsNull() {
    var hosts []string
    data.Hosts.ElementsAs(ctx, &hosts, false)
    createReq["hosts"] = hosts
}
// If null, hosts is not sent - API rejects this!
```

**After (v0.2.8)**:
```go
if !data.Hosts.IsNull() && !data.Hosts.IsUnknown() {
    var hosts []string
    data.Hosts.ElementsAs(ctx, &hosts, false)
    createReq["hosts"] = hosts
} else {
    createReq["hosts"] = []string{}  // Default: allow all hosts
}
```

---

### Fix 2: Snapshot Task Resource

**File**: `internal/provider/resource_periodic_snapshot_task.go`

**Added Helper Functions**:
```go
func parseCronSchedule(cronStr string) (map[string]interface{}, error) {
    parts := strings.Fields(cronStr)
    if len(parts) != 5 {
        return nil, fmt.Errorf("invalid cron format")
    }
    return map[string]interface{}{
        "minute": parts[0],
        "hour":   parts[1],
        "dom":    parts[2],
        "month":  parts[3],
        "dow":    parts[4],
    }, nil
}

func scheduleToCron(schedule map[string]interface{}) string {
    return fmt.Sprintf("%s %s %s %s %s",
        schedule["minute"], schedule["hour"],
        schedule["dom"], schedule["month"], schedule["dow"])
}
```

**Create Function**:
```go
// Before (v0.2.7)
var schedule map[string]interface{}
json.Unmarshal([]byte(data.Schedule.ValueString()), &schedule)

// After (v0.2.8)
schedule, err := parseCronSchedule(data.Schedule.ValueString())
```

**Read Function**:
```go
// Before (v0.2.7)
scheduleJSON, _ := json.Marshal(schedule)
data.Schedule = types.StringValue(string(scheduleJSON))

// After (v0.2.8)
cronStr := scheduleToCron(schedule)
data.Schedule = types.StringValue(cronStr)
```

---

### Fix 3: VM Resource

**File**: `internal/provider/resource_vm.go`

**Before (v0.2.7)**:
```go
if !data.CPUMode.IsNull() {
    createReq["cpu_mode"] = data.CPUMode.ValueString()  // Could be ""
}
```

**After (v0.2.8)**:
```go
if !data.CPUMode.IsNull() && data.CPUMode.ValueString() != "" {
    createReq["cpu_mode"] = data.CPUMode.ValueString()
}
```

---

## 📊 Release Statistics

| Metric | Value |
|--------|-------|
| **Version** | 0.2.8 |
| **Release Type** | Critical Bug Fix |
| **Resources Fixed** | 3 (NFS shares, snapshot tasks, VMs) |
| **Breaking Changes** | 0 |
| **Files Changed** | 4 |
| **Lines Changed** | +103, -23 |
| **Platforms** | 5 |

---

## ⚠️ Important Notice

**All v0.2.7 users should upgrade to v0.2.8 immediately.**

v0.2.7 has critical bugs that prevent NFS shares, snapshot tasks, and VMs from being created. v0.2.8 fixes all three issues.

---

## 🔄 Version History

### v0.2.6 → v0.2.7 → v0.2.8
- **v0.2.6**: Fixed Read() to read all dataset properties
- **v0.2.7**: Fixed integer property parsing for datasets
- **v0.2.8**: Fixed NFS shares, snapshot tasks, and VM creation

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

**🎉 v0.2.8 completes the core resource functionality for datasets, NFS shares, snapshots, and VMs!** 🎉

**Recommendation**: All users should upgrade to v0.2.8 for full resource support.

