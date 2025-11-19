---
page_title: "truenas_vm Resource - terraform-provider-truenas"
subcategory: "Virtual Machines"
description: |-
  Manages a TrueNAS Scale virtual machine.
---

# truenas_vm (Resource)

Manages a virtual machine on TrueNAS Scale. This resource handles the complete VM lifecycle including creation, configuration updates, and deletion.

## Example Usage

### Basic VM

```terraform
resource "truenas_vm" "ubuntu" {
  name        = "ubuntu-server"
  description = "Ubuntu 22.04 Server"
  
  vcpus  = 2
  memory = 4096
  
  autostart = true
}
```

### VM with Custom CPU Configuration

```terraform
resource "truenas_vm" "windows" {
  name        = "windows-server"
  description = "Windows Server 2022"
  
  vcpus    = 4
  cores    = 2
  threads  = 2
  memory   = 8192
  
  cpu_mode  = "HOST-PASSTHROUGH"
  cpu_model = null
  
  autostart = false
  time      = "LOCAL"
}
```

### VM with Display Configuration

```terraform
resource "truenas_vm" "desktop" {
  name        = "ubuntu-desktop"
  description = "Ubuntu Desktop with VNC"
  
  vcpus  = 4
  memory = 8192
  
  display = {
    port       = 5900
    resolution = "1920x1080"
    bind       = "0.0.0.0"
    password   = var.vnc_password
    web        = true
  }
  
  autostart = true
}
```

### Complete VM with All Options

```terraform
resource "truenas_vm" "production" {
  name        = "production-app"
  description = "Production application server"
  
  vcpus    = 8
  cores    = 4
  threads  = 2
  memory   = 16384
  
  cpu_mode  = "HOST-PASSTHROUGH"
  cpu_model = null
  
  bootloader           = "UEFI"
  grubconfig           = ""
  shutdown_timeout     = 90
  ensure_display_device = true
  hide_from_msr        = false
  
  time = "LOCAL"
  
  autostart = true
  
  display = {
    port       = 5900
    resolution = "1920x1080"
    bind       = "0.0.0.0"
    web        = true
  }
}
```

### VM with Cloud-Init

```terraform
resource "truenas_vm" "cloud_vm" {
  name        = "ubuntu-cloud"
  description = "Ubuntu with Cloud-Init"
  vcpus       = 2
  memory      = 4096
  autostart   = true

  cloud_init {
    user_data = <<EOF
#cloud-config
hostname: ubuntu-server
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
runcmd:
  - systemctl enable --now qemu-guest-agent
EOF
    
    meta_data = <<EOF
instance-id: ubuntu-cloud-001
local-hostname: ubuntu-server
EOF
    
    # Optional: Customize boot order (default is 10000)
    # Lower numbers boot first
    device_order = 5000
  }
}
```

## Schema

### Required

- `name` (String) The name of the virtual machine. Must be unique.
- `vcpus` (Number) Number of virtual CPUs to allocate. Minimum: 1
- `memory` (Number) Amount of memory in MB to allocate. Minimum: 256

### Optional

