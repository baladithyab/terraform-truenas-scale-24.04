# Basic iSCSI target
resource "truenas_iscsi_target" "example" {
  name  = "target1"
  alias = "Example Target"
  mode  = "ISCSI"
}

# iSCSI target with portal groups and auth networks
resource "truenas_iscsi_target" "advanced" {
  name           = "target2"
  alias          = "Advanced Target"
  mode           = "ISCSI"
  groups         = [truenas_iscsi_portal.example.id]
  auth_networks  = ["192.168.1.0/24", "10.0.0.0/8"]
}

# Import an existing target
# terraform import truenas_iscsi_target.existing 1

