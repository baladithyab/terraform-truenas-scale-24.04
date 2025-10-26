# TrueNAS Scale 24.04 API Coverage

This document tracks the implementation status of TrueNAS Scale 24.04 REST API endpoints in the Terraform provider.

## Overview

The TrueNAS Scale 24.04 API contains **148,765 lines** of OpenAPI specification with hundreds of endpoints across 80+ categories.

**Current Coverage:** 14 resources, 2 data sources (~2.2% of 643 total endpoints)

## ⚠️ Important Version Information

**Latest Code Version**: `main` branch (all features implemented)
**Published Registry Version**: `v0.1.0` (may have missing features)
**Upcoming Release**: `v0.2.0` (includes all features, ETA: 24-48 hours)

### Known Issues with v0.1.0

If you're using `v0.1.0` from the Terraform Registry, you may encounter:
- ❌ Data sources not working (`data.truenas_pool`, `data.truenas_dataset`)
- ❌ Some import functionality missing
- ❌ Schema validation errors for snapshot resources

**Solution**:
1. **Immediate**: Build from source (see [GAPS_ANALYSIS_RESPONSE.md](GAPS_ANALYSIS_RESPONSE.md))
2. **Recommended**: Wait for v0.2.0 release (24-48 hours)

All features listed below ARE implemented in the codebase and will be available in v0.2.0.

### Quick Stats

**Implementation Summary:**
- ✅ **Fully Implemented Categories (6)**:
  - Storage: Datasets ✅, Snapshots ✅, Periodic Snapshot Tasks ✅
  - Sharing: NFS ✅, SMB ✅
  - Users & Groups: Users ✅, Groups ✅
  - Network: Interfaces ✅, Static Routes ✅

- 🟡 **Partially Implemented Categories (3)**:
  - Virtual Machines: Basic VM ✅ (devices, lifecycle operations planned)
  - iSCSI: Target ✅, Extent ✅, Portal ✅ (initiator, auth, associations planned)
  - Kubernetes: Chart Releases ✅ (cluster config, catalogs planned)

- 🔜 **High Priority Planned (5)**: Replication, Cloud Sync, Services, Certificates, Cron Jobs

**Metrics:**
- 📊 **Total Resources**: 14 (started with 5, added 9)
- 🎯 **Import Support**: 100% (all 14 resources)
- 📚 **Documentation**: 10 comprehensive guides
- 🚀 **Special Features**: Kubernetes migration to external clusters
- 📈 **API Coverage**: ~2.2% (14 of 643 endpoints)

## Implementation Status

### ✅ Fully Implemented (14 resources, 2 data sources)

All resources include:
- ✅ Full CRUD operations (Create, Read, Update, Delete)
- ✅ Import support
- ✅ Comprehensive examples
- ✅ Documentation

#### Storage & File Sharing (3 resources)
1. **`truenas_dataset`** - ZFS dataset management
   - API: `/pool/dataset`
   - Features: Compression, quotas, reservations, ZFS properties
   - Import: By dataset name (e.g., `tank/mydata`)

2. **`truenas_nfs_share`** - NFS share management
   - API: `/sharing/nfs`
   - Features: Network ACLs, security, user mapping
   - Import: By share ID

3. **`truenas_smb_share`** - SMB/CIFS share management
   - API: `/sharing/smb`
   - Features: Guest access, recycle bin, shadow copies
   - Import: By share ID

#### User Management (2 resources)
4. **`truenas_user`** - User account management
   - API: `/user`
   - Features: Passwords, SSH keys, home directories, sudo
   - Import: By user ID

5. **`truenas_group`** - Group management
   - API: `/group`
   - Features: User assignments, sudo, SMB settings
   - Import: By group ID

#### Virtual Machines (1 resource)
6. **`truenas_vm`** - Virtual machine management
   - API: `/vm`
   - Features: CPU/memory config, bootloader, autostart
   - Import: By VM name

#### iSCSI (3 resources)
7. **`truenas_iscsi_target`** - iSCSI target management
   - API: `/iscsi/target`
   - Features: IQN-based targets, portal associations
   - Import: By target ID

