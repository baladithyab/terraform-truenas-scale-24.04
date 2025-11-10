---
page_title: "truenas_vm Data Source - terraform-provider-truenas"
subcategory: "Virtual Machines"
description: |-
  Fetches information about a specific VM on the TrueNAS system.
---

# truenas_vm (Data Source)

Fetches information about a specific virtual machine on the TrueNAS system. This data source can query VM information by either ID or name and returns detailed configuration and status data.

## Example Usage

### Query VM by Name

```terraform
data "truenas_vm" "web_server" {
  name = "web-server"
}

output "vm_info" {
  value = {
    id          = data.truenas_vm.web_server.id
    name        = data.truenas_vm.web_server.name
    description = data.truenas_vm.web_server.description
    vcpus       = data.truenas_vm.web_server.vcpus
    memory_gb   = data.truenas_vm.web_server.memory / 1024
    status      = data.truenas_vm.web_server.status
  }
}
```

### Query VM by ID

```terraform
data "truenas_vm" "by_id" {
  id = "1"
}

output "vm_config" {
  value = {
    name        = data.truenas_vm.by_id.name
    vcpus       = data.truenas_vm.by_id.vcpus
    cores       = data.truenas_vm.by_id.cores
    threads     = data.truenas_vm.by_id.threads
    memory      = data.truenas_vm.by_id.memory
    min_memory  = data.truenas_vm.by_id.min_memory
    autostart   = data.truenas_vm.by_id.autostart
    bootloader  = data.truenas_vm.by_id.bootloader
    cpu_mode    = data.truenas_vm.by_id.cpu_mode
    cpu_model   = data.truenas_vm.by_id.cpu_model
    status      = data.truenas_vm.by_id.status
  }
}
```

### Conditional Resource Based on VM Status

```terraform
data "truenas_vm" "database" {
  name = "database-server"
}

# Only create backup if VM is running
resource "truenas_snapshot" "db_backup" {
  count = data.truenas_vm.database.status == "RUNNING" ? 1 : 0
  
  dataset = "tank/vms/${data.truenas_vm.database.name}"
  name = "auto-backup-${formatdate("YYYY-MM-DD-hhmm", timestamp())}"
}
```

### VM Resource Validation

```terraform
data "truenas_vm" "production" {
  name = "production-app"
}

locals {
  # Validate production VM configuration
  is_valid = (
    data.truenas_vm.production.vcpus >= 4 &&
    data.truenas_vm.production.memory >= 8192 &&
    data.truenas_vm.production.autostart == true
  )
}

resource "null_resource" "validation" {
  count = local.is_valid ? 0 : 1
  
  provisioner "local-exec" {
    command = <<-EOT
      echo "Error: Production VM validation failed"
      echo "VM: ${data.truenas_vm.production.name}"
      echo "vCPUs: ${data.truenas_vm.production.vcpus} (required: >=4)"
      echo "Memory: ${data.truenas_vm.production.memory}MiB (required: >=8192)"
      echo "Autostart: ${data.truenas_vm.production.autostart} (required: true)"
      exit 1
    EOT
  }
}
```

### Get VM Guest IP Information

```terraform
data "truenas_vm" "target" {
  name = "target-vm"
}

data "truenas_vm_guest_info" "target_ip" {
  vm_id = data.truenas_vm.target.id
}

output "vm_network_info" {
  value = {
    name = data.truenas_vm.target.name
    status = data.truenas_vm.target.status
    ipv4_addresses = data.truenas_vm_guest_info.target_ip.ipv4_addresses
    ipv6_addresses = data.truenas_vm_guest_info.target_ip.ipv6_addresses
  }
}
```

### VM Configuration Comparison

```terraform
data "truenas_vm" "template" {
  name = "vm-template"
}

data "truenas_vm" "instance" {
  name = "vm-instance"
}

locals {
  config_match = (
    data.truenas_vm.template.vcpus == data.truenas_vm.instance.vcpus &&
    data.truenas_vm.template.memory == data.truenas_vm.instance.memory &&
    data.truenas_vm.template.cores == data.truenas_vm.instance.cores &&
    data.truenas_vm.template.threads == data.truenas_vm.instance.threads
  )
}

output "config_comparison" {
  value = {
    template = {
      vcpus = data.truenas_vm.template.vcpus
      memory = data.truenas_vm.template.memory
      cores = data.truenas_vm.template.cores
      threads = data.truenas_vm.template.threads
    }
    instance = {
      vcpus = data.truenas_vm.instance.vcpus
      memory = data.truenas_vm.instance.memory
      cores = data.truenas_vm.instance.cores
      threads = data.truenas_vm.instance.threads
    }
    config_match = local.config_match
  }
}
```

