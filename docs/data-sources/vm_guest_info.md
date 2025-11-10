---
page_title: "truenas_vm_guest_info Data Source - terraform-provider-truenas"
subcategory: "Virtual Machines"
description: |-
  Retrieves guest information from a TrueNAS Scale virtual machine, including IP addresses.
---

# truenas_vm_guest_info (Data Source)

Retrieves guest information from a running TrueNAS Scale virtual machine. This data source is particularly useful for obtaining the VM's IP address after provisioning, enabling dynamic infrastructure configuration.

**Note:** This data source requires the QEMU Guest Agent to be installed and running inside the VM.

## Example Usage

### Get VM IP Address

```terraform
resource "truenas_vm" "server" {
  name   = "app-server"
  vcpus  = 2
  memory = 4096
  
  autostart = true
}

# Wait for VM to boot and guest agent to start
data "truenas_vm_guest_info" "server" {
  vm_id = truenas_vm.server.id
  
  depends_on = [truenas_vm.server]
}

output "server_ip" {
  value = data.truenas_vm_guest_info.server.ipv4_addresses[0]
}
```

### Use IP in Another Resource

```terraform
resource "truenas_vm" "database" {
  name   = "postgres-db"
  vcpus  = 4
  memory = 8192
}

data "truenas_vm_guest_info" "database" {
  vm_id = truenas_vm.database.id
}

# Configure application with database IP
resource "null_resource" "configure_app" {
  provisioner "local-exec" {
    command = "echo 'DB_HOST=${data.truenas_vm_guest_info.database.ipv4_addresses[0]}' > .env"
  }
}
```

### Multiple Network Interfaces

```terraform
data "truenas_vm_guest_info" "multi_nic" {
  vm_id = truenas_vm.multi_nic_server.id
}

output "all_interfaces" {
  value = data.truenas_vm_guest_info.multi_nic.interfaces
}

output "primary_ip" {
  value = data.truenas_vm_guest_info.multi_nic.ipv4_addresses[0]
}
```

### Complete Example with Polling

```terraform
resource "truenas_vm" "web_server" {
  name   = "web-01"
  vcpus  = 2
  memory = 4096
  
  autostart = true
}

# Poll until guest agent is available
data "truenas_vm_guest_info" "web_server" {
  vm_id = truenas_vm.web_server.id
  
  # Terraform will retry if guest agent not ready
  depends_on = [truenas_vm.web_server]
  
  lifecycle {
    postcondition {
      condition     = length(self.ipv4_addresses) > 0
      error_message = "VM does not have an IPv4 address yet."
    }
  }
}

output "web_server_url" {
  value = "http://${data.truenas_vm_guest_info.web_server.ipv4_addresses[0]}"
}
```

## Schema

### Required

- `vm_id` (String) The ID or name of the virtual machine.

### Read-Only

- `id` (String) The ID of the data source (same as vm_id).
- `hostname` (String) The hostname reported by the guest.
- `ipv4_addresses` (List of String) List of IPv4 addresses configured on the guest.
- `ipv6_addresses` (List of String) List of IPv6 addresses configured on the guest.
- `interfaces` (List of Object) Detailed information about network interfaces. Each interface contains:
  - `name` (String) Interface name (e.g., `eth0`, `ens18`)
  - `ip_addresses` (List of String) IP addresses on this interface
  - `mac_address` (String) MAC address
  - `stats` (Object) Network statistics (rx_bytes, tx_bytes, etc.)

## Notes

### Requirements

**Guest Agent Installation:**

The QEMU Guest Agent must be installed and running inside the VM:

- **Ubuntu/Debian:**
  ```bash
  sudo apt-get install qemu-guest-agent
  sudo systemctl enable qemu-guest-agent
  sudo systemctl start qemu-guest-agent
  ```

- **CentOS/RHEL/Rocky:**
  ```bash
  sudo yum install qemu-guest-agent
  sudo systemctl enable qemu-guest-agent
  sudo systemctl start qemu-guest-agent
  ```

- **Windows:**
  - Download virtio-win drivers from Fedora project
  - Install QEMU Guest Agent from the virtio-win ISO

### Timing Considerations

- The data source will fail if the VM is not running
- Guest agent may take 30-60 seconds to start after VM boot
- Use `depends_on` to ensure VM is created first
- Consider adding `time_sleep` resource for newly provisioned VMs

### IP Address Selection

IP addresses are returned as a list. Typically:
- `ipv4_addresses[0]` is the primary IPv4 address
- List may include loopback (127.0.0.1) and link-local addresses
- Filter as needed for your use case

### Example with Delay

```terraform
resource "truenas_vm" "app" {
  name   = "application"
  vcpus  = 2
  memory = 4096
}

# Wait for VM to fully boot
resource "time_sleep" "wait_for_vm" {
  create_duration = "60s"
  
  depends_on = [truenas_vm.app]
}

data "truenas_vm_guest_info" "app" {
  vm_id = truenas_vm.app.id
  
  depends_on = [time_sleep.wait_for_vm]
}
```

### Filtering IP Addresses

```terraform
locals {
  # Filter out loopback and link-local addresses
  public_ips = [
    for ip in data.truenas_vm_guest_info.server.ipv4_addresses :
    ip if !startswith(ip, "127.") && !startswith(ip, "169.254.")
  ]
}

output "public_ip" {
  value = local.public_ips[0]
}
```

### Error Handling

If the data source fails:

1. **Verify guest agent is installed:**
   ```bash
   # Inside the VM
   systemctl status qemu-guest-agent
   ```

2. **Check VM is running:**
   ```bash
   vm status <vm-name>
   ```

3. **View guest agent info manually:**
   ```bash
   vm guest_info <vm-name>
   ```

4. **Wait longer for boot:**
   - Increase `time_sleep` duration
   - Check VM console for boot issues

### Best Practices

- Always include `depends_on` to ensure VM exists
- Use `lifecycle` postconditions to validate IP presence
- Handle multiple IPs appropriately for your use case
- Consider network configuration (DHCP vs static)
- Test guest agent functionality before production use

### Dynamic Infrastructure Patterns

#### Register VM in DNS

```terraform
data "truenas_vm_guest_info" "app" {
  vm_id = truenas_vm.app.id
}

resource "dns_a_record" "app" {
  zone  = "example.com"
  name  = "app"
  value = data.truenas_vm_guest_info.app.ipv4_addresses[0]
}
```

#### Configure Load Balancer

```terraform
data "truenas_vm_guest_info" "backend" {
  count  = length(truenas_vm.backends)
  vm_id  = truenas_vm.backends[count.index].id
}

resource "loadbalancer_pool" "backend" {
  name = "backend-pool"
  
  members = [
    for info in data.truenas_vm_guest_info.backend :
    {
      ip   = info.ipv4_addresses[0]
      port = 8080
    }
  ]
}
```

#### Ansible Inventory

```terraform
output "ansible_inventory" {
  value = yamlencode({
    all = {
      hosts = {
        for vm in data.truenas_vm_guest_info.servers :
        vm.hostname => {
          ansible_host = vm.ipv4_addresses[0]
        }
      }
    }
  })
}
```

## See Also

- [truenas_vm Resource](../resources/vm) - Create and manage VMs
- [truenas_vm Data Source](vm) - Query VM configuration
- [VM IP Discovery Guide](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/guides/vm_ip_discovery) - Detailed IP discovery guide