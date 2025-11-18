terraform {
  required_providers {
    truenas = {
      source  = "baladithyab/truenas"
      version = "~> 0.2"
    }
  }
}

provider "truenas" {
  base_url = var.truenas_url
  api_key  = var.truenas_api_key
}

# Create a virtual machine with cloud-init for static IP configuration
resource "truenas_vm" "ubuntu_static" {
  name        = "ubuntu-static-ip"
  description = "Ubuntu 22.04 with static IP configuration via cloud-init"
  vcpus       = 2
  memory      = 4096
  autostart   = true

  cloud_init {
    user_data      = file("${path.module}/user-data")
    meta_data      = file("${path.module}/meta-data")
    network_config = file("${path.module}/network-config")
    filename       = "cloud-init-${self.name}.iso"
    upload_path    = "/mnt/${var.pool_name}/isos/"
  }
}

# Create a virtual machine with cloud-init for DHCP configuration
resource "truenas_vm" "ubuntu_dhcp" {
  name        = "ubuntu-dhcp"
  description = "Ubuntu 22.04 with DHCP configuration via cloud-init"
  vcpus       = 2
  memory      = 4096
  autostart   = true

  cloud_init {
    user_data   = file("${path.module}/user-data-dhcp")
    meta_data   = file("${path.module}/meta-data-dhcp")
    filename    = "cloud-init-${self.name}.iso"
    upload_path = "/mnt/${var.pool_name}/isos/"
  }
}

# Output VM information
output "ubuntu_static_vm_info" {
  value = {
    id     = truenas_vm.ubuntu_static.id
    name   = truenas_vm.ubuntu_static.name
    status = truenas_vm.ubuntu_static.status
  }
}

output "ubuntu_dhcp_vm_info" {
  value = {
    id     = truenas_vm.ubuntu_dhcp.id
    name   = truenas_vm.ubuntu_dhcp.name
    status = truenas_vm.ubuntu_dhcp.status
  }
}
