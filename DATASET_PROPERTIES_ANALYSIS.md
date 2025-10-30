# TrueNAS Scale 24.04 Dataset Properties Analysis

**Date**: October 30, 2025  
**API Version**: v2.0  
**Purpose**: Deep dive analysis of valid properties for FILESYSTEM vs VOLUME datasets

---

## Executive Summary

Based on OpenAPI spec analysis and live API testing, TrueNAS Scale 24.04 has **DIFFERENT** property requirements for FILESYSTEM vs VOLUME datasets. The current provider implementation in v0.2.2 incorrectly sends FILESYSTEM-only properties to VOLUME datasets, causing 422 errors.

---

## Key Findings

### Properties Valid for BOTH Types

These properties are accepted by the API for both FILESYSTEM and VOLUME datasets:

| Property | Type | Description |
|----------|------|-------------|
| `name` | string | Dataset name (required) |
| `type` | string | "FILESYSTEM" or "VOLUME" (required) |
| `comments` | string | User comments |
| `compression` | string | Compression algorithm (LZ4, GZIP, ZSTD, etc.) |
| `sync` | string | Sync mode (STANDARD, ALWAYS, DISABLED) |
| `deduplication` | string | Deduplication (ON, OFF) |
| `snapdev` | string | Snapshot device visibility (HIDDEN, VISIBLE) |
| `checksum` | string | Checksum algorithm |
| `readonly` | string | Read-only flag (ON, OFF) |
| `copies` | integer | Number of copies |
| `reservation` | integer | Space reservation |
| `refreservation` | integer | Referenced space reservation |

### Properties ONLY for VOLUME

These properties are **REQUIRED** or **ONLY VALID** for VOLUME datasets:

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `volsize` | integer | **YES** | Volume size in bytes |
| `volblocksize` | string | No | Block size (512, 1K, 2K, 4K, 8K, 16K, 32K, 64K, 128K) |
| `sparse` | boolean | No | Sparse volume flag |

### Properties ONLY for FILESYSTEM

These properties are **REJECTED** by the API for VOLUME datasets:

| Property | Type | Description |
|----------|------|-------------|
| `atime` | string | Access time updates (ON, OFF) |
| `exec` | string | Execute permission (ON, OFF) |
| `recordsize` | string | Record size (512, 1K, 2K, ..., 1M) |
| `quota` | integer | Dataset quota |
| `refquota` | integer | Referenced quota |
| `snapdir` | string | Snapshot directory visibility (HIDDEN, VISIBLE) |
| `aclmode` | string | ACL mode |
| `acltype` | string | ACL type (POSIX, NFSV4) |
| `xattr` | string | Extended attributes |
| `casesensitivity` | string | Case sensitivity |
| `special_small_block_size` | integer | Special small block size |

---

## API Testing Results

### Test 1: VOLUME with Only Required Fields ✅

**Request:**
```json
{
  "name": "Dagr/test-volume-api",
  "type": "VOLUME",
  "volsize": 1073741824,
  "comments": "Test VOLUME with only required fields"
}
```

**Result:** ✅ **SUCCESS** - Dataset created successfully

### Test 2: VOLUME with FILESYSTEM Properties ❌

**Request:**
```json
{
  "name": "Dagr/test-volume-with-fs-props",
  "type": "VOLUME",
  "volsize": 1073741824,
  "comments": "Test VOLUME with FILESYSTEM properties",
  "atime": "OFF",
  "recordsize": "128K",
  "quota": 0
}
```

**Result:** ❌ **FAILED** - HTTP 422 Unprocessable Entity

**Error:**
```json
{
  "pool_dataset_create.atime": [
    {
      "message": "This field is not valid for VOLUME",
      "errno": 22
    }
  ],
  "pool_dataset_create.quota": [
    {
      "message": "This field is not valid for VOLUME",
      "errno": 22
    }
  ],
  "pool_dataset_create.recordsize": [
    {
      "message": "This field is not valid for VOLUME",
      "errno": 22
    }
  ]
}
```

### Test 3: VOLUME with Valid Shared Properties ✅

**Request:**
```json
{
  "name": "Dagr/test-volume-with-valid-props",
  "type": "VOLUME",
  "volsize": 1073741824,
  "comments": "Test VOLUME with valid properties",
  "compression": "LZ4",
  "sync": "STANDARD",
  "deduplication": "OFF"
}
```

**Result:** ✅ **SUCCESS** - Dataset created successfully with all properties

---

## Current Provider Issues (v0.2.2)

### Issue 1: Incorrect Property Categorization

The v0.2.2 implementation treats these properties as FILESYSTEM-only:
- `compression` ❌ **WRONG** - Valid for BOTH types
- `atime` ✅ Correct - FILESYSTEM only
- `deduplication` ❌ **WRONG** - Valid for BOTH types
- `exec` ✅ Correct - FILESYSTEM only
- `readonly` ❌ **WRONG** - Valid for BOTH types
- `sync` ❌ **WRONG** - Valid for BOTH types
- `snapdir` ❌ **WRONG** - Valid for BOTH types (but different meaning)
- `recordsize` ✅ Correct - FILESYSTEM only
- `quota` ✅ Correct - FILESYSTEM only

