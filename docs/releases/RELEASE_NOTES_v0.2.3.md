# Release Notes - v0.2.3

**Release Date**: October 30, 2025  
**Provider**: TrueNAS Terraform Provider  
**Repository**: https://github.com/baladithyab/terraform-provider-truenas  
**Compatibility**: TrueNAS Scale 24.04

---

## 🔧 Critical Bug Fix Release

v0.2.3 is a **critical bug fix release** that resolves a major issue discovered in v0.2.2 where VOLUME datasets could not be created due to incorrect property categorization.

---

## 🐛 What Was Fixed

### Critical Issue: Incorrect Property Categorization

**Problem in v0.2.2:**
```hcl
resource "truenas_dataset" "vm_disk" {
  name    = "tank/vms/vm01-disk0"
  type    = "VOLUME"
  volsize = 107374182400  # 100GB
}
```

**Error in v0.2.2:**
```
Error: Client Error
Unable to create dataset, got error: API returned status 422: Unprocessable Entity
```

**Root Cause:**
- v0.2.2 incorrectly categorized `compression`, `sync`, `deduplication`, `readonly`, `copies`, `reservation`, and `refreservation` as FILESYSTEM-only properties
- The provider was NOT sending these properties to VOLUME datasets
- However, these properties are actually **valid for BOTH** FILESYSTEM and VOLUME datasets
- TrueNAS API expects these properties for VOLUME datasets (they inherit from parent if not specified)

**Fix in v0.2.3:**
- Properly categorized properties into three groups based on TrueNAS Scale 24.04 API specification
- VOLUME datasets now receive all applicable shared properties
- Read() function now correctly handles properties based on dataset type

**Result:**
✅ VOLUME datasets can now be created successfully  
✅ FILESYSTEM datasets continue to work correctly  
✅ Both dataset types are fully functional

---

## 📊 Property Categorization

### Properties Valid for BOTH Types ✅

These properties work for both FILESYSTEM and VOLUME datasets:

| Property | Description |
|----------|-------------|
| `comments` | User comments |
| `compression` | Compression algorithm (LZ4, GZIP, ZSTD, etc.) |
| `sync` | Sync mode (STANDARD, ALWAYS, DISABLED) |
| `deduplication` | Deduplication (ON, OFF) |
| `readonly` | Read-only flag (ON, OFF) |
| `copies` | Number of copies (1-3) |
| `reservation` | Space reservation in bytes |
| `refreservation` | Referenced space reservation in bytes |

### VOLUME-Specific Properties 📦

These properties are ONLY for VOLUME datasets:

| Property | Required | Description |
|----------|----------|-------------|
| `volsize` | **YES** | Volume size in bytes |
| `volblocksize` | No | Block size (512, 1K, 2K, 4K, 8K, 16K, 32K, 64K, 128K) |
| `sparse` | No | Sparse volume flag |

### FILESYSTEM-Specific Properties 📁

These properties are ONLY for FILESYSTEM datasets:

| Property | Description |
|----------|-------------|
| `atime` | Access time updates (ON, OFF) |
| `exec` | Execute permission (ON, OFF) |
| `recordsize` | Record size (512, 1K, ..., 1M) |
| `quota` | Dataset quota in bytes |
| `refquota` | Referenced quota in bytes |
| `snapdir` | Snapshot directory visibility (HIDDEN, VISIBLE) |

---

## ✅ Verified Working Configurations

### FILESYSTEM Dataset (Works!)
```hcl
resource "truenas_dataset" "data" {
  name        = "tank/data"
  type        = "FILESYSTEM"
  compression = "LZ4"
  recordsize  = "128K"
  atime       = "OFF"
  sync        = "STANDARD"
  comments    = "Data storage"
}
```

### VOLUME Dataset (Now Works!)
```hcl
resource "truenas_dataset" "vm_disk" {
  name        = "tank/vms/vm01-disk0"
  type        = "VOLUME"
  volsize     = 107374182400  # 100GB in bytes
  compression = "LZ4"
  sync        = "STANDARD"
  comments    = "VM disk for vm01"
}
```

### Both Types Together (Now Works!)
```hcl
# FILESYSTEM for file storage
resource "truenas_dataset" "shares" {
  name        = "tank/shares"
  type        = "FILESYSTEM"
  compression = "LZ4"
  atime       = "OFF"
  recordsize  = "128K"
}

# VOLUME for VM disk
resource "truenas_dataset" "vm_disk" {
  name        = "tank/vms/vm01-disk0"
  type        = "VOLUME"
  volsize     = 107374182400  # 100GB
  compression = "LZ4"
  sync        = "STANDARD"
}
```

