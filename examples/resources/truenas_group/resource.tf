# Create a group
resource "truenas_group" "example" {
  name = "developers"
  sudo = false
  smb  = true
}

# Create a group with users
resource "truenas_group" "admins" {
  name  = "admins"
  sudo  = true
  smb   = false
  users = [1000, 1001] # User IDs
}

# Import an existing group
# terraform import truenas_group.existing 1000

