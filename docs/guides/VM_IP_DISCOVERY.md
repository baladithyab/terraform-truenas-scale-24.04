# VM IP Address Discovery Guide

## Overview

This guide explains how to discover VM IP addresses using the TrueNAS Terraform Provider.

**Problem**: TrueNAS API does not expose VM IP addresses or guest agent information.

**Solution**: Two complementary methods:
1. **MAC Address Export** - Works for ALL VMs (including Talos)
2. **Guest Agent Query** - Works for VMs with guest agent installed (Ubuntu, Debian, etc.)
3. **Cloud-Init Static IP** - Proactive configuration (Best for automation)

---

## Method 1: MAC Address Export

### Quick Start

```hcl
resource "truenas_vm" "example" {
  name   = "my-vm"
  memory = 4096
  vcpus  = 2
}

output "mac_addresses" {
  value = truenas_vm.example.mac_addresses
}
```

### DHCP Lookup

Once you have the MAC address, look up the IP in your DHCP server:

**pfSense/OPNsense:**
```
Status > DHCP Leases > Search for MAC address
```

**Linux DHCP Server:**
```bash
cat /var/lib/dhcp/dhcpd.leases | grep -A 5 "00:a0:98:66:a6:bd"
```

**Network Scanning:**
```bash
nmap -sn 10.0.0.0/24 | grep -B 2 "00:a0:98:66:a6:bd"
sudo arp-scan --localnet | grep "00:a0:98:66:a6:bd"
```

### Pros & Cons

✅ **Pros:**
- Works for ALL VMs (including Talos)
- No guest agent required
- No SSH access required
- Simple and reliable

❌ **Cons:**
- Requires manual DHCP lookup
- Only works if VM uses DHCP
- Doesn't work for static IPs configured in guest OS

---

## Method 2: Guest Agent Query

### Important: Authentication Validation

**New in this release**: The provider now validates SSH authentication **before** attempting to query the guest agent. This means:

- ✅ Clear error messages if authentication fails
- ✅ Faster feedback on configuration issues
- ✅ No wasted time attempting guest agent queries with wrong credentials
- ✅ Better security through early validation

If authentication fails, you'll see an error like:
```
Error: Failed to authenticate with TrueNAS host: ssh: handshake failed: ssh: unable to authenticate
```

This helps you fix authentication issues before attempting costly guest agent operations.

### Prerequisites

1. **Install QEMU Guest Agent in VM:**

```bash
# Ubuntu/Debian
sudo apt-get install qemu-guest-agent
sudo systemctl start qemu-guest-agent
sudo systemctl enable qemu-guest-agent

# RHEL/CentOS/Rocky
sudo yum install qemu-guest-agent
sudo systemctl start qemu-guest-agent
sudo systemctl enable qemu-guest-agent

# Arch Linux
sudo pacman -S qemu-guest-agent
sudo systemctl start qemu-guest-agent
sudo systemctl enable qemu-guest-agent
```

2. **Setup SSH Access to TrueNAS:**

```bash
# Generate SSH key
ssh-keygen -t rsa -b 4096 -f ~/.ssh/truenas_key

# Copy to TrueNAS
ssh-copy-id -i ~/.ssh/truenas_key.pub root@10.0.0.83

# Test
ssh -i ~/.ssh/truenas_key root@10.0.0.83 "virsh list --all"
```

### Quick Start

```hcl
data "truenas_vm_guest_info" "ubuntu" {
  vm_name      = "ubuntu-vm"
  truenas_host = "10.0.0.83"
  ssh_user     = "root"
  ssh_key_path = "~/.ssh/truenas_key"
  
  # Optional: Enhanced security and timeout options
  ssh_strict_host_key_checking = true   # Enable strict host key checking
  ssh_timeout_seconds          = 30     # Set SSH timeout (default: 10)
}

output "ubuntu_info" {
  value = {
    ips      = data.truenas_vm_guest_info.ubuntu.ip_addresses
    hostname = data.truenas_vm_guest_info.ubuntu.hostname
    os       = "${data.truenas_vm_guest_info.ubuntu.os_name} ${data.truenas_vm_guest_info.ubuntu.os_version}"
  }
}
```

### Configuration Options

The `truenas_vm_guest_info` data source supports the following authentication and security options:

**Required:**
- `vm_name` - Name of the VM to query
- `truenas_host` - TrueNAS hostname or IP address
- `ssh_user` - SSH user (typically "root")

**Authentication (choose one):**
- `ssh_key_path` - Path to SSH private key file
- `ssh_password` - SSH password (sensitive)

**Optional Security & Timeout:**
- `ssh_strict_host_key_checking` - Enable strict SSH host key checking (default: `false`)
  - When `true`: SSH will fail if host key is not in known_hosts
  - When `false`: SSH will accept any host key (less secure, but easier for automation)
  - **Use `true` in production** for better security
- `ssh_timeout_seconds` - SSH connection timeout in seconds (default: `10`)
  - Increase if TrueNAS is slow to respond
  - Useful for high-latency networks or busy systems

### Pros & Cons