### VM Migration Planning

```terraform
data "truenas_vm" "migrate_me" {
  name = "vm-to-migrate"
}

locals {
  vm_config = {
    name = data.truenas_vm.migrate_me.name
    description = data.truenas_vm.migrate_me.description
    vcpus = data.truenas_vm.migrate_me.vcpus
    cores = data.truenas_vm.migrate_me.cores
    threads = data.truenas_vm.migrate_me.threads
    memory = data.truenas_vm.migrate_me.memory
    min_memory = data.truenas_vm.migrate_me.min_memory
    autostart = data.truenas_vm.migrate_me.autostart
    bootloader = data.truenas_vm.migrate_me.bootloader
    cpu_mode = data.truenas_vm.migrate_me.cpu_mode
    cpu_model = data.truenas_vm.migrate_me.cpu_model
  }
}

output "migration_config" {
  value = local.vm_config
}

# Example of creating VM on another system with same config
# resource "truenas_vm" "migrated" {
#   for_each = data.truenas_vm.migrate_me.status == "STOPPED" ? toset(["migrate"]) : toset([])
#   
#   name = "${local.vm_config.name}-migrated"
#   description = local.vm_config.description
#   
#   vcpus = local.vm_config.vcpus
#   cores = local.vm_config.cores
#   threads = local.vm_config.threads
#   memory = local.vm_config.memory
#   
#   autostart = local.vm_config.autostart
#   bootloader = local.vm_config.bootloader
#   cpu_mode = local.vm_config.cpu_mode
#   cpu_model = local.vm_config.cpu_model
# }
```

## Schema

### Optional

- `id` (String) VM ID (numeric) - specify either id or name.
- `name` (String) VM name - specify either id or name.

**At least one of `id` or `name` must be specified.**

### Read-Only

- `id` (String) VM ID.
- `name` (String) VM name.
- `description` (String) VM description.
- `vcpus` (Number) Number of virtual CPUs.
- `cores` (Number) Number of cores per socket.
- `threads` (Number) Number of threads per core.
- `memory` (Number) Memory in MiB.
- `min_memory` (Number) Minimum memory in MiB.
- `autostart` (Boolean) Whether VM starts automatically on boot.
- `bootloader` (String) Bootloader type (UEFI, GRUB, etc.).
- `cpu_mode` (String) CPU mode (HOST-PASSTHROUGH, etc.).
- `cpu_model` (String) CPU model.
- `status` (String) VM status (RUNNING, STOPPED, etc.).

## Notes

### VM Identification

You can query a VM using either:
- **ID**: Numeric VM identifier (e.g., "1", "2", "3")
- **Name**: VM name as configured (e.g., "web-server", "database")

Using the VM name is recommended for better readability.

### VM Status Values

Common VM statuses:
- **RUNNING**: VM is currently running
- **STOPPED**: VM is stopped
- **PAUSED**: VM is paused
- **LOADING**: VM is starting up
- **STOPPING**: VM is shutting down

### Resource Units

- **Memory**: Provided in MiB (mebibytes)
  - 1 GiB = 1024 MiB
  - To convert to GB: `memory / 1024`
- **vCPUs**: Total number of virtual CPUs
- **Cores**: Number of cores per socket
- **Threads**: Number of threads per core

### CPU Configuration

The relationship between CPU attributes:
- `vcpus` = `cores` × `threads`
- Common configurations:
  - 2 vCPUs: 2 cores × 1 thread
  - 4 vCPUs: 4 cores × 1 thread or 2 cores × 2 threads
  - 8 vCPUs: 8 cores × 1 thread, 4 cores × 2 threads, or 2 cores × 4 threads

### Bootloader Types

Common bootloader options:
- **UEFI**: Modern UEFI firmware (recommended)
- **UEFI_CSM**: UEFI with Compatibility Support Module
- **GRUB**: GRUB bootloader (for specific use cases)

### CPU Modes

Available CPU modes:
- **HOST-PASSTHROUGH**: Best performance, passes host CPU features
- **HOST-MODEL**: Emulates host CPU model
- **CUSTOM**: Uses specified CPU model

### Use Cases

