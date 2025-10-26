# Create a user
resource "truenas_user" "example" {
  username  = "john"
  full_name = "John Doe"
  email     = "john@example.com"
  password  = "SecurePassword123!"
  home      = "/mnt/tank/home/john"
  shell     = "/bin/bash"
  sudo      = false
  smb       = true
}

# Create a user with SSH key
resource "truenas_user" "admin" {
  username  = "admin"
  full_name = "Admin User"
  password  = "AdminPassword123!"
  home      = "/mnt/tank/home/admin"
  shell     = "/bin/bash"
  sudo      = true
  sshpubkey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."
}

# Import an existing user
# terraform import truenas_user.existing 1000

