# VM IP Address Discovery Guide

## Overview

This guide explains how to discover VM IP addresses using the TrueNAS Terraform Provider.

**Problem**: TrueNAS API does not expose VM IP addresses or guest agent information.

**Solution**: Two complementary methods:
1. **MAC Address Export** - Works for ALL VMs (including Talos)
2. **Guest Agent Query** - Works for VMs with guest agent installed (Ubuntu, Debian, etc.)

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
}

output "ubuntu_info" {
  value = {
    ips      = data.truenas_vm_guest_info.ubuntu.ip_addresses
    hostname = data.truenas_vm_guest_info.ubuntu.hostname
    os       = "${data.truenas_vm_guest_info.ubuntu.os_name} ${data.truenas_vm_guest_info.ubuntu.os_version}"
  }
}
```

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

| Feature | MAC Address Export | Guest Agent Query |
|---------|-------------------|-------------------|
| **Works for Talos** | ✅ Yes | ❌ No |
| **Works for Ubuntu** | ✅ Yes | ✅ Yes |
| **Requires SSH** | ❌ No | ✅ Yes |
| **Requires Guest Agent** | ❌ No | ✅ Yes |
| **Automatic IP Discovery** | ❌ No | ✅ Yes |
| **Gets Hostname** | ❌ No | ✅ Yes |
| **Gets OS Info** | ❌ No | ✅ Yes |
| **Setup Complexity** | Low | Medium |
| **Reliability** | High | Medium |

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
**For Ubuntu/Debian/etc**: Use **Guest Agent Query** for automatic discovery  
**For Mixed Environment**: Use both methods as shown in the examples

---

## Additional Resources

- [TrueNAS VM Documentation](https://www.truenas.com/docs/scale/scaletutorials/virtualization/)
- [QEMU Guest Agent Documentation](https://wiki.qemu.org/Features/GuestAgent)
- [Talos Network Configuration](https://www.talos.dev/v1.5/reference/configuration/#machineconfig)
- [Provider Examples](examples/vm-ip-discovery/)

