# Release Notes - v0.2.4

**Release Date**: October 30, 2025  
**Provider**: TrueNAS Terraform Provider  
**Repository**: https://github.com/baladithyab/terraform-provider-truenas  
**Compatibility**: TrueNAS Scale 24.04

---

## 🔧 Critical Bug Fix Release

v0.2.4 is a **critical bug fix release** that resolves an issue discovered in v0.2.3 where the provider sent empty strings to the TrueNAS API, causing "Invalid choice: " errors.

---

## 🐛 What Was Fixed

### Critical Issue: Empty String Properties Sent to API

**Problem in v0.2.3:**
```hcl
resource "truenas_dataset" "data" {
  name = "tank/data"
  type = "FILESYSTEM"
  # Optional properties not specified, but provider sent empty strings
}
```

**Error in v0.2.3:**
```
Error: Client Error
Unable to create dataset, got error: API returned status 422: 
{
  "pool_dataset_create.compression": [
    {
      "message": "Invalid choice: ",
      "errno": 22
    }
  ]
}
```

**Root Cause:**
- v0.2.3 checked `!IsNull()` to determine if a property should be sent
- However, Terraform sets optional string properties to empty strings (`""`) by default
- The provider was sending these empty strings to the API
- TrueNAS API rejected empty strings with "Invalid choice: " errors

**Fix in v0.2.4:**
- Added empty string check: `&& data.PropertyName.ValueString() != ""`
- Properties are now only included in API requests if they have **non-empty** values
- Applied to all string properties in both `Create()` and `Update()` functions

**Result:**
✅ Datasets can now be created without specifying all optional properties  
✅ Empty string properties are correctly omitted from API requests  
✅ Non-empty properties are sent correctly

---

## 📊 Properties Affected

### String Properties (Fixed)
These properties now require non-empty values to be sent to the API:

- `comments` - User comments
- `compression` - Compression algorithm
- `sync` - Sync mode
- `deduplication` - Deduplication setting
- `readonly` - Read-only flag
- `atime` - Access time updates (FILESYSTEM only)
- `exec` - Execute permission (FILESYSTEM only)
- `recordsize` - Record size (FILESYSTEM only)
- `snapdir` - Snapshot directory visibility (FILESYSTEM only)

### Integer Properties (No Change)
These properties already worked correctly:

- `volsize` - Volume size (VOLUME only)
- `quota` - Dataset quota (FILESYSTEM only)
- `refquota` - Referenced quota (FILESYSTEM only)
- `reservation` - Space reservation
- `refreservation` - Referenced space reservation
- `copies` - Number of copies

---

## ✅ Verified Working Configurations

### Minimal FILESYSTEM Dataset (Now Works!)
```hcl
resource "truenas_dataset" "data" {
  name = "tank/data"
  type = "FILESYSTEM"
  # No optional properties - works perfectly!
}
```

### FILESYSTEM with Some Properties (Now Works!)
```hcl
resource "truenas_dataset" "data" {
  name        = "tank/data"
  type        = "FILESYSTEM"
  compression = "LZ4"
  # Other properties omitted - works perfectly!
}
```

### VOLUME Dataset (Now Works!)
```hcl
resource "truenas_dataset" "vm_disk" {
  name    = "tank/vms/vm01-disk0"
  type    = "VOLUME"
  volsize = 107374182400  # 100GB
  # Optional properties omitted - works perfectly!
}
```

---

## 🔄 Upgrading from v0.2.3

### No Breaking Changes

**Step 1**: Update version:
```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.4"  # Changed from 0.2.3
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

**All v0.2.3 configurations will work in v0.2.4!**

---

## 🔍 Technical Details

### Code Changes

**File**: `internal/provider/resource_dataset.go`

**Change 1: Create Function (Lines 215-268)**
```go
// Before (v0.2.3) - Sent empty strings
if !data.Compression.IsNull() {
    createReq["compression"] = data.Compression.ValueString()  // Could be ""
}

// After (v0.2.4) - Omits empty strings
if !data.Compression.IsNull() && data.Compression.ValueString() != "" {
    createReq["compression"] = data.Compression.ValueString()  // Only non-empty
}
```

**Change 2: Update Function (Lines 323-376)**
- Applied same empty string check as Create function

**Affected Properties:**
- All string properties now check for empty strings before sending to API
- Integer properties unchanged (already working correctly)

---

## 🧪 Testing Results

### Build
- ✅ Compiles without errors
- ✅ No warnings

### Dataset Creation
- ✅ Minimal FILESYSTEM dataset (no optional properties)
- ✅ FILESYSTEM dataset with some properties
- ✅ FILESYSTEM dataset with all properties
- ✅ Minimal VOLUME dataset (only volsize)
- ✅ VOLUME dataset with optional properties

### API Behavior
- ✅ Empty string properties are omitted from requests
- ✅ Non-empty properties are sent correctly
- ✅ No "Invalid choice: " errors

---

## 📚 Documentation

- **CHANGELOG.md** - Updated with v0.2.4 section
- **RELEASE_NOTES_v0.2.4.md** - This file
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
| **Version** | 0.2.4 |
| **Release Type** | Critical Bug Fix |
| **Resources** | 14 (all working) |
| **Data Sources** | 2 (all working) |
| **Bugs Fixed** | 1 (critical) |
| **Breaking Changes** | 0 |
| **Binary Size** | ~25MB per platform |

---

## ⚠️ Important Notice

**All v0.2.3 users should upgrade to v0.2.4 immediately.**

v0.2.3 has a critical bug that sends empty strings to the API, causing "Invalid choice: " errors. v0.2.4 fixes this issue by properly omitting empty string properties.

---

**🎉 v0.2.4 fixes the empty string handling bug and allows datasets to be created with minimal configuration!** 🎉

**Recommendation**: All users should upgrade to v0.2.4 for proper empty string handling.

