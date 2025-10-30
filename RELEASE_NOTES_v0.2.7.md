# Release Notes - v0.2.7

**Release Date**: October 30, 2025  
**Provider**: TrueNAS Terraform Provider  
**Repository**: https://github.com/baladithyab/terraform-truenas-scale-24.04  
**Compatibility**: TrueNAS Scale 24.04

---

## 🔧 Critical Bug Fix Release

v0.2.7 is a **critical bug fix release** that resolves integer property parsing issues discovered in v0.2.6 where the provider couldn't correctly parse integer properties from the TrueNAS API response.

---

## 🐛 What Was Fixed

### Critical Issue: Integer Properties Not Parsed Correctly

**Problem in v0.2.6:**
```
Error: Provider produced inconsistent result after apply

When applying changes to truenas_dataset.data, provider produced an unexpected new value:
.copies: was unknown, but now known.

This is a bug in the provider, which should be reported in the provider's own issue tracker.
```

**Root Causes:**
1. **`copies` property**: v0.2.6 tried to parse as `float64`, but TrueNAS API returns it as a **string** (`"1"`)
2. **Other integer properties**: v0.2.6 didn't handle `null` values for quota, refquota, reservation, refreservation

**Fix in v0.2.7:**
1. Parse `copies` as string and convert to int64 using `strconv.ParseInt()`
2. Handle null values for all integer properties using type switch
3. Set to `types.Int64Null()` when value is null or missing

**Result:**
✅ All integer properties correctly parsed from TrueNAS API  
✅ Null values handled correctly  
✅ Terraform can properly track integer property values  
✅ No more "unknown values" errors for integer properties

---

## 📊 Properties Fixed

### Integer Properties Now Correctly Parsed

**Shared Properties (FILESYSTEM and VOLUME)**:
- ✅ `copies` - Parsed as string, converted to int64
- ✅ `reservation` - Handles null and numeric values
- ✅ `refreservation` - Handles null and numeric values

**FILESYSTEM-Specific Properties**:
- ✅ `quota` - Handles null and numeric values
- ✅ `refquota` - Handles null and numeric values

---

## ✅ Verified Working Behavior

### Before v0.2.7 (v0.2.6)
```bash
terraform apply
# Datasets created successfully
# But Terraform reports: "unknown values" for integer properties
# State file missing integer property values
```

### After v0.2.7
```bash
terraform apply
# Datasets created successfully
# All integer properties parsed correctly
# State file contains all property values
# No "unknown values" errors
```

---

## 🔄 Upgrading from v0.2.6

### No Breaking Changes

**Step 1**: Update version:
```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.7"  # Changed from 0.2.6
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

**All v0.2.6 configurations will work in v0.2.7!**

---

## 🔍 Technical Details

### Code Changes

**File**: `internal/provider/resource_dataset.go`

**Change 1: Added strconv import (Lines 3-18)**
```go
import (
    "strconv"  // Added for string-to-int conversion
    // ... other imports
)
```

**Change 2: Fixed copies parsing (Lines 480-493)**

**Before (v0.2.6)** - Tried to parse as float64:
```go
if copies, ok := result["copies"].(map[string]interface{}); ok {
    if value, ok := copies["value"].(float64); ok {  // Wrong type!
        data.Copies = types.Int64Value(int64(value))
    }
} else {
    data.Copies = types.Int64Null()
}
```

**After (v0.2.7)** - Parse as string and convert:
```go
if copies, ok := result["copies"].(map[string]interface{}); ok {
    if value, ok := copies["value"].(string); ok && value != "" {
        // copies is returned as a string, convert to int64
        if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
            data.Copies = types.Int64Value(intValue)
        } else {
            data.Copies = types.Int64Null()
        }
    } else {
        data.Copies = types.Int64Null()
    }
} else {
    data.Copies = types.Int64Null()
}
```

**Change 3: Fixed null handling for other integer properties (Lines 494-527, 573-605)**

**Before (v0.2.6)** - Didn't handle null:
```go
if reservation, ok := result["reservation"].(map[string]interface{}); ok {
    if value, ok := reservation["parsed"].(float64); ok {  // Fails if null!
        data.Reservation = types.Int64Value(int64(value))
    }
} else {
    data.Reservation = types.Int64Null()
}
```

**After (v0.2.7)** - Handle null with type switch:
```go
if reservation, ok := result["reservation"].(map[string]interface{}); ok {
    if parsed := reservation["parsed"]; parsed != nil {
        switch v := parsed.(type) {
        case float64:
            data.Reservation = types.Int64Value(int64(v))
        case int64:
            data.Reservation = types.Int64Value(v)
        default:
            data.Reservation = types.Int64Null()
        }
    } else {
        data.Reservation = types.Int64Null()
    }
} else {
    data.Reservation = types.Int64Null()
}
```

**Key Improvements:**
- `copies`: Parse as string, convert to int64
- All integer properties: Handle null values correctly
- Type switch: Handle both float64 and int64 from JSON
- Proper null handling: Set to `types.Int64Null()` when missing

---

## 🧪 Testing Results

### Build
- ✅ Compiles without errors
- ✅ No warnings

### API Response Parsing
- ✅ `copies` parsed as string "1" → int64(1)
- ✅ `quota` null → types.Int64Null()
- ✅ `refquota` null → types.Int64Null()
- ✅ `reservation` null → types.Int64Null()
- ✅ `refreservation` null → types.Int64Null()

### State Tracking
- ✅ All integer properties correctly stored in state file
- ✅ No "unknown values" errors
- ✅ Terraform can detect drift correctly
- ✅ `terraform plan` shows no changes after apply

---

## 📚 Documentation

- **CHANGELOG.md** - Updated with v0.2.7 section
- **RELEASE_NOTES_v0.2.7.md** - This file
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
| **Version** | 0.2.7 |
| **Release Type** | Critical Bug Fix |
| **Resources** | 14 (all working) |
| **Data Sources** | 2 (all working) |
| **Bugs Fixed** | 1 (integer property parsing) |
| **Breaking Changes** | 0 |
| **Binary Size** | ~25MB per platform |

---

## ⚠️ Important Notice

**All v0.2.6 users should upgrade to v0.2.7 immediately.**

v0.2.6 has a critical bug where integer properties cannot be parsed correctly from the TrueNAS API, causing "unknown values" errors. v0.2.7 fixes this issue by properly parsing `copies` as a string and handling null values for all integer properties.

---

## 🔄 Version History

### v0.2.5 → v0.2.6 → v0.2.7
- **v0.2.5**: Fixed zero value handling for integer properties in Create/Update
- **v0.2.6**: Fixed Read() to read all properties, but integer parsing was incorrect
- **v0.2.7**: Fixed integer property parsing to handle string values and nulls

All three fixes are now in place - datasets can be created AND tracked properly with correct integer values!

---

**🎉 v0.2.7 fixes the integer property parsing bug and enables proper state tracking!** 🎉

**Recommendation**: All users should upgrade to v0.2.7 for proper integer property tracking.