✅ **Pros:**
- Automatic IP discovery
- Gets hostname, OS info, and other guest data
- Works for static IPs configured in guest OS
- Real-time information

❌ **Cons:**
- Requires QEMU guest agent installed in VM
- Requires SSH access to TrueNAS host
- Doesn't work for Talos (no guest agent support)
- More complex setup

---

## Method 3: Cloud-Init Static IP

### Overview

Instead of discovering the IP after the VM boots, you can use Cloud-Init to assign a static IP address during provisioning. This eliminates the need for discovery entirely.

### Quick Start

```hcl
resource "truenas_vm" "static_vm" {
  name   = "ubuntu-static"
  vcpus  = 2
  memory = 4096
  
  cloud_init {
    # Network configuration for static IP
    network_config = <<EOF
version: 2
ethernets:
  eth0:
    dhcp4: no
    addresses: [192.168.1.150/24]
    gateway4: 192.168.1.1
    nameservers:
      addresses: [8.8.8.8, 1.1.1.1]
EOF

    # User configuration
    user_data = <<EOF
#cloud-config
hostname: ubuntu-static
users:
  - name: ubuntu
    sudo: ALL=(ALL) NOPASSWD:ALL
    ssh_authorized_keys:
      - ssh-rsa AAAAB3...
EOF
  }
}

output "vm_ip" {
  value = "192.168.1.150"
}
```

### Pros & Cons

✅ **Pros:**
- No discovery required - you know the IP in advance
- Immediate access after boot
- Perfect for automation and CI/CD pipelines
- Works with any Cloud-Init enabled image

❌ **Cons:**
- Requires Cloud-Init enabled OS image
- Requires managing IP allocation (avoiding conflicts)
- Requires `network_config` support in the OS (Netplan/NetworkManager)

---

## Use Case: Talos Kubernetes Cluster

### Problem

When deploying Talos Kubernetes on TrueNAS:
1. Need to know what IPs are already in use
2. Need to configure static IPs for Talos nodes
3. Need to avoid IP conflicts
4. Talos doesn't support QEMU guest agent

### Solution

**Step 1**: Query existing VMs to see what IPs are in use

```hcl
data "truenas_vm_guest_info" "ubuntu" {
  vm_name      = "ubuntu-vm"
  truenas_host = "10.0.0.83"
  ssh_user     = "root"
  ssh_key_path = "~/.ssh/truenas_key"
}

data "truenas_vm_guest_info" "plex" {
  vm_name      = "plex-server"
  truenas_host = "10.0.0.83"
  ssh_user     = "root"
  ssh_key_path = "~/.ssh/truenas_key"
}

locals {
  existing_ips = concat(
    data.truenas_vm_guest_info.ubuntu.ip_addresses,
    data.truenas_vm_guest_info.plex.ip_addresses,
  )
}

output "existing_ips" {
  description = "IPs currently in use - avoid these for Talos"
  value       = local.existing_ips
}
```

**Step 2**: Create Talos VMs and get MAC addresses

```hcl
resource "truenas_vm" "talos_worker" {
  count  = 3
  name   = "talos-worker-${count.index + 1}"
  memory = 4096
  vcpus  = 2
  cores  = 2
  threads = 1
  autostart = true
  start_on_create = true
}

output "talos_worker_macs" {
  description = "MAC addresses for DHCP lookup"
  value = [for vm in truenas_vm.talos_worker : vm.mac_addresses]
}
```

**Step 3**: Define static IPs for Talos (avoiding existing IPs)

```hcl
locals {
  talos_control_plane_ips = [
    "10.0.0.101",  # Make sure these don't conflict
    "10.0.0.102",  # with existing_ips
    "10.0.0.103",
  ]

  talos_worker_ips = [
    "10.0.0.111",
    "10.0.0.112",
    "10.0.0.113",
  ]
}

output "talos_ip_plan" {
  value = {
    existing_ips    = local.existing_ips
    control_plane   = local.talos_control_plane_ips
    workers         = local.talos_worker_ips
  }
}
```

**Step 4**: Configure Talos with static IPs

```yaml
# talos-worker-01.yaml
machine:
  network:
    interfaces:
      - interface: eth0
        addresses:
          - 10.0.0.111/24
        routes:
          - network: 0.0.0.0/0
            gateway: 10.0.0.1
    nameservers:
      - 10.0.0.1
      - 8.8.8.8
```

---

## Comparison Table

| Feature | MAC Address Export | Guest Agent Query | Cloud-Init Static IP |
|---------|-------------------|-------------------|----------------------|
| **Works for Talos** | ✅ Yes | ❌ No | ❌ No (Uses machine config) |
| **Works for Ubuntu** | ✅ Yes | ✅ Yes | ✅ Yes |
| **Requires SSH** | ❌ No | ✅ Yes | ❌ No |
| **Requires Guest Agent** | ❌ No | ✅ Yes | ❌ No |
| **Automatic IP Discovery** | ❌ No | ✅ Yes | ✅ (Pre-defined) |
| **Gets Hostname** | ❌ No | ✅ Yes | ✅ (Pre-defined) |
| **Gets OS Info** | ❌ No | ✅ Yes | ❌ No |
| **Setup Complexity** | Low | Medium | Medium |
| **Reliability** | High | Medium | High |

