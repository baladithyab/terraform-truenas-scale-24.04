# File-based iSCSI extent
resource "truenas_iscsi_extent" "file_extent" {
  name     = "file-extent-1"
  type     = "FILE"
  path     = "/mnt/tank/iscsi/extent1"
  filesize = 10737418240  # 10GB
  comment  = "File-based extent"
  enabled  = true
  readonly = false
}

# Disk-based iSCSI extent
resource "truenas_iscsi_extent" "disk_extent" {
  name      = "disk-extent-1"
  type      = "DISK"
  disk      = "zvol/tank/iscsi-disk"
  comment   = "Disk-based extent"
  enabled   = true
  readonly  = false
  blocksize = 512
}

# Read-only extent with custom settings
resource "truenas_iscsi_extent" "readonly_extent" {
  name            = "readonly-extent"
  type            = "FILE"
  path            = "/mnt/tank/iscsi/readonly"
  filesize        = 5368709120  # 5GB
  enabled         = true
  readonly        = true
  blocksize       = 4096
  avail_threshold = 80
  rpm             = "SSD"
  serial          = "CUSTOM123456"
}

# Import an existing extent
# terraform import truenas_iscsi_extent.existing 1