#### VM Discovery
```terraform
data "truenas_vm" "existing" {
  name = "existing-vm"
}

# Use VM information for other resources
resource "truenas_snapshot" "vm_snapshot" {
  dataset = "tank/vms/${data.truenas_vm.existing.name}"
  name = "manual-snapshot"
}
```

#### Configuration Validation
```terraform
data "truenas_vm" "validate" {
  name = "critical-vm"
}

locals {
  meets_requirements = (
    data.truenas_vm.validate.vcpus >= 2 &&
    data.truenas_vm.validate.memory >= 4096 &&
    data.truenas_vm.validate.autostart == true
  )
}

# Alert if VM doesn't meet requirements
resource "null_resource" "alert" {
  count = local.meets_requirements ? 0 : 1
  
  provisioner "local-exec" {
    command = "echo 'VM ${data.truenas_vm.validate.name} does not meet requirements!'"
  }
}
```

#### Resource Planning
```terraform
data "truenas_vm" "analyze" {
  name = "vm-to-analyze"
}

locals {
  cpu_efficiency = data.truenas_vm.analyze.vcpus / (data.truenas_vm.analyze.cores * data.truenas_vm.analyze.threads)
  memory_per_vcpu = data.truenas_vm.analyze.memory / data.truenas_vm.analyze.vcpus
}

output "analysis" {
  value = {
    cpu_efficiency = local.cpu_efficiency
    memory_per_vcpu_mib = local.memory_per_vcpu
    memory_per_vcpu_gb = local.memory_per_vcpu / 1024
  }
}
```

### Integration with Other Resources

The VM data source is commonly used with:

```terraform
data "truenas_vm" "target" {
  name = "target-vm"
}

# Get VM IP information
data "truenas_vm_guest_info" "network" {
  vm_id = data.truenas_vm.target.id
}

# Create snapshots
resource "truenas_snapshot" "backup" {
  dataset = "tank/vms/${data.truenas_vm.target.name}"
  name = "backup-${formatdate("YYYY-MM-DD", timestamp())}"
}

# Monitor VM status
resource "null_resource" "monitor" {
  triggers = {
    status = data.truenas_vm.target.status
  }
  
  provisioner "local-exec" {
    command = "echo 'VM ${data.truenas_vm.target.name} status: ${data.truenas_vm.target.status}'"
  }
}
```

### Best Practices

1. **Use VM names** instead of IDs for better readability
2. **Validate VM status** before performing operations
3. **Monitor resource usage** for capacity planning
4. **Document VM configurations** for consistency
5. **Use for discovery** when managing existing infrastructure

### Common Patterns

#### VM Health Check
```terraform
data "truenas_vm" "health_check" {
  name = "production-vm"
}

locals {
  is_healthy = data.truenas_vm.health_check.status == "RUNNING"
}

output "vm_health" {
  value = {
    name = data.truenas_vm.health_check.name
    status = data.truenas_vm.health_check.status
    healthy = local.is_healthy
  }
}
```

#### Configuration Export
```terraform
data "truenas_vm" "export" {
  name = "vm-to-export"
}

output "vm_config" {
  value = {
    name = data.truenas_vm.export.name
    description = data.truenas_vm.export.description
    vcpus = data.truenas_vm.export.vcpus
    cores = data.truenas_vm.export.cores
    threads = data.truenas_vm.export.threads
    memory = data.truenas_vm.export.memory
    autostart = data.truenas_vm.export.autostart
    bootloader = data.truenas_vm.export.bootloader
    cpu_mode = data.truenas_vm.export.cpu_mode
    cpu_model = data.truenas_vm.export.cpu_model
  }
}
```

### Troubleshooting

**VM Not Found:**
- Verify VM name spelling
- Check if VM exists in TrueNAS web interface
- Try using numeric VM ID instead

**Status Not Updating:**
- VM status may have a delay
- Check TrueNAS web interface for current status
- Verify VM is not in transition state

**Missing Attributes:**
- Some attributes may be null if not configured
- Check VM configuration in TrueNAS
- Verify TrueNAS version supports all features

## See Also

- [truenas_vms](vms) - Query all VMs
- [truenas_vm Resource](../resources/vm) - Manage VMs
- [truenas_vm_guest_info](vm_guest_info) - Get VM IP addresses
- [truenas_vm_device](../resources/vm_device) - Manage VM devices
- [VM Management Guide](https://www.truenas.com/docs/scale/virtualmachines/) - TrueNAS documentation