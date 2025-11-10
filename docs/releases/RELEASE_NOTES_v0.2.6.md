# Release Notes - v0.2.6

**Release Date**: October 30, 2025  
**Provider**: TrueNAS Terraform Provider  
**Repository**: https://github.com/baladithyab/terraform-provider-truenas  
**Compatibility**: TrueNAS Scale 24.04

---

## 🔧 Critical Bug Fix Release

v0.2.6 is a **critical bug fix release** that resolves an issue discovered in v0.2.5 where the provider's `Read()` function only read a few properties from the TrueNAS API, causing Terraform to report "unknown values" after apply.

---

## 🐛 What Was Fixed

### Critical Issue: Read() Function Not Reading All Properties

**Problem in v0.2.5:**
```
Error: Provider produced inconsistent result after apply

When applying changes to truenas_dataset.data, provider produced an unexpected new value:
.compression: was unknown, but now known.

This is a bug in the provider, which should be reported in the provider's own issue tracker.
```

**Root Cause:**
- v0.2.5 `Read()` function only read 4 properties: comments, compression, atime, volsize
- All other properties were set to null in the state
- Terraform expected all values to be known after apply
- This caused "unknown values" errors and prevented proper state tracking

**Fix in v0.2.6:**
- Completely rewrote `Read()` function to parse and populate **ALL** properties from API response
- Added proper parsing for all shared properties
- Added proper parsing for all FILESYSTEM-specific properties
- Added proper null handling when properties don't exist in API response

**Result:**
✅ All properties are correctly read from TrueNAS API  
✅ Terraform can properly track dataset state  
✅ No more "unknown values" errors  
✅ State file contains all property values

---

## 📊 Properties Now Correctly Read

### Shared Properties (FILESYSTEM and VOLUME)
- ✅ `comments` - User comments
- ✅ `compression` - Compression algorithm
- ✅ `sync` - Sync mode
- ✅ `deduplication` - Deduplication setting
- ✅ `readonly` - Read-only flag
- ✅ `copies` - Number of copies
- ✅ `reservation` - Space reservation
- ✅ `refreservation` - Referenced space reservation

### VOLUME-Specific Properties
- ✅ `volsize` - Volume size

### FILESYSTEM-Specific Properties
- ✅ `atime` - Access time updates
- ✅ `exec` - Execute permission
- ✅ `recordsize` - Record size
- ✅ `quota` - Dataset quota
- ✅ `refquota` - Referenced quota
- ✅ `snapdir` - Snapshot directory visibility

---

## ✅ Verified Working Behavior

### Before v0.2.6 (v0.2.5)
```bash
terraform apply
# Datasets created successfully
# But Terraform reports: "unknown values" error
# State file missing most property values
```

### After v0.2.6
```bash
terraform apply
# Datasets created successfully
# All properties read from API
# State file contains all property values
# No "unknown values" errors
```

---

## 🔄 Upgrading from v0.2.5

### No Breaking Changes

**Step 1**: Update version:
```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.6"  # Changed from 0.2.5
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

**All v0.2.5 configurations will work in v0.2.6!**

---

## 🔍 Technical Details

### Code Changes

**File**: `internal/provider/resource_dataset.go`

**Change: Read() Function (Lines 439-574)**

**Before (v0.2.5)** - Only read 4 properties:
```go
// Read properties valid for BOTH FILESYSTEM and VOLUME
if comments, ok := result["comments"].(map[string]interface{}); ok {
    if value, ok := comments["value"].(string); ok {
        data.Comments = types.StringValue(value)
    }
}
if compression, ok := result["compression"].(map[string]interface{}); ok {
    if value, ok := compression["value"].(string); ok {
        data.Compression = types.StringValue(value)
    }
}
// Only atime and volsize were read - all others set to null!
```

**After (v0.2.6)** - Read ALL properties:
```go
// Read ALL shared properties
if sync, ok := result["sync"].(map[string]interface{}); ok {
    if value, ok := sync["value"].(string); ok {
        data.Sync = types.StringValue(value)
    }
} else {
    data.Sync = types.StringNull()
}

if dedup, ok := result["deduplication"].(map[string]interface{}); ok {
    if value, ok := dedup["value"].(string); ok {
        data.Dedup = types.StringValue(value)
    }
} else {
    data.Dedup = types.StringNull()
}

// ... and so on for ALL properties
```

**Key Improvements:**
- Added parsing for all shared properties
- Added parsing for all FILESYSTEM properties
- Added proper null handling with `else` clauses
- All properties now correctly populated in state

---

## 🧪 Testing Results

### Build
- ✅ Compiles without errors
- ✅ No warnings

### Dataset Operations
- ✅ Create datasets (FILESYSTEM and VOLUME)
- ✅ Read all properties from API
- ✅ Update datasets
- ✅ Delete datasets

### State Tracking
- ✅ All properties correctly stored in state file
- ✅ No "unknown values" errors
- ✅ Terraform can detect drift correctly
- ✅ `terraform plan` shows no changes after apply

---

## 📚 Documentation

- **CHANGELOG.md** - Updated with v0.2.6 section
- **RELEASE_NOTES_v0.2.6.md** - This file
- All existing examples continue to work

---

## 🚀 What's Next

### Planned for v0.3.0
- Replication task management
- Cloud sync task management
- Service management (start/stop/configure)
- Certificate management
- Cron job management

See `../api/API_COVERAGE.md` for the complete roadmap.

---

## 📞 Support

- **GitHub Issues**: https://github.com/baladithyab/terraform-provider-truenas/issues
- **Repository**: https://github.com/baladithyab/terraform-provider-truenas
- **TrueNAS Version**: Scale 24.04 (REST API)

---

## 📊 Release Statistics

| Metric | Value |
|--------|-------|
| **Version** | 0.2.6 |
| **Release Type** | Critical Bug Fix |
| **Resources** | 14 (all working) |
| **Data Sources** | 2 (all working) |
| **Bugs Fixed** | 1 (critical Read() function) |
| **Breaking Changes** | 0 |
| **Binary Size** | ~25MB per platform |

---

## ⚠️ Important Notice

**All v0.2.5 users should upgrade to v0.2.6 immediately.**

v0.2.5 has a critical bug where the `Read()` function doesn't read all properties, causing "unknown values" errors and preventing proper state tracking. v0.2.6 fixes this issue by properly reading all properties from the TrueNAS API.

---

## 🔄 Version History

### v0.2.4 → v0.2.5 → v0.2.6
- **v0.2.4**: Fixed empty string handling for string properties
- **v0.2.5**: Fixed zero value handling for integer properties
- **v0.2.6**: Fixed Read() function to read all properties from API

All three fixes are now in place - datasets can be created AND tracked properly!

---

**🎉 v0.2.6 fixes the Read() function bug and enables proper state tracking!** 🎉

**Recommendation**: All users should upgrade to v0.2.6 for proper state tracking and drift detection.

