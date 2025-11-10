# Known Limitations

This document describes the known limitations of the TrueNAS Scale Terraform Provider and provides workarounds where available.

## TrueNAS Version Compatibility

### ⚠️ TrueNAS Scale 24.04 Only

**Limitation**: This provider is **only compatible with TrueNAS Scale 24.04**.

**Reason**: 
- TrueNAS Scale 24.04 uses a REST API that this provider is built upon
- TrueNAS Scale 25.x switched to JSON-RPC over WebSocket, which is a completely different protocol
- The provider cannot be adapted to work with both versions simultaneously

**Impact**:
- ❌ Will not work with TrueNAS Scale 25.0 or later
- ❌ Will not work with TrueNAS Scale 23.x or earlier (untested, may work)
- ✅ Works perfectly with TrueNAS Scale 24.04 (24.04.0, 24.04.1, 24.04.2)

**Workaround**:
- **Stay on TrueNAS Scale 24.04** if you want to use this Terraform provider
- Do not upgrade to TrueNAS Scale 25.x if you rely on this provider
- Consider this when planning your infrastructure automation strategy

**Future Plans**:
- A new provider version for TrueNAS Scale 25.x would need to be created from scratch
- This would be a separate provider due to the fundamentally different API architecture

---

## VM IP Address Discovery

### ⚠️ No Native IP Discovery via REST API

**Limitation**: The TrueNAS Scale 24.04 REST API does not expose VM IP addresses or guest agent information.

**Reason**:
- The TrueNAS REST API focuses on VM configuration, not runtime guest information
- IP addresses are managed by the guest OS, not by the hypervisor
- Guest agent data is available via `virsh` command-line tool, but not through the REST API

**Impact**:
- ❌ Cannot query VM IP addresses directly from the API
- ❌ Cannot get hostname, OS version, or other guest information from the API
- ❌ Must use alternative methods to discover VM IPs

**Workarounds**:

#### Method 1: MAC Address Export (Recommended for ALL VMs)

The provider exports MAC addresses from VM network devices, which you can use to look up IPs in your DHCP server.

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

**Advantages**:
- ✅ Works for ALL VMs (including Talos, Alpine, etc.)
- ✅ No guest agent required
- ✅ No SSH access required
- ✅ Simple and reliable

**How to use**:
1. Get MAC address from Terraform output
2. Look up IP in your DHCP server (pfSense, OPNsense, router, etc.)
3. Use the IP in your configuration

**Limitation**: Only works if VM uses DHCP. If VM has a static IP configured in the guest OS, this won't help.

#### Method 2: Guest Agent Query (For VMs with Guest Agent)

The provider includes a data source that queries the QEMU guest agent via SSH to the TrueNAS host.

```hcl
data "truenas_vm_guest_info" "ubuntu" {
  vm_name      = "ubuntu-vm"
  truenas_host = "10.0.0.83"
  ssh_user     = "root"
  ssh_key_path = "~/.ssh/truenas_key"
  
  # Optional security settings
  ssh_strict_host_key_checking = true
  ssh_timeout_seconds          = 30
}

output "vm_info" {
  value = {
    ips      = data.truenas_vm_guest_info.ubuntu.ip_addresses
    hostname = data.truenas_vm_guest_info.ubuntu.hostname
    os       = "${data.truenas_vm_guest_info.ubuntu.os_name} ${data.truenas_vm_guest_info.ubuntu.os_version}"
  }
}
```

**Requirements**:
- QEMU guest agent installed in the VM
- SSH access to the TrueNAS host
- VM must be running

**Advantages**:
- ✅ Automatic IP discovery
- ✅ Gets hostname, OS info, and other guest data
- ✅ Works for static IPs configured in guest OS
- ✅ Real-time information

**Disadvantages**:
- ❌ Requires QEMU guest agent in VM
- ❌ Doesn't work for Talos (no guest agent support)
- ❌ Requires SSH access to TrueNAS
- ❌ More complex setup

**See**: [`VM_IP_DISCOVERY.md`](docs/guides/VM_IP_DISCOVERY.md) for complete guide with examples.

---

## Static IP Configuration in Guest OS

### ⚠️ Cannot Configure Guest OS Network Settings via API

**Limitation**: You cannot configure static IP addresses, DNS servers, or other network settings inside the guest OS through the TrueNAS API.

**Reason**:
- The TrueNAS API only manages the hypervisor layer (VM configuration, virtual devices)
- Guest OS configuration is outside the scope of the hypervisor API
- This is by design and consistent with other hypervisor APIs (VMware, Proxmox, etc.)

**Impact**:
- ❌ Cannot set static IP from Terraform
- ❌ Cannot configure DNS servers from Terraform
- ❌ Cannot configure network routes from Terraform
- ❌ Must configure network settings inside the guest OS manually or via other tools

