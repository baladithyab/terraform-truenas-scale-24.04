# Quick Start Guide

This guide will help you get started with the TrueNAS Terraform provider.

## Prerequisites

1. **TrueNAS Scale 24.04** server running and accessible
2. **Terraform** installed (version 1.0 or later)
3. **Go** installed (version 1.21 or later) for building the provider
4. **TrueNAS API Key** (see below for how to generate)

## Step 1: Generate a TrueNAS API Key

1. Log in to your TrueNAS web interface
2. Click on the settings icon in the top-right corner
3. Select **API Keys**
4. Click **Add**
5. Enter a name for the key (e.g., "Terraform")
6. Click **Add**
7. **Copy the API key** - you won't be able to see it again!

## Step 2: Build and Install the Provider

```bash
# Clone the repository (or navigate to your local copy)
cd terraform-provider-truenas

# Build the provider
make build

# Install the provider locally
make install
```

This will install the provider to `~/.terraform.d/plugins/terraform-providers/truenas/1.0.0/linux_amd64/`

## Step 3: Create Your First Terraform Configuration

Create a new directory for your Terraform configuration:

```bash
mkdir ~/truenas-terraform
cd ~/truenas-terraform
```

Create a file named `main.tf`:

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
  base_url = "http://10.0.0.83:81"  # Replace with your TrueNAS IP
  api_key  = "your-api-key-here"      # Replace with your API key
}

# Create a simple dataset
resource "truenas_dataset" "test" {
  name        = "tank/terraform-test"  # Replace 'tank' with your pool name
  type        = "FILESYSTEM"
  compression = "LZ4"
  comments    = "Created by Terraform"
}

# Output the dataset name
output "dataset_name" {
  value = truenas_dataset.test.name
}
```

## Step 4: Initialize Terraform

```bash
terraform init
```

This will initialize Terraform and download any required providers.

## Step 5: Plan Your Changes

```bash
terraform plan
```

This will show you what Terraform will create without actually making any changes.

## Step 6: Apply Your Configuration

```bash
terraform apply
```

Type `yes` when prompted to confirm the changes.

## Step 7: Verify the Dataset Was Created

1. Log in to your TrueNAS web interface
2. Navigate to **Storage** → **Pools**
3. You should see the new dataset `tank/terraform-test`

## Step 8: Clean Up (Optional)

To remove the resources created by Terraform:

```bash
terraform destroy
```

Type `yes` when prompted to confirm.

## Using Environment Variables

Instead of hardcoding credentials in your Terraform files, you can use environment variables:

```bash
export TRUENAS_BASE_URL="http://10.0.0.83:81"
export TRUENAS_API_KEY="your-api-key-here"
```

Then simplify your `main.tf`:

```hcl
provider "truenas" {
  # Configuration will be read from environment variables
}
```

## Next Steps

- Check out the [examples](examples/) directory for more complex configurations
- Read the [README.md](README.md) for detailed documentation
- Explore the complete example in [examples/complete/](examples/complete/)

## Common Issues

### Provider Not Found

If you get an error about the provider not being found:

1. Make sure you ran `make install`
2. Check that the provider binary exists in `~/.terraform.d/plugins/terraform-providers/truenas/1.0.0/linux_amd64/`
3. Run `terraform init` again

### API Connection Errors

If you get connection errors:

1. Verify your TrueNAS server is accessible at the URL you specified
2. Check that the API key is correct
3. Make sure your TrueNAS server is running version 24.04
4. Try accessing the API docs at `http://your-truenas-ip/api/docs/`

### Dataset Already Exists

If you get an error that a dataset already exists:

1. Either delete the existing dataset from TrueNAS
2. Or import it into Terraform: `terraform import truenas_dataset.test tank/terraform-test`

## Tips

- Always test in a non-production environment first
- Use `terraform plan` before `terraform apply` to preview changes
- Keep your API key secure and never commit it to version control
- Use Terraform workspaces for managing multiple environments
- Enable state locking if working in a team

## Getting Help

- Check the [README.md](README.md) for detailed documentation
- Review the [examples](examples/) directory
- Open an issue on GitHub if you encounter problems

