# Example: NIC Device
resource "truenas_vm_device" "nic" {
  vm_id       = truenas_vm.myvm.id
  device_type = "NIC"
  order       = 1000

  nic_config {
    type                   = "VIRTIO"
    mac                    = "00:a0:98:66:a6:bd"
    nic_attach             = "br0"
    trust_guest_rx_filters = false
  }
}

# Example: Disk Device
resource "truenas_vm_device" "disk" {
  vm_id       = truenas_vm.myvm.id
  device_type = "DISK"
  order       = 2000

  disk_config {
    path                = "/dev/zvol/pool/vm-disk"
    type                = "VIRTIO"
    iotype              = "THREADS"
    physical_sectorsize = 512
    logical_sectorsize  = 512
  }
}

# Example: CDROM Device
resource "truenas_vm_device" "cdrom" {
  vm_id       = truenas_vm.myvm.id
  device_type = "CDROM"
  order       = 3000

  cdrom_config {
    path = "/mnt/pool/isos/ubuntu-22.04.iso"
  }
}

# Example: PCI Passthrough Device
resource "truenas_vm_device" "gpu" {
  vm_id       = truenas_vm.myvm.id
  device_type = "PCI"
  order       = 4000

  pci_config {
    pptdev = "pci_0000_3b_00_0"
  }
}

# Example: Display Device (SPICE)
resource "truenas_vm_device" "display" {
  vm_id       = truenas_vm.myvm.id
  device_type = "DISPLAY"
  order       = 5000

  display_config {
    type       = "SPICE"
    port       = 5900
    bind       = "0.0.0.0"
    web        = true
    resolution = "1920x1080"
  }
}

# Import existing device
# terraform import truenas_vm_device.nic <device_id>
