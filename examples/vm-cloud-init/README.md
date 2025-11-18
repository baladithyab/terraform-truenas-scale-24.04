# VM Cloud-Init Example

This example demonstrates how to create TrueNAS Scale virtual machines with cloud-init support for automated initialization and network configuration.

## Overview

This example creates two Ubuntu 22.04 virtual machines:

1. **ubuntu-static-ip**: VM with static IP configuration
2. **ubuntu-dhcp**: VM with DHCP configuration

Both VMs are configured with:
- 2 vCPUs and 4GB RAM
- Ubuntu user with sudo access
- SSH key authentication
- QEMU guest agent
- Custom MOTD (Message of the Day)

## Files

- [`main.tf`](main.tf) - Terraform configuration for creating VMs with cloud-init
- [`variables.tf`](variables.tf) - Variable definitions
- [`terraform.tfvars.example`](terraform.tfvars.example) - Example variable values
- [`user-data`](user-data) - Cloud-init user-data for static IP VM
- [`meta-data`](meta-data) - Cloud-init meta-data for static IP VM
- [`network-config`](network-config) - Network configuration for static IP VM
- [`user-data-dhcp`](user-data-dhcp) - Cloud-init user-data for DHCP VM
- [`meta-data-dhcp`](meta-data-dhcp) - Cloud-init meta-data for DHCP VM

## Prerequisites

1. TrueNAS Scale 24.04 or later
2. Terraform 1.0 or later
3. TrueNAS API key with appropriate permissions
4. Ubuntu 22.04 cloud image available in TrueNAS

## Usage

1. **Copy the example variables**:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. **Edit the variables**:
   ```bash
   nano terraform.tfvars
   ```
   Update the following values:
   - `truenas_url` - Your TrueNAS server URL
   - `truenas_api_key` - Your TrueNAS API key
   - `pool_name` - Your storage pool name
   - `static_ip` - Desired static IP address
   - `static_gateway` - Your network gateway
   - `dns_servers` - DNS servers
   - `ssh_public_key` - Your SSH public key

3. **Initialize Terraform**:
   ```bash
   terraform init
   ```

4. **Plan the deployment**:
   ```bash
   terraform plan
   ```

5. **Apply the configuration**:
   ```bash
   terraform apply
   ```

## Cloud-Init Configuration

### Static IP Configuration

The static IP VM uses the following cloud-init configuration:

- **Network**: Static IP configuration with custom gateway and DNS
- **User**: Ubuntu user with sudo access and SSH key
- **Packages**: QEMU guest agent, network tools, curl, wget
- **Files**: Custom MOTD and netplan configuration
- **Commands**: Enable QEMU guest agent

### DHCP Configuration

The DHCP VM uses a simpler configuration:

- **Network**: DHCP configuration (automatic IP assignment)
- **User**: Ubuntu user with sudo access and SSH key
- **Packages**: QEMU guest agent, network tools, curl, wget
- **Files**: Custom MOTD
- **Commands**: Enable QEMU guest agent

## Customization

### Network Configuration

To customize the network configuration, edit the following files:

- For static IP: [`network-config`](network-config)
- For DHCP: No network-config file needed (uses default DHCP)

### User Configuration

To customize user settings, edit the `users` section in:
- [`user-data`](user-data) for static IP VM
- [`user-data-dhcp`](user-data-dhcp) for DHCP VM

### Package Installation

To add or remove packages, edit the `packages` section in the user-data files.

### Custom Scripts

To add custom initialization scripts, modify the `runcmd` section in the user-data files.

## Verification

After deployment, you can verify the VMs:

1. **Check VM status in TrueNAS web interface**
2. **Connect via SSH**:
   ```bash
   ssh ubuntu@192.168.1.100  # For static IP VM
   ```
3. **Check cloud-init logs**:
   ```bash
   sudo cat /var/log/cloud-init.log
   ```
4. **Verify network configuration**:
   ```bash
   ip addr show eth0
   ```

## Troubleshooting

### Common Issues

1. **Cloud-init not running**: Ensure the VM has the cloud-init package installed
2. **Network not configured**: Check the network-config file syntax
3. **SSH access denied**: Verify the SSH public key is correctly configured

### Debugging

To debug cloud-init issues:

1. Check cloud-init logs: `/var/log/cloud-init.log`
2. Check cloud-init output: `/var/log/cloud-init-output.log`
3. Verify ISO generation in TrueNAS: Check the ISO directory
4. Check VM console output for boot messages

## Cleanup

To destroy the created resources:

```bash
terraform destroy
```

## Additional Resources

- [Cloud-Init Documentation](https://cloudinit.readthedocs.io/)
- [Netplan Configuration](https://netplan.io/)
- [TrueNAS Provider Documentation](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs)
- [VM Resource Documentation](../../docs/resources/vm.md)