# Release Notes - v0.2.15

**Release Date:** 2025-11-03

## 🎉 Major Feature Release

This is a **major feature release** that adds comprehensive resource discovery, GPU/PCI passthrough support, enhanced dataset deletion, and VM IP discovery capabilities to the TrueNAS Terraform Provider.

### 📊 Release Summary

**5 Major Features Added:**
1. **Resource Discovery** - 4 new data sources (NFS, SMB, VMs)
2. **GPU/PCI Passthrough** - 3 new data sources + VM enhancements
3. **Enhanced Dataset Deletion** - force_destroy & recursive_destroy
4. **Pool Data Source Enhancement** - Accept names or IDs
5. **VM IP Discovery** - Password authentication for guest agent

**Total Changes:**
- ✅ 7 new data sources
- ✅ 3 VM resource enhancements (pci_devices, hide_from_msr, ensure_display_device)
- ✅ 2 dataset resource enhancements (force_destroy, recursive_destroy)
- ✅ 1 pool data source enhancement (ID support)
- ✅ 1 guest info enhancement (password auth)
- ✅ 1 new client method (DeleteWithBody)
- ✅ 2 new examples (GPU passthrough, data sources discovery)
- ✅ 1 bug fix (schema validation)

**Eliminates Workarounds:**
- ❌ No more HTTP data sources for resource discovery
- ❌ No more null_resource for PCI device attachment
- ❌ No more null_resource for dataset cleanup
- ❌ No more manual snapshot deletion

---

## ✨ What's New

### 1. Resource Discovery Data Sources

Four new data sources for discovering existing TrueNAS resources:

**`truenas_nfs_shares`** - List all NFS shares
```hcl
data "truenas_nfs_shares" "all" {}

output "nfs_paths" {
  value = [for share in data.truenas_nfs_shares.all.shares : share.path]
}
```

**`truenas_smb_shares`** - List all SMB/CIFS shares
```hcl
data "truenas_smb_shares" "all" {}

output "smb_names" {
  value = [for share in data.truenas_smb_shares.all.shares : share.name]
}
```

**`truenas_vms`** - List all VMs
```hcl
data "truenas_vms" "all" {}

output "running_vms" {
  value = [for vm in data.truenas_vms.all.vms : vm.name if vm.status == "RUNNING"]
}
```

**`truenas_vm`** - Query specific VM by name or ID
```hcl
data "truenas_vm" "my_vm" {
  name = "ubuntu-server"
}

output "vm_status" {
  value = data.truenas_vm.my_vm.status
}
```

**Benefits:**
- ✅ Eliminates HTTP data source workarounds
- ✅ Type-safe with IDE auto-completion
- ✅ Better error messages
- ✅ Consistent with Terraform best practices

---

### 2. GPU/PCI Passthrough Support

Three new data sources and VM resource enhancements for GPU passthrough:

**`truenas_gpu_pci_choices`** - Discover available GPUs
```hcl
data "truenas_gpu_pci_choices" "gpus" {}

output "available_gpus" {
  value = data.truenas_gpu_pci_choices.gpus.choices
}
```

**`truenas_vm_pci_passthrough_devices`** - List PCI devices with IOMMU info
```hcl
data "truenas_vm_pci_passthrough_devices" "devices" {}

output "nvidia_gpus" {
  value = [for dev in data.truenas_vm_pci_passthrough_devices.devices.devices :
    dev if can(regex("NVIDIA", dev.description))
  ]
}
```

**`truenas_vm_iommu_enabled`** - Check IOMMU status
```hcl
data "truenas_vm_iommu_enabled" "check" {}

output "iommu_ready" {
  value = data.truenas_vm_iommu_enabled.check.enabled
}
```

**VM Resource Enhancements:**
```hcl
resource "truenas_vm" "gpu_vm" {
  name   = "ml-workstation"
  memory = 16384
  vcpus  = 8

  # GPU passthrough
  pci_devices = [{
    pci_slot = "0000:01:00.0"  # GPU PCI address
  }]

  # Hide hypervisor from MSR (for NVIDIA drivers)
  hide_from_msr = true

  # Disable virtual display (GPU provides display)
  ensure_display_device = false
}
```

**Use Cases:**
- Machine learning workloads
- GPU-accelerated rendering
- Gaming VMs
- CUDA/OpenCL compute

---

### 3. Enhanced Dataset Deletion

Dataset resource now supports automatic cleanup:

**New Attributes:**
- `force_destroy` - Force delete even if busy or has snapshots
- `recursive_destroy` - Recursively delete child datasets

```hcl
resource "truenas_dataset" "vm_storage" {
  name = "tank/vms/storage"
  type = "FILESYSTEM"

  # Auto-cleanup on destroy
  force_destroy     = true
  recursive_destroy = true
}
```

**Benefits:**
- ✅ No more manual snapshot cleanup
- ✅ No more "dataset is busy" errors
- ✅ Eliminates null_resource workarounds
- ✅ Clean terraform destroy

**Technical Details:**
- Automatically deletes snapshots before dataset deletion
- Passes `recursive=true` and `force=true` to TrueNAS API
- Uses new `DeleteWithBody` client method for JSON body in DELETE requests

---

### 4. Enhanced Pool Data Source

Pool data source now accepts both names and numeric IDs:

**Before:**
```hcl
data "truenas_pool" "tank" {
  name = "tank"  # Only names worked
}
```

**After:**
```hcl
# By name
data "truenas_pool" "tank" {
  name = "tank"
}

# By ID
data "truenas_pool" "tank" {
  id = 1
}
```

---

### 5. Password Authentication for Guest Info

The `truenas_vm_guest_info` data source now supports both SSH key and password authentication:

**Before v0.2.15** - Only SSH keys:
```hcl
data "truenas_vm_guest_info" "vm" {
  vm_name      = "my-vm"
  truenas_host = "10.0.0.83"
  ssh_user     = "root"
  ssh_key_path = "~/.ssh/id_rsa"  # Required
}
```

**After v0.2.15** - SSH keys OR password:
```hcl
data "truenas_vm_guest_info" "vm" {
  vm_name      = "my-vm"
  truenas_host = "10.0.0.83"
  ssh_user     = "root"
  ssh_password = var.truenas_ssh_password  # New!
}
```

### New Example: Talos Linux VM with Guest Agent

Added comprehensive example showing how to:
- Create a Talos Linux VM with QEMU guest agent
- Retrieve VM IP address automatically
- Get OS information and hostname
- Use password authentication for guest queries

See `examples/talos-vm-with-guest-agent/` for complete working example.

---

## 🔧 Changes

### Enhanced Data Source: `truenas_vm_guest_info`

**New Attribute:**
- `ssh_password` (Optional, Sensitive) - SSH password for authentication

**Updated Behavior:**
- Supports both `ssh_key_path` and `ssh_password`
- Uses `sshpass` for password authentication
- Falls back to key-based auth if password not provided
- Maintains backward compatibility

**Example Output:**
```hcl
output "vm_ip" {
  value = data.truenas_vm_guest_info.talos.ip_addresses[0]
  # Output: "10.0.0.59"
}

output "vm_os" {
  value = data.truenas_vm_guest_info.talos.os_version
  # Output: "6.12.52-talos"
}
```

---

## 📖 How to Get VM IP Addresses

### Prerequisites

1. **QEMU Guest Agent** must be installed in the VM:
   - **Talos Linux**: Built-in, starts automatically ✅
   - **Ubuntu/Debian**: `sudo apt-get install qemu-guest-agent`
   - **RHEL/CentOS**: `sudo yum install qemu-guest-agent`
   - **Windows**: Install QEMU Guest Agent from VirtIO drivers

2. **sshpass** must be installed on the Terraform host (for password auth):
   ```bash
   # Ubuntu/Debian
   sudo apt-get install sshpass
   
   # macOS
   brew install hudochenkov/sshpass/sshpass
   ```

3. **SSH access** to TrueNAS host (password or key)

### Step-by-Step Example

**1. Create VM with guest agent support:**

```hcl
resource "truenas_vm" "talos" {
  name   = "talos-demo"
  memory = 2048
  vcpus  = 2
  
  bootloader      = "UEFI"
  start_on_create = true
  
  cdrom_devices = [{
    path  = "/mnt/pool/isos/talos-v1.11.3-metal-amd64.iso"
    order = 1001
  }]
  
  nic_devices = [{
    type       = "VIRTIO"
    nic_attach = "eno1"
    order      = 1000
  }]
  
  disk_devices = [{
    path  = "/dev/zvol/pool/talos-disk"
    type  = "VIRTIO"
    order = 1002
  }]
}
```

**2. Query guest information:**

```hcl
data "truenas_vm_guest_info" "talos" {
  vm_name      = truenas_vm.talos.name
  truenas_host = "10.0.0.83"
  ssh_user     = "root"
  ssh_password = var.truenas_ssh_password
  
  depends_on = [truenas_vm.talos]
}
```

**3. Use the IP address:**

```hcl
output "talos_ip" {
  value = length(data.truenas_vm_guest_info.talos.ip_addresses) > 0 ? (
    data.truenas_vm_guest_info.talos.ip_addresses[0]
  ) : "Waiting for DHCP..."
}

# Use in other resources
resource "null_resource" "configure_talos" {
  provisioner "local-exec" {
    command = "talosctl config endpoint ${data.truenas_vm_guest_info.talos.ip_addresses[0]}"
  }
}
```

---

## 🎯 Use Cases

### 1. Automated Kubernetes Cluster Setup

```hcl
# Create 3 Talos nodes
resource "truenas_vm" "k8s_nodes" {
  count = 3
  name  = "k8s-node-${count.index}"
  # ... VM configuration ...
}

# Get IPs for all nodes
data "truenas_vm_guest_info" "k8s_nodes" {
  count        = 3
  vm_name      = truenas_vm.k8s_nodes[count.index].name
  truenas_host = var.truenas_host
  ssh_password = var.truenas_ssh_password
}

# Configure Talos cluster
resource "null_resource" "bootstrap_k8s" {
  provisioner "local-exec" {
    command = <<-EOT
      talosctl config endpoint ${join(" ", data.truenas_vm_guest_info.k8s_nodes[*].ip_addresses[0])}
      talosctl bootstrap --nodes ${data.truenas_vm_guest_info.k8s_nodes[0].ip_addresses[0]}
    EOT
  }
}
```

### 2. Dynamic Inventory for Ansible

```hcl
output "ansible_inventory" {
  value = templatefile("${path.module}/inventory.tpl", {
    vms = {
      for i, vm in data.truenas_vm_guest_info.nodes :
      vm.hostname => vm.ip_addresses[0]
    }
  })
}
```

### 3. VM Health Monitoring

```hcl
data "truenas_vm_guest_info" "app_server" {
  vm_name      = "app-server"
  truenas_host = var.truenas_host
  ssh_password = var.truenas_ssh_password
}

resource "datadog_monitor" "vm_health" {
  name    = "App Server Health"
  type    = "service check"
  message = "App server at ${data.truenas_vm_guest_info.app_server.ip_addresses[0]} is down"
  
  query = "\"http.can_connect\".over(\"instance:${data.truenas_vm_guest_info.app_server.ip_addresses[0]}\").last(2).count_by_status()"
}
```

---

## 📚 New Example: `talos-vm-with-guest-agent`

Complete working example demonstrating:

**Features:**
- ✅ Talos Linux VM creation
- ✅ QEMU guest agent integration
- ✅ Automatic IP discovery
- ✅ OS information retrieval
- ✅ Password-based SSH authentication
- ✅ Comprehensive documentation

**Files:**
- `main.tf` - Complete VM configuration with guest info
- `README.md` - Detailed usage guide and troubleshooting
- `terraform.tfvars.example` - Example variables

**Quick Start:**
```bash
cd examples/talos-vm-with-guest-agent
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your credentials
terraform init
terraform apply

# Wait 60 seconds for Talos to boot
sleep 60
terraform refresh

# View VM IP
terraform output guest_info
```

---

## 🔄 Migration Guide

### From v0.2.14 to v0.2.15

**No breaking changes!** This release is fully backward compatible.

**If you're using SSH keys** - No changes needed:
```hcl
# This still works exactly the same
data "truenas_vm_guest_info" "vm" {
  vm_name      = "my-vm"
  truenas_host = "10.0.0.83"
  ssh_user     = "root"
  ssh_key_path = "~/.ssh/id_rsa"
}
```

**If you want to use passwords** - Add `ssh_password`:
```hcl
data "truenas_vm_guest_info" "vm" {
  vm_name      = "my-vm"
  truenas_host = "10.0.0.83"
  ssh_user     = "root"
  ssh_password = var.truenas_ssh_password  # New option
}
```

---

## 🐛 Bug Fixes

**Schema Validation Fix:**
- Fixed `force_destroy` and `recursive_destroy` attributes in dataset resource
- These attributes now properly marked as `Computed: true` (required when using `Default` values)
- Resolves Terraform plugin framework validation error

---

## 📦 Installation

### Local Installation

```bash
git clone https://github.com/your-org/terraform-provider-truenas.git
cd terraform-provider-truenas
make install
```

This installs the provider to:
```
~/.terraform.d/plugins/terraform-providers/truenas/truenas/0.2.15/linux_amd64/
```

### Terraform Configuration

```hcl
terraform {
  required_providers {
    truenas = {
      source  = "terraform-providers/truenas/truenas"
      version = "0.2.15"
    }
  }
}
```

---

## 🧪 Testing

Comprehensive end-to-end testing performed:

**Test Environment:**
- TrueNAS Scale 24.04
- Talos Linux v1.11.3
- Ubuntu 22.04 (Terraform host)

**Test Results:**
- ✅ **Resource Discovery**: All data sources return accurate information
- ✅ **GPU Passthrough**: PCI devices discovered and attachable to VMs
- ✅ **Dataset Deletion**: force_destroy and recursive_destroy work correctly
- ✅ **Pool Data Source**: Accepts both names and numeric IDs
- ✅ **VM Creation**: Talos VM created with all device types
- ✅ **Guest Agent**: Responds within 60 seconds, returns IP/OS info
- ✅ **Password Auth**: sshpass authentication working
- ✅ **IP Discovery**: Retrieved 10.0.0.59 successfully
- ✅ **Network**: VM pingable and accessible
- ✅ **Cleanup**: Clean destruction with no orphaned resources

---

## 📝 Known Limitations

1. **Boot Time**: VMs need 60-90 seconds to boot before guest agent is available
   - **Workaround**: Use `depends_on` and run `terraform refresh` after initial apply

2. **sshpass Required**: Password authentication requires `sshpass` utility
   - **Workaround**: Use SSH key authentication instead

3. **Guest Agent Required**: Data source only works if QEMU guest agent is installed in VM
   - **Workaround**: Install guest agent in VM or use alternative IP discovery methods

---

## 🙏 Acknowledgments

- Talos Linux team for excellent QEMU guest agent integration
- TrueNAS Scale team for robust libvirt/virsh support
- Community feedback on IP discovery requirements

---

## 📞 Support

- **Issues**: https://github.com/your-org/terraform-provider-truenas/issues
- **Discussions**: https://github.com/your-org/terraform-provider-truenas/discussions
- **Documentation**: See `examples/` directory

---

## 🔜 What's Next (v0.2.16)

Planned features:
- Cloud-init support for VM configuration
- Snapshot management improvements
- Enhanced error messages for guest agent failures
- Support for Windows guest agent

---

**Full Changelog**: https://github.com/your-org/terraform-provider-truenas/compare/v0.2.14...v0.2.15

