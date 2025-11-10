# Release Notes - v0.2.1

**Release Date**: October 30, 2025  
**Provider**: TrueNAS Terraform Provider  
**Repository**: https://github.com/baladithyab/terraform-provider-truenas  
**Compatibility**: TrueNAS Scale 24.04

---

## 🎉 What's New in v0.2.1

This release adds support for **VOLUME type datasets (zvols)** with the new `volsize` attribute, enabling full VM disk and iSCSI extent management via Terraform.

### ✨ New Feature: `volsize` Attribute

The `truenas_dataset` resource now supports the `volsize` attribute for creating VOLUME type datasets (zvols).

**Key Features:**
- ✅ Create VOLUME datasets for VM disks
- ✅ Create VOLUME datasets for iSCSI extents
- ✅ Automatic validation ensures correct usage
- ✅ Clear error messages for misconfigurations
- ✅ Full CRUD support (Create, Read, Update, Delete)

---

## 📖 Usage Examples

### Creating a VOLUME Dataset for VM Disk

```hcl
resource "truenas_dataset" "vm_disk" {
  name        = "tank/vms/vm01-disk0"
  type        = "VOLUME"
  volsize     = 107374182400  # 100GB in bytes
  compression = "LZ4"
  comments    = "VM disk for vm01"
}
```

### Creating a VOLUME Dataset for iSCSI

```hcl
resource "truenas_dataset" "iscsi_extent" {
  name        = "tank/iscsi/target01-lun0"
  type        = "VOLUME"
  volsize     = 1099511627776  # 1TB in bytes
  compression = "LZ4"
  comments    = "iSCSI LUN for target01"
}
```

### FILESYSTEM Dataset (No Change)

```hcl
resource "truenas_dataset" "data" {
  name        = "tank/data"
  type        = "FILESYSTEM"
  compression = "LZ4"
  recordsize  = "128K"
  # volsize not applicable for FILESYSTEM
}
```

---

## 🔧 Validation

The provider now validates `volsize` usage:

### ✅ Valid Configurations

```hcl
# VOLUME with volsize - OK
resource "truenas_dataset" "volume_ok" {
  name    = "tank/volume"
  type    = "VOLUME"
  volsize = 10737418240  # 10GB
}

# FILESYSTEM without volsize - OK
resource "truenas_dataset" "filesystem_ok" {
  name = "tank/filesystem"
  type = "FILESYSTEM"
}
```

### ❌ Invalid Configurations

```hcl
# VOLUME without volsize - ERROR
resource "truenas_dataset" "volume_fail" {
  name = "tank/volume"
  type = "VOLUME"
  # Missing volsize - will fail with clear error
}
# Error: Missing Required Attribute
# volsize is required when type is VOLUME. Please specify the volume size in bytes.

# FILESYSTEM with volsize - ERROR
resource "truenas_dataset" "filesystem_fail" {
  name    = "tank/filesystem"
  type    = "FILESYSTEM"
  volsize = 10737418240
  # volsize not valid for FILESYSTEM - will fail with clear error
}
# Error: Invalid Attribute
# volsize is not valid for FILESYSTEM type datasets. Remove the volsize attribute or change type to VOLUME.
```

---

## 📏 Size Conversion Reference

| Size | Bytes | Example |
|------|-------|---------|
| 1 GB | 1073741824 | `volsize = 1073741824` |
| 10 GB | 10737418240 | `volsize = 10737418240` |
| 50 GB | 53687091200 | `volsize = 53687091200` |
| 100 GB | 107374182400 | `volsize = 107374182400` |
| 500 GB | 536870912000 | `volsize = 536870912000` |
| 1 TB | 1099511627776 | `volsize = 1099511627776` |

**Helper for dynamic sizing:**
```hcl
variable "disk_size_gb" {
  description = "Disk size in GB"
  type        = number
  default     = 100
}

resource "truenas_dataset" "vm_disk" {
  name    = "tank/vms/disk"
  type    = "VOLUME"
  volsize = var.disk_size_gb * 1024 * 1024 * 1024  # Convert GB to bytes
}
```

---

## 🔄 Upgrading from v0.2.0

### No Breaking Changes

This is a feature addition release with no breaking changes. Upgrading is simple:

**Step 1: Update your `terraform` block:**

```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.1"  # Changed from 0.2.0
    }
  }
}
```

**Step 2: Upgrade the provider:**

```bash
terraform init -upgrade
```

**Step 3: Verify everything works:**

```bash
terraform plan
```

### Existing Configurations

All existing configurations will continue to work without modification. The `volsize` attribute is:
- **Optional** for FILESYSTEM datasets (and will be rejected if provided)
- **Required** for VOLUME datasets (new functionality)

---

## 🐛 What Was Fixed

### Issue: Cannot Create VOLUME Datasets

**Before v0.2.1:**
```hcl
# This would fail or require manual creation
resource "truenas_dataset" "vm_disk" {
  name = "tank/vms/disk"
  type = "VOLUME"
  # No way to specify volsize
}
```

**After v0.2.1:**
```hcl
# Now works perfectly!
resource "truenas_dataset" "vm_disk" {
  name    = "tank/vms/disk"
  type    = "VOLUME"
  volsize = 107374182400  # 100GB
}
```

---

## 📊 Statistics

| Metric | Value |
|--------|-------|
| Resources | 14 (all working) |
| Data Sources | 2 (all working) |
| Import Support | 100% (all resources) |
| New Attributes | 1 (`volsize`) |
| Breaking Changes | 0 |
| Binary Size | ~25MB |

---

## 🧪 Testing

This release includes comprehensive test cases in the `test-volsize/` directory:

1. ✅ FILESYSTEM dataset without volsize
2. ✅ VOLUME dataset with volsize
3. ✅ Validation: VOLUME without volsize (should fail)
4. ✅ Validation: FILESYSTEM with volsize (should fail)

See `test-volsize/README.md` for detailed testing instructions.

---

## 🔗 Use Cases

### VM Management

```hcl
module "vm_disk" {
  source = "./modules/vm-disk"
  
  pool_name = "tank"
  vm_name   = "ubuntu-server"
  disk_size = 100  # GB
}

# In module:
resource "truenas_dataset" "vm_disk" {
  name    = "${var.pool_name}/vms/${var.vm_name}-disk0"
  type    = "VOLUME"
  volsize = var.disk_size * 1024 * 1024 * 1024
}
```

### iSCSI Storage

```hcl
resource "truenas_dataset" "iscsi_lun" {
  name    = "tank/iscsi/lun-${var.target_name}"
  type    = "VOLUME"
  volsize = var.lun_size_bytes
}

resource "truenas_iscsi_extent" "extent" {
  name = "extent-${var.target_name}"
  type = "DISK"
  disk = "zvol/${truenas_dataset.iscsi_lun.name}"
}
```

---

## 📞 Support

- **GitHub Issues**: https://github.com/baladithyab/terraform-provider-truenas/issues
- **Documentation**: https://github.com/baladithyab/terraform-provider-truenas
- **TrueNAS Version**: Scale 24.04 (REST API)

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

**Full Changelog**: https://github.com/baladithyab/terraform-provider-truenas/blob/main/CHANGELOG.md

