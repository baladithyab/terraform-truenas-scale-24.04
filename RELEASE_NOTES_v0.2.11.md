# Release Notes - v0.2.11

**Release Date**: October 30, 2025  
**Provider**: TrueNAS Terraform Provider  
**Repository**: https://github.com/baladithyab/terraform-truenas-scale-24.04  
**Compatibility**: TrueNAS Scale 24.04

---

## 🚀 New Feature Release

v0.2.11 adds **automatic VM startup** functionality to eliminate the manual step of starting VMs after Terraform creates them.

---

## ✨ What's New

### Feature: `start_on_create` Attribute for VMs

**Problem**:
- v0.2.10 created VMs successfully but left them in STOPPED state
- Users had to manually start VMs after creation
- This added an extra manual step to infrastructure deployment

**Solution**:
- New `start_on_create` attribute automatically starts VMs after creation
- Calls TrueNAS API `/vm/id/{id}/start` endpoint
- Graceful error handling - start failures don't block VM creation

---

## 📝 Usage

### Basic Example
```hcl
resource "truenas_vm" "worker" {
  name            = "talos_worker_01"
  memory          = 16384
  vcpus           = 8
  cores           = 1
  threads         = 1
  bootloader      = "UEFI"
  cpu_mode        = "CUSTOM"
  time            = "LOCAL"
  autostart       = true
  start_on_create = true  # ✨ NEW: Start VM immediately after creation
}
```

### Before v0.2.11 (Manual Start Required)
```bash
terraform apply
# ✅ VM created successfully
# ❌ VM is in STOPPED state
# ⚠️  Manual step required: Start VM in TrueNAS UI or via API
```

### After v0.2.11 (Automatic Start)
```hcl
resource "truenas_vm" "worker" {
  name            = "talos_worker_01"
  memory          = 16384
  start_on_create = true  # ✨ Automatically start after creation
}
```

```bash
terraform apply
# ✅ VM created successfully
# ✅ VM started automatically
# ✅ VM is in RUNNING state
# ✅ No manual steps required!
```

---

## 🔧 Technical Details

### API Endpoints Used

**VM Start**:
```
POST /vm/id/{id}/start
```

**Response**:
- Success: Returns job ID (e.g., `804`)
- Already running: Returns error with errno 14
- Failure: Returns error message

### Implementation

**Model Changes**:
```go
type VMResourceModel struct {
    // ... existing fields ...
    StartOnCreate types.Bool `tfsdk:"start_on_create"`  // ✨ NEW
}
```

**Schema Changes**:
```go
"start_on_create": schema.BoolAttribute{
    MarkdownDescription: "Start VM immediately after creation (default: false)",
    Optional:            true,
},
```

**Create Logic**:
```go
// After VM creation
if !data.StartOnCreate.IsNull() && data.StartOnCreate.ValueBool() {
    startEndpoint := fmt.Sprintf("/vm/id/%s/start", data.ID.ValueString())
    _, err := r.client.Post(startEndpoint, nil)
    if err != nil {
        resp.Diagnostics.AddWarning(
            "VM Start Warning",
            fmt.Sprintf("VM created successfully but failed to start: %s", err),
        )
    }
}
```

---

## ⚙️ Behavior

### Default Behavior (start_on_create not set or false)
```hcl
resource "truenas_vm" "example" {
  name   = "my_vm"
  memory = 4096
  # start_on_create defaults to false
}
```
- ✅ VM is created
- ❌ VM is NOT started
- ℹ️  VM remains in STOPPED state

### Explicit Start (start_on_create = true)
```hcl
resource "truenas_vm" "example" {
  name            = "my_vm"
  memory          = 4096
  start_on_create = true  # Explicitly start after creation
}
```
- ✅ VM is created
- ✅ VM is started automatically
- ✅ VM is in RUNNING state

### Error Handling

