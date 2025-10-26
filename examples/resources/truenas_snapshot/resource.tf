# Create a simple snapshot
resource "truenas_snapshot" "backup" {
  dataset = "tank/mydata"
  name    = "backup-2024-01-15"
}

# Create a recursive snapshot
resource "truenas_snapshot" "full_backup" {
  dataset   = "tank"
  name      = "full-backup-2024-01-15"
  recursive = true
}

# Create snapshot with VMware sync
resource "truenas_snapshot" "vm_backup" {
  dataset     = "tank/vms"
  name        = "vm-backup-2024-01-15"
  recursive   = true
  vmware_sync = "CONTINUE"
}

# Snapshot before major changes
resource "truenas_snapshot" "pre_upgrade" {
  dataset = "tank/production"
  name    = "pre-upgrade-${formatdate("YYYY-MM-DD-hhmm", timestamp())}"
}

# Import an existing snapshot
# terraform import truenas_snapshot.existing tank/mydata@backup-2024-01-15