8. **`truenas_iscsi_extent`** - iSCSI extent management
   - API: `/iscsi/extent`
   - Features: FILE/DISK types, block sizes, read-only
   - Import: By extent ID

9. **`truenas_iscsi_portal`** - iSCSI portal management
   - API: `/iscsi/portal`
   - Features: Listen addresses, CHAP auth
   - Import: By portal ID

#### Network (2 resources)
10. **`truenas_interface`** - Network interface management
    - API: `/interface`
    - Features: PHYSICAL, VLAN, BRIDGE, LAG types
    - Import: By interface name

11. **`truenas_static_route`** - Static route management
    - API: `/staticroute`
    - Features: CIDR destinations, gateway IPs
    - Import: By route ID

#### Kubernetes/Apps (1 resource) ✨
12. **`truenas_chart_release`** - Kubernetes application deployment
    - API: `/chart/release`
    - Features: Catalog apps, JSON values, version management
    - **Special**: Migration support to external K8s clusters
    - Import: By release name

#### Snapshots (2 resources) ✨
13. **`truenas_snapshot`** - ZFS snapshot management
    - API: `/zfs/snapshot`
    - Features: Recursive snapshots, VMware sync
    - Import: By `dataset@snapshotname`

14. **`truenas_periodic_snapshot_task`** - Automated snapshot scheduling
    - API: `/pool/snapshottask`
    - Features: Cron scheduling, retention policies, exclusions
    - Import: By task ID

#### Data Sources (2)
- **`truenas_dataset`** - Query dataset information
  - API: `/pool/dataset/id/{id}`

- **`truenas_pool`** - Query pool information
  - API: `/pool/id/{id}`

### 🎯 Special Features

#### Kubernetes Migration Support ✨ NEW
The provider includes comprehensive Kubernetes app migration capabilities:

- **Export to External K8s**: Migrate TrueNAS apps to EKS, GKE, AKS, or any Kubernetes cluster
- **PVC Data Migration**: Automated tools for migrating persistent volume data
- **Backup & Restore**: Snapshot-based backup with full data preservation
- **Migration Automation**: `export-apps.sh` script generates migration manifests and scripts
- **Complete Examples**: Production-ready examples with Plex, Nextcloud, Sonarr, Radarr, Home Assistant

**Documentation:**
- [KUBERNETES_MIGRATION.md](KUBERNETES_MIGRATION.md) - Complete migration guide (5 workflows)
- [examples/complete-kubernetes/](examples/complete-kubernetes/) - Production examples
- [IMPORT_GUIDE.md](IMPORT_GUIDE.md) - Import guide for all resources

**Use Cases:**
- Migrate from TrueNAS K8s to cloud Kubernetes
- Backup apps with data before major changes
- Replicate apps across multiple TrueNAS instances
- Version control all app configurations

### � Partially Implemented - Additional Features Available

#### Virtual Machines (46 endpoints) - Basic VM ✅, Advanced Features 🔜
**Implemented:**
- `/vm` - VM CRUD operations ✅ IMPLEMENTED

**Planned:**
- `/vm/device` - VM device management 🔜 PLANNED
- `/vm/id/{id}/start` - Start VM 🔜 PLANNED
- `/vm/id/{id}/stop` - Stop VM 🔜 PLANNED
- `/vm/id/{id}/restart` - Restart VM 🔜 PLANNED
- `/vm/id/{id}/suspend` - Suspend VM 🔜 PLANNED
- `/vm/id/{id}/resume` - Resume VM 🔜 PLANNED
- `/vm/id/{id}/clone` - Clone VM 🔜 PLANNED
- `/vm/get_console` - Get console access 🔜 PLANNED
- `/vm/device/disk_choices` - Available disks 🔜 PLANNED
- `/vm/device/nic_attach_choices` - Network options 🔜 PLANNED
- `/vm/device/passthrough_device` - PCI passthrough 🔜 PLANNED
- `/vm/device/usb_passthrough_device` - USB passthrough 🔜 PLANNED

**Terraform Resources:**
- `truenas_vm` ✅ IMPLEMENTED (basic VM management)
- `truenas_vm_device` 🔜 PLANNED (advanced device management)