**If VM creation succeeds but start fails**:
```
Warning: VM Start Warning

VM created successfully but failed to start: <error message>. 
You can start it manually.
```

- ✅ VM is still created successfully
- ⚠️  Warning is shown (not an error)
- ℹ️  Terraform apply succeeds
- 👉 User can start VM manually if needed

---

## 🔄 Upgrading from v0.2.10

### No Breaking Changes

**Step 1**: Update version:
```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.11"  # Changed from 0.2.10
    }
  }
}
```

**Step 2**: (Optional) Add `start_on_create` to VMs:
```hcl
resource "truenas_vm" "worker" {
  name            = "talos_worker_01"
  memory          = 16384
  start_on_create = true  # ✨ NEW: Add this line to auto-start
}
```

**Step 3**: Upgrade:
```bash
terraform init -upgrade
terraform plan
terraform apply
```

**All v0.2.10 configurations will work in v0.2.11 without changes!**

---

## 📊 Use Cases

### 1. Kubernetes Worker Nodes
```hcl
resource "truenas_vm" "talos_worker" {
  count           = 3
  name            = "talos_worker_${count.index + 1}"
  memory          = 16384
  vcpus           = 8
  start_on_create = true  # Start all workers automatically
}
```

### 2. Development VMs
```hcl
resource "truenas_vm" "dev_vm" {
  name            = "dev_environment"
  memory          = 8192
  start_on_create = true  # Ready to use immediately
}
```

### 3. Production VMs (Manual Start)
```hcl
resource "truenas_vm" "prod_vm" {
  name            = "production_app"
  memory          = 32768
  start_on_create = false  # Create but don't start (manual verification)
}
```

---

## 🎯 Benefits

### Before v0.2.11
1. Run `terraform apply`
2. VMs created in STOPPED state
3. **Manual step**: Log into TrueNAS UI
4. **Manual step**: Start each VM individually
5. **Manual step**: Wait for VMs to boot
6. Continue with deployment

### After v0.2.11
1. Run `terraform apply` with `start_on_create = true`
2. VMs created AND started automatically
3. ✅ **No manual steps required!**
4. Continue with deployment immediately

**Time Saved**: Eliminates 3 manual steps per deployment!

---

## 📈 Release Statistics

| Metric | Value |
|--------|-------|
| **Version** | 0.2.11 |
| **Release Type** | Feature Addition |
| **New Features** | 1 (VM auto-start) |
| **Breaking Changes** | 0 |
| **Files Changed** | 2 |
| **Lines Changed** | +51, -1 |
| **Platforms** | 5 |

---

## 🔗 Related Attributes

### `autostart` vs `start_on_create`

**`autostart`** (existing):
- Controls whether VM starts automatically when TrueNAS boots
- Persistent setting stored in TrueNAS
- Affects VM behavior on TrueNAS host reboots

**`start_on_create`** (new):
- Controls whether VM starts immediately after Terraform creates it
- One-time action during resource creation
- Does not affect VM behavior on TrueNAS host reboots

**Recommended Configuration**:
```hcl
resource "truenas_vm" "example" {
  name            = "my_vm"
  memory          = 4096
  autostart       = true   # Start on TrueNAS boot
  start_on_create = true   # Start immediately after creation
}
```

---

## 🚀 What's Next

### Planned for v0.3.0
- VM stop/restart operations
- Replication task management
- Cloud sync task management
- Service management
- Certificate management

---

## 📞 Support

- **Download v0.2.11**: https://github.com/baladithyab/terraform-truenas-scale-24.04/releases/tag/v0.2.11
- **GitHub Issues**: https://github.com/baladithyab/terraform-truenas-scale-24.04/issues
- **Repository**: https://github.com/baladithyab/terraform-truenas-scale-24.04

---

**🎉 v0.2.11 eliminates manual VM startup steps!** 🎉

**Recommendation**: All users deploying VMs should upgrade to v0.2.11 and use `start_on_create = true` for automatic VM startup.

