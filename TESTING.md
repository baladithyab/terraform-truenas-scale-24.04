# Testing Guide for TrueNAS Terraform Provider

This guide provides instructions for testing the TrueNAS Terraform provider.

## Prerequisites

1. TrueNAS Scale 24.04 server (accessible at http://10.0.0.83:81 or your server IP)
2. TrueNAS API key
3. At least one ZFS pool created (e.g., "tank")
4. Provider built and installed locally

## Setup

### 1. Build and Install the Provider

```bash
cd /mnt/e/CS/terraform-truenas-scale-24.04
make build
make install
```

### 2. Set Environment Variables

```bash
export TRUENAS_BASE_URL="http://10.0.0.83:81"
export TRUENAS_API_KEY="your-api-key-here"
```

## Test 1: Basic Provider Configuration

Create a test directory:

```bash
mkdir -p ~/test-truenas-provider
cd ~/test-truenas-provider
```

Create `test-provider.tf`:

```hcl
terraform {
  required_providers {
    truenas = {
      source  = "terraform-providers/truenas"
      version = "1.0.0"
    }
  }
}

provider "truenas" {
  # Will use environment variables
}

# Test data source
data "truenas_pool" "tank" {
  id = "tank"  # Replace with your pool name
}

output "pool_status" {
  value = data.truenas_pool.tank.status
}
```

Run:

```bash
terraform init
terraform plan
terraform apply
```

Expected output: Pool information should be displayed.

## Test 2: Dataset Creation

Create `test-dataset.tf`:

```hcl
resource "truenas_dataset" "test" {
  name        = "tank/terraform-test"
  type        = "FILESYSTEM"
  compression = "LZ4"
  atime       = "OFF"
  comments    = "Test dataset created by Terraform"
}

output "dataset_id" {
  value = truenas_dataset.test.id
}
```

Run:

```bash
terraform apply
```

Verify in TrueNAS UI:
1. Navigate to Storage → Pools
2. Expand your pool
3. Verify "terraform-test" dataset exists

Clean up:

```bash
terraform destroy
```

## Test 3: NFS Share Creation

Create `test-nfs.tf`:

```hcl
resource "truenas_dataset" "nfs_test" {
  name        = "tank/nfs-test"
  type        = "FILESYSTEM"
  compression = "LZ4"
}

resource "truenas_nfs_share" "test" {
  path     = "/mnt/${truenas_dataset.nfs_test.name}"
  comment  = "Test NFS share"
  networks = ["192.168.1.0/24"]  # Adjust to your network
  readonly = false
  enabled  = true
}

output "nfs_share_id" {
  value = truenas_nfs_share.test.id
}
```

Run:

```bash
terraform apply
```

Verify in TrueNAS UI:
1. Navigate to Shares → Unix (NFS) Shares
2. Verify the share exists

Clean up:

```bash
terraform destroy
```

## Test 4: SMB Share Creation

Create `test-smb.tf`:

```hcl
resource "truenas_dataset" "smb_test" {
  name        = "tank/smb-test"
  type        = "FILESYSTEM"
  compression = "LZ4"
}

resource "truenas_smb_share" "test" {
  name       = "test-share"
  path       = "/mnt/${truenas_dataset.smb_test.name}"
  comment    = "Test SMB share"
  enabled    = true
  browsable  = true
  guestok    = false
  recyclebin = true
  shadowcopy = true
}

output "smb_share_id" {
  value = truenas_smb_share.test.id
}
```

Run:

```bash
terraform apply
```

Verify in TrueNAS UI:
1. Navigate to Shares → Windows (SMB) Shares
2. Verify the share exists

Clean up:

```bash
terraform destroy
```

## Test 5: User and Group Creation

Create `test-users.tf`:

```hcl
resource "truenas_group" "test" {
  name = "terraform-test-group"
  sudo = false
  smb  = true
}

resource "truenas_user" "test" {
  username  = "tftest"
  full_name = "Terraform Test User"
  password  = "TestPassword123!"
  group     = truenas_group.test.gid
  home      = "/mnt/tank/home/tftest"
  shell     = "/bin/bash"
  sudo      = false
  smb       = true
}

output "user_id" {
  value = truenas_user.test.id
}

output "group_id" {
  value = truenas_group.test.id
}
```

Run:

```bash
terraform apply
```

Verify in TrueNAS UI:
1. Navigate to Credentials → Local Users
2. Verify "tftest" user exists
3. Navigate to Credentials → Local Groups
4. Verify "terraform-test-group" exists

Clean up:

```bash
terraform destroy
```

## Test 6: Import Existing Resources

### Import a Dataset

First, create a dataset manually in TrueNAS UI:
1. Storage → Pools → Add Dataset
2. Name: "manual-dataset"

Then import it:

```hcl
resource "truenas_dataset" "imported" {
  name        = "tank/manual-dataset"
  type        = "FILESYSTEM"
  compression = "LZ4"
}
```

```bash
terraform import truenas_dataset.imported tank/manual-dataset
terraform plan  # Should show no changes
```

### Import an NFS Share

Create an NFS share manually, note its ID, then:

```hcl
resource "truenas_nfs_share" "imported" {
  path    = "/mnt/tank/some-path"
  enabled = true
}
```

```bash
terraform import truenas_nfs_share.imported 1  # Use actual share ID
terraform plan
```

## Test 7: Complete Infrastructure

Use the complete example:

```bash
cd /mnt/e/CS/terraform-truenas-scale-24.04/examples/complete
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your values
terraform init
terraform plan
terraform apply
```

This will create:
- Multiple datasets
- NFS and SMB shares
- Users and groups
- Complete project structure

Verify all resources in TrueNAS UI, then:

```bash
terraform destroy
```

## Test 8: State Management

Test state operations:

```bash
# Create a resource
terraform apply

# Show state
terraform state list
terraform state show truenas_dataset.test

# Move a resource
terraform state mv truenas_dataset.test truenas_dataset.renamed

# Remove from state (doesn't delete resource)
terraform state rm truenas_dataset.renamed
```

## Test 9: Update Operations

Create a dataset:

```hcl
resource "truenas_dataset" "update_test" {
  name        = "tank/update-test"
  compression = "LZ4"
  quota       = 10737418240  # 10GB
}
```

Apply, then modify:

```hcl
resource "truenas_dataset" "update_test" {
  name        = "tank/update-test"
  compression = "GZIP"  # Changed
  quota       = 21474836480  # Changed to 20GB
  comments    = "Updated via Terraform"  # Added
}
```

Run:

```bash
terraform plan  # Should show changes
terraform apply
```

Verify changes in TrueNAS UI.

## Test 10: Error Handling

Test various error conditions:

### Invalid API Key

```bash
export TRUENAS_API_KEY="invalid-key"
terraform plan  # Should fail with authentication error
```

### Invalid Server URL

```bash
export TRUENAS_BASE_URL="http://invalid-server:81"
terraform plan  # Should fail with connection error
```

### Duplicate Resource

Try creating a dataset that already exists:

```bash
# Create manually in TrueNAS UI first
terraform apply  # Should fail with appropriate error
```

## Troubleshooting

### Provider Not Found

```bash
# Verify installation
ls -la ~/.terraform.d/plugins/terraform-providers/truenas/1.0.0/linux_amd64/

# Reinstall
cd /mnt/e/CS/terraform-truenas-scale-24.04
make install
```

### API Errors

Enable debug logging:

```bash
export TF_LOG=DEBUG
terraform apply
```

### State Issues

If state gets corrupted:

```bash
# Backup state
cp terraform.tfstate terraform.tfstate.backup

# Remove problematic resource
terraform state rm problematic_resource

# Re-import
terraform import problematic_resource id
```

## Performance Testing

Test with multiple resources:

```hcl
resource "truenas_dataset" "perf_test" {
  count = 10
  name  = "tank/perf-test-${count.index}"
  type  = "FILESYSTEM"
}
```

Monitor:
- Creation time
- API response times
- Resource usage

## Cleanup

After all tests:

```bash
# Destroy all Terraform-managed resources
terraform destroy

# Manually verify in TrueNAS UI that resources are removed
# Clean up any orphaned resources
```

## Continuous Testing

For ongoing development, create a test script:

```bash
#!/bin/bash
set -e

echo "Building provider..."
make build
make install

echo "Running basic tests..."
cd ~/test-truenas-provider
terraform init -upgrade
terraform plan
terraform apply -auto-approve

echo "Verifying resources..."
# Add verification commands

echo "Cleaning up..."
terraform destroy -auto-approve

echo "All tests passed!"
```

## Reporting Issues

When reporting issues, include:
1. Terraform version: `terraform version`
2. Provider version
3. TrueNAS version
4. Full error message
5. Relevant configuration
6. Steps to reproduce

