# Terraform Provider for TrueNAS Scale 24.04

This is a Terraform provider for managing TrueNAS Scale 24.04 resources using the REST API. This provider is specifically designed for TrueNAS Scale 24.04, as version 25.04 transitions from REST to JSON-RPC over WebSocket.

## ⚠️ Known Limitations

**Please read these important limitations before using this provider:**

### TrueNAS Version Compatibility
- **✅ TrueNAS Scale 24.04 ONLY**: This provider uses the REST API which is only available in version 24.04
- **❌ TrueNAS Scale 25.x NOT SUPPORTED**: Version 25.x switched to JSON-RPC over WebSocket and is incompatible with this provider
- **Recommendation**: Stay on TrueNAS Scale 24.04 if you want to use this Terraform provider

### VM IP Address Discovery
- **API Limitation**: The TrueNAS 24.04 REST API does not expose VM IP addresses or guest agent information
- **Workarounds Available**:
  - Use MAC address export + DHCP lookup (works for ALL VMs including Talos)
  - Use SSH-based guest agent query (requires guest agent installed in VM)
- **📖 See**: [`VM_IP_DISCOVERY.md`](docs/guides/VM_IP_DISCOVERY.md) for complete guide and examples

### Static IP Configuration
- **API Limitation**: You cannot configure static IP addresses in the guest OS through the TrueNAS API
- **Workaround**: Configure static IPs manually in the guest OS or use cloud-init/user-data
- **For Talos Linux**: Use Talos machine configuration to set static IPs

### Full Details
For comprehensive information about limitations and workarounds, see [`KNOWN_LIMITATIONS.md`](KNOWN_LIMITATIONS.md).

---

## Features

- **Full REST API Support**: Leverages all available TrueNAS Scale 24.04 REST API endpoints
- **Resource Management**: Create, read, update, and delete TrueNAS resources
- **Import Support**: Import existing TrueNAS resources into Terraform state
- **Comprehensive Resources**:
  - ZFS Datasets & Snapshots with periodic scheduling
  - NFS & SMB Shares
  - Users & Groups
  - **Virtual Machines** with lifecycle management (`desired_state` for started/stopped control)
  - **VM Devices** - Standalone device management (NICs, disks, CDROMs, PCI passthrough)
  - iSCSI (Targets, Extents, Portals)
  - Network (Interfaces, VLANs, Bridges, LAGs, Static Routes)
  - Kubernetes Apps (Chart Releases) with **Migration Support**
- **VM IP Discovery**: Multiple methods for discovering VM IP addresses (see [`VM_IP_DISCOVERY.md`](docs/guides/VM_IP_DISCOVERY.md))
  - MAC address export for DHCP lookup
  - Guest agent queries with enhanced security options
- **Data Sources**:
  - Dataset, Pool, VM information
  - VM guest info (IP addresses, hostname, OS details)
  - Resource discovery (VMs, NFS shares, SMB shares)
  - GPU/PCI device discovery for passthrough
- **Migration Capabilities**:
  - Export Kubernetes apps to external K8s clusters
  - Backup and restore with PVCs
  - Automated migration scripts
  - See [`KUBERNETES_MIGRATION.md`](KUBERNETES_MIGRATION.md)

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- TrueNAS Scale 24.04
- TrueNAS API Key

## Building the Provider

Clone the repository and build the provider:

```bash
git clone https://github.com/baladithyab/terraform-provider-truenas
cd terraform-provider-truenas
go build -o terraform-provider-truenas
```

## Installation

### Using from GitHub (Recommended)

You can use the provider directly from GitHub. First, you need to configure Terraform to use GitHub as a provider source.

Create or update `~/.terraformrc`:

```hcl
provider_installation {
  filesystem_mirror {
    path    = "/home/YOUR_USER/.terraform.d/plugins"
    include = ["registry.terraform.io/*/*"]
  }
  direct {
    exclude = ["registry.terraform.io/*/*"]
  }
}
```

Then in your Terraform configuration:

```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.22"
    }
  }
}

provider "truenas" {
  base_url = "http://10.0.0.83:81"
  api_key  = var.truenas_api_key
}
```

Build and install the provider:

```bash
git clone https://github.com/baladithyab/terraform-provider-truenas.git
cd terraform-provider-truenas
make install
```

Then run:

```bash
terraform init
terraform plan
```

### Building from Source

If you want to build the provider yourself:

```bash
# Clone the repository
git clone https://github.com/baladithyab/terraform-provider-truenas.git
cd terraform-provider-truenas

# Build the provider
go build -o terraform-provider-truenas

# Install locally
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/YOUR_USERNAME/truenas/1.0.0/linux_amd64/
cp terraform-provider-truenas ~/.terraform.d/plugins/registry.terraform.io/YOUR_USERNAME/truenas/1.0.0/linux_amd64/
```

