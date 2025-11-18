---
page_title: "TrueNAS Provider"
subcategory: ""
description: |-
  The TrueNAS provider enables Terraform to manage TrueNAS Scale resources through the TrueNAS REST API.
---

# TrueNAS Provider

The TrueNAS provider allows you to manage [TrueNAS Scale](https://www.truenas.com/truenas-scale/) resources using Terraform. It provides comprehensive resource management for storage, virtualization, networking, and application deployment on TrueNAS Scale 24.04 and later.

## What's New

🎉 **Complete Documentation Coverage**: This release includes comprehensive documentation for all 15 resources and 10 data sources, with categorized navigation, practical examples, and detailed configuration guides.

## Features

- **Storage Management**: ZFS datasets, snapshots, and automated snapshot tasks
- **File Sharing**: NFS and SMB shares with full configuration support
- **Virtual Machines**: Complete VM lifecycle management with device attachment, GPU passthrough, and cloud-init support
- **User Management**: Users and groups with comprehensive permission controls
- **Network Configuration**: Static routes and interface management
- **iSCSI Storage**: Complete iSCSI infrastructure with targets, extents, and portals
- **Kubernetes Applications**: Chart deployment and management for TrueNAS apps
- **Hardware Discovery**: PCI passthrough, GPU selection, and IOMMU capability detection
- **Cloud-Init Support**: Automated VM initialization with user-data, meta-data, and network configuration
- **Import Support**: Import existing resources into Terraform state
- **Comprehensive Data Sources**: Query and discover existing TrueNAS resources

## Supported TrueNAS Versions

- TrueNAS Scale 24.04 (Dragonfish)
- TrueNAS Scale 24.10 (Electric Eel)

## Authentication

The provider requires API credentials for your TrueNAS Scale instance. API keys can be generated from the TrueNAS web interface under **System Settings > API Keys**.

## Quick Start Examples

### Basic Storage and File Sharing Setup

```terraform
terraform {
  required_providers {
    truenas = {
      source  = "baladithyab/truenas"
      version = "~> 0.2"
    }
  }
}

provider "truenas" {
  base_url = "https://truenas.example.com"
  api_key  = var.truenas_api_key
}

# Create a ZFS dataset for application data
resource "truenas_dataset" "app_data" {
  name = "tank/applications"
  type = "FILESYSTEM"
  
  quota {
    quota = 107374182400  # 100GB
    quota_type = "DATASET"
  }
}

# Create an NFS share for application data
resource "truenas_nfs_share" "app_nfs" {
  path    = "/mnt/${truenas_dataset.app_data.name}"
  comment = "Application data share managed by Terraform"
  
  networks = ["192.168.1.0/24"]
  security {
    maproot_user = "root"
  }
}

# Create an SMB share for Windows clients
resource "truenas_smb_share" "app_smb" {
  path    = "/mnt/${truenas_dataset.app_data.name}"
  name    = "app-data"
  comment = "Application data SMB share"
  
  hostsallow = ["192.168.1.0/24"]
}
```

### Virtual Machine with GPU Passthrough

```terraform
# Discover available GPU devices
data "truenas_gpu_pci_choices" "available_gpus" {}

# Check IOMMU capability
data "truenas_vm_iommu_enabled" "iommu_check" {}

# Create a virtual machine with GPU passthrough
resource "truenas_vm" "gpu_vm" {
  name        = "gpu-workstation"
  description = "VM with GPU passthrough for graphics workloads"
  
  vcpus  = 4
  memory = 8192
  
  autostart = true
  
  boot_device {
    type = "CDROM"
    file = "/mnt/tank/iso/ubuntu-22.04.iso"
  }
}

# Attach GPU device to VM
resource "truenas_vm_device" "gpu_device" {
  vm_id = truenas_vm.gpu_vm.id
  
  pci_config {
    pci_address = data.truenas_gpu_pci_choices.available_gpus.choices[0].value
  }
}

# Get VM guest IP information
data "truenas_vm_guest_info" "vm_info" {
  vm_id = truenas_vm.gpu_vm.id
}
```

### Complete iSCSI Infrastructure

```terraform
# Create iSCSI portal
resource "truenas_iscsi_portal" "main_portal" {
  listen_addresses = ["0.0.0.0:3260"]
  comment          = "Main iSCSI portal"
}

# Create iSCSI extent
resource "truenas_iscsi_extent" "data_extent" {
  name        = "data-extent"
  type        = "FILE"
  path        = "/mnt/tank/iscsi/data-extent"
  filesize    = 10737418240  # 10GB
  blocksize   = 4096
  comment     = "Data storage extent"
}

# Create iSCSI target
resource "truenas_iscsi_target" "data_target" {
  name     = "data-target"
  alias    = "Data Storage Target"
  comment  = "Target for data storage"
  
  groups = ["authenticated"]
  auth_networks = ["192.168.1.0/24"]
  
  extents {
    extent_id = truenas_iscsi_extent.data_extent.id
  }
  
  portals {
    portal_id = truenas_iscsi_portal.main_portal.id
  }
}
```

### Kubernetes Application Deployment

```terraform
# Deploy a Kubernetes application using TrueNAS charts
resource "truenas_chart_release" "nextcloud" {
  name        = "nextcloud"
  catalog     = "OFFICIAL"
  train       = "stable"
  version     = "latest"
  
  values = jsonencode({
    nextcloud: {
      host: "nextcloud.example.com"
      username: "admin"
      password: "secure-password"
    }
  })
}
```

## Configuration Reference

### Provider Configuration

- `base_url` (String, Required) - The base URL of your TrueNAS Scale instance (e.g., `https://truenas.example.com`)
- `api_key` (String, Required, Sensitive) - API key for authentication. Can also be set via `TRUENAS_API_KEY` environment variable
- `skip_tls_verify` (Boolean, Optional) - Skip TLS certificate verification. Useful for self-signed certificates. Default: `false`

### Environment Variables

- `TRUENAS_BASE_URL` - Alternative to `base_url` configuration
- `TRUENAS_API_KEY` - Alternative to `api_key` configuration

## Getting Started

1. **Generate an API Key**
   - Log into your TrueNAS Scale web interface
   - Navigate to **System Settings > API Keys**
   - Click **Add** and create a new API key
   - Save the generated key securely

2. **Configure the Provider**
   ```terraform
   provider "truenas" {
     base_url = "https://your-truenas-host"
     api_key  = var.truenas_api_key
   }
   ```

3. **Use Resources**
   - See the [Resources](#resources) section for available resource types
   - Check the [Guides](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/guides) for common use cases

## Additional Resources

- [GitHub Repository](https://github.com/baladithyab/terraform-provider-truenas)
- [Import Guide](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/guides/import_guide)
- [Known Limitations](https://github.com/baladithyab/terraform-provider-truenas/blob/main/KNOWN_LIMITATIONS.md)
- [Issue Tracker](https://github.com/baladithyab/terraform-provider-truenas/issues)

## Resources

The following resources are supported:

### Storage
- [truenas_dataset](./resources/dataset) - ZFS dataset management
- [truenas_snapshot](./resources/snapshot) - Manual snapshot creation
- [truenas_periodic_snapshot_task](./resources/periodic_snapshot_task) - Automated snapshot scheduling

### File Sharing
- [truenas_nfs_share](./resources/nfs_share) - NFS share configuration
- [truenas_smb_share](./resources/smb_share) - SMB/CIFS share configuration

### Virtual Machines
- [truenas_vm](./resources/vm) - Virtual machine management
- [truenas_vm_device](./resources/vm_device) - VM device management (NICs, disks, PCI passthrough)

### User Management
- [truenas_user](./resources/user) - User account management
- [truenas_group](./resources/group) - User group management

### Network
- [truenas_interface](./resources/interface) - Network interface management
- [truenas_static_route](./resources/static_route) - Static route configuration

### iSCSI
- [truenas_iscsi_target](./resources/iscsi_target) - iSCSI target configuration
- [truenas_iscsi_extent](./resources/iscsi_extent) - iSCSI storage extent management
- [truenas_iscsi_portal](./resources/iscsi_portal) - iSCSI portal management

### Kubernetes
- [truenas_chart_release](./resources/chart_release) - Kubernetes chart releases management

## Data Sources

The following data sources are supported:

### Storage & Discovery
- [truenas_dataset](./data-sources/dataset) - Dataset information
- [truenas_pool](./data-sources/pool) - Storage pool information

### Virtual Machines
- [truenas_vm](./data-sources/vm) - VM information by ID or name
- [truenas_vms](./data-sources/vms) - List all VMs
- [truenas_vm_guest_info](./data-sources/vm_guest_info) - VM guest IP information

### Hardware Discovery
- [truenas_vm_pci_passthrough_devices](./data-sources/vm_pci_passthrough_devices) - PCI passthrough device discovery
- [truenas_vm_iommu_enabled](./data-sources/vm_iommu_enabled) - IOMMU capability checking
- [truenas_gpu_pci_choices](./data-sources/gpu_pci_choices) - GPU PCI device selection

### File Sharing
- [truenas_nfs_shares](./data-sources/nfs_shares) - NFS shares listing
- [truenas_smb_shares](./data-sources/smb_shares) - SMB shares listing

## Examples

The provider includes comprehensive examples for various use cases:

### Complete Setups
- [Complete Infrastructure Setup](../../examples/complete/) - Full TrueNAS infrastructure with storage, networking, and VMs
- [Complete iSCSI Setup](../../examples/complete-iscsi/) - Complete iSCSI infrastructure example
- [Complete Kubernetes Setup](../../examples/complete-kubernetes/) - Kubernetes application deployment
- [Complete Network Setup](../../examples/complete-network/) - Network configuration and routing

### Virtual Machine Examples
- [VM with Devices](../../examples/vm-with-devices/) - VM with multiple device attachments
- [VM Boot Order](../../examples/vm-boot-order/) - Custom boot order configuration
- [Boot Order Testing](../../examples/boot-order/) - Boot order verification and testing
- [VM GPU Passthrough](../../examples/vm-gpu-passthrough/) - GPU passthrough setup
- [VM IP Discovery](../../examples/vm-ip-discovery/) - Dynamic IP discovery for VMs

### Individual Resource Examples
- [Dataset Management](../../examples/resources/truenas_dataset/) - ZFS dataset creation and management
- [NFS Share](../../examples/resources/truenas_nfs_share/) - NFS share configuration
- [SMB Share](../../examples/resources/truenas_smb_share/) - SMB share setup
- [User Management](../../examples/resources/truenas_user/) - User account creation
- [Group Management](../../examples/resources/truenas_group/) - User group configuration
- [VM Creation](../../examples/resources/truenas_vm/) - Basic VM setup
- [VM Device Management](../../examples/resources/truenas_vm_device/) - VM device attachment
- [Interface Configuration](../../examples/resources/truenas_interface/) - Network interface setup
- [iSCSI Target](../../examples/resources/truenas_iscsi_target/) - iSCSI target configuration
- [iSCSI Extent](../../examples/resources/truenas_iscsi_extent/) - iSCSI extent management
- [iSCSI Portal](../../examples/resources/truenas_iscsi_portal/) - iSCSI portal setup
- [Snapshot Management](../../examples/resources/truenas_snapshot/) - Manual snapshot creation
- [Periodic Snapshot Tasks](../../examples/resources/truenas_periodic_snapshot_task/) - Automated snapshot scheduling
- [Static Routes](../../examples/resources/truenas_static_route/) - Network routing configuration
- [Chart Releases](../../examples/resources/truenas_chart_release/) - Kubernetes application deployment

### Data Source Examples
- [VM Guest Info](../../examples/data-sources/truenas_vm_guest_info/) - VM guest information discovery
- [Data Sources Discovery](../../examples/data-sources-discovery/) - Comprehensive data source usage

### Specialized Setups
- [Talos Minimal](../../examples/talos-minimal/) - Minimal Talos OS setup
- [Volume Size Management](../../examples/volsize/) - Dynamic volume sizing

## Guides

- [Quick Start Guide](./guides/QUICKSTART.md) - Getting started with the provider
- [Import Guide](./guides/IMPORT_GUIDE.md) - Import existing resources
- [Kubernetes Migration](./guides/KUBERNETES_MIGRATION.md) - Migrating Kubernetes applications
- [VM IP Discovery](./guides/VM_IP_DISCOVERY.md) - Dynamic IP discovery for VMs
- [Testing](./guides/TESTING.md) - Testing provider functionality