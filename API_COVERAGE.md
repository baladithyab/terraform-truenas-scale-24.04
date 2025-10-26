# TrueNAS Scale 24.04 API Coverage

This document tracks the implementation status of TrueNAS Scale 24.04 REST API endpoints in the Terraform provider.

## Overview

The TrueNAS Scale 24.04 API contains **148,765 lines** of OpenAPI specification with hundreds of endpoints across 80+ categories.

## Implementation Status

### ✅ Implemented (5 resources, 2 data sources)

#### Resources
- `truenas_dataset` - ZFS dataset management
- `truenas_nfs_share` - NFS share management  
- `truenas_smb_share` - SMB/CIFS share management
- `truenas_user` - User account management
- `truenas_group` - Group management

#### Data Sources
- `truenas_dataset` - Query dataset information
- `truenas_pool` - Query pool information

### 🚧 Planned - High Priority

#### Virtual Machines (46 endpoints)
- `/vm` - VM CRUD operations
- `/vm/device` - VM device management
- `/vm/id/{id}/start` - Start VM
- `/vm/id/{id}/stop` - Stop VM
- `/vm/id/{id}/restart` - Restart VM
- `/vm/id/{id}/suspend` - Suspend VM
- `/vm/id/{id}/resume` - Resume VM
- `/vm/id/{id}/clone` - Clone VM
- `/vm/get_console` - Get console access
- `/vm/device/disk_choices` - Available disks
- `/vm/device/nic_attach_choices` - Network options
- `/vm/device/passthrough_device` - PCI passthrough
- `/vm/device/usb_passthrough_device` - USB passthrough

**Terraform Resources:**
- `truenas_vm` - Virtual machine
- `truenas_vm_device` - VM devices (disk, NIC, USB, PCI)

#### iSCSI (32 endpoints)
- `/iscsi/target` - iSCSI targets
- `/iscsi/extent` - Storage extents
- `/iscsi/portal` - Network portals
- `/iscsi/initiator` - Initiator groups
- `/iscsi/auth` - Authentication
- `/iscsi/targetextent` - Target-extent associations
- `/iscsi/global` - Global iSCSI configuration

**Terraform Resources:**
- `truenas_iscsi_target`
- `truenas_iscsi_extent`
- `truenas_iscsi_portal`
- `truenas_iscsi_initiator`
- `truenas_iscsi_auth`
- `truenas_iscsi_targetextent`

#### Kubernetes/Apps (10+ endpoints)
- `/kubernetes` - K8s cluster management
- `/kubernetes/status` - Cluster status
- `/kubernetes/backup_chart_releases` - Backup apps
- `/kubernetes/restore_backup` - Restore apps
- `/chart/release` - Application management
- `/chart/release/upgrade` - Upgrade apps
- `/chart/release/rollback` - Rollback apps
- `/chart/release/scale` - Scale apps
- `/catalog` - App catalogs

**Terraform Resources:**
- `truenas_kubernetes_config`
- `truenas_chart_release`
- `truenas_catalog`

#### Network Configuration (21 endpoints)
- `/interface` - Network interfaces
- `/interface/bridge_members_choices` - Bridge configuration
- `/interface/vlan_setup` - VLAN setup
- `/interface/lag_setup` - Link aggregation
- `/staticroute` - Static routes
- `/network/configuration` - Network settings

**Terraform Resources:**
- `truenas_interface`
- `truenas_vlan`
- `truenas_bridge`
- `truenas_lagg`
- `truenas_static_route`
- `truenas_network_config`

#### Snapshots & Replication (12+ endpoints)
- `/zfs/snapshot` - ZFS snapshots
- `/replication` - Replication tasks
- `/pool/dataset/destroy_snapshots` - Destroy snapshots
- `/pool/snapshottask` - Periodic snapshot tasks

**Terraform Resources:**
- `truenas_snapshot`
- `truenas_replication_task`
- `truenas_periodic_snapshot_task`

#### Cloud Sync (15 endpoints)
- `/cloudsync` - Cloud sync tasks
- `/cloudsync/credentials` - Cloud credentials
- `/cloudsync/create_bucket` - Create cloud bucket
- `/cloudsync/id/{id}/sync` - Run sync
- `/cloudsync/id/{id}/abort` - Abort sync

**Terraform Resources:**
- `truenas_cloudsync_credentials`
- `truenas_cloudsync_task`

#### Services (10 endpoints)
- `/service` - Service management
- `/service/start` - Start service
- `/service/stop` - Stop service
- `/service/restart` - Restart service
- `/service/reload` - Reload service

**Terraform Resources:**
- `truenas_service`

#### Cron Jobs (4 endpoints)
- `/cronjob` - Cron job management
- `/cronjob/run` - Execute cron job

**Terraform Resources:**
- `truenas_cronjob`

#### Certificates (20+ endpoints)
- `/certificate` - SSL certificates
- `/certificateauthority` - Certificate authorities
- `/acme/dns/authenticator` - ACME DNS authenticators

**Terraform Resources:**
- `truenas_certificate`
- `truenas_certificate_authority`
- `truenas_acme_dns_authenticator`

### 🔜 Planned - Medium Priority

#### Storage Pool Management (67 endpoints)
- `/pool` - Pool CRUD
- `/pool/attach` - Attach vdev
- `/pool/detach` - Detach vdev
- `/pool/expand` - Expand pool
- `/pool/scrub` - Scrub pool

**Terraform Resources:**
- `truenas_pool`
- `truenas_pool_scrub_task`

#### Directory Services
- `/activedirectory` - Active Directory
- `/ldap` - LDAP configuration
- `/kerberos` - Kerberos settings
- `/idmap` - ID mapping

**Terraform Resources:**
- `truenas_activedirectory`
- `truenas_ldap`
- `truenas_kerberos_realm`
- `truenas_kerberos_keytab`

#### Alerts & Monitoring
- `/alert` - Alert management
- `/alertservice` - Alert services
- `/alertclasses` - Alert classes
- `/reporting` - Reporting configuration

**Terraform Resources:**
- `truenas_alert_service`
- `truenas_alert_policy`

#### Backup Tasks
- `/cloud_backup` - Cloud backup
- `/rsynctask` - Rsync tasks

**Terraform Resources:**
- `truenas_cloud_backup`
- `truenas_rsync_task`

### 📋 Planned - Lower Priority

#### System Configuration
- `/system/general` - General settings
- `/system/advanced` - Advanced settings
- `/system/ntpserver` - NTP servers
- `/tunable` - System tunables
- `/bootenv` - Boot environments

#### Hardware
- `/disk` - Disk management
- `/smart` - SMART monitoring
- `/enclosure` - Enclosure management

#### Other Services
- `/ftp` - FTP service
- `/ssh` - SSH service
- `/snmp` - SNMP service
- `/ups` - UPS configuration
- `/vmware` - VMware integration

## API Categories Summary

Total API categories: **80+**

| Category | Endpoints | Status |
|----------|-----------|--------|
| pool | 67 | Partial (dataset only) |
| vm | 46 | Planned |
| iscsi | 32 | Planned |
| interface | 21 | Planned |
| certificate | 20+ | Planned |
| cloudsync | 15 | Planned |
| replication | 12 | Planned |
| kubernetes | 10 | Planned |
| service | 10 | Planned |
| sharing | 9 | ✅ Implemented |
| user | 8 | ✅ Implemented |
| group | 6 | ✅ Implemented |
| cronjob | 4 | Planned |
| network | 3 | Planned |

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

