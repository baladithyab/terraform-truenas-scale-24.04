# Minimal Talos VM Test

This test creates a minimal Talos VM with just the essentials to verify boot order and IP assignment.

## Configuration Based on Proxmox Setup

The Proxmox VM configuration uses:
- `boot_order = ["scsi0", "ide2"]` - Boot from disk first, then CDROM
- Machine type: `q35`
- SCSI controller: `virtio-scsi-pci`
- Network: `virtio`

This TrueNAS configuration mimics that setup:
- Disk device with `order = 1` (boot first)
- CDROM device with `order = 2` (boot second)
- Machine type: `Q35`
- Network: `VIRTIO`

## TrueNAS API Logging

### How to Enable Debug Logging

TrueNAS middleware supports debug logging via command-line arguments:

```bash
middlewared --debug-level TRACE --log-handler file
```

Available log levels (most to least verbose):
- `TRACE` - Most detailed
- `DEBUG`
- `INFO`
- `WARN`
- `ERROR`

### Log Files

- `/var/log/middlewared.log` - Main middleware log (only logs failures by default)
- `/var/log/fallback-middlewared.log` - Fallback when syslog-ng is unavailable

### Enabling in Production

To enable debug logging on a running TrueNAS system, you'd need to:
1. SSH into TrueNAS
2. Modify the middlewared service to start with `--debug-level TRACE`
3. Restart the middleware service
4. Monitor `/var/log/middlewared.log`

**Note:** This is not officially documented, so use with caution in production.

## Usage

1. Source environment variables:
   ```bash
   source .envrc
   # or
   direnv allow
   ```

2. Initialize Terraform:
   ```bash
   terraform init
   ```

3. Plan and apply:
   ```bash
   terraform plan
   terraform apply
   ```

4. Check VM status via TrueNAS UI or API

5. Look for IP assignment and verify Talos boots properly

## Expected Behavior

If boot order is working correctly:
1. VM should attempt to boot from disk first (order=1)
2. Since disk is empty, it should fall through to CDROM (order=2)
3. Talos ISO should boot
4. Talos should get an IP via DHCP
5. You should be able to see the IP in TrueNAS VM details or via network scan

## Troubleshooting

If the VM doesn't boot:
1. Check TrueNAS UI to see the actual device order
2. Compare with VMs created through the UI
3. Enable middleware debug logging to see API calls
4. Check if boot order is being respected or if there's a UI vs API difference
