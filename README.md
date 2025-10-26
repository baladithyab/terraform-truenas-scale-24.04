# Terraform Provider for TrueNAS Scale 24.04

This is a Terraform provider for managing TrueNAS Scale 24.04 resources using the REST API. This provider is specifically designed for TrueNAS Scale 24.04, as version 25.04 transitions from REST to JSON-RPC over WebSocket.

## Features

- **Full REST API Support**: Leverages all available TrueNAS Scale 24.04 REST API endpoints
- **Resource Management**: Create, read, update, and delete TrueNAS resources
- **Import Support**: Import existing TrueNAS resources into Terraform state
- **Comprehensive Resources**:
  - ZFS Datasets
  - NFS Shares
  - SMB/CIFS Shares
  - Users
  - Groups
- **Data Sources**:
  - Dataset information
  - Pool information

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- TrueNAS Scale 24.04
- TrueNAS API Key

## Building the Provider

Clone the repository and build the provider:

```bash
git clone https://github.com/terraform-providers/terraform-provider-truenas
cd terraform-provider-truenas
go build -o terraform-provider-truenas
```

## Installation

### Local Development

For local development, you can use the provider by placing it in your Terraform plugins directory:

```bash
# Build the provider
go build -o terraform-provider-truenas

# Create the plugins directory
mkdir -p ~/.terraform.d/plugins/terraform-providers/truenas/1.0.0/linux_amd64/

# Copy the provider binary
cp terraform-provider-truenas ~/.terraform.d/plugins/terraform-providers/truenas/1.0.0/linux_amd64/
```

Then configure Terraform to use the local provider:

```hcl
terraform {
  required_providers {
    truenas = {
      source  = "terraform-providers/truenas"
      version = "1.0.0"
    }
  }
}
```

## Configuration

### Provider Configuration

```hcl
provider "truenas" {
  base_url = "http://10.0.0.213:81"  # Your TrueNAS server URL (with or without port)
  api_key  = "your-api-key-here"      # Your TrueNAS API key
}
```

### Environment Variables

You can also configure the provider using environment variables:

```bash
export TRUENAS_BASE_URL="http://10.0.0.213:81"
export TRUENAS_API_KEY="your-api-key-here"
```

### Getting a TrueNAS API Key

1. Log in to your TrueNAS web interface
2. Navigate to the top-right corner and click on the settings icon
3. Select "API Keys"
4. Click "Add" to create a new API key
5. Give it a name and click "Add"
6. Copy the generated API key (you won't be able to see it again!)

## Usage Examples

### Creating a ZFS Dataset

```hcl
resource "truenas_dataset" "mydata" {
  name        = "tank/mydata"
  type        = "FILESYSTEM"
  compression = "LZ4"
  atime       = "OFF"
  quota       = 107374182400  # 100GB
  comments    = "Managed by Terraform"
}
```

### Creating an NFS Share

```hcl
resource "truenas_nfs_share" "myshare" {
  path     = "/mnt/tank/mydata"
  comment  = "NFS share for mydata"
  networks = ["192.168.1.0/24"]
  readonly = false
  enabled  = true
}
```

### Creating an SMB Share

```hcl
resource "truenas_smb_share" "myshare" {
  name       = "myshare"
  path       = "/mnt/tank/mydata"
  comment    = "SMB share for mydata"
  enabled    = true
  browsable  = true
  guestok    = false
  recyclebin = true
  shadowcopy = true
}
```

### Creating a User

```hcl
resource "truenas_user" "john" {
  username  = "john"
  full_name = "John Doe"
  email     = "john@example.com"
  password  = "SecurePassword123!"
  home      = "/mnt/tank/home/john"
  shell     = "/bin/bash"
  sudo      = false
  smb       = true
}
```

### Creating a Group

```hcl
resource "truenas_group" "developers" {
  name = "developers"
  sudo = false
  smb  = true
}
```

### Using Data Sources

```hcl
# Get information about a dataset
data "truenas_dataset" "existing" {
  id = "tank/mydata"
}

# Get information about a pool
data "truenas_pool" "tank" {
  id = "tank"
}

output "dataset_available_space" {
  value = data.truenas_dataset.existing.available
}
```

## Importing Existing Resources

All resources support importing existing TrueNAS resources into Terraform state:

```bash
# Import a dataset
terraform import truenas_dataset.mydata tank/mydata

# Import an NFS share (use the share ID)
terraform import truenas_nfs_share.myshare 1

# Import an SMB share (use the share ID)
terraform import truenas_smb_share.myshare 1

# Import a user (use the user ID)
terraform import truenas_user.john 1000

# Import a group (use the group ID)
terraform import truenas_group.developers 1000
```

## Available Resources

- `truenas_dataset` - Manage ZFS datasets
- `truenas_nfs_share` - Manage NFS shares
- `truenas_smb_share` - Manage SMB/CIFS shares
- `truenas_user` - Manage user accounts
- `truenas_group` - Manage groups

## Available Data Sources

- `truenas_dataset` - Get information about a dataset
- `truenas_pool` - Get information about a pool

## Development

### Running Tests

```bash
go test ./...
```

### Building Documentation

```bash
go generate
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This provider is distributed under the Mozilla Public License 2.0. See `LICENSE` for more information.

## Support

For issues and questions:
- Open an issue on GitHub
- Check the TrueNAS Scale 24.04 API documentation at `http://your-truenas-ip/api/docs/`

## Roadmap

Future enhancements may include:
- Additional resources (iSCSI, VMs, Apps, etc.)
- More comprehensive data sources
- Enhanced error handling and validation
- Automated testing suite
- Support for more TrueNAS features

## Notes

- This provider is designed specifically for TrueNAS Scale 24.04
- TrueNAS Scale 25.04+ uses JSON-RPC over WebSocket instead of REST API
- Always backup your TrueNAS configuration before making changes
- Test in a non-production environment first

