# TrueNAS Scale 24.04 Terraform Provider - Project Summary

## Overview

This project is a complete, production-ready Terraform provider for TrueNAS Scale 24.04. It enables infrastructure-as-code management of TrueNAS resources using Terraform.

## Why TrueNAS Scale 24.04?

TrueNAS Scale 24.04 uses a REST API, while version 25.04 transitions to JSON-RPC over WebSocket. This provider is specifically designed for the REST API available in version 24.04.

## Project Structure

```
terraform-truenas-scale-24.04/
├── main.go                                    # Provider entry point
├── go.mod                                     # Go module definition
├── go.sum                                     # Go dependencies checksums
├── Makefile                                   # Build automation
├── .goreleaser.yml                            # Release configuration
├── .gitignore                                 # Git ignore rules
│
├── Documentation/
│   ├── README.md                              # Main documentation
│   ├── QUICKSTART.md                          # Quick start guide
│   ├── PROJECT_SUMMARY.md                     # This file
│   ├── TESTING.md                             # Testing guide
│   ├── CONTRIBUTING.md                        # Contributor guide
│   ├── API_COVERAGE.md                        # API implementation status
│   ├── API_ENDPOINTS.md                       # Complete API reference
│   ├── IMPORT_GUIDE.md                        # Import guide for all resources
│   ├── KUBERNETES_MIGRATION.md                # K8s migration workflows ✨ NEW
│   └── RELEASE_NOTES.md                       # Release notes
│
├── internal/
│   ├── provider/
│   │   ├── provider.go                        # Provider implementation
│   │   ├── resource_dataset.go                # ZFS dataset resource
│   │   ├── resource_nfs_share.go              # NFS share resource
│   │   ├── resource_smb_share.go              # SMB share resource
│   │   ├── resource_user.go                   # User resource
│   │   ├── resource_group.go                  # Group resource
│   │   ├── resource_vm.go                     # Virtual machine resource
│   │   ├── resource_iscsi_target.go           # iSCSI target resource
│   │   ├── resource_iscsi_extent.go           # iSCSI extent resource
│   │   ├── resource_iscsi_portal.go           # iSCSI portal resource
│   │   ├── resource_interface.go              # Network interface resource
│   │   ├── resource_static_route.go           # Static route resource
│   │   ├── resource_chart_release.go          # Kubernetes app resource ✨
│   │   ├── resource_snapshot.go               # ZFS snapshot resource ✨
│   │   ├── resource_periodic_snapshot_task.go # Snapshot task resource ✨
│   │   ├── datasource_dataset.go              # Dataset data source
│   │   └── datasource_pool.go                 # Pool data source
│   │
│   └── truenas/
│       └── client.go                          # TrueNAS API client
│
└── examples/
    ├── provider/
    │   └── provider.tf                        # Provider configuration
    ├── resources/                             # Individual resource examples (14)
    │   ├── truenas_dataset/
    │   ├── truenas_nfs_share/
    │   ├── truenas_smb_share/
    │   ├── truenas_user/
    │   ├── truenas_group/
    │   ├── truenas_vm/
    │   ├── truenas_iscsi_target/
    │   ├── truenas_iscsi_extent/
    │   ├── truenas_iscsi_portal/
    │   ├── truenas_interface/
    │   ├── truenas_static_route/
    │   ├── truenas_chart_release/             # K8s app examples ✨
    │   ├── truenas_snapshot/                  # Snapshot examples ✨
    │   └── truenas_periodic_snapshot_task/    # Snapshot task examples ✨
    ├── complete/                              # Complete infrastructure
    ├── complete-iscsi/                        # Complete iSCSI setup
    ├── complete-network/                      # Complete network setup
    └── complete-kubernetes/                   # Complete K8s stack ✨ NEW
        ├── main.tf                            # 5 production apps
        ├── variables.tf                       # Configuration
        ├── README.md                          # Deployment guide
        └── export-apps.sh                     # Migration automation ✨
```

