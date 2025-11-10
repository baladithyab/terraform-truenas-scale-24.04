---
page_title: "truenas_user Resource - terraform-provider-truenas"
subcategory: "User & Group Management"
description: |-
  Manages a user account on TrueNAS Scale.
---

# truenas_user (Resource)

Manages a user account on TrueNAS Scale. This resource handles user creation, configuration updates, and deletion including authentication settings and SSH access.

## Example Usage

### Basic User

```terraform
resource "truenas_user" "basic" {
  username = "johndoe"
  full_name = "John Doe"
  email = "john.doe@example.com"
  
  password = var.user_password
}
```

### User with Custom UID and Group

```terraform
resource "truenas_user" "developer" {
  username = "developer"
  full_name = "Developer User"
  email = "dev@company.com"
  
  uid = 2000
  group = 2000
  
  home = "/home/developer"
  shell = "/bin/bash"
  
  password = var.dev_password
}
```

### User with SSH Access and Sudo

```terraform
resource "truenas_user" "admin" {
  username = "admin"
  full_name = "Administrator"
  email = "admin@company.com"
  
  password = var.admin_password
  
  sshpubkey = var.admin_ssh_public_key
  
  sudo = true
  smb = true
}
```

### Service User

```terraform
resource "truenas_user" "service" {
  username = "service-account"
  full_name = "Service Account"
  
  password = var.service_password
  
  # Service account with no shell access
  shell = "/usr/sbin/nologin"
  home = "/nonexistent"
  
  # No sudo access
  sudo = false
  
  # Enable SMB for file sharing
  smb = true
}
```

### Complete User Configuration

```terraform
resource "truenas_user" "complete" {
  username = "fulluser"
  full_name = "Complete User Example"
  email = "complete@example.com"
  
  uid = 3000
  group = 3000
  
  home = "/home/fulluser"
  shell = "/usr/bin/zsh"
  
  password = var.complete_password
  
  sshpubkey = var.complete_ssh_public_key
  
  locked = false
  sudo = true
  smb = true
}
```

## Schema

### Required

- `username` (String) Username for the user account. Must be unique.

### Optional

- `uid` (Number) User ID (UID). If not specified, next available UID will be used.
- `full_name` (String) Full name of the user.
- `email` (String) Email address of the user.
- `password` (String, Sensitive) User password for authentication.
- `group` (Number) Primary group ID (GID).
- `home` (String) Home directory path.
- `shell` (String) Login shell (e.g., `/bin/bash`, `/usr/bin/zsh`, `/usr/sbin/nologin`).
- `sshpubkey` (String) SSH public key for key-based authentication.
- `locked` (Boolean) Lock the account to prevent login. Default: `false`.
- `sudo` (Boolean) Allow sudo access for the user. Default: `false`.
- `smb` (Boolean) Enable SMB authentication for the user. Default: `false`.

### Read-Only

- `id` (String) The ID of the user account.

## Import

Users can be imported using their ID:

```shell
terraform import truenas_user.example 1
```

To find the user ID, list all users via the TrueNAS API or web interface.

## Notes

### User Management

- Usernames must be unique across the system
- UID conflicts will cause creation to fail
- Password changes take effect immediately
- SSH keys are added to `~/.ssh/authorized_keys`

### UID/GID Management

- If `uid` is not specified, the next available UID is used
- If `group` is not specified, a new group with the same name is created
- Use specific UIDs/GIDs for consistency across systems
- Avoid UIDs below 1000 (typically reserved for system users)

### Shell Options

Common shell choices:
- `/bin/bash` - Standard Bash shell
- `/bin/sh` - POSIX shell
- `/usr/bin/zsh` - Z shell (if installed)
- `/usr/sbin/nologin` - No login access (for service accounts)
- `/bin/false` - Prevents login

### SSH Key Management

- SSH public keys are automatically added to the user's authorized_keys
- Multiple SSH keys can be added by concatenating them with newlines
- SSH keys provide passwordless authentication
- Remove the `sshpubkey` attribute to remove all SSH keys

### Account Security

- Use `locked = true` to temporarily disable an account
- Set `sudo = true` only for users who need administrative privileges
- Consider using `smb = false` for users who don't need Windows file sharing
- Regular password changes are recommended for security

### Home Directory

- Home directories are created automatically if they don't exist
- Use `/nonexistent` for service accounts that don't need a home directory
- Ensure proper permissions on custom home directories

### Common Patterns

#### Developer User
```terraform
resource "truenas_user" "developer" {
  username = "devuser"
  full_name = "Developer User"
  email = "dev@company.com"
  
  group = 1000
  shell = "/bin/bash"
  
  sshpubkey = var.dev_ssh_key
  sudo = true
}
```

#### Service Account
```terraform
resource "truenas_user" "service" {
  username = "svc-app"
  full_name = "Application Service Account"
  
  shell = "/usr/sbin/nologin"
  home = "/nonexistent"
  
  password = random_password.service.result
  
  sudo = false
  smb = true
}
```

#### Limited User
```terraform
resource "truenas_user" "limited" {
  username = "limited"
  full_name = "Limited Access User"
  email = "limited@company.com"
  
  password = var.limited_password
  
  # No shell access
  shell = "/usr/sbin/nologin"
  
  # No administrative privileges
  sudo = false
  
  # Account can be locked if needed
  locked = false
}
```

### Troubleshooting

**Login Issues:**
- Verify password is correctly set
- Check if account is locked (`locked = true`)
- Ensure shell is valid and exists
- Verify home directory permissions

**SSH Access Issues:**
- Confirm SSH public key format is correct
- Check SSH service is running on TrueNAS
- Verify `~/.ssh/authorized_keys` file permissions

**Permission Issues:**
- Check primary group membership
- Verify sudo configuration if needed
- Ensure user is member of required groups

## See Also

- [truenas_group](group) - Manage user groups
- [truenas_dataset](dataset) - Create home directories
- [truenas_nfs_share](nfs_share) - Share directories via NFS
- [truenas_smb_share](smb_share) - Share directories via SMB