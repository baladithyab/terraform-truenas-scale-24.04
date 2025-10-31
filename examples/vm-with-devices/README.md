# VM with Devices Example

This example demonstrates how to create TrueNAS VMs with network interfaces, disks, and CDROM devices using the TrueNAS Terraform provider.

## Features

- **NIC Devices**: Configure network interfaces with VIRTIO or other types
- **Disk Devices**: Attach zvol or file-based disks to VMs
- **CDROM Devices**: Mount ISO files for installation or boot
- **MAC Address Export**: Automatically retrieve MAC addresses for network configuration

## Prerequisites

1. TrueNAS Scale 24.04 or later
2. Terraform 1.0 or later
3. TrueNAS API key with VM management permissions
4. Pre-created zvols for VM disks (or use existing ones)
5. ISO files uploaded to TrueNAS storage

## Creating Zvols for VM Disks

Before creating VMs, you need to create zvols for the disk devices:

```bash
# Create a 32GB zvol for test VM
zfs create -V 32G Loki/vms/test-vm-disk0

# Create a 64GB zvol for Talos worker
zfs create -V 64G Loki/vms/talos-worker-test-disk0
```

Or use the TrueNAS web UI:
1. Storage → Pools → Your Pool → Add Zvol
2. Name: `vms/test-vm-disk0`
3. Size: `32 GiB`
4. Click Save

## Usage

1. Copy the example tfvars file:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. Edit `terraform.tfvars` with your TrueNAS credentials:
   ```hcl
   truenas_endpoint = "http://10.0.0.83"
   truenas_api_key  = "your-api-key-here"
   ```

3. Initialize Terraform:
   ```bash
   terraform init
   ```

4. Review the plan:
   ```bash
   terraform plan
   ```

5. Apply the configuration:
   ```bash
   terraform apply
   ```

## Device Configuration

### NIC Devices

```hcl
nic_devices = [
  {
    type       = "VIRTIO"           # NIC type: VIRTIO (recommended), E1000, etc.
    nic_attach = "eno1"             # Physical interface to attach to
    mac        = ""                 # Optional: leave empty for auto-generation
    trust_guest_rx_filters = false  # Optional: default is false
  }
]
```

**Supported NIC Types:**
- `VIRTIO` (recommended for best performance)
- `E1000`
- `E1000E`
- `RTL8139`

### Disk Devices

```hcl
disk_devices = [
  {
    path   = "/dev/zvol/pool/vms/vm-disk0"  # Path to zvol or file
    type   = "VIRTIO"                        # Disk type: VIRTIO, AHCI, etc.
    iotype = "THREADS"                       # IO type: THREADS or NATIVE
    physical_sectorsize = null               # Optional: physical sector size
    logical_sectorsize  = null               # Optional: logical sector size
  }
]
```

**Supported Disk Types:**
- `VIRTIO` (recommended for best performance)
- `AHCI`
- `SCSI`

**IO Types:**
- `THREADS` (recommended)
- `NATIVE`

### CDROM Devices

```hcl
cdrom_devices = [
  {
    path = "/mnt/pool/isos/ubuntu.iso"  # Path to ISO file
  }
]
```

## Network Configuration

The provider automatically exports MAC addresses for all NICs:

```hcl
output "vm_mac_addresses" {
  value = truenas_vm.my_vm.mac_addresses
}
```

Use these MAC addresses to:
- Configure DHCP reservations
- Set up static IP mappings
- Configure network monitoring

## Example: Talos Linux Worker Node

The example includes a complete Talos Linux worker node configuration:

```hcl
resource "truenas_vm" "talos_worker" {
  name   = "talos-worker-test"
  vcpus  = 4
  memory = 8192  # 8GB

  nic_devices = [{
    type       = "VIRTIO"
    nic_attach = "eno1"
  }]

  disk_devices = [{
    path = "/dev/zvol/Loki/vms/talos-worker-test-disk0"
    type = "VIRTIO"
  }]

  cdrom_devices = [{
    path = "/mnt/Loki/isos/talos-v1.10.6-metal-amd64.iso"
  }]

  start_on_create = true
}
```

## Troubleshooting

### VM has no network connectivity

**Problem:** VM is running but cannot reach the network

**Solution:** Check that:
1. NIC device is configured with correct `nic_attach` value
2. Physical interface exists on TrueNAS host
3. Network bridge is properly configured
4. VM has started successfully

### Disk not found error

**Problem:** `Unable to create disk device: path not found`

**Solution:**
1. Verify zvol exists: `zfs list | grep your-zvol-name`
2. Check path format: `/dev/zvol/pool/dataset/zvol-name`
3. Ensure zvol is not already attached to another VM

### ISO file not found

**Problem:** `Unable to create CDROM device: path not found`

**Solution:**
1. Verify ISO file exists on TrueNAS
2. Check path format: `/mnt/pool/path/to/file.iso`
3. Ensure file has correct permissions

## Cleanup

To destroy the VMs and clean up:

```bash
terraform destroy
```

**Note:** This will delete the VMs but NOT the zvols. To delete zvols:

```bash
zfs destroy Loki/vms/test-vm-disk0
zfs destroy Loki/vms/talos-worker-test-disk0
```

## Next Steps

- Configure static IPs using MAC addresses
- Set up VM monitoring
- Configure backup schedules
- Deploy Kubernetes clusters with Talos

