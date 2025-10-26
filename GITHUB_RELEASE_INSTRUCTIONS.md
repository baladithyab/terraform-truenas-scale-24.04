# GitHub Release Instructions for v0.2.0

## ✅ Pre-Release Checklist (COMPLETED)

- [x] Code verified and tested
- [x] Documentation updated
- [x] Version references updated in README.md
- [x] CHANGELOG.md updated
- [x] Release notes created (RELEASE_NOTES_v0.2.0.md)
- [x] All changes committed and pushed
- [x] Tag v0.2.0 created and pushed
- [x] Binaries built for all platforms
- [x] SHA256 checksums generated

## 📦 Built Artifacts

All binaries are in the `dist/` directory:

```
dist/
├── terraform-provider-truenas_v0.2.0_darwin_amd64 (26MB)
├── terraform-provider-truenas_v0.2.0_darwin_arm64 (25MB)
├── terraform-provider-truenas_v0.2.0_linux_amd64 (25MB)
├── terraform-provider-truenas_v0.2.0_linux_arm64 (24MB)
├── terraform-provider-truenas_v0.2.0_windows_amd64.exe (26MB)
└── terraform-provider-truenas_v0.2.0_SHA256SUMS
```

**SHA256 Checksums:**
```
70ba6a50d39029387c59090ae71c664bb91303e3e78d45aaf73237b95c25fd41  terraform-provider-truenas_v0.2.0_darwin_amd64
b853562922ea0a4e0456af5543055b5d9ef24aa13ccbe5c5c3ffb31643fcafa6  terraform-provider-truenas_v0.2.0_darwin_arm64
06645e188b85dab97f1bab7bfd6eb0b61228ff8c5c6b0662b1ca45de8b45a1b3  terraform-provider-truenas_v0.2.0_linux_amd64
f7b2b0b37c5d085434950d2f1a75a2a733df9c9378c2b0ea632fc070a06f824d  terraform-provider-truenas_v0.2.0_linux_arm64
6297aa3621ac13f3e5b012fc3b7e83ab9a0d2098badf23ae3c49312f3b3a2dfc  terraform-provider-truenas_v0.2.0_windows_amd64.exe
```

## 🚀 Creating the GitHub Release

### Step 1: Navigate to Releases

1. Go to: https://github.com/baladithyab/terraform-truenas-scale-24.04/releases
2. Click "Draft a new release"

### Step 2: Configure Release

**Tag**: `v0.2.0` (already created and pushed)

**Release Title**: `v0.2.0 - Data Sources and Import Fixes`

**Description**: Copy the content from `RELEASE_NOTES_v0.2.0.md` or use this:

