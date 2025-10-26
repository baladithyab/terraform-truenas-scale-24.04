# Create an SMB/CIFS share
resource "truenas_smb_share" "example" {
  name       = "myshare"
  path       = "/mnt/tank/mydata"
  comment    = "SMB share for mydata"
  enabled    = true
  browsable  = true
  guestok    = false
  readonly   = false
  recyclebin = true
  shadowcopy = true
}

# Create a guest-accessible SMB share
resource "truenas_smb_share" "public" {
  name      = "public"
  path      = "/mnt/tank/public"
  comment   = "Public share"
  enabled   = true
  browsable = true
  guestok   = true
  readonly  = true
}

# Import an existing SMB share
# terraform import truenas_smb_share.existing 1