---

## 🔄 Upgrading from v0.2.2

### No Breaking Changes

**Step 1**: Update version:
```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.3"  # Changed from 0.2.2
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

**All v0.2.2 configurations will work in v0.2.3!**

---

## 🔍 Technical Details

### Code Changes

**File**: `internal/provider/resource_dataset.go`

**Change 1**: Create Function (Lines 209-268)
```go
// Properties valid for BOTH FILESYSTEM and VOLUME
if !data.Compression.IsNull() {
    createReq["compression"] = data.Compression.ValueString()
}
if !data.Sync.IsNull() {
    createReq["sync"] = data.Sync.ValueString()
}
if !data.Dedup.IsNull() {
    createReq["deduplication"] = data.Dedup.ValueString()
}
// ... other shared properties

// VOLUME-specific properties
if datasetType == "VOLUME" {
    if !data.Volsize.IsNull() && !data.Volsize.IsUnknown() {
        createReq["volsize"] = data.Volsize.ValueInt64()
    }
}

// FILESYSTEM-specific properties
if datasetType == "FILESYSTEM" {
    if !data.Atime.IsNull() {
        createReq["atime"] = data.Atime.ValueString()
    }
    // ... other FILESYSTEM properties
}
```

**Change 2**: Update Function (Lines 320-376)
- Applied same conditional logic as Create function

**Change 3**: Read Function (Lines 426-477)
```go
// Determine dataset type first
datasetType := "FILESYSTEM"
if dtype, ok := result["type"].(string); ok {
    datasetType = dtype
}

// Read properties based on dataset type
if datasetType == "VOLUME" {
    // Read VOLUME-specific properties
    // Set FILESYSTEM-only properties to null
    data.Atime = types.StringNull()
    data.Exec = types.StringNull()
    data.RecordSize = types.StringNull()
    // ...
} else {
    // Read FILESYSTEM-specific properties
    // Set VOLUME-only properties to null
    data.Volsize = types.Int64Null()
}
```

---

## 🧪 Testing Results

### API Testing

**Test 1: VOLUME with Shared Properties** ✅
```bash
curl -X POST "http://10.0.0.83:81/api/v2.0/pool/dataset" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d '{
    "name": "Dagr/test-volume",
    "type": "VOLUME",
    "volsize": 1073741824,
    "compression": "LZ4",
    "sync": "STANDARD",
    "deduplication": "OFF"
  }'
```
**Result**: ✅ SUCCESS - Dataset created with all properties

**Test 2: VOLUME with FILESYSTEM Properties** ❌
```bash
curl -X POST "http://10.0.0.83:81/api/v2.0/pool/dataset" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d '{
    "name": "Dagr/test-volume",
    "type": "VOLUME",
    "volsize": 1073741824,
    "atime": "OFF",
    "recordsize": "128K"
  }'
```
**Result**: ❌ FAILED - HTTP 422: "This field is not valid for VOLUME"

### Provider Testing

**Build:**
- ✅ Compiles without errors
- ✅ No warnings

**FILESYSTEM Datasets:**
- ✅ Create with all properties
- ✅ Update works correctly
- ✅ Delete works correctly
- ✅ Import works correctly

**VOLUME Datasets:**
- ✅ Create with shared properties (compression, sync, etc.)
- ✅ Update works correctly
- ✅ Delete works correctly
- ✅ Import works correctly

---

## 📚 Documentation

- **DATASET_PROPERTIES_ANALYSIS.md** - Complete analysis of property categorization
- **CHANGELOG.md** - Updated with v0.2.3 section
- **RELEASE_NOTES_v0.2.3.md** - This file
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
| **Version** | 0.2.3 |
| **Release Type** | Critical Bug Fix |
| **Resources** | 14 (all working) |
| **Data Sources** | 2 (all working) |
| **Bugs Fixed** | 1 (critical) |
| **Breaking Changes** | 0 |
| **Binary Size** | ~25MB per platform |

---

## ⚠️ Important Notice

**All v0.2.2 users should upgrade to v0.2.3 immediately.**

v0.2.2 has a critical bug that prevents VOLUME datasets from being created. v0.2.3 fixes this issue and makes both FILESYSTEM and VOLUME datasets fully functional.

---

**🎉 v0.2.3 fixes the critical property categorization bug and makes VOLUME datasets fully functional!** 🎉

**Recommendation**: All users should upgrade to v0.2.3 for full dataset functionality.

