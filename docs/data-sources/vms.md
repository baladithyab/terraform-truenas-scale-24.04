---
page_title: "truenas_vms Data Source - terraform-provider-truenas"
subcategory: "Virtual Machines"
description: |-
  Fetches information about all VMs on the TrueNAS system.
---

# truenas_vms (Data Source)

Fetches information about all virtual machines on the TrueNAS system. This data source returns a list of all VMs with their configuration and status information.

## Example Usage

### List All VMs

```terraform
data "truenas_vms" "all" {}

output "vm_list" {
  value = data.truenas_vms.all.vms
}
```

### Filter Running VMs

```terraform
data "truenas_vms" "all" {}

locals {
  running_vms = [
    for vm in data.truenas_vms.all.vms :
    vm
    if vm.status == "RUNNING"
  ]
}

output "running_vms" {
  value = local.running_vms
}
```

### Calculate Total Resources

```terraform
data "truenas_vms" "inventory" {}

locals {
  total_vcpus = sum([
    for vm in data.truenas_vms.inventory.vms :
    vm.vcpus
  ])
  
  total_memory = sum([
    for vm in data.truenas_vms.inventory.vms :
    vm.memory
  ])
  
  autostart_count = length([
    for vm in data.truenas_vms.inventory.vms :
    vm
    if vm.autostart
  ])
}

output "resource_summary" {
  value = {
    vm_count       = length(data.truenas_vms.inventory.vms)
    total_vcpus    = local.total_vcpus
    total_memory   = local.total_memory
    autostart_vms  = local.autostart_count
  }
}
```

### VM Inventory Report

```terraform
data "truenas_vms" "report" {}

output "vm_inventory" {
  value = {
    for vm in data.truenas_vms.report.vms :
    vm.name => {
      id          = vm.id
      description = vm.description
      vcpus       = vm.vcpus
      cores       = vm.cores
      threads     = vm.threads
      memory_mb   = vm.memory
      memory_gb   = vm.memory / 1024
      autostart   = vm.autostart
      status      = vm.status
    }
  }
}
```

### Find VMs by Status

```terraform
data "truenas_vms" "status_check" {}

locals {
  stopped_vms = [
    for vm in data.truenas_vms.status_check.vms :
    {
      name = vm.name
      id = vm.id
      memory = vm.memory
    }
    if vm.status == "STOPPED"
  ]
  
  running_vms = [
    for vm in data.truenas_vms.status_check.vms :
    {
      name = vm.name
      id = vm.id
      memory = vm.memory
    }
    if vm.status == "RUNNING"
  ]
}

output "vm_status_summary" {
  value = {
    stopped = local.stopped_vms
    running = local.running_vms
  }
}
```

### Resource Planning

```terraform
data "truenas_vms" "current" {}

locals {
  # Calculate current resource usage
  used_vcpus = sum([
    for vm in data.truenas_vms.current.vms :
    vm.vcpus
  ])
  
  used_memory = sum([
    for vm in data.truenas_vms.current.vms :
    vm.memory
  ])
  
  # Define available resources
  available_vcpus = 16
  available_memory = 32768  # 32GB
  
  # Calculate remaining capacity
  remaining_vcpus = local.available_vcpus - local.used_vcpus
  remaining_memory = local.available_memory - local.used_memory
}

output "capacity_planning" {
  value = {
    used_vcpus = local.used_vcpus
    used_memory_gb = local.used_memory / 1024
    remaining_vcpus = local.remaining_vcpus
    remaining_memory_gb = local.remaining_memory / 1024
  }
}

# Alert if resources are running low
resource "null_resource" "resource_alert" {
  count = local.remaining_vcpus < 2 || local.remaining_memory < 2048 ? 1 : 0
  
  provisioner "local-exec" {
    command = <<-EOT
      echo "Warning: Low VM resources available!"
      echo "Remaining vCPUs: ${local.remaining_vcpus}"
      echo "Remaining Memory: ${local.remaining_memory / 1024}GB"
    EOT
  }
}
```

### VM Configuration Analysis

```terraform
data "truenas_vms" "analysis" {}

locals {
  # Analyze CPU configurations
  single_core_vms = length([
    for vm in data.truenas_vms.analysis.vms :
    vm
    if vm.cores == 1
  ])
  
  multi_core_vms = length([
    for vm in data.truenas_vms.analysis.vms :
    vm
    if vm.cores > 1
  ])
  
  # Analyze memory usage
  small_vms = length([
    for vm in data.truenas_vms.analysis.vms :
    vm
    if vm.memory <= 2048
  ])
  
  large_vms = length([
    for vm in data.truenas_vms.analysis.vms :
    vm
    if vm.memory > 8192
  ])
}

output "vm_analysis" {
  value = {
    total_vms = length(data.truenas_vms.analysis.vms)
    single_core_vms = local.single_core_vms
    multi_core_vms = local.multi_core_vms
    small_vms = local.small_vms
    large_vms = local.large_vms
  }
}
```

## Schema

### Read-Only

- `vms` (List of Object) List of VMs with the following attributes:
  - `id` (String) VM ID
  - `name` (String) VM name
  - `description` (String) VM description
  - `vcpus` (Number) Number of virtual CPUs
  - `cores` (Number) Number of cores per socket
  - `threads` (Number) Number of threads per core
  - `memory` (Number) Memory in MiB
  - `autostart` (Boolean) Whether VM starts automatically on boot
  - `status` (String) VM status (RUNNING, STOPPED, etc.)