```markdown
## 🎉 What's New in v0.2.0

This release fixes critical issues identified during community testing and ensures all documented features are fully functional.

### 🐛 Critical Fixes

#### Data Sources Now Working ✅
- **`data.truenas_pool`** - Query pool information (status, health, capacity)
- **`data.truenas_dataset`** - Query dataset information
- **Fixed**: "no schema available" errors that prevented data source usage

#### Import Functionality Verified ✅
- All 14 resources now support import
- **NFS shares** - Import by ID works correctly
- **SMB shares** - Import by ID works correctly
- **Snapshots** - Import with custom format (`dataset@snapshotname`)

#### Snapshot Resources Operational ✅
- **`truenas_snapshot`** - Manual snapshot creation
- **`truenas_periodic_snapshot_task`** - Automated snapshot scheduling
- **Fixed**: Schema validation errors

### ✅ All Features Verified

**Resources (14)**
- ✅ `truenas_dataset` - ZFS dataset management
- ✅ `truenas_nfs_share` - NFS share management
- ✅ `truenas_smb_share` - SMB/CIFS share management
- ✅ `truenas_user` - User account management
- ✅ `truenas_group` - Group management
- ✅ `truenas_vm` - Virtual machine management
- ✅ `truenas_iscsi_target` - iSCSI target management
- ✅ `truenas_iscsi_extent` - iSCSI extent management
- ✅ `truenas_iscsi_portal` - iSCSI portal management
- ✅ `truenas_interface` - Network interface management
- ✅ `truenas_static_route` - Static route management
- ✅ `truenas_chart_release` - Kubernetes application deployment
- ✅ `truenas_snapshot` - ZFS snapshot management
- ✅ `truenas_periodic_snapshot_task` - Automated snapshot scheduling

**Data Sources (2)**
- ✅ `data.truenas_pool` - Query pool information
- ✅ `data.truenas_dataset` - Query dataset information

## 🔄 Upgrading from v0.1.0

**No Breaking Changes** - Simple upgrade:

```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.0"  # Changed from 0.1.0
    }
  }
}
```

Then run:
```bash
terraform init -upgrade
```

## 📚 Documentation

- [Release Notes](RELEASE_NOTES_v0.2.0.md)
- [Changelog](CHANGELOG.md)
- [Import Guide](IMPORT_GUIDE.md)
- [API Coverage](API_COVERAGE.md)
- [Gaps Analysis Response](GAPS_ANALYSIS_RESPONSE.md)

## 🙏 Acknowledgments

Special thanks to the Yggdrasil Infrastructure Team for comprehensive testing and detailed gap analysis!

## 📦 Installation

### Option 1: Terraform Registry (Recommended)
```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.0"
    }
  }
}
```

### Option 2: Manual Installation

Download the appropriate binary for your platform from the assets below, then:

**Linux/macOS:**
```bash
# Extract and install
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/baladithyab/truenas/0.2.0/linux_amd64/
mv terraform-provider-truenas_v0.2.0_linux_amd64 ~/.terraform.d/plugins/registry.terraform.io/baladithyab/truenas/0.2.0/linux_amd64/terraform-provider-truenas_v0.2.0
chmod +x ~/.terraform.d/plugins/registry.terraform.io/baladithyab/truenas/0.2.0/linux_amd64/terraform-provider-truenas_v0.2.0
```

**Windows:**
```powershell
# Extract and install
mkdir $env:APPDATA\terraform.d\plugins\registry.terraform.io\baladithyab\truenas\0.2.0\windows_amd64\
move terraform-provider-truenas_v0.2.0_windows_amd64.exe $env:APPDATA\terraform.d\plugins\registry.terraform.io\baladithyab\truenas\0.2.0\windows_amd64\terraform-provider-truenas_v0.2.0.exe
```

## 🔍 Verification

Verify the download with SHA256 checksums (see `terraform-provider-truenas_v0.2.0_SHA256SUMS`).

**Full Changelog**: https://github.com/baladithyab/terraform-truenas-scale-24.04/blob/main/CHANGELOG.md
```

### Step 3: Upload Binaries

Upload these files from the `dist/` directory:

1. `terraform-provider-truenas_v0.2.0_linux_amd64`
2. `terraform-provider-truenas_v0.2.0_linux_arm64`
3. `terraform-provider-truenas_v0.2.0_darwin_amd64`
4. `terraform-provider-truenas_v0.2.0_darwin_arm64`
5. `terraform-provider-truenas_v0.2.0_windows_amd64.exe`
6. `terraform-provider-truenas_v0.2.0_SHA256SUMS`

### Step 4: Publish Release

- [ ] Check "Set as the latest release"
- [ ] Uncheck "Set as a pre-release" (this is a stable release)
- [ ] Click "Publish release"

## 📢 Post-Release Actions

### 1. Announce Release

Create announcements in:
- GitHub Discussions (if enabled)
- Project README.md (already updated)
- Any community channels

### 2. Monitor for Issues

- Watch GitHub Issues for bug reports
- Respond to community feedback
- Prepare v0.2.1 if critical bugs found

### 3. Update Terraform Registry (if applicable)

If publishing to the official Terraform Registry:
1. Follow Terraform Registry publishing guidelines
2. Ensure GPG signing is set up
3. Create registry manifest
4. Submit for review

## 🎯 Success Criteria

Release is successful when:
- [x] Tag v0.2.0 created and pushed
- [x] Binaries built for all platforms
- [x] SHA256 checksums generated
- [ ] GitHub release published
- [ ] Binaries uploaded
- [ ] Release notes visible
- [ ] Community can download and use

## 📞 Support

If issues arise during release:
- Check GitHub Actions (if configured)
- Verify tag exists: `git tag -l`
- Verify binaries: `ls -lh dist/`
- Check checksums: `shasum -c dist/terraform-provider-truenas_v0.2.0_SHA256SUMS`

## 🔗 Quick Links

- **Repository**: https://github.com/baladithyab/terraform-truenas-scale-24.04
- **Releases**: https://github.com/baladithyab/terraform-truenas-scale-24.04/releases
- **Issues**: https://github.com/baladithyab/terraform-truenas-scale-24.04/issues
- **Tag**: https://github.com/baladithyab/terraform-truenas-scale-24.04/releases/tag/v0.2.0

---

**Status**: Ready for GitHub release publication  
**Next Step**: Create release on GitHub and upload binaries  
**ETA**: Can be done immediately

