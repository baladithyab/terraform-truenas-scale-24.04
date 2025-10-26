# Basic iSCSI portal listening on all interfaces
resource "truenas_iscsi_portal" "example" {
  comment = "Default Portal"
  
  listen {
    ip   = "0.0.0.0"
    port = 3260
  }
}

# Portal with multiple listen addresses
resource "truenas_iscsi_portal" "multi_listen" {
  comment                = "Multi-interface Portal"
  discovery_authmethod   = "NONE"
  
  listen {
    ip   = "192.168.1.10"
    port = 3260
  }
  
  listen {
    ip   = "10.0.0.10"
    port = 3260
  }
}

# Portal with CHAP authentication
resource "truenas_iscsi_portal" "secure" {
  comment                = "Secure Portal"
  discovery_authmethod   = "CHAP"
  discovery_authgroup    = 1
  
  listen {
    ip   = "192.168.1.10"
    port = 3260
  }
}

# Import an existing portal
# terraform import truenas_iscsi_portal.existing 1