---

## Troubleshooting

### MAC Address is Null

**Problem**: `mac_addresses = []` or contains `null` values

**Cause**: TrueNAS auto-generates MAC addresses, and they may not be set in the API response

**Solution**: MAC addresses are generated when VM starts. Import the VM after it has been started at least once.

### Guest Agent Query Fails

**Problem**: "Failed to query guest agent"

**Possible Causes:**
1. Guest agent not installed in VM
2. Guest agent not running
3. SSH access not configured
4. VM name doesn't match exactly

**Solutions:**
```bash
# Check if guest agent is running in VM
sudo systemctl status qemu-guest-agent

# Test SSH access
ssh -i ~/.ssh/truenas_key root@10.0.0.83

# Test virsh command manually
ssh -i ~/.ssh/truenas_key root@10.0.0.83 \
  "virsh qemu-agent-command ubuntu-vm '{\"execute\":\"guest-network-get-interfaces\"}'"

# List all VMs to verify name
ssh -i ~/.ssh/truenas_key root@10.0.0.83 "virsh list --all"
```

### SSH Authentication Errors

**Problem**: "Failed to authenticate with TrueNAS host"

**Error Message**: Provider now validates authentication **before attempting to query** the guest agent

**Possible Causes:**
1. Wrong SSH key path
2. Wrong SSH password
3. Wrong username
4. SSH service not running on TrueNAS
5. Firewall blocking SSH access

**Solutions:**
```bash
# Test SSH key authentication
ssh -i ~/.ssh/truenas_key root@10.0.0.83 "echo 'Authentication successful'"

# Test SSH password authentication
sshpass -p 'your-password' ssh root@10.0.0.83 "echo 'Authentication successful'"

# Check SSH service on TrueNAS
ssh root@10.0.0.83 "systemctl status ssh"

# Test from Terraform
terraform plan -target=data.truenas_vm_guest_info.example
```

### Host Key Verification Failed

**Problem**: "Host key verification failed"

**Cause**: TrueNAS host key is not in known_hosts file, and `ssh_strict_host_key_checking = true`

**Solutions:**

**Option 1: Add host key to known_hosts (recommended for production)**
```bash
# Add TrueNAS host key to known_hosts
ssh-keyscan -H 10.0.0.83 >> ~/.ssh/known_hosts

# Or connect once manually to accept
ssh root@10.0.0.83
```

**Option 2: Disable strict host key checking (for development only)**
```hcl
data "truenas_vm_guest_info" "example" {
  vm_name                      = "my-vm"
  truenas_host                 = "10.0.0.83"
  ssh_user                     = "root"
  ssh_key_path                 = "~/.ssh/truenas_key"
  ssh_strict_host_key_checking = false  # Accept any host key
}
```

### SSH Timeout

**Problem**: "SSH connection timed out"

**Cause**: TrueNAS is slow to respond, or network latency is high

**Solution**: Increase the SSH timeout
```hcl
data "truenas_vm_guest_info" "example" {
  vm_name             = "my-vm"
  truenas_host        = "10.0.0.83"
  ssh_user            = "root"
  ssh_key_path        = "~/.ssh/truenas_key"
  ssh_timeout_seconds = 30  # Increase from default 10 seconds
}
```

### Common Error Messages

| Error Message | Meaning | Solution |
|---------------|---------|----------|
| "Failed to authenticate with TrueNAS host" | SSH authentication failed before querying guest agent | Check SSH credentials and connectivity |
| "Failed to execute SSH command" | SSH connected but command execution failed | Check SSH user permissions and virsh access |
| "Failed to query guest agent" | Guest agent not responding or VM not found | Check guest agent installation and VM name |
| "Failed to parse guest agent response" | Guest agent returned invalid JSON | Check guest agent version compatibility |
| "No IP addresses found in guest agent response" | VM doesn't have network interfaces or IPs assigned | Check VM network configuration |
| "Host key verification failed" | SSH host key not trusted | Add host key to known_hosts or disable strict checking |
| "Connection timed out" | SSH connection took too long | Increase `ssh_timeout_seconds` |

---

## Examples

See `examples/vm-ip-discovery/` for complete examples including:
- MAC address export
- Guest agent queries
- Talos static IP configuration
- DHCP lookup scripts

---

## Recommendation

**For Talos**: Use **MAC Address Export** + DHCP lookup
**For Ubuntu/Debian/etc**: Use **Cloud-Init Static IP** for deterministic deployments, or **Guest Agent Query** for DHCP environments
**For Mixed Environment**: Use the method that best fits your network architecture

---

## Additional Resources

- [TrueNAS VM Documentation](https://www.truenas.com/docs/scale/scaletutorials/virtualization/)
- [QEMU Guest Agent Documentation](https://wiki.qemu.org/Features/GuestAgent)
- [Talos Network Configuration](https://www.talos.dev/v1.5/reference/configuration/#machineconfig)
- [Provider Examples](examples/vm-ip-discovery/)