### Issue 2: Read() Function Doesn't Filter

The `readDataset()` function reads ALL properties regardless of dataset type, which causes computed attributes to have values even when they shouldn't.

---

## Correct Property Categorization

### Category A: Valid for BOTH Types
```go
// These can be sent for both FILESYSTEM and VOLUME
- comments
- compression
- sync
- deduplication
- snapdev
- checksum
- readonly
- copies
- reservation
- refreservation
```

### Category B: VOLUME-Specific
```go
// Only send these for VOLUME datasets
- volsize (REQUIRED)
- volblocksize
- sparse
```

### Category C: FILESYSTEM-Specific
```go
// Only send these for FILESYSTEM datasets
- atime
- exec
- recordsize
- quota
- refquota
- snapdir (different from snapdev!)
- aclmode
- acltype
- xattr
- casesensitivity
- special_small_block_size
```

---

## Recommended Fix for v0.2.3

### Fix 1: Update Create() Function

```go
// Properties valid for BOTH FILESYSTEM and VOLUME
if !data.Comments.IsNull() {
    createReq["comments"] = data.Comments.ValueString()
}
if !data.Compression.IsNull() {
    createReq["compression"] = data.Compression.ValueString()
}
if !data.Sync.IsNull() {
    createReq["sync"] = data.Sync.ValueString()
}
if !data.Dedup.IsNull() {
    createReq["deduplication"] = data.Dedup.ValueString()
}
if !data.ReadOnly.IsNull() {
    createReq["readonly"] = data.ReadOnly.ValueString()
}
if !data.Copies.IsNull() {
    createReq["copies"] = data.Copies.ValueInt64()
}
if !data.Reservation.IsNull() {
    createReq["reservation"] = data.Reservation.ValueInt64()
}
if !data.RefReserv.IsNull() {
    createReq["refreservation"] = data.RefReserv.ValueInt64()
}

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
    if !data.Exec.IsNull() {
        createReq["exec"] = data.Exec.ValueString()
    }
    if !data.RecordSize.IsNull() {
        createReq["recordsize"] = data.RecordSize.ValueString()
    }
    if !data.Quota.IsNull() {
        createReq["quota"] = data.Quota.ValueInt64()
    }
    if !data.RefQuota.IsNull() {
        createReq["refquota"] = data.RefQuota.ValueInt64()
    }
    if !data.SnapDir.IsNull() {
        createReq["snapdir"] = data.SnapDir.ValueString()
    }
}
```

### Fix 2: Update Read() Function

```go
// Determine dataset type first
datasetType := "FILESYSTEM"
if dtype, ok := result["type"].(string); ok {
    datasetType = dtype
    data.Type = types.StringValue(dtype)
}

// Read properties valid for BOTH types
if compression, ok := result["compression"].(map[string]interface{}); ok {
    if value, ok := compression["value"].(string); ok {
        data.Compression = types.StringValue(value)
    }
}
// ... other shared properties

// Read VOLUME-specific properties
if datasetType == "VOLUME" {
    if volsize, ok := result["volsize"].(map[string]interface{}); ok {
        if value, ok := volsize["parsed"].(float64); ok {
            data.Volsize = types.Int64Value(int64(value))
        }
    }
    
    // Set FILESYSTEM-only properties to null
    data.Atime = types.StringNull()
    data.Exec = types.StringNull()
    data.RecordSize = types.StringNull()
    data.Quota = types.Int64Null()
    data.RefQuota = types.Int64Null()
    data.SnapDir = types.StringNull()
}

// Read FILESYSTEM-specific properties
if datasetType == "FILESYSTEM" {
    if atime, ok := result["atime"].(map[string]interface{}); ok {
        if value, ok := atime["value"].(string); ok {
            data.Atime = types.StringValue(value)
        }
    }
    // ... other FILESYSTEM properties
    
    // Set VOLUME-only properties to null
    data.Volsize = types.Int64Null()
}
```

---

## Summary

**Root Cause**: The provider incorrectly categorizes properties as FILESYSTEM-only when they're actually valid for both types.

**Impact**: VOLUME datasets fail to create because the provider sends FILESYSTEM-only properties like `atime`, `exec`, `recordsize`, `quota`, etc.

**Solution**: Properly categorize properties into three groups:
1. Valid for BOTH types (compression, sync, deduplication, readonly, etc.)
2. VOLUME-specific (volsize, volblocksize, sparse)
3. FILESYSTEM-specific (atime, exec, recordsize, quota, refquota, snapdir, etc.)

**Version**: This fix should be released as v0.2.3

