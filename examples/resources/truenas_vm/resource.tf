# Create a basic virtual machine
resource "truenas_vm" "example" {
  name        = "ubuntu-vm"
  description = "Ubuntu Server VM"
  memory      = 4096  # 4GB RAM
  vcpus       = 2
  cores       = 1
  threads     = 1
  autostart   = true
  bootloader  = "UEFI"
  cpu_mode    = "HOST-MODEL"
  time        = "LOCAL"
}

# Create a VM with custom CPU configuration
resource "truenas_vm" "custom_cpu" {
  name         = "custom-vm"
  description  = "VM with custom CPU"
  memory       = 8192  # 8GB RAM
  vcpus        = 4
  cores        = 2
  threads      = 2
  autostart    = false
  bootloader   = "UEFI"
  cpu_mode     = "CUSTOM"
  cpu_model    = "Haswell"
  machine_type = "q35"
  time         = "UTC"
}

# Create a VM with memory ballooning
resource "truenas_vm" "ballooning" {
  name        = "flexible-vm"
  description = "VM with memory ballooning"
  memory      = 8192   # 8GB max
  min_memory  = 2048   # 2GB min
  vcpus       = 2
  autostart   = true
  bootloader  = "UEFI"
}

# Import an existing VM
# terraform import truenas_vm.existing 1

