# VM IP Address Discovery

This example demonstrates how to discover VM IP addresses using two methods:

1. **MAC Address Export** - Get MAC addresses from VMs for DHCP lookup
2. **Guest Agent Query** - Query QEMU guest agent for IP addresses (requires SSH)

## Overview

### Why Two Methods?

**TrueNAS API Limitation**: The TrueNAS REST API does not expose VM IP addresses or guest agent information.

**Solution**:
- **Method 1 (MAC Addresses)**: Works for ALL VMs, no guest agent required
- **Method 2 (Guest Agent)**: Works only for VMs with guest agent installed, requires SSH

## Method 1: MAC Address Export

### How It Works

1. Terraform reads VM configuration from TrueNAS API
2. Extracts MAC addresses from NIC devices
3. You use MAC addresses to look up DHCP leases

### Advantages

✅ Works for ALL VMs (including Talos, which doesn't support guest agent)  
✅ No SSH access required  
✅ No guest agent installation needed  
✅ Simple and reliable  

### Disadvantages

❌ Requires manual DHCP lookup  
❌ Only works if VM uses DHCP  
❌ Doesn't work for static IPs configured in guest OS  

### Example Usage

```hcl
resource "truenas_vm" "talos_worker" {
  name   = "talos-worker-01"
  memory = 4096
  vcpus  = 2
}

output "mac_addresses" {
  value = truenas_vm.talos_worker.mac_addresses
}
```

### DHCP Lookup Methods

#### Option A: pfSense/OPNsense
1. Navigate to **Status > DHCP Leases**
2. Search for the MAC address
3. Find the assigned IP

#### Option B: Linux DHCP Server
```bash
# Check DHCP leases file
cat /var/lib/dhcp/dhcpd.leases | grep -A 5 "00:a0:98:66:a6:bd"
```

#### Option C: Network Scanning
```bash
# Scan network and find MAC address
nmap -sn 10.0.0.0/24 | grep -B 2 "00:a0:98:66:a6:bd"

# Or use arp-scan
sudo arp-scan --localnet | grep "00:a0:98:66:a6:bd"
```

#### Option D: Router Web Interface
Most routers show DHCP client list with MAC addresses and IPs

---

## Method 2: Guest Agent Query (v0.2.15+)

**New in v0.2.15**: Password authentication support!

### How It Works

1. Terraform connects to TrueNAS host via SSH
2. Runs `virsh qemu-agent-command` to query guest agent
3. Parses JSON response to extract IP addresses

### Advantages

✅ Automatic IP discovery  
✅ Gets hostname, OS info, and other guest data  
✅ Works for static IPs configured in guest OS  
✅ Real-time information  

### Disadvantages

❌ Requires QEMU guest agent installed in VM
❌ Requires SSH access to TrueNAS host
✅ **Now works with Talos!** (v1.11.3+ has built-in guest agent)
❌ More complex setup

### Prerequisites

1. **QEMU Guest Agent Installed in VM**
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

   # Talos Linux
   # No installation needed - guest agent is built-in! ✅
   ```

2. **SSH Access to TrueNAS Host**

   **Option A: SSH Key (Recommended for production)**
   ```bash
   # Generate SSH key if you don't have one
   ssh-keygen -t rsa -b 4096 -f ~/.ssh/truenas_key

   # Copy public key to TrueNAS
   ssh-copy-id -i ~/.ssh/truenas_key.pub root@10.0.0.83

   # Test SSH access
   ssh -i ~/.ssh/truenas_key root@10.0.0.83 "virsh list --all"
   ```

   **Option B: Password (New in v0.2.15, easier for testing)**
   ```bash
   # Install sshpass
   sudo apt-get install sshpass  # Ubuntu/Debian
   brew install hudochenkov/sshpass/sshpass  # macOS

   # Test SSH access
   sshpass -p 'your-password' ssh -o StrictHostKeyChecking=no root@10.0.0.83 "virsh list --all"
   ```

### Example Usage

**With SSH Key:**
```hcl
data "truenas_vm_guest_info" "ubuntu" {
  vm_name      = "ubuntu-vm"
  truenas_host = "10.0.0.83"
  ssh_user     = "root"
  ssh_key_path = "~/.ssh/truenas_key"
}
```

**With Password (v0.2.15+):**
```hcl
data "truenas_vm_guest_info" "talos" {
  vm_name      = "talos-demo"
  truenas_host = "10.0.0.83"
  ssh_user     = "root"
  ssh_password = var.truenas_ssh_password  # Sensitive!
}

output "ubuntu_ips" {
  value = data.truenas_vm_guest_info.ubuntu.ip_addresses
}

output "ubuntu_hostname" {
  value = data.truenas_vm_guest_info.ubuntu.hostname
}
```

### Troubleshooting

#### Error: "Failed to query guest agent"

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

## Use Case: Talos Kubernetes Cluster

### Problem

You're deploying a Talos Kubernetes cluster and need to:
1. Know what IPs are already in use
2. Configure static IPs for Talos nodes
3. Avoid IP conflicts

### Solution

**Step 1**: Query existing VMs with guest agent to see what IPs are in use

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
```

**Step 2**: Define static IPs for Talos (avoiding existing IPs)

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
```

**Step 3**: Create Talos VMs and get their MAC addresses

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

## Recommendation

**For Talos**: Use **MAC Address Export** + DHCP lookup  
**For Ubuntu/Debian/etc**: Use **Guest Agent Query** for automatic discovery  
**For Mixed Environment**: Use both methods as shown in the examples

---

## Running the Example

```bash
# Set your API key
export TRUENAS_API_KEY="your-api-key-here"

# Initialize Terraform
terraform init

# Plan
terraform plan

# Apply
terraform apply

# View outputs
terraform output
```

## Additional Resources

- [TrueNAS VM Documentation](https://www.truenas.com/docs/scale/scaletutorials/virtualization/)
- [QEMU Guest Agent Documentation](https://wiki.qemu.org/Features/GuestAgent)
- [Talos Network Configuration](https://www.talos.dev/v1.5/reference/configuration/#machineconfig)

