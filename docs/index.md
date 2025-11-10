---
page_title: "TrueNAS Provider"
subcategory: ""
description: |-
  The TrueNAS provider enables Terraform to manage TrueNAS Scale resources through the TrueNAS REST API.
---

# TrueNAS Provider

The TrueNAS provider allows you to manage [TrueNAS Scale](https://www.truenas.com/truenas-scale/) resources using Terraform. It provides comprehensive resource management for storage, virtualization, networking, and application deployment on TrueNAS Scale 24.04 and later.

## Features

- **Storage Management**: ZFS datasets, snapshots, and snapshot tasks
- **File Sharing**: NFS and SMB shares with full configuration support
- **Virtual Machines**: Complete VM lifecycle management with device attachment
- **User Management**: Users and groups with permissions
- **Network Configuration**: Static routes and interface management
- **iSCSI Storage**: Targets, extents, and portal configuration
- **Kubernetes Applications**: Chart deployment and management
- **Import Support**: Import existing resources into Terraform state

## Supported TrueNAS Versions

- TrueNAS Scale 24.04 (Dragonfish)
- TrueNAS Scale 24.10 (Electric Eel)

## Authentication

The provider requires API credentials for your TrueNAS Scale instance. API keys can be generated from the TrueNAS web interface under **System Settings > API Keys**.

## Example Usage

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
  
  # Optional: Disable SSL verification for self-signed certificates
  # skip_tls_verify = true
}

# Create a ZFS dataset
resource "truenas_dataset" "example" {
  name = "tank/terraform-managed"
  type = "FILESYSTEM"
  
  quota {
    quota = 107374182400  # 100GB
    quota_type = "DATASET"
  }
}

# Create an NFS share
resource "truenas_nfs_share" "example" {
  path    = "/mnt/${truenas_dataset.example.name}"
  comment = "Managed by Terraform"
  
  networks = ["192.168.1.0/24"]
}

# Deploy a virtual machine
resource "truenas_vm" "example" {
  name        = "ubuntu-vm"
  description = "Ubuntu Server managed by Terraform"
  
  vcpus  = 2
  memory = 4096
  
  autostart = true
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

### Storage & File Sharing
- [truenas_dataset](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/resources/dataset)
- [truenas_snapshot](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/resources/snapshot)
- [truenas_periodic_snapshot_task](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/resources/periodic_snapshot_task)
- [truenas_nfs_share](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/resources/nfs_share)
- [truenas_smb_share](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/resources/smb_share)

### Virtual Machines
- [truenas_vm](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/resources/vm)
- [truenas_vm_device](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/resources/vm_device)

### User Management
- [truenas_user](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/resources/user)
- [truenas_group](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/resources/group)

### Network
- [truenas_interface](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/resources/interface)
- [truenas_static_route](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/resources/static_route)

### iSCSI
- [truenas_iscsi_target](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/resources/iscsi_target)
- [truenas_iscsi_extent](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/resources/iscsi_extent)
- [truenas_iscsi_portal](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/resources/iscsi_portal)

### Applications
- [truenas_chart_release](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/resources/chart_release)

## Data Sources

- [truenas_dataset](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/data-sources/dataset)
- [truenas_pool](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/data-sources/pool)
- [truenas_vm](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/data-sources/vm)
- [truenas_vms](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/data-sources/vms)
- [truenas_vm_guest_info](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/data-sources/vm_guest_info)
- [truenas_vm_pci_passthrough_devices](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/data-sources/vm_pci_passthrough_devices)
- [truenas_vm_iommu_enabled](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/data-sources/vm_iommu_enabled)
- [truenas_gpu_pci_choices](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/data-sources/gpu_pci_choices)
- [truenas_nfs_shares](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/data-sources/nfs_shares)
- [truenas_smb_shares](https://registry.terraform.io/providers/baladithyab/truenas/latest/docs/data-sources/smb_shares)