## Implemented Features

### Resources (14 Total - All with Full CRUD + Import)

#### Storage & File Sharing (3)

1. **truenas_dataset** - ZFS Dataset Management
   - Create, read, update, delete datasets
   - Configure compression, quotas, reservations
   - Set ZFS properties (atime, dedup, sync, etc.)
   - Import existing datasets

2. **truenas_nfs_share** - NFS Share Management
   - Create and manage NFS exports
   - Configure network/host access controls
   - Set security mechanisms
   - Map users (maproot, mapall)
   - Import existing shares

3. **truenas_smb_share** - SMB/CIFS Share Management
   - Create and manage SMB shares
   - Configure browsability and guest access
   - Enable recycle bin and shadow copies
   - Host allow/deny lists
   - Import existing shares

#### User Management (2)

4. **truenas_user** - User Account Management
   - Create and manage user accounts
   - Set passwords, SSH keys
   - Configure home directories and shells
   - Enable sudo and SMB authentication
   - Import existing users

5. **truenas_group** - Group Management
   - Create and manage groups
   - Assign users to groups
   - Configure sudo and SMB settings
   - Import existing groups

#### Virtual Machines (1)

6. **truenas_vm** - Virtual Machine Management
   - Create and manage VMs
   - Configure CPU, memory, cores, threads
   - Set bootloader (UEFI, GRUB)
   - Enable autostart and memory ballooning
   - Import existing VMs

#### iSCSI (3)

7. **truenas_iscsi_target** - iSCSI Target Management
   - Create IQN-based targets
   - Portal group associations
   - Network-based authentication
   - Import existing targets

8. **truenas_iscsi_extent** - iSCSI Extent Management
   - FILE and DISK extent types
   - Configurable block sizes
   - Read-only and Xen compatibility
   - Import existing extents

9. **truenas_iscsi_portal** - iSCSI Portal Management
   - Multiple listen addresses
   - CHAP authentication
   - Discovery auth methods
   - Import existing portals

#### Network (2)

10. **truenas_interface** - Network Interface Management
    - PHYSICAL, VLAN, BRIDGE, LINK_AGGREGATION types
    - IP address configuration
    - MTU and VLAN tag settings
    - Import existing interfaces

11. **truenas_static_route** - Static Route Management
    - Destination CIDR configuration
    - Gateway IP settings
    - Route descriptions
    - Import existing routes

#### Kubernetes/Apps (1) ✨ NEW

12. **truenas_chart_release** - Kubernetes Application Deployment
    - Deploy apps from TrueNAS catalogs
    - JSON-based values configuration
    - Version management
    - Status tracking
    - **Migration support** to external K8s clusters
    - Import existing chart releases

#### Snapshots (2) ✨ NEW

13. **truenas_snapshot** - ZFS Snapshot Management
    - On-demand snapshot creation
    - Recursive snapshots
    - VMware sync integration
    - Immutable (no updates)
    - Import via dataset@snapshotname format

14. **truenas_periodic_snapshot_task** - Automated Snapshot Scheduling
    - Cron-based scheduling
    - Configurable retention (HOUR, DAY, WEEK, MONTH, YEAR)
    - Recursive snapshots with exclusions
    - Naming schema templates
    - Enable/disable tasks
    - Import existing tasks

### Data Sources

1. **truenas_dataset** - Query dataset information
   - Get dataset properties
   - Check available/used space
   - Retrieve compression settings

2. **truenas_pool** - Query pool information
   - Get pool status and health
   - Check available space
   - Retrieve pool size

## Technical Implementation

### Provider Configuration

