# Create an NFS share
resource "truenas_nfs_share" "example" {
  path     = "/mnt/tank/mydata"
  comment  = "NFS share for mydata"
  networks = ["192.168.1.0/24"]
  readonly = false
  enabled  = true
}

# Create an NFS share with specific host access
resource "truenas_nfs_share" "restricted" {
  path         = "/mnt/tank/secure"
  comment      = "Restricted NFS share"
  hosts        = ["192.168.1.100", "192.168.1.101"]
  readonly     = true
  maproot_user = "root"
  security     = ["SYS"]
  enabled      = true
}

# Import an existing NFS share
# terraform import truenas_nfs_share.existing 1

