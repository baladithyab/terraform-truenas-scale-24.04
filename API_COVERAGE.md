# TrueNAS Scale 24.04 API Coverage

This document tracks the implementation status of TrueNAS Scale 24.04 REST API endpoints in the Terraform provider.

## Overview

The TrueNAS Scale 24.04 API contains **148,765 lines** of OpenAPI specification with hundreds of endpoints across 80+ categories.

**Current Coverage:** 14 resources, 2 data sources (~2.2% of 643 total endpoints)

### Quick Stats

- ✅ **Fully Implemented Categories**: Sharing (NFS, SMB), Users, Groups, Network (Interface, Routes), Snapshots
- 🟡 **Partially Implemented**: Storage (datasets, snapshots), VMs, iSCSI, Kubernetes (apps)
- 🔜 **High Priority Planned**: Replication, Cloud Sync, Services, Certificates
- 📊 **Total Resources**: 14 (started with 5)
- 🎯 **Import Support**: All 14 resources support import
- 📚 **Documentation**: 10 comprehensive guides including migration workflows

## Implementation Status

### ✅ Implemented (14 resources, 2 data sources)

#### Resources
- `truenas_dataset` - ZFS dataset management
- `truenas_nfs_share` - NFS share management
- `truenas_smb_share` - SMB/CIFS share management
- `truenas_user` - User account management
- `truenas_group` - Group management
- `truenas_vm` - Virtual machine management
- `truenas_iscsi_target` - iSCSI target management
- `truenas_iscsi_extent` - iSCSI extent (storage) management
- `truenas_iscsi_portal` - iSCSI portal (network listener) management
- `truenas_interface` - Network interface management
- `truenas_static_route` - Static route management
- `truenas_chart_release` - Kubernetes application deployment
- `truenas_snapshot` - ZFS snapshot management
- `truenas_periodic_snapshot_task` - Automated snapshot scheduling

#### Data Sources
- `truenas_dataset` - Query dataset information
- `truenas_pool` - Query pool information

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

### 🚧 Planned - High Priority

#### Virtual Machines (46 endpoints)
- `/vm` - VM CRUD operations ✅ IMPLEMENTED
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
- `truenas_vm` ✅ IMPLEMENTED
- `truenas_vm_device` 🔜 PLANNED

#### iSCSI (32 endpoints)
- `/iscsi/target` - iSCSI targets ✅ IMPLEMENTED
- `/iscsi/extent` - Storage extents ✅ IMPLEMENTED
- `/iscsi/portal` - Network portals ✅ IMPLEMENTED
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

#### Kubernetes/Apps (10+ endpoints)
- `/kubernetes` - K8s cluster management 🔜 PLANNED
- `/kubernetes/status` - Cluster status 🔜 PLANNED
- `/kubernetes/backup_chart_releases` - Backup apps 🔜 PLANNED
- `/kubernetes/restore_backup` - Restore apps 🔜 PLANNED
- `/chart/release` - Application management ✅ IMPLEMENTED
- `/chart/release/upgrade` - Upgrade apps (part of chart_release)
- `/chart/release/rollback` - Rollback apps (part of chart_release)
- `/chart/release/scale` - Scale apps (part of chart_release)
- `/catalog` - App catalogs 🔜 PLANNED

**Terraform Resources:**
- `truenas_kubernetes_config` 🔜 PLANNED
- `truenas_chart_release` ✅ IMPLEMENTED
- `truenas_catalog` 🔜 PLANNED

#### Network Configuration (21 endpoints)
- `/interface` - Network interfaces ✅ IMPLEMENTED
- `/interface/bridge_members_choices` - Bridge configuration (part of interface)
- `/interface/vlan_setup` - VLAN setup (part of interface)
- `/interface/lag_setup` - Link aggregation (part of interface)
- `/staticroute` - Static routes ✅ IMPLEMENTED
- `/network/configuration` - Network settings 🔜 PLANNED

**Terraform Resources:**
- `truenas_interface` ✅ IMPLEMENTED (supports PHYSICAL, VLAN, BRIDGE, LINK_AGGREGATION)
- `truenas_static_route` ✅ IMPLEMENTED
- `truenas_network_config` 🔜 PLANNED

#### Snapshots & Replication (12+ endpoints)
- `/zfs/snapshot` - ZFS snapshots ✅ IMPLEMENTED
- `/replication` - Replication tasks 🔜 PLANNED
- `/pool/dataset/destroy_snapshots` - Destroy snapshots (part of snapshot)
- `/pool/snapshottask` - Periodic snapshot tasks ✅ IMPLEMENTED

**Terraform Resources:**
- `truenas_snapshot` ✅ IMPLEMENTED
- `truenas_replication_task` 🔜 PLANNED
- `truenas_periodic_snapshot_task` ✅ IMPLEMENTED

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

## API Categories Summary

Total API categories: **80+**

| Category | Endpoints | Status |
|----------|-----------|--------|
| pool | 67 | Partial (dataset ✅, snapshots ✅) |
| vm | 46 | Partial (vm ✅, devices planned) |
| iscsi | 32 | Partial (target ✅, extent ✅, portal ✅) |
| interface | 21 | ✅ Implemented |
| certificate | 20+ | Planned |
| cloudsync | 15 | Planned |
| replication | 12 | Planned |
| kubernetes | 10 | Partial (chart_release ✅, cluster planned) |
| service | 10 | Planned |
| sharing | 9 | ✅ Implemented (NFS ✅, SMB ✅) |
| user | 8 | ✅ Implemented |
| group | 6 | ✅ Implemented |
| cronjob | 4 | Planned |
| network | 3 | ✅ Implemented (interface ✅, static_route ✅) |
| snapshot | 4 | ✅ Implemented (snapshot ✅, periodic_task ✅) |

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