Or use the Makefile:

```bash
make build
make install
```

### Using a Specific Version

To use a specific version or commit:

```hcl
terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.22"
    }
  }
}
```

## Configuration

### Provider Configuration

```hcl
provider "truenas" {
  base_url = "http://10.0.0.83:81"  # Your TrueNAS server URL (with or without port)
  api_key  = "your-api-key-here"      # Your TrueNAS API key
}
```

### Environment Variables

You can also configure the provider using environment variables:

```bash
export TRUENAS_BASE_URL="http://10.0.0.83:81"
export TRUENAS_API_KEY="your-api-key-here"
```

### Getting a TrueNAS API Key

1. Log in to your TrueNAS web interface
2. Navigate to the top-right corner and click on the settings icon
3. Select "API Keys"
4. Click "Add" to create a new API key
5. Give it a name and click "Add"
6. Copy the generated API key (you won't be able to see it again!)

## Usage Examples

### Creating a ZFS Dataset

```hcl
resource "truenas_dataset" "mydata" {
  name        = "tank/mydata"
  type        = "FILESYSTEM"
  compression = "LZ4"
  atime       = "OFF"
  quota       = 107374182400  # 100GB
  comments    = "Managed by Terraform"
}
```

### Creating an NFS Share

```hcl
resource "truenas_nfs_share" "myshare" {
  path     = "/mnt/tank/mydata"
  comment  = "NFS share for mydata"
  networks = ["192.168.1.0/24"]
  readonly = false
  enabled  = true
}
```

### Creating an SMB Share

```hcl
resource "truenas_smb_share" "myshare" {
  name       = "myshare"
  path       = "/mnt/tank/mydata"
  comment    = "SMB share for mydata"
  enabled    = true
  browsable  = true
  guestok    = false
  recyclebin = true
  shadowcopy = true
}
```

### Creating a User

```hcl
resource "truenas_user" "john" {
  username  = "john"
  full_name = "John Doe"
  email     = "john@example.com"
  password  = "SecurePassword123!"
  home      = "/mnt/tank/home/john"
  shell     = "/bin/bash"
  sudo      = false
  smb       = true
}
```

### Creating a Group

```hcl
resource "truenas_group" "developers" {
  name = "developers"
  sudo = false
  smb  = true
}
```

### Using Data Sources

```hcl
# Get information about a dataset
data "truenas_dataset" "existing" {
  id = "tank/mydata"
}

# Get information about a pool
data "truenas_pool" "tank" {
  id = "tank"
}

output "dataset_available_space" {
  value = data.truenas_dataset.existing.available
}
```

## Importing Existing Resources

All resources support importing existing TrueNAS resources into Terraform state:

```bash
# Import a dataset
terraform import truenas_dataset.mydata tank/mydata

# Import an NFS share (use the share ID)
terraform import truenas_nfs_share.myshare 1

# Import an SMB share (use the share ID)
terraform import truenas_smb_share.myshare 1

# Import a user (use the user ID)
terraform import truenas_user.john 1000

# Import a group (use the group ID)
terraform import truenas_group.developers 1000
```

## Available Resources

### Storage & File Sharing
- [`truenas_dataset`](examples/resources/truenas_dataset/resource.tf) - ZFS dataset management
- [`truenas_nfs_share`](examples/resources/truenas_nfs_share/resource.tf) - NFS share management
- [`truenas_smb_share`](examples/resources/truenas_smb_share/resource.tf) - SMB/CIFS share management
- [`truenas_snapshot`](examples/resources/truenas_snapshot/resource.tf) - ZFS snapshot management
- [`truenas_periodic_snapshot_task`](examples/resources/truenas_periodic_snapshot_task/resource.tf) - Automated snapshot scheduling

### Virtual Machines
- [`truenas_vm`](examples/resources/truenas_vm/resource.tf) - Virtual machine management with **lifecycle control** (desired_state)
- [`truenas_vm_device`](examples/resources/truenas_vm_device/resource.tf) - Standalone VM device management

### User Management
- [`truenas_user`](examples/resources/truenas_user/resource.tf) - User account management
- [`truenas_group`](examples/resources/truenas_group/resource.tf) - Group management

### iSCSI
- [`truenas_iscsi_target`](examples/resources/truenas_iscsi_target/resource.tf) - iSCSI target management
- [`truenas_iscsi_extent`](examples/resources/truenas_iscsi_extent/resource.tf) - iSCSI extent management
- [`truenas_iscsi_portal`](examples/resources/truenas_iscsi_portal/resource.tf) - iSCSI portal management

### Network
- [`truenas_interface`](examples/resources/truenas_interface/resource.tf) - Network interface management
- [`truenas_static_route`](examples/resources/truenas_static_route/resource.tf) - Static route management

### Kubernetes/Apps
- [`truenas_chart_release`](examples/resources/truenas_chart_release/resource.tf) - Kubernetes application deployment

## Available Data Sources

- [`truenas_dataset`](examples/data-sources/) - Get information about a dataset
- [`truenas_pool`](examples/data-sources/) - Get information about a pool
- [`truenas_vm_guest_info`](examples/data-sources/truenas_vm_guest_info/) - Query VM guest agent for IP addresses and OS info
- [`truenas_vms`](examples/data-sources/) - List all VMs with status
- [`truenas_vm`](examples/data-sources/) - Get specific VM information
- [`truenas_nfs_shares`](examples/data-sources/) - List all NFS shares
- [`truenas_smb_shares`](examples/data-sources/) - List all SMB shares
- [`truenas_gpu_pci_choices`](examples/data-sources/) - Discover available GPUs
- [`truenas_vm_pci_passthrough_devices`](examples/data-sources/) - List PCI passthrough devices
- [`truenas_vm_iommu_enabled`](examples/data-sources/) - Check IOMMU status

## Development

### Running Tests

```bash
go test ./...
```

### Building Documentation

```bash
go generate
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This provider is distributed under the Mozilla Public License 2.0. See `LICENSE` for more information.

## Support

For issues and questions:
- Open an issue on GitHub
- Check the TrueNAS Scale 24.04 API documentation at `http://your-truenas-ip/api/docs/`

## Roadmap

The TrueNAS Scale 24.04 API has 148,000+ lines of OpenAPI specification with hundreds of endpoints. Here are the planned additions:

### High Priority Resources

- **Virtual Machines** (`truenas_vm`) ✅ **IMPLEMENTED**
  - ✅ Create and manage VMs
  - ✅ Configure CPU, memory, storage
  - ✅ Manage VM devices (NICs, disks, USB, PCI passthrough)
  - ✅ VM lifecycle management via `desired_state` attribute (started/stopped)
  - ✅ Standalone device management via `truenas_vm_device` resource

- **iSCSI**
  - `truenas_iscsi_target` - iSCSI targets
  - `truenas_iscsi_extent` - Storage extents
  - `truenas_iscsi_portal` - Network portals
  - `truenas_iscsi_initiator` - Initiator groups
  - `truenas_iscsi_auth` - Authentication

- **Kubernetes/Apps**
  - `truenas_kubernetes_config` - K8s cluster configuration
  - `truenas_chart_release` - Application deployments
  - Kubernetes backup/restore

- **Network Configuration**
  - `truenas_interface` - Network interfaces
  - `truenas_static_route` - Static routes
  - `truenas_vlan` - VLAN configuration
  - `truenas_bridge` - Network bridges
  - `truenas_lagg` - Link aggregation

- **Snapshots & Replication**
  - `truenas_snapshot` - ZFS snapshots
  - `truenas_replication_task` - Replication tasks
  - `truenas_periodic_snapshot_task` - Snapshot schedules

- **Cloud Sync**
  - `truenas_cloudsync_credentials` - Cloud credentials
  - `truenas_cloudsync_task` - Sync tasks

- **System Services**
  - `truenas_service` - Service management (start/stop/enable)
  - `truenas_cronjob` - Cron jobs
  - `truenas_init_shutdown_script` - Init/shutdown scripts

- **Certificates & Security**
  - `truenas_certificate` - SSL certificates
  - `truenas_certificate_authority` - Certificate authorities
  - `truenas_acme_dns_authenticator` - ACME DNS authenticators

- **Storage**
  - `truenas_pool` - ZFS pool creation/management
  - `truenas_disk` - Disk management
  - `truenas_zvol` - ZFS volumes

### Medium Priority

- **Directory Services**
  - Active Directory integration
  - LDAP configuration
  - Kerberos settings

- **Monitoring & Alerts**
  - Alert services
  - Alert policies
  - Reporting configuration

- **Backup**
  - Cloud backup tasks
  - Rsync tasks

### Lower Priority

- VMware integration
- Boot environments
- System dataset configuration
- Tunable parameters
- UPS configuration

### Infrastructure Improvements

- Comprehensive acceptance tests
- Enhanced error handling and validation
- Retry logic for transient failures
- Pagination support for large datasets
- Bulk operations optimization
- Better state management
- Terraform Cloud/Enterprise support

## Notes

- This provider is designed specifically for TrueNAS Scale 24.04
- TrueNAS Scale 25.04+ uses JSON-RPC over WebSocket instead of REST API
- Always backup your TrueNAS configuration before making changes
- Test in a non-production environment first

