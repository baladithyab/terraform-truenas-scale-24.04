---
page_title: "truenas_group Resource - terraform-provider-truenas"
subcategory: "User & Group Management"
description: |-
  Manages a user group on TrueNAS Scale.
---

# truenas_group (Resource)

Manages a user group on TrueNAS Scale. Groups provide a way to organize users and apply permissions and access controls to multiple users at once.

## Example Usage

### Basic Group

```terraform
resource "truenas_group" "developers" {
  name = "developers"
}
```

### Group with Custom GID

```terraform
resource "truenas_group" "staff" {
  name = "staff"
  gid = 2000
}
```

### Group with Sudo Access

```terraform
resource "truenas_group" "admins" {
  name = "admins"
  gid = 1000
  
  sudo = true
  smb = true
}
```

### Group with Members

```terraform
resource "truenas_user" "user1" {
  username = "user1"
  full_name = "User One"
  password = var.user1_password
}

resource "truenas_user" "user2" {
  username = "user2"
  full_name = "User Two"
  password = var.user2_password
}

resource "truenas_group" "team" {
  name = "team"
  gid = 3000
  
  users = [
    truenas_user.user1.id,
    truenas_user.user2.id
  ]
  
  smb = true
}
```

### Complete Group Configuration

```terraform
resource "truenas_group" "production" {
  name = "production"
  gid = 4000
  
  # Grant sudo privileges to all group members
  sudo = true
  
  # Enable SMB authentication for Windows file sharing
  smb = true
  
  # Add existing users to the group
  users = [
    1001,  # User ID
    1002,  # User ID
    1003   # User ID
  ]
}
```

## Schema

### Required

- `name` (String) Group name. Must be unique.

### Optional

- `gid` (Number) Group ID (GID). If not specified, next available GID will be used.
- `sudo` (Boolean) Allow sudo access for group members. Default: `false`.
- `smb` (Boolean) Enable SMB authentication for group. Default: `false`.
- `users` (List of Number) List of user IDs that are members of this group.

### Read-Only

- `id` (String) The ID of the group.

## Import

Groups can be imported using their ID:

```shell
terraform import truenas_group.example 1
```

To find the group ID, list all groups via the TrueNAS API or web interface.

## Notes

### Group Management

- Group names must be unique across the system
- GID conflicts will cause creation to fail
- Groups can be used for file permissions and access control
- Primary group of users is set in the user resource

### GID Management

- If `gid` is not specified, the next available GID is used
- Use specific GIDs for consistency across systems
- Avoid GIDs below 1000 (typically reserved for system groups)
- GID ranges to consider:
  - 0-99: System groups
  - 100-999: System service groups
  - 1000+: User groups

### User Membership

- The `users` attribute contains user IDs (not usernames)
- Users can be members of multiple groups
- Primary group is set in the user resource
- Secondary groups are managed through the group resource

### Sudo Access

- When `sudo = true`, all group members gain sudo privileges
- This is equivalent to adding the group to `/etc/sudoers`
- Use with caution - only grant sudo to trusted groups
- Consider using more specific sudo rules in production

### SMB Authentication

- `smb = true` enables group members to authenticate via SMB/CIFS
- Required for Windows file sharing access
- Works with `truenas_smb_share` resources
- Group members can access SMB shares with their credentials

### Common Patterns

#### Development Team Group
```terraform
resource "truenas_group" "developers" {
  name = "developers"
  gid = 2000
  
  sudo = true
  smb = true
  
  users = [
    truenas_user.dev1.id,
    truenas_user.dev2.id,
    truenas_user.dev3.id
  ]
}
```

#### Read-Only Access Group
```terraform
resource "truenas_group" "readonly" {
  name = "readonly"
  gid = 3000
  
  # No sudo access
  sudo = false
  
  # Enable for file sharing
  smb = true
  
  users = [
    truenas_user.ro_user1.id,
    truenas_user.ro_user2.id
  ]
}
```

#### Service Account Group
```terraform
resource "truenas_group" "services" {
  name = "services"
  gid = 4000
  
  # Service accounts don't need sudo
  sudo = false
  
  # May need SMB for file access
  smb = true
  
  users = [
    truenas_user.service_app.id,
    truenas_user.service_db.id
  ]
}
```

### Group Hierarchy

While Unix doesn't have nested groups, you can simulate group hierarchies:

```terraform
# Base group
resource "truenas_group" "staff" {
  name = "staff"
  gid = 2000
  smb = true
}

# Specialized groups
resource "truenas_group" "developers" {
  name = "developers"
  gid = 2001
  sudo = true
  smb = true
}

resource "truenas_group" "designers" {
  name = "designers"
  gid = 2002
  smb = true
}

# Users can be in multiple groups
resource "truenas_user" "fullstack" {
  username = "fullstack"
  full_name = "Full Stack Developer"
  
  # Primary group
  group = truenas_group.developers.gid
  
  # Add to additional groups via group resources
}
```

### Best Practices

1. **Use descriptive group names** that clearly indicate purpose
2. **Assign specific GIDs** for important groups to maintain consistency
3. **Grant sudo sparingly** - only to groups that truly need it
4. **Use groups for file permissions** rather than individual users
5. **Document group purposes** in comments or external documentation

### Integration with Other Resources

Groups work seamlessly with other TrueNAS resources:

```terraform
# Create a group for shared access
resource "truenas_group" "shared_users" {
  name = "shared_users"
  gid = 5000
  smb = true
}

# Create a dataset for shared data
resource "truenas_dataset" "shared_data" {
  name = "tank/shared"
  type = "FILESYSTEM"
}

# Share the dataset via SMB
resource "truenas_smb_share" "shared" {
  path = "/mnt/tank/shared"
  comment = "Shared data for team"
  
  # Group members can access this share
}
```

### Troubleshooting

**Permission Issues:**
- Verify users are actually in the group
- Check file/directory permissions
- Ensure primary group is set correctly

**Sudo Not Working:**
- Confirm `sudo = true` is set on the group
- Check sudoers configuration
- Verify user is logged in as correct user

**SMB Access Issues:**
- Ensure `smb = true` is set
- Check SMB service is running
- Verify share permissions

**Group Creation Fails:**
- Check for duplicate group names
- Verify GID is not already in use
- Ensure group name follows naming conventions

## See Also

- [truenas_user](user) - Manage user accounts
- [truenas_dataset](dataset) - Create datasets with group permissions
- [truenas_nfs_share](nfs_share) - Share with group permissions
- [truenas_smb_share](smb_share) - SMB shares with group access