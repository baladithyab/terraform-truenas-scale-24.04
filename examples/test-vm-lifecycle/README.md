# VM Lifecycle Test

This example tests the complete VM lifecycle with the TrueNAS provider:
- ✅ VM creation with devices (CDROM, NIC, DISK)
- ✅ VM start on creation
- ✅ Data source queries (by name, by ID, list all)
- ✅ Guest agent information
- ✅ VM destruction

## What This Tests

### 1. **VM Resource**
- Creates a VM with Talos 1.11.3 ISO
- Configures CPU, memory, bootloader
- Attaches CDROM, NIC, and DISK devices
- Starts VM automatically (`start_on_create = true`)

### 2. **Data Sources**
- `truenas_vm` by name
- `truenas_vm` by ID
- `truenas_vms` list all VMs
- `truenas_vm_guest_info` for guest agent data

### 3. **Cleanup**
- Tests proper VM destruction
- Verifies zvol cleanup

## Prerequisites

- TrueNAS Scale 24.04
- Talos 1.11.3 ISO at `/mnt/Loki/isos/talos-v1.11.3-metal-amd64.iso`
- Pool `Loki` with available space
- Bridge interface `br0` configured
- SSH access to TrueNAS host

## Usage

### 1. Set API Key
```bash
export TF_VAR_truenas_api_key="your-api-key"
```

### 2. Initialize
```bash
cd examples/test-vm-lifecycle
terraform init
```

### 3. Plan
```bash
terraform plan
```

### 4. Create VM
```bash
terraform apply
```

Expected output:
```
vm_id = "123"
vm_status = "RUNNING"
vm_by_name = {
  id     = "123"
  name   = "terraform-test-vm"
  status = "RUNNING"
  vcpus  = 2
  memory = 2048
}
all_vms_count = 8
```

### 5. Verify VM is Running
```bash
# Check TrueNAS UI or via API
curl -H "Authorization: Bearer $TF_VAR_truenas_api_key" \
  http://10.0.0.83:81/api/v2.0/vm | jq '.[] | select(.name == "terraform-test-vm")'
```

### 6. Wait for Guest Agent (Optional)
Talos boots quickly and starts the QEMU guest agent. Wait ~30 seconds, then refresh:
```bash
terraform refresh
terraform output guest_info
```

Expected output once guest agent is running:
```
guest_info = {
  hostname     = "talos-..."
  ip_addresses = ["10.0.0.x"]
  os_name      = "Talos"
  os_version   = "1.11.3"
}
```

### 7. Destroy VM
```bash
terraform destroy
```

This should:
- Stop the VM
- Delete all devices
- Delete the zvol
- Remove the VM

### 8. Verify Cleanup
```bash
# VM should be gone
curl -H "Authorization: Bearer $TF_VAR_truenas_api_key" \
  http://10.0.0.83:81/api/v2.0/vm | jq '.[] | select(.name == "terraform-test-vm")'

# Zvol should be gone
curl -H "Authorization: Bearer $TF_VAR_truenas_api_key" \
  http://10.0.0.83:81/api/v2.0/pool/dataset | \
  jq '.[] | select(.name == "Loki/terraform-test-vm-disk0")'
```

Both should return empty results.

## Troubleshooting

### VM Doesn't Start
Check TrueNAS logs:
```bash
ssh root@10.0.0.83 "tail -f /var/log/middlewared.log"
```

### Guest Agent Not Responding
Talos includes QEMU guest agent by default, but it may take 30-60 seconds to boot.

Check if VM is running:
```bash
ssh root@10.0.0.83 "virsh list --all"
```

Check guest agent status:
```bash
ssh root@10.0.0.83 "virsh qemu-agent-command terraform-test-vm '{\"execute\":\"guest-ping\"}'"
```

### Zvol Already Exists
If a previous test failed, manually delete:
```bash
curl -X DELETE \
  -H "Authorization: Bearer $TF_VAR_truenas_api_key" \
  http://10.0.0.83:81/api/v2.0/pool/dataset/id/Loki%2Fterraform-test-vm-disk0
```

## Success Criteria

✅ **VM Creation**
- VM appears in TrueNAS UI
- Status is "RUNNING"
- All devices attached (CDROM, NIC, DISK)

✅ **Data Sources**
- `truenas_vm` by name returns correct data
- `truenas_vm` by ID returns correct data
- `truenas_vms` includes the test VM
- `truenas_vm_guest_info` returns data (after boot)

✅ **VM Destruction**
- VM is stopped
- VM is deleted
- Zvol is deleted
- No errors in Terraform

## Next Steps

After successful testing:
1. Use this pattern for real VM deployments
2. Customize CPU, memory, disk sizes
3. Add multiple NICs or disks
4. Configure GPU passthrough (see `vm-gpu-passthrough` example)

