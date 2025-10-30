# Release Notes - v0.2.5

**Release Date**: October 30, 2025  
**Provider**: TrueNAS Terraform Provider  
**Repository**: https://github.com/baladithyab/terraform-truenas-scale-24.04  
**Compatibility**: TrueNAS Scale 24.04

---

## 🔧 Critical Bug Fix Release

v0.2.5 is a **critical bug fix release** that resolves an issue discovered in v0.2.4 where the provider sent zero values for unset integer properties to the TrueNAS API, causing "'copies' must be one of '1 | 2 | 3'" errors.

---

## 🐛 What Was Fixed

### Critical Issue: Zero Values Sent for Integer Properties

**Problem in v0.2.4:**
```hcl
resource "truenas_dataset" "data" {
  name = "tank/data"
  type = "FILESYSTEM"
  # Optional integer properties not specified
  # Terraform sets them to 0 by default
}
```

**Error in v0.2.4:**
```
Error: Client Error
Unable to create dataset, got error: API returned status 422: 
{
  "pool_dataset_create.copies": [
    {
      "message": "'copies' must be one of '1 | 2 | 3'",
      "errno": 22
    }
  ]
}
```

**Root Cause:**
- v0.2.4 checked `!IsNull()` to determine if an integer property should be sent
- However, Terraform sets optional integer properties to **zero (`0`)** by default
- The provider was sending these zero values to the API
- TrueNAS API rejected zero values (e.g., `copies` must be 1, 2, or 3)

**Fix in v0.2.5:**
- Added value validation: `&& data.PropertyName.ValueInt64() > 0`
- Integer properties are now only included in API requests if they have **positive (non-zero)** values
- Applied to all integer properties in both `Create()` and `Update()` functions

**Result:**
✅ Datasets can now be created without specifying all optional integer properties  
✅ Zero-value integer properties are correctly omitted from API requests  
✅ Positive integer values are sent correctly

---

## 📊 Properties Affected

### Integer Properties (Fixed in v0.2.5)
These properties now require positive values to be sent to the API:

**Shared Properties (FILESYSTEM and VOLUME):**
- `copies` - Number of copies (must be 1, 2, or 3 if specified)
- `reservation` - Space reservation in bytes
- `refreservation` - Referenced space reservation in bytes

**VOLUME-Specific:**
- `volsize` - Volume size in bytes (required for VOLUME, but now validated)

**FILESYSTEM-Specific:**
- `quota` - Dataset quota in bytes
- `refquota` - Referenced quota in bytes

### String Properties (Fixed in v0.2.4)
These properties already work correctly (no change in v0.2.5):

- `comments`, `compression`, `sync`, `deduplication`, `readonly`
- `atime`, `exec`, `recordsize`, `snapdir` (FILESYSTEM only)

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
  copies      = 2  # Only sent if > 0
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

## 🔄 Upgrading from v0.2.4

### No Breaking Changes

**Step 1**: Update version:
```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.5"  # Changed from 0.2.4
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

**All v0.2.4 configurations will work in v0.2.5!**

---

## 🔍 Technical Details

### Code Changes

**File**: `internal/provider/resource_dataset.go`

**Change 1: Create Function (Lines 231-268)**
```go
// Before (v0.2.4) - Sent zero values
if !data.Copies.IsNull() {
    createReq["copies"] = data.Copies.ValueInt64()  // Could be 0
}

// After (v0.2.5) - Omits zero values
if !data.Copies.IsNull() && data.Copies.ValueInt64() > 0 {
    createReq["copies"] = data.Copies.ValueInt64()  // Only positive values
}
```

**Change 2: Update Function (Lines 339-376)**
- Applied same value validation as Create function

**Affected Properties:**
- All integer properties now check for positive values before sending to API
- String properties unchanged (already fixed in v0.2.4)

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
- ✅ Zero-value integer properties are omitted from requests
- ✅ Positive integer values are sent correctly
- ✅ No "'copies' must be one of '1 | 2 | 3'" errors

---

## 📚 Documentation

- **CHANGELOG.md** - Updated with v0.2.5 section
- **RELEASE_NOTES_v0.2.5.md** - This file
- All existing examples continue to work

---

## 🚀 What's Next

### Planned for v0.3.0
- Replication task management
- Cloud sync task management
- Service management (start/stop/configure)
- Certificate management
- Cron job management

See `API_COVERAGE.md` for the complete roadmap.

---

## 📞 Support

- **GitHub Issues**: https://github.com/baladithyab/terraform-truenas-scale-24.04/issues
- **Repository**: https://github.com/baladithyab/terraform-truenas-scale-24.04
- **TrueNAS Version**: Scale 24.04 (REST API)

---

## 📊 Release Statistics

| Metric | Value |
|--------|-------|
| **Version** | 0.2.5 |
| **Release Type** | Critical Bug Fix |
| **Resources** | 14 (all working) |
| **Data Sources** | 2 (all working) |
| **Bugs Fixed** | 1 (critical integer validation) |
| **Breaking Changes** | 0 |
| **Binary Size** | ~25MB per platform |

---

## ⚠️ Important Notice

**All v0.2.4 users should upgrade to v0.2.5 immediately.**

v0.2.4 has a critical bug that sends zero values for integer properties to the API, causing "'copies' must be one of '1 | 2 | 3'" errors. v0.2.5 fixes this issue by properly omitting zero-value integer properties.

---

## 🔄 Version History

### v0.2.4 → v0.2.5
- **v0.2.4**: Fixed empty string handling for string properties
- **v0.2.5**: Fixed zero value handling for integer properties

Both fixes are now in place - datasets can be created with minimal configuration!

---

**🎉 v0.2.5 fixes the integer property validation bug and allows datasets to be created with minimal configuration!** 🎉

**Recommendation**: All users should upgrade to v0.2.5 for proper integer property handling.