The provider accepts two configuration parameters:
- `base_url` - TrueNAS server URL (e.g., http://10.0.0.213:81)
- `api_key` - TrueNAS API authentication key

Both can be set via environment variables:
- `TRUENAS_BASE_URL`
- `TRUENAS_API_KEY`

### API Client

The `internal/truenas/client.go` implements a robust HTTP client that:
- Handles authentication via Bearer token
- Supports all HTTP methods (GET, POST, PUT, PATCH, DELETE)
- Provides proper error handling
- Uses configurable timeouts
- Formats requests/responses as JSON

### Resource Implementation

All resources follow Terraform best practices:
- Implement full CRUD operations
- Support state import
- Use proper plan modifiers
- Handle computed vs. required attributes
- Provide comprehensive error messages
- Support partial updates

## Build and Installation

### Prerequisites
- Go 1.21 or later
- Terraform 1.0 or later
- TrueNAS Scale 24.04

### Building
```bash
make build
```

### Installing Locally
```bash
make install
```

This installs to: `~/.terraform.d/plugins/terraform-providers/truenas/1.0.0/linux_amd64/`

### Testing
```bash
make test
```

## Usage Example

```hcl
terraform {
  required_providers {
    truenas = {
      source  = "terraform-providers/truenas"
      version = "1.0.0"
    }
  }
}

provider "truenas" {
  base_url = "http://10.0.0.213:81"
  api_key  = var.truenas_api_key
}

resource "truenas_dataset" "mydata" {
  name        = "tank/mydata"
  type        = "FILESYSTEM"
  compression = "LZ4"
  quota       = 107374182400  # 100GB
}

resource "truenas_nfs_share" "myshare" {
  path     = "/mnt/${truenas_dataset.mydata.name}"
  networks = ["192.168.1.0/24"]
  enabled  = true
}
```

## Import Support

All 14 resources support import! See [IMPORT_GUIDE.md](IMPORT_GUIDE.md) for complete documentation.

```bash
# Storage & Sharing
terraform import truenas_dataset.mydata tank/mydata
terraform import truenas_nfs_share.myshare 1
terraform import truenas_smb_share.myshare 1

# Users & Groups
terraform import truenas_user.john 1000
terraform import truenas_group.developers 1000

# Virtual Machines
terraform import truenas_vm.ubuntu ubuntu-vm

# iSCSI
terraform import truenas_iscsi_target.target1 1
terraform import truenas_iscsi_extent.extent1 1
terraform import truenas_iscsi_portal.portal1 1

# Network
terraform import truenas_interface.eth0 eth0
terraform import truenas_static_route.route1 1

# Kubernetes
terraform import truenas_chart_release.plex plex

# Snapshots
terraform import truenas_snapshot.backup tank/mydata@backup-2024-01-15
terraform import truenas_periodic_snapshot_task.hourly 1
```

## Kubernetes Migration Capabilities ✨ NEW

The provider includes comprehensive Kubernetes app migration support:

### Features
- **Export to External K8s**: Migrate TrueNAS apps to EKS, GKE, AKS, or any Kubernetes cluster
- **PVC Data Migration**: Automated tools for migrating persistent volume data
- **Backup & Restore**: Snapshot-based backup with full data preservation
- **Migration Automation**: `export-apps.sh` script generates migration manifests and scripts
- **Complete Examples**: Production-ready examples with Plex, Nextcloud, Sonarr, Radarr, Home Assistant

### Documentation
- [KUBERNETES_MIGRATION.md](KUBERNETES_MIGRATION.md) - Complete migration guide with 5 workflows
- [examples/complete-kubernetes/](examples/complete-kubernetes/) - Production examples
- [examples/complete-kubernetes/export-apps.sh](examples/complete-kubernetes/export-apps.sh) - Migration automation

### Use Cases
- Migrate from TrueNAS K8s to cloud Kubernetes
- Backup apps with data before major changes
- Replicate apps across multiple TrueNAS instances
- Version control all app configurations
- Disaster recovery with automated snapshots

## API Coverage

The provider currently implements 14 resources covering:
- ✅ `/pool/dataset` - Dataset management
- ✅ `/sharing/nfs` - NFS shares
- ✅ `/sharing/smb` - SMB shares
- ✅ `/user` - User accounts
- ✅ `/group` - Groups
- ✅ `/vm` - Virtual machines
- ✅ `/iscsi/target` - iSCSI targets
- ✅ `/iscsi/extent` - iSCSI extents
- ✅ `/iscsi/portal` - iSCSI portals
- ✅ `/interface` - Network interfaces
- ✅ `/staticroute` - Static routes
- ✅ `/chart/release` - Kubernetes applications
- ✅ `/zfs/snapshot` - ZFS snapshots
- ✅ `/pool/snapshottask` - Periodic snapshot tasks

**Coverage**: ~2.2% of 643 total endpoints (14 implemented)

Additional APIs available for future implementation:
- Replication tasks
- Cloud sync tasks
- Services management
- Certificates
- Cron jobs
- Alert services
- And many more... (see [API_COVERAGE.md](API_COVERAGE.md))

## Future Enhancements

High-priority additions:
1. **Replication tasks** - ZFS replication to remote systems
2. **Cloud sync** - Sync to S3, Azure, Google Cloud
3. **Services management** - Start/stop/configure TrueNAS services
4. **Certificates** - SSL certificate management
5. **Cron jobs** - Scheduled task management

Medium-priority additions:
6. More data sources (VMs, shares, snapshots)
7. Acceptance tests
8. Enhanced validation
9. Better error messages
10. Retry logic for transient failures
11. Pagination support for list operations

## Testing Recommendations

Before using in production:
1. Test in a non-production TrueNAS environment
2. Verify API key has appropriate permissions
3. Test import functionality with existing resources
4. Validate backup/restore procedures
5. Test state management and locking
6. Verify idempotency of operations

## Security Considerations

1. **API Key Storage**: Never commit API keys to version control
2. **Use Environment Variables**: Store credentials in environment variables
3. **Terraform State**: State files contain sensitive data - secure them appropriately
4. **Network Security**: Use HTTPS when possible (TrueNAS supports SSL)
5. **Access Control**: Limit API key permissions to minimum required

## Known Limitations

1. Provider is specific to TrueNAS Scale 24.04 (REST API)
2. Not compatible with TrueNAS Scale 25.04+ (uses WebSocket)
3. Some advanced ZFS properties may not be exposed
4. Bulk operations are not optimized
5. No built-in retry logic for transient failures

## Contributing

To add new resources:
1. Create resource file in `internal/provider/`
2. Implement CRUD operations
3. Add import support
4. Register in `provider.go`
5. Create examples in `examples/resources/`
6. Update documentation

## License

Mozilla Public License 2.0

## Support

- GitHub Issues: For bug reports and feature requests
- TrueNAS API Docs: `http://your-truenas-ip/api/docs/`
- Terraform Provider Development: https://developer.hashicorp.com/terraform/plugin

## Conclusion

This provider offers a comprehensive, production-ready solution for managing TrueNAS Scale 24.04 infrastructure with Terraform. With 14 resources covering storage, networking, virtualization, iSCSI, Kubernetes apps, and snapshots, plus complete import support and Kubernetes migration capabilities, it's suitable for:

- **New Deployments**: Define entire infrastructure as code
- **Existing Infrastructure**: Import and manage existing resources
- **Kubernetes Migration**: Migrate apps to external K8s clusters with data
- **Disaster Recovery**: Automated snapshot scheduling and backup
- **Multi-Environment**: Manage dev/staging/prod TrueNAS instances
- **Version Control**: Track all configurations in Git

**Key Highlights:**
- ✅ 14 resources with full CRUD operations
- ✅ 100% import support across all resources
- ✅ Kubernetes app migration to external clusters
- ✅ 10 comprehensive documentation guides
- ✅ Production-ready examples
- ✅ Automated migration tooling
- ✅ Multi-tier snapshot strategies