## Notes

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

### Use Cases

#### VM Discovery
```terraform
data "truenas_vms" "discovery" {}

locals {
  vm_names = [
    for vm in data.truenas_vms.discovery.vms :
    vm.name
  ]
}

output "discovered_vms" {
  value = local.vm_names
}
```

#### Capacity Monitoring
```terraform
data "truenas_vms" "monitor" {}

locals {
  total_memory = sum([
    for vm in data.truenas_vms.monitor.vms :
    vm.memory
  ])
  
  running_memory = sum([
    for vm in data.truenas_vms.monitor.vms :
    vm.memory
    if vm.status == "RUNNING"
  ])
}

output "memory_usage" {
  value = {
    total_allocated_gb = local.total_memory / 1024
    running_gb = local.running_memory / 1024
  }
}
```

#### VM Backup Planning
```terraform
data "truenas_vms" "backup_targets" {}

locals {
  backup_candidates = [
    for vm in data.truenas_vms.backup_targets.vms :
    {
      name = vm.name
      id = vm.id
      memory_gb = vm.memory / 1024
    }
    if vm.status == "RUNNING" && vm.autostart == true
  ]
}

output "backup_targets" {
  value = local.backup_candidates
}
```

#### Configuration Audit
```terraform
data "truenas_vms" "audit" {}

locals {
  # Find VMs without descriptions
  undocumented_vms = [
    for vm in data.truenas_vms.audit.vms :
    vm.name
    if vm.description == null || vm.description == ""
  ]
  
  # Find VMs without autostart
  manual_start_vms = [
    for vm in data.truenas_vms.audit.vms :
    vm.name
    if vm.autostart == false
  ]
}

output "audit_results" {
  value = {
    undocumented_vms = local.undocumented_vms
    manual_start_vms = local.manual_start_vms
  }
}
```

### Integration with Other Resources

The VMs data source is commonly used with:

```terraform
data "truenas_vms" "existing" {}

# Create snapshots for all running VMs
resource "truenas_snapshot" "vm_snapshots" {
  for_each = {
    for vm in data.truenas_vms.existing.vms :
    vm.id => vm
    if vm.status == "RUNNING"
  }
  
  dataset = "tank/vms/${each.value.name}"
  name = "auto-${formatdate("YYYY-MM-DD-hhmm", timestamp())}"
}

# Get IP information for running VMs
data "truenas_vm_guest_info" "vm_ips" {
  for_each = {
    for vm in data.truenas_vms.existing.vms :
    vm.id => vm
    if vm.status == "RUNNING"
  }
  
  vm_id = each.value.id
}

output "running_vm_ips" {
  value = {
    for id, info in data.truenas_vm_guest_info.vm_ips :
    info.vm_name => info.ipv4_addresses
  }
}
```

### Best Practices

1. **Monitor resource usage** to prevent overallocation
2. **Track VM status** for operational awareness
3. **Use for discovery** when managing existing infrastructure
4. **Plan capacity** based on current usage patterns
5. **Audit configurations** for consistency and compliance

### Common Patterns

#### VM Resource Summary
```terraform
data "truenas_vms" "summary" {}

locals {
  vm_summary = {
    total = length(data.truenas_vms.summary.vms)
    running = length([
      for vm in data.truenas_vms.summary.vms :
      vm
      if vm.status == "RUNNING"
    ])
    stopped = length([
      for vm in data.truenas_vms.summary.vms :
      vm
      if vm.status == "STOPPED"
    ])
    total_vcpus = sum([
      for vm in data.truenas_vms.summary.vms :
      vm.vcpus
    ])
    total_memory_gb = sum([
      for vm in data.truenas_vms.summary.vms :
      vm.memory
    ]) / 1024
  }
}

output "vm_summary" {
  value = local.vm_summary
}
```

#### Find Large VMs
```terraform
data "truenas_vms" "large" {}

locals {
  large_vms = [
    for vm in data.truenas_vms.large.vms :
    {
      name = vm.name
      vcpus = vm.vcpus
      memory_gb = vm.memory / 1024
    }
    if vm.memory > 8192 || vm.vcpus > 4
  ]
}

output "large_vms" {
  value = local.large_vms
}
```

### Troubleshooting

**Empty VM List:**
- Check if VM service is enabled in TrueNAS
- Verify user has permissions to view VMs
- Ensure VMs exist on the system

**Incorrect Status:**
- VM status may have a delay in updating
- Check TrueNAS web interface for current status
- Verify VM is not in transition state

**Resource Calculation Issues:**
- Memory is in MiB, not MB
- vCPus include all threads
- Check for VMs with unusual configurations

## See Also

- [truenas_vm](vm) - Query a specific VM
- [truenas_vm Resource](../resources/vm) - Manage VMs
- [truenas_vm_guest_info](vm_guest_info) - Get VM IP addresses
- [truenas_vm_device](../resources/vm_device) - Manage VM devices
- [VM Management Guide](https://www.truenas.com/docs/scale/virtualmachines/) - TrueNAS documentation