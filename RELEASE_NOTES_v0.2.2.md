# Release Notes - v0.2.2

**Release Date**: October 30, 2025  
**Provider**: TrueNAS Terraform Provider  
**Repository**: https://github.com/baladithyab/terraform-truenas-scale-24.04  
**Compatibility**: TrueNAS Scale 24.04

---

## 🐛 Bug Fix Release

v0.2.2 is a **critical bug fix release** that resolves two major issues discovered in v0.2.1 that prevented both FILESYSTEM and VOLUME datasets from being created successfully.

---

## 🔧 What Was Fixed

### Bug #1: False Positive Validation Error on FILESYSTEM Datasets

**Problem:**
```hcl
resource "truenas_dataset" "data" {
  name = "tank/data"
  type = "FILESYSTEM"
  # No volsize specified
}
```

**Error in v0.2.1:**
```
Error: Invalid Attribute
volsize is not valid for FILESYSTEM type datasets. Remove the volsize attribute or change type to VOLUME.
```

**Root Cause:**
- The `volsize` attribute has `Computed: true` in the schema
- When Terraform evaluates computed attributes, `IsNull()` returns `false` even when the user didn't specify the value
- The validation logic only checked `!IsNull()`, causing false positives

**Fix:**
- Added `!IsUnknown()` check in addition to `!IsNull()` in validation logic
- Now correctly distinguishes between user-specified values and computed values

**Result:**
✅ FILESYSTEM datasets can now be created without spurious validation errors

---

### Bug #2: API 422 Errors When Creating VOLUME Datasets

**Problem:**
```hcl
resource "truenas_dataset" "vm_disk" {
  name    = "tank/vms/vm01-disk0"
  type    = "VOLUME"
  volsize = 107374182400  # 100GB
}
```

**Error in v0.2.1:**
```
Error: Client Error
Unable to create dataset, got error: API returned status 422: Unprocessable Entity
```

**Root Cause:**
- The provider was sending FILESYSTEM-only properties to the TrueNAS API for VOLUME datasets
- Properties like `compression`, `atime`, `deduplication`, `exec`, `readonly`, `sync`, `snapdir`, `recordsize`, `quota`, etc. are not valid for VOLUME type datasets
- TrueNAS Scale 24.04 API rejects these properties with HTTP 422 Unprocessable Entity

**Fix:**
- Implemented conditional property sending based on dataset type
- **VOLUME datasets** now only send: `name`, `type`, `volsize`, `comments`
- **FILESYSTEM datasets** send all applicable properties
- Applied same logic to both `Create()` and `Update()` functions

**Result:**
✅ VOLUME datasets can now be created successfully via Terraform

---

## 📊 Impact

| Dataset Type | v0.2.1 Status | v0.2.2 Status |
|--------------|---------------|---------------|
| FILESYSTEM | ❌ Validation Error | ✅ Works |
| VOLUME | ❌ API 422 Error | ✅ Works |

**Both dataset types now work correctly!**

---

## 🔄 Upgrading from v0.2.1

### No Breaking Changes

This is a **bug fix release** with no breaking changes. All v0.2.1 configurations will work in v0.2.2.

**Step 1**: Update version in `terraform` block:
```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.2"  # Changed from 0.2.1
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

---

## ✅ Verified Working Configurations

### FILESYSTEM Dataset (Now Works!)
```hcl
resource "truenas_dataset" "data" {
  name        = "tank/data"
  type        = "FILESYSTEM"
  compression = "LZ4"
  recordsize  = "128K"
  atime       = "OFF"
  comments    = "Data storage"
}
```

### VOLUME Dataset (Now Works!)
```hcl
resource "truenas_dataset" "vm_disk" {
  name     = "tank/vms/vm01-disk0"
  type     = "VOLUME"
  volsize  = 107374182400  # 100GB in bytes
  comments = "VM disk for vm01"
}
```

### Both Types Together (Now Works!)
```hcl
# FILESYSTEM for file storage
resource "truenas_dataset" "shares" {
  name        = "tank/shares"
  type        = "FILESYSTEM"
  compression = "LZ4"
}

# VOLUME for VM disk
resource "truenas_dataset" "vm_disk" {
  name    = "tank/vms/vm01-disk0"
  type    = "VOLUME"
  volsize = 107374182400  # 100GB
}
```

---

## 🧪 Testing

### Test Results

**Build:**
- ✅ Compiles without errors
- ✅ No warnings

**FILESYSTEM Datasets:**
- ✅ Create without validation errors
- ✅ Update works correctly
- ✅ Delete works correctly
- ✅ Import works correctly

**VOLUME Datasets:**
- ✅ Create without API 422 errors
- ✅ Update works correctly
- ✅ Delete works correctly
- ✅ Import works correctly

---

## 🔍 Technical Details

### Code Changes

**File**: `internal/provider/resource_dataset.go`

**Change 1**: Validation Logic (Line 200)
```go
// BEFORE (v0.2.1):
if datasetType == "FILESYSTEM" && !data.Volsize.IsNull() {

// AFTER (v0.2.2):
if datasetType == "FILESYSTEM" && !data.Volsize.IsNull() && !data.Volsize.IsUnknown() {
```

**Change 2**: Create Function (Lines 209-269)
```go
// Properties valid for BOTH FILESYSTEM and VOLUME
if !data.Comments.IsNull() {
    createReq["comments"] = data.Comments.ValueString()
}

// Only send VOLUME-specific properties for VOLUME datasets
if datasetType == "VOLUME" {
    if !data.Volsize.IsNull() && !data.Volsize.IsUnknown() {
        createReq["volsize"] = data.Volsize.ValueInt64()
    }
}

// Only send FILESYSTEM-specific properties for FILESYSTEM datasets
if datasetType == "FILESYSTEM" {
    if !data.Compression.IsNull() {
        createReq["compression"] = data.Compression.ValueString()
    }
    // ... all other FILESYSTEM properties
}
```

**Change 3**: Update Function (Lines 307-377)
- Applied same conditional logic as Create function

---

## 📚 Documentation

- **CHANGELOG.md** - Updated with v0.2.2 section
- **RELEASE_NOTES_v0.2.2.md** - This file
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
| **Version** | 0.2.2 |
| **Release Type** | Bug Fix |
| **Resources** | 14 (all working) |
| **Data Sources** | 2 (all working) |
| **Bugs Fixed** | 2 (critical) |
| **Breaking Changes** | 0 |
| **Binary Size** | ~25MB per platform |

---

**🎉 v0.2.2 fixes critical bugs and makes both FILESYSTEM and VOLUME datasets fully functional!** 🎉

**Recommendation**: All v0.2.1 users should upgrade to v0.2.2 immediately.