**Workarounds**:

#### For Linux VMs: Cloud-Init

Use cloud-init with a seed ISO to configure network settings on first boot:

```hcl
resource "truenas_vm" "ubuntu" {
  name   = "ubuntu-vm"
  memory = 4096
  vcpus  = 2
  
  # Attach cloud-init ISO
  cdrom_devices = [{
    path = "/mnt/tank/cloud-init/ubuntu-vm-seed.iso"
  }]
}
```

Create the cloud-init ISO with network configuration:
```yaml
# network-config
version: 2
ethernets:
  eth0:
    addresses:
      - 10.0.0.100/24
    gateway4: 10.0.0.1
    nameservers:
      addresses: [8.8.8.8, 1.1.1.1]
```

**Advantages**:
- ✅ Automated configuration on first boot
- ✅ Standard Linux approach
- ✅ Works with all major distributions

**Disadvantages**:
- ❌ Requires cloud-init support in the OS
- ❌ Requires creating and managing seed ISOs
- ❌ Only runs on first boot (immutable afterward)

#### For Talos Linux: Machine Configuration

Use Talos machine configuration files to set static IPs:

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

Apply configuration:
```bash
talosctl apply-config --insecure \
  --nodes 10.0.0.111 \
  --file talos-worker-01.yaml
```

**Advantages**:
- ✅ Native Talos approach
- ✅ Can be reapplied anytime
- ✅ GitOps-friendly

**Disadvantages**:
- ❌ Talos-specific
- ❌ Requires manual application after VM creation

#### For Windows VMs: Sysprep Answer File

Use Windows Sysprep with an answer file to configure network settings:

```xml
<!-- unattend.xml -->
<unattend>
  <settings pass="specialize">
    <component name="Microsoft-Windows-TCPIP">
      <Interfaces>
        <Interface wcm:action="add">
          <Ipv4Settings>
            <DhcpEnabled>false</DhcpEnabled>
          </Ipv4Settings>
          <UnicastIpAddresses>
            <IpAddress wcm:action="add">10.0.0.100</IpAddress>
          </UnicastIpAddresses>
        </Interface>
      </Interfaces>
    </component>
  </settings>
</unattend>
```

**Advantages**:
- ✅ Native Windows approach
- ✅ Automated configuration

**Disadvantages**:
- ❌ Complex XML configuration
- ❌ Windows-specific

#### Manual Configuration

For simple setups, configure network settings manually after VM creation:

1. Create VM with Terraform
2. Boot VM
3. Log in and configure network manually
4. VM will keep configuration permanently

**Advantages**:
- ✅ Works for all OSes
- ✅ Simple and straightforward
- ✅ No additional tools needed

**Disadvantages**:
- ❌ Not automated
- ❌ Not reproducible
- ❌ Doesn't scale

---

## Other Limitations

### VM Lifecycle Management

**Limitation**: Some VM lifecycle operations are not available as direct API calls.

**Available**:
- ✅ Start VM via `desired_state = "RUNNING"`
- ✅ Stop VM via `desired_state = "STOPPED"`

**Not Available**:
- ❌ Restart VM
- ❌ Suspend VM
- ❌ Resume VM
- ❌ Clone VM

**Workaround**: Use the TrueNAS web UI or CLI for these operations.

### Import Limitations

**Limitation**: Some resources have import limitations:

- **VMs**: Can import VM configuration, but not devices (devices must be re-added to Terraform config)
- **Snapshots**: Must use special format `dataset@snapshotname` for import

**Workaround**: See [`IMPORT_GUIDE.md`](IMPORT_GUIDE.md) for detailed import instructions.

---

## Summary

| Limitation | Severity | Workaround Available | Workaround Complexity |
|------------|----------|---------------------|---------------------|
| TrueNAS 24.04 only | High | Stay on 24.04 | Easy |
| No native IP discovery | Medium | MAC export + DHCP lookup OR Guest agent query | Medium |
| No static IP via API | Medium | Cloud-init, Talos config, or manual | Medium-High |
| Limited VM lifecycle | Low | Use web UI or CLI | Easy |
| Import limitations | Low | Follow import guide | Easy |

---

## Getting Help

If you encounter issues related to these limitations:

1. Review this document and the linked guides
2. Check [`VM_IP_DISCOVERY.md`](docs/guides/VM_IP_DISCOVERY.md) for IP discovery help
3. Check [`IMPORT_GUIDE.md`](IMPORT_GUIDE.md) for import help
4. Open an issue on GitHub with details about your specific use case

---

## Contributing

If you find a workaround for any of these limitations that isn't documented here, please:

1. Test it thoroughly
2. Document it clearly
3. Submit a pull request to update this document

Your contributions help the entire community!