#### iSCSI (32 endpoints) - Core Features ✅, Advanced Features 🔜
**Implemented:**
- `/iscsi/target` - iSCSI targets ✅ IMPLEMENTED
- `/iscsi/extent` - Storage extents ✅ IMPLEMENTED
- `/iscsi/portal` - Network portals ✅ IMPLEMENTED

**Planned:**
- `/iscsi/initiator` - Initiator groups 🔜 PLANNED
- `/iscsi/auth` - Authentication 🔜 PLANNED
- `/iscsi/targetextent` - Target-extent associations 🔜 PLANNED
- `/iscsi/global` - Global iSCSI configuration 🔜 PLANNED

**Terraform Resources:**
- `truenas_iscsi_target` ✅ IMPLEMENTED
- `truenas_iscsi_extent` ✅ IMPLEMENTED
- `truenas_iscsi_portal` ✅ IMPLEMENTED
- `truenas_iscsi_initiator` 🔜 PLANNED
- `truenas_iscsi_auth` 🔜 PLANNED
- `truenas_iscsi_targetextent` 🔜 PLANNED

#### Kubernetes/Apps (10+ endpoints) - Apps ✅, Cluster Management 🔜
**Implemented:**
- `/chart/release` - Application management ✅ IMPLEMENTED
- `/chart/release/upgrade` - Upgrade apps ✅ (part of chart_release)
- `/chart/release/rollback` - Rollback apps ✅ (part of chart_release)
- `/chart/release/scale` - Scale apps ✅ (part of chart_release)

**Planned:**
- `/kubernetes` - K8s cluster management 🔜 PLANNED
- `/kubernetes/status` - Cluster status 🔜 PLANNED
- `/kubernetes/backup_chart_releases` - Backup apps 🔜 PLANNED
- `/kubernetes/restore_backup` - Restore apps 🔜 PLANNED
- `/catalog` - App catalogs 🔜 PLANNED

**Terraform Resources:**
- `truenas_chart_release` ✅ IMPLEMENTED (full app lifecycle)
- `truenas_kubernetes_config` 🔜 PLANNED (cluster configuration)
- `truenas_catalog` 🔜 PLANNED (catalog management)

### 🚧 Planned - High Priority

#### Replication (12+ endpoints)
- `/replication` - Replication tasks 🔜 PLANNED
- `/replication/id/{id}/run` - Run replication 🔜 PLANNED
- `/replication/count_eligible_manual_snapshots` - Count snapshots 🔜 PLANNED

**Terraform Resources:**
- `truenas_replication_task` 🔜 PLANNED

#### Cloud Sync (15 endpoints)
- `/cloudsync` - Cloud sync tasks 🔜 PLANNED
- `/cloudsync/credentials` - Cloud credentials 🔜 PLANNED
- `/cloudsync/create_bucket` - Create cloud bucket 🔜 PLANNED
- `/cloudsync/id/{id}/sync` - Run sync 🔜 PLANNED
- `/cloudsync/id/{id}/abort` - Abort sync 🔜 PLANNED

**Terraform Resources:**
- `truenas_cloudsync_credentials` 🔜 PLANNED
- `truenas_cloudsync_task` 🔜 PLANNED

#### Services (10 endpoints)
- `/service` - Service management 🔜 PLANNED
- `/service/start` - Start service 🔜 PLANNED
- `/service/stop` - Stop service 🔜 PLANNED
- `/service/restart` - Restart service 🔜 PLANNED
- `/service/reload` - Reload service 🔜 PLANNED

**Terraform Resources:**
- `truenas_service` 🔜 PLANNED

#### Cron Jobs (4 endpoints)
- `/cronjob` - Cron job management 🔜 PLANNED
- `/cronjob/run` - Execute cron job 🔜 PLANNED

**Terraform Resources:**
- `truenas_cronjob` 🔜 PLANNED

#### Certificates (20+ endpoints)
- `/certificate` - SSL certificates 🔜 PLANNED
- `/certificateauthority` - Certificate authorities 🔜 PLANNED
- `/acme/dns/authenticator` - ACME DNS authenticators 🔜 PLANNED

**Terraform Resources:**
- `truenas_certificate` 🔜 PLANNED
- `truenas_certificate_authority` 🔜 PLANNED
- `truenas_acme_dns_authenticator` 🔜 PLANNED

### 🔜 Planned - Medium Priority

#### Storage Pool Management (67 endpoints)
- `/pool` - Pool CRUD 🔜 PLANNED
- `/pool/attach` - Attach vdev 🔜 PLANNED
- `/pool/detach` - Detach vdev 🔜 PLANNED
- `/pool/expand` - Expand pool 🔜 PLANNED
- `/pool/scrub` - Scrub pool 🔜 PLANNED

**Terraform Resources:**
- `truenas_pool` 🔜 PLANNED
- `truenas_pool_scrub_task` 🔜 PLANNED

#### Directory Services
- `/activedirectory` - Active Directory 🔜 PLANNED
- `/ldap` - LDAP configuration 🔜 PLANNED
- `/kerberos` - Kerberos settings 🔜 PLANNED
- `/idmap` - ID mapping 🔜 PLANNED

**Terraform Resources:**
- `truenas_activedirectory` 🔜 PLANNED
- `truenas_ldap` 🔜 PLANNED
- `truenas_kerberos_realm` 🔜 PLANNED
- `truenas_kerberos_keytab` 🔜 PLANNED

#### Alerts & Monitoring
- `/alert` - Alert management 🔜 PLANNED
- `/alertservice` - Alert services 🔜 PLANNED
- `/alertclasses` - Alert classes 🔜 PLANNED
- `/reporting` - Reporting configuration 🔜 PLANNED

**Terraform Resources:**
- `truenas_alert_service` 🔜 PLANNED
- `truenas_alert_policy` 🔜 PLANNED

#### Backup Tasks
- `/cloud_backup` - Cloud backup 🔜 PLANNED
- `/rsynctask` - Rsync tasks 🔜 PLANNED

**Terraform Resources:**
- `truenas_cloud_backup` 🔜 PLANNED
- `truenas_rsync_task` 🔜 PLANNED

### 📋 Planned - Lower Priority

#### System Configuration
- `/system/general` - General settings 🔜 PLANNED
- `/system/advanced` - Advanced settings 🔜 PLANNED
- `/system/ntpserver` - NTP servers 🔜 PLANNED
- `/tunable` - System tunables 🔜 PLANNED
- `/bootenv` - Boot environments 🔜 PLANNED

**Terraform Resources:**
- `truenas_system_general` 🔜 PLANNED
- `truenas_system_advanced` 🔜 PLANNED
- `truenas_ntp_server` 🔜 PLANNED
- `truenas_tunable` 🔜 PLANNED
- `truenas_boot_environment` 🔜 PLANNED

#### Hardware
- `/disk` - Disk management 🔜 PLANNED
- `/smart` - SMART monitoring 🔜 PLANNED
- `/enclosure` - Enclosure management 🔜 PLANNED

**Terraform Resources:**
- `truenas_disk` 🔜 PLANNED
- `truenas_smart_test` 🔜 PLANNED

#### Other Services
- `/ftp` - FTP service 🔜 PLANNED
- `/ssh` - SSH service 🔜 PLANNED
- `/snmp` - SNMP service 🔜 PLANNED
- `/ups` - UPS configuration 🔜 PLANNED
- `/vmware` - VMware integration 🔜 PLANNED

**Terraform Resources:**
- `truenas_ftp_config` 🔜 PLANNED
- `truenas_ssh_config` 🔜 PLANNED
- `truenas_snmp_config` 🔜 PLANNED
- `truenas_ups_config` 🔜 PLANNED

## Implementation Roadmap

### Phase 1: Foundation ✅ COMPLETE
**Goal**: Core infrastructure management
- ✅ Datasets (storage)
- ✅ NFS/SMB shares (file sharing)
- ✅ Users & Groups (access control)
- ✅ Basic documentation

### Phase 2: Advanced Infrastructure ✅ COMPLETE
**Goal**: Virtualization and block storage
- ✅ Virtual Machines
- ✅ iSCSI (target, extent, portal)
- ✅ Network (interfaces, routes)
- ✅ Complete examples

### Phase 3: Kubernetes & Snapshots ✅ COMPLETE
**Goal**: Application management and data protection
- ✅ Kubernetes chart releases
- ✅ ZFS snapshots
- ✅ Periodic snapshot tasks
- ✅ **Migration capabilities** (TrueNAS K8s → External K8s)
- ✅ Migration automation scripts
- ✅ Complete documentation (10 guides)

### Phase 4: Data Management 🔜 NEXT
**Goal**: Replication and cloud integration
- 🔜 Replication tasks
- 🔜 Cloud sync tasks
- 🔜 Cloud credentials
- 🔜 Rsync tasks

### Phase 5: System Management 🔜 PLANNED
**Goal**: Services and monitoring
- 🔜 Service management
- 🔜 Cron jobs
- 🔜 Certificates
- 🔜 Alert services

### Phase 6: Advanced Features 🔜 FUTURE
**Goal**: Enterprise features
- 🔜 Active Directory integration
- 🔜 LDAP configuration
- 🔜 Advanced VM features (devices, lifecycle)
- 🔜 Advanced iSCSI (initiators, auth)

## API Categories Summary

Total API categories: **80+**
Total endpoints: **643**
Implemented: **14 resources** (~2.2% coverage)

| Category | Endpoints | Implemented | Status | Priority |
|----------|-----------|-------------|--------|----------|
| **sharing** | 9 | 2 | ✅ Complete (NFS ✅, SMB ✅) | ✅ Done |
| **user** | 8 | 1 | ✅ Complete | ✅ Done |
| **group** | 6 | 1 | ✅ Complete | ✅ Done |
| **interface** | 21 | 1 | ✅ Complete | ✅ Done |
| **network** | 3 | 1 | ✅ Complete (static_route ✅) | ✅ Done |
| **snapshot** | 4 | 2 | ✅ Complete (snapshot ✅, periodic_task ✅) | ✅ Done |
| **pool** | 67 | 1 | 🟡 Partial (dataset ✅, snapshots ✅) | Medium |
| **vm** | 46 | 1 | 🟡 Partial (vm ✅, devices planned) | Medium |
| **iscsi** | 32 | 3 | 🟡 Partial (target ✅, extent ✅, portal ✅) | Medium |
| **kubernetes** | 10 | 1 | 🟡 Partial (chart_release ✅, cluster planned) | Medium |
| **replication** | 12 | 0 | 🔜 Planned | High |
| **cloudsync** | 15 | 0 | 🔜 Planned | High |
| **service** | 10 | 0 | 🔜 Planned | High |
| **certificate** | 20+ | 0 | 🔜 Planned | High |
| **cronjob** | 4 | 0 | 🔜 Planned | High |
| **alertservice** | 8 | 0 | 🔜 Planned | Medium |
| **activedirectory** | 6 | 0 | 🔜 Planned | Medium |
| **ldap** | 5 | 0 | 🔜 Planned | Medium |
| **system** | 30+ | 0 | 🔜 Planned | Low |
| **disk** | 15 | 0 | 🔜 Planned | Low |
| **Other** | 350+ | 0 | 🔜 Future | Low |

**Legend:**
- ✅ Complete: All core features implemented
- 🟡 Partial: Basic features implemented, advanced features planned
- 🔜 Planned: Not yet implemented

## Contributing

To add support for a new resource:

1. Review the OpenAPI spec: `http://your-truenas-ip/api/v2.0`
2. Create resource file in `internal/provider/`
3. Implement CRUD operations
4. Add import support
5. Register in `provider.go`
6. Create examples
7. Update this document

## Testing New Resources

When implementing new resources, test against your TrueNAS server:

```bash
# Get the full OpenAPI spec
curl http://10.0.0.213:81/api/v2.0 > openapi.json

# Search for specific endpoints
cat openapi.json | jq '.paths | keys[] | select(contains("vm"))'

# View endpoint details
cat openapi.json | jq '.paths["/vm"]'
```

## Notes

- The OpenAPI spec is 148,765 lines - this is a massive API surface
- Priority is given to most commonly used features
- Some endpoints are for internal use only
- Not all endpoints make sense as Terraform resources
- Some resources may be better as data sources

