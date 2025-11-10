---
page_title: "truenas_nfs_share Resource - terraform-provider-truenas"
subcategory: "Storage & File Sharing"
description: |-
  Manages an NFS share on TrueNAS Scale.
---

# truenas_nfs_share (Resource)

Manages an NFS (Network File System) share on TrueNAS Scale. NFS shares allow Unix/Linux systems to access files over the network with native performance and permissions.

## Example Usage

### Basic NFS Share

```terraform
resource "truenas_nfs_share" "data" {
  path    = "/mnt/tank/data"
  comment = "General data share"
}
```

### NFS Share with Network Restrictions

```terraform
resource "truenas_nfs_share" "private" {
  path    = "/mnt/tank/private"
  comment = "Private network share"
  
  networks = [
    "192.168.1.0/24",
    "10.0.0.0/8"
  ]
  
  hosts = [
    "server1.example.com",
    "server2.example.com"
  ]
}
```

### Read-Only NFS Share

```terraform
resource "truenas_nfs_share" "readonly" {
  path    = "/mnt/tank/readonly"
  comment = "Read-only share"
  
  ro       = true
  maproot_user  = "nobody"
  maproot_group = "nogroup"
}
```

### NFS Share with Security Options

```terraform
resource "truenas_nfs_share" "secure" {
  path    = "/mnt/tank/secure"
  comment = "Secure NFS share with Kerberos"
  
  security = ["KRB5", "KRB5I", "KRB5P"]
  
  networks = ["192.168.1.0/24"]
}
```

### Complete NFS Share Configuration

```terraform
resource "truenas_dataset" "nfs_data" {
  name = "tank/nfs-share"
  type = "FILESYSTEM"
  
  acltype = "NFSV4"
}

resource "truenas_nfs_share" "complete" {
  path    = "/mnt/${truenas_dataset.nfs_data.name}"
  comment = "Production NFS share"
  
  # Access control
  enabled  = true
  ro       = false
  
  # Network restrictions
  networks = [
    "192.168.1.0/24",
    "10.0.0.0/16"
  ]
  
  hosts = [
    "appserver1.example.com",
    "appserver2.example.com"
  ]
  
  # User mapping
  maproot_user  = "root"
  maproot_group = "wheel"
  mapall_user   = null
  mapall_group  = null
  
  # Security
  security = ["SYS"]
  
  depends_on = [truenas_dataset.nfs_data]
}
```

## Schema

### Required

- `path` (String) Full path to the directory to share (e.g., `/mnt/tank/data`)

### Optional

- `comment` (String) Description of the share
- `enabled` (Boolean) Enable the NFS share. Default: `true`
- `ro` (Boolean) Read-only access. Default: `false`
- `maproot_user` (String) Map root user to this local user. Default: `"root"`
- `maproot_group` (String) Map root group to this local group. Default: `"wheel"`  
- `mapall_user` (String) Map all users to this local user
- `mapall_group` (String) Map all groups to this local group
- `security` (List of String) Security flavors. Options: `SYS`, `KRB5`, `KRB5I`, `KRB5P`. Default: `["SYS"]`
- `networks` (List of String) Allowed networks in CIDR notation (e.g., `192.168.1.0/24`)
- `hosts` (List of String) Allowed hostnames or IP addresses

### Read-Only

- `id` (Number) The ID of the NFS share.

## Import

NFS shares can be imported using their ID:

```shell
terraform import truenas_nfs_share.example 1
```

To find the share ID, list all shares via the TrueNAS API or web interface.

## Notes

### Path Requirements

- Path must exist before creating the share
- Create the dataset first, then the share
- Use `depends_on` to ensure proper ordering

### Access Control

Multiple access control methods are available:

1. **networks**: Restrict by IP network (CIDR notation)
2. **hosts**: Restrict by hostname or individual IP
3. Combine both for fine-grained control

If neither is specified, share is accessible from any network.

### User Mapping

- **maproot**: Maps root (UID 0) to a specific local user
  - Use for administrative access
  - Default: `root`/`wheel`

- **mapall**: Maps all users to a single local user
  - Use for anonymous/simplified access
  - Overrides maproot when set

### Security Flavors

- **SYS**: Standard Unix authentication (default)
- **KRB5**: Kerberos authentication
- **KRB5I**: Kerberos with integrity checking
- **KRB5P**: Kerberos with privacy (encryption)

Multiple flavors can be specified. Kerberos requires additional configuration.

### Performance Considerations

- Enable async I/O on the dataset for better performance
- Use appropriate network MTU (jumbo frames for 10GbE)
- Consider dataset recordsize based on workload

### Common Patterns

#### Development Share
```terraform
resource "truenas_nfs_share" "dev" {
  path    = "/mnt/tank/dev"
  comment = "Development environment"
  
  maproot_user  = "developer"
  maproot_group = "developers"
  networks      = ["192.168.1.0/24"]
}
```

#### Backup Target
```terraform
resource "truenas_nfs_share" "backup" {
  path    = "/mnt/tank/backups"
  comment = "Backup storage"
  
  hosts = ["backup-server.example.com"]
  
  # Map all writes as limited user
  mapall_user  = "backup"
  mapall_group = "backup"
}
```

#### Read-Only Data Distribution
```terraform
resource "truenas_nfs_share" "software" {
  path    = "/mnt/tank/software"
  comment = "Software repository"
  
  ro = true
  
  # Public access
  networks = [
    "192.168.0.0/16",
    "10.0.0.0/8"
  ]
}
```

### Troubleshooting

**Permission Denied Errors:**
- Check dataset ACL configuration
- Verify user mapping (maproot/mapall)
- Ensure NFS service is running

**Mount Errors:**
- Verify path exists
- Check network/host restrictions
- Confirm firewall allows NFS ports (2049, 111)

**Performance Issues:**
- Enable async on dataset
- Check network speed and MTU
- Consider NFS version (NFSv4 recommended)

## See Also

- [truenas_dataset](dataset) - Create datasets to share
- [truenas_smb_share](smb_share) - SMB/CIFS shares for Windows
- [truenas_nfs_shares Data Source](../data-sources/nfs_shares) - Query all NFS shares