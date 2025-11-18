# Release Notes v0.2.23 - Cloud-Init Support for VMs

**Release Date:** 2025-11-18
**Type:** Feature Release

## ☁️ New Feature: Cloud-Init Support

This release introduces native support for **Cloud-Init** configuration in Virtual Machines. You can now provision VMs with custom user-data and meta-data directly from Terraform, enabling automated configuration of operating systems upon first boot.

### What's New

#### Cloud-Init Configuration Block
The `truenas_vm` resource now supports a `cloud_init` block that allows you to specify:
- `user_data`: Custom cloud-init user-data (YAML format)
- `meta_data`: Custom cloud-init meta-data (YAML format)
- `filename`: (Optional) Name of the generated ISO file
- `upload_path`: (Optional) Path where the ISO will be stored

The provider automatically:
1. Generates a valid Cloud-Init ISO containing your configuration
2. Uploads it to the specified location on your TrueNAS system
3. Attaches it as a CD-ROM device to the VM
4. Handles cleanup when the VM is destroyed

### Example Usage

```terraform
resource "truenas_vm" "ubuntu_cloud" {
  name        = "ubuntu-cloud-init"
  description = "Ubuntu with Cloud-Init"
  vcpus       = 2
  memory      = 4096
  autostart   = true

  cloud_init {
    user_data = <<EOF
#cloud-config
hostname: ubuntu-server
manage_etc_hosts: true
users:
  - name: ubuntu
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: users, admin
    home: /home/ubuntu
    shell: /bin/bash
    lock_passwd: false
    ssh_authorized_keys:
      - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC...
packages:
  - qemu-guest-agent
  - nginx
runcmd:
  - systemctl enable --now qemu-guest-agent
EOF
    
    meta_data = <<EOF
instance-id: ubuntu-cloud-init-001
local-hostname: ubuntu-server
EOF
  }
  
  # ... other configuration ...
}
```

### 🐛 Bug Fixes & Improvements

- **VM Device Ordering**: Improved handling of device boot order
- **ISO Management**: Better handling of ISO file uploads and cleanup

### 📚 Documentation

- Updated `truenas_vm` resource documentation with Cloud-Init examples and parameter reference.

---

**Download:** Available from the Terraform Registry and GitHub Releases
**Documentation:** [Complete Documentation](../../docs/)
**Support:** [GitHub Issues](https://github.com/baladithyab/terraform-provider-truenas/issues)