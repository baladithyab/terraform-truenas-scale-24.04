# Create a basic virtual machine that starts automatically
resource "truenas_vm" "example" {
  name          = "ubuntu-vm"
  description   = "Ubuntu Server VM"
  memory        = 4096 # 4GB RAM
  vcpus         = 2
  cores         = 1
  threads       = 1
  autostart     = true
  bootloader    = "UEFI"
  cpu_mode      = "HOST-MODEL"
  time          = "LOCAL"
  desired_state = "RUNNING" # Start the VM immediately
}

# Create a VM with custom CPU configuration (stopped)
resource "truenas_vm" "custom_cpu" {
  name          = "custom-vm"
  description   = "VM with custom CPU"
  memory        = 8192 # 8GB RAM
  vcpus         = 4
  cores         = 2
  threads       = 2
  autostart     = false
  bootloader    = "UEFI"
  cpu_mode      = "CUSTOM"
  cpu_model     = "Haswell"
  machine_type  = "q35"
  time          = "UTC"
  desired_state = "STOPPED" # Keep VM stopped (default)
}

# Create a VM with memory ballooning
resource "truenas_vm" "ballooning" {
  name          = "flexible-vm"
  description   = "VM with memory ballooning"
  memory        = 8192 # 8GB max
  min_memory    = 2048 # 2GB min
  vcpus         = 2
  autostart     = true
  bootloader    = "UEFI"
  desired_state = "RUNNING"
}

# Create a suspended VM for quick resume
resource "truenas_vm" "suspended" {
  name          = "suspended-vm"
  description   = "VM in suspended state"
  memory        = 4096
  vcpus         = 2
  bootloader    = "UEFI"
  desired_state = "SUSPENDED" # Suspend the VM
}

# Legacy example using deprecated start_on_create (still supported)
resource "truenas_vm" "legacy" {
  name            = "legacy-vm"
  description     = "VM using deprecated start_on_create"
  memory          = 4096
  vcpus           = 2
  bootloader      = "UEFI"
  start_on_create = true # Deprecated: use desired_state = "RUNNING" instead
}

# Import an existing VM
# terraform import truenas_vm.existing 1

