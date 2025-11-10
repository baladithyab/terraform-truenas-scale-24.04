# Release Notes - v0.2.12

## 🎉 VM IP Address Discovery Features

This release adds two powerful methods for discovering VM IP addresses, addressing a critical need for network configuration and Talos Kubernetes deployments.

---

## 🆕 New Features

### 1. **MAC Address Export from VMs** ✅

The `truenas_vm` resource now exports MAC addresses from all NIC devices.

**New Computed Attribute:**
- `mac_addresses` - List of MAC addresses for all NIC devices

**Use Cases:**
- Look up DHCP leases by MAC address
- Configure static IPs for Talos nodes
- Network inventory and documentation
- Avoid IP conflicts

**Example:**
```hcl
resource "truenas_vm" "talos_worker" {
  name   = "talos-worker-01"
  memory = 4096
  vcpus  = 2
}

output "mac_addresses" {
  value = truenas_vm.talos_worker.mac_addresses
}

# Output:
# mac_addresses = ["00:a0:98:66:a6:bd"]
```

**DHCP Lookup:**
```bash
# pfSense/OPNsense: Status > DHCP Leases
# Linux: cat /var/lib/dhcp/dhcpd.leases | grep "00:a0:98:66:a6:bd"
# nmap: nmap -sn 10.0.0.0/24 | grep -B 2 "00:a0:98:66:a6:bd"
```

---

### 2. **Guest Agent Data Source** ✅

New `truenas_vm_guest_info` data source queries QEMU guest agent for VM information.

**Attributes:**
- `ip_addresses` - List of IP addresses from guest agent
- `hostname` - Hostname reported by guest
- `os_name` - Operating system name
- `os_version` - Operating system version

**Requirements:**
- QEMU guest agent installed in VM
- SSH access to TrueNAS host
- SSH key-based authentication

**Example:**
```hcl
data "truenas_vm_guest_info" "ubuntu" {
  vm_name      = "ubuntu-vm"
  truenas_host = "10.0.0.83"
  ssh_user     = "root"
  ssh_key_path = "~/.ssh/id_rsa"
}

output "ubuntu_info" {
  value = {
    ips      = data.truenas_vm_guest_info.ubuntu.ip_addresses
    hostname = data.truenas_vm_guest_info.ubuntu.hostname
    os       = "${data.truenas_vm_guest_info.ubuntu.os_name} ${data.truenas_vm_guest_info.ubuntu.os_version}"
  }
}

# Output:
# ubuntu_info = {
#   ips      = ["10.0.0.50", "fe80::2a0:98ff:fe66:a6bd"]
#   hostname = "ubuntu-server"
#   os       = "Ubuntu 22.04.3 LTS"
# }
```

**Setup Guest Agent:**
```bash
# Ubuntu/Debian
sudo apt-get install qemu-guest-agent
sudo systemctl start qemu-guest-agent
sudo systemctl enable qemu-guest-agent

# RHEL/CentOS/Rocky
sudo yum install qemu-guest-agent
sudo systemctl start qemu-guest-agent
sudo systemctl enable qemu-guest-agent
```

**Setup SSH Access:**
```bash
# Generate SSH key
ssh-keygen -t rsa -b 4096 -f ~/.ssh/truenas_key

# Copy to TrueNAS
ssh-copy-id -i ~/.ssh/truenas_key.pub root@10.0.0.83

# Test
ssh -i ~/.ssh/truenas_key root@10.0.0.83 "virsh list --all"
```

---

## 🎯 Use Case: Talos Kubernetes Cluster

### Problem
When deploying Talos Kubernetes on TrueNAS VMs:
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
  value = local.existing_ips
}
```

**Step 2**: Create Talos VMs and get MAC addresses

```hcl
resource "truenas_vm" "talos_worker" {
  count  = 3
  name   = "talos-worker-${count.index + 1}"
  memory = 4096
  vcpus  = 2
}

output "talos_worker_macs" {
  value = [for vm in truenas_vm.talos_worker : vm.mac_addresses]
}
```

**Step 3**: Look up DHCP leases or configure static IPs

```yaml
# talos-worker-01.yaml
machine:
  network:
    interfaces:
      - interface: eth0
        addresses:
          - 10.0.0.111/24  # Choose IP not in existing_ips
        routes:
          - network: 0.0.0.0/0
            gateway: 10.0.0.1
    nameservers:
      - 10.0.0.1
```

---

## 📊 Comparison: MAC Address vs Guest Agent

| Feature | MAC Address Export | Guest Agent Query |
|---------|-------------------|-------------------|
| **Works for Talos** | ✅ Yes | ❌ No (no guest agent) |
| **Works for Ubuntu** | ✅ Yes | ✅ Yes |
| **Requires SSH** | ❌ No | ✅ Yes |
| **Requires Guest Agent** | ❌ No | ✅ Yes |
| **Automatic IP Discovery** | ❌ No (manual DHCP lookup) | ✅ Yes |
| **Gets Hostname** | ❌ No | ✅ Yes |
| **Gets OS Info** | ❌ No | ✅ Yes |
| **Setup Complexity** | Low | Medium |
| **Reliability** | High | Medium (depends on guest agent) |

**Recommendation:**
- **For Talos**: Use MAC Address Export + DHCP lookup
- **For Ubuntu/Debian/etc**: Use Guest Agent Query for automatic discovery
- **For Mixed Environment**: Use both methods

---

## 🔧 Technical Details

### Why Two Methods?

**TrueNAS API Limitation**: The TrueNAS REST API does not expose:
- VM IP addresses
- Guest agent information
- Network interface details beyond MAC addresses

**Our Solutions**:
1. **MAC Address Export**: Read from VM device configuration (available in API)
2. **Guest Agent Query**: SSH to TrueNAS host and run `virsh qemu-agent-command`

### Implementation Details

**MAC Address Export:**
- Reads `devices` array from VM API response
- Filters for `dtype == "NIC"`
- Extracts `mac` attribute from each NIC
- Returns as computed list attribute

**Guest Agent Query:**
- Connects to TrueNAS via SSH
- Runs `virsh qemu-agent-command <vm-name> '{"execute":"guest-network-get-interfaces"}'`
- Parses JSON response
- Filters out loopback and link-local addresses
- Also queries for hostname and OS info

---

## 📚 Examples

See the new `examples/vm-ip-discovery/` directory for:
- Complete Terraform configuration
- DHCP lookup methods
- Talos static IP configuration
- Troubleshooting guide

---

## 🐛 Bug Fixes

None in this release.

---

## ⚠️ Breaking Changes

None. This release is fully backward compatible.

---

## 📦 Installation

### GitHub Release

```bash
# Download from GitHub releases
wget https://github.com/baladithyab/terraform-provider-truenas/releases/download/v0.2.12/terraform-provider-truenas_v0.2.12_linux_amd64

# Install
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/baladithyab/truenas/0.2.12/linux_amd64/
mv terraform-provider-truenas_v0.2.12_linux_amd64 \
   ~/.terraform.d/plugins/registry.terraform.io/baladithyab/truenas/0.2.12/linux_amd64/terraform-provider-truenas_v0.2.12
chmod +x ~/.terraform.d/plugins/registry.terraform.io/baladithyab/truenas/0.2.12/linux_amd64/terraform-provider-truenas_v0.2.12
```

### Build from Source

```bash
git clone https://github.com/baladithyab/terraform-provider-truenas.git
cd terraform-provider-truenas
git checkout v0.2.12
make build
make install
```

---

## 🎯 What's Next?

Potential future enhancements:
- Display device configuration (SPICE settings)
- VM device management (add/remove disks, NICs)
- VM snapshot management
- VM cloning support

---

## 🙏 Acknowledgments

This feature was developed in response to user feedback about Talos Kubernetes deployments on TrueNAS.

---

## 📝 Changelog

See [CHANGELOG.md](CHANGELOG.md) for complete version history.

---

## 🐛 Report Issues

Found a bug? Have a feature request?
- **GitHub Issues**: https://github.com/baladithyab/terraform-provider-truenas/issues
- **Discussions**: https://github.com/baladithyab/terraform-provider-truenas/discussions

---

**Happy Terraforming! 🚀**