- `description` (String) Description of the virtual machine.
- `cores` (Number) Number of cores per socket. If not specified, calculated from vcpus.
- `threads` (Number) Number of threads per core. If not specified, calculated from vcpus.
- `cpu_mode` (String) CPU mode. Options: `CUSTOM`, `HOST-MODEL`, `HOST-PASSTHROUGH`. Default: `HOST-PASSTHROUGH`
- `cpu_model` (String) CPU model to emulate when cpu_mode is CUSTOM. Default: `null`
- `bootloader` (String) Boot loader type. Options: `UEFI`, `UEFI_CSM`, `GRUB`. Default: `UEFI`
- `grubconfig` (String) GRUB configuration for GRUB bootloader.
- `shutdown_timeout` (Number) Timeout in seconds for graceful shutdown. Default: 90
- `autostart` (Boolean) Whether to automatically start the VM on boot. Default: false
- `time` (String) Time synchronization mode. Options: `LOCAL`, `UTC`. Default: `LOCAL`
- `ensure_display_device` (Boolean) Ensure a display device exists. Default: true
- `hide_from_msr` (Boolean) Hide the VM from MSR (Windows specific). Default: false
- `display` (Block) Display/VNC configuration. See [Display Configuration](#display-configuration) below.
- `cloud_init` (Block) Cloud-Init configuration. See [Cloud-Init Configuration](#cloud-init-configuration) below.

### Display Configuration

The `display` block supports:

- `port` (Number) VNC port number. Default: 5900
- `resolution` (String) Display resolution. Options: `1920x1200`, `1920x1080`, `1600x1200`, `1600x900`, `1400x1050`, `1280x1024`, `1280x720`, `1024x768`, `800x600`, `640x480`. Default: `1024x768`
- `bind` (String) IP address to bind VNC server. Default: `0.0.0.0`
- `password` (String, Sensitive) VNC password for authentication.
- `web` (Boolean) Enable web VNC access. Default: false

### Cloud-Init Configuration

The `cloud_init` block supports:

- `user_data` (String) Cloud-init user-data configuration (YAML format).
- `meta_data` (String) Cloud-init meta-data configuration (YAML format).
- `network_config` (String) Cloud-init network-config configuration (YAML format) for static IP assignment.
- `filename` (String, Optional) Name of the generated ISO file. Defaults to `cloud-init-{vm_name}.iso`.
- `upload_path` (String, Optional) Directory to upload the ISO to. Defaults to `/mnt/{first_pool}/isos/`.
- `device_order` (Number, Optional) Boot order for the cloud-init ISO device. Defaults to `10000` to ensure it boots after regular devices. Lower values boot first.

#### Static IP Assignment Example

```terraform
resource "truenas_vm" "static_ip_vm" {
  name        = "ubuntu-static"
  description = "Ubuntu with static IP via cloud-init"
  vcpus       = 2
  memory      = 4096
  autostart   = true

  cloud_init {
    user_data = <<EOF
#cloud-config
hostname: ubuntu-static
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
runcmd:
  - systemctl enable --now qemu-guest-agent
EOF
    
    network_config = <<EOF
version: 2
ethernets:
  eth0:
    dhcp4: no
    addresses: [192.168.1.100/24]
    gateway4: 192.168.1.1
    nameservers:
      addresses: [8.8.8.8, 8.8.4.4]
EOF
  }
}
```

### Read-Only

- `id` (String) The ID of the virtual machine.
- `status` (Object) Current status of the VM including state and PID.

## Import

Virtual machines can be imported using the VM name:

```shell
terraform import truenas_vm.example ubuntu-server
```

## Notes

### VM Lifecycle

- The VM is created in a stopped state by default
- Use `autostart = true` to automatically start the VM when TrueNAS boots
- Changing most attributes requires stopping the VM first
- The resource manages the VM configuration but not its running state

### CPU Configuration

- `vcpus` is the total number of virtual CPUs
- If `cores` and `threads` are not specified, they default to: `cores = vcpus`, `threads = 1`
- For optimal performance: `vcpus = cores × threads`

### Memory Management

- Memory is specified in MB
- Minimum is 256 MB
- Ensure sufficient host memory is available
- Memory changes require VM restart

### Display/VNC Access

- VNC is enabled by default on port 5900
- Use `display.password` to secure VNC access
- `display.web` enables access through TrueNAS web interface
- Multiple VMs should use different VNC ports

### Boot Loader

- `UEFI` is recommended for modern operating systems
- `UEFI_CSM` provides legacy BIOS compatibility
- `GRUB` is for specific use cases (BSD, older Linux)

## See Also

- [truenas_vm_device](vm_device) - Attach devices (disks, NICs) to VMs
- [truenas_vm Data Source](../data-sources/vm) - Query VM information
- [truenas_vm_guest_info Data Source](../data-sources/vm_guest_info) - Get VM IP addresses
- [VM IP Discovery Guide](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/guides/vm_ip_discovery)