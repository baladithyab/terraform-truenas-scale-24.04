# Volsize Feature Test

This directory contains test configurations for the new `volsize` attribute in the `truenas_dataset` resource.

## What's Being Tested

1. **FILESYSTEM dataset** - Should work without `volsize`
2. **VOLUME dataset** - Should work with `volsize` (required)
3. **VOLUME without volsize** - Should fail with validation error (commented out)
4. **FILESYSTEM with volsize** - Should fail with validation error (commented out)

## Prerequisites

1. TrueNAS Scale 24.04 server running at `http://10.0.0.83:81`
2. API key with permissions to create datasets
3. Storage pool named "Loki" (or change `pool_name` variable)

## How to Test

### Step 1: Set up the provider

```bash
# Copy the built provider to the local plugins directory
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/baladithyab/truenas/0.2.1/linux_amd64/
cp ../terraform-provider-truenas ~/.terraform.d/plugins/registry.terraform.io/baladithyab/truenas/0.2.1/linux_amd64/terraform-provider-truenas_v0.2.1
chmod +x ~/.terraform.d/plugins/registry.terraform.io/baladithyab/truenas/0.2.1/linux_amd64/terraform-provider-truenas_v0.2.1
```

### Step 2: Create terraform.tfvars

```bash
cat > terraform.tfvars <<EOF
truenas_api_key = "your-api-key-here"
pool_name       = "Loki"  # Or your pool name
EOF
```

### Step 3: Initialize and plan

```bash
terraform init
terraform plan
```

### Step 4: Apply (creates the datasets)

```bash
terraform apply
```

### Step 5: Verify the datasets were created

Check in TrueNAS UI:
- Storage → Pools → Loki → test-filesystem (should be FILESYSTEM)
- Storage → Pools → Loki → test-volume (should be VOLUME with 1GB size)

Or via API:
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
  http://10.0.0.83:81/api/v2.0/pool/dataset/id/Loki%2Ftest-filesystem

curl -H "Authorization: Bearer YOUR_API_KEY" \
  http://10.0.0.83:81/api/v2.0/pool/dataset/id/Loki%2Ftest-volume
```

### Step 6: Test validation (optional)

Uncomment test 3 or test 4 in `main.tf` and run `terraform plan` to verify validation errors.

### Step 7: Clean up

```bash
terraform destroy
```

## Expected Results

### Successful Creation

```
Apply complete! Resources: 2 added, 0 changed, 0 destroyed.

Outputs:

filesystem_dataset = {
  "id" = "Loki/test-filesystem"
  "name" = "Loki/test-filesystem"
  "type" = "FILESYSTEM"
}
volume_dataset = {
  "id" = "Loki/test-volume"
  "name" = "Loki/test-volume"
  "type" = "VOLUME"
  "volsize" = 1073741824
}
```

### Validation Errors (when uncommented)

**Test 3 (VOLUME without volsize):**
```
Error: Missing Required Attribute

volsize is required when type is VOLUME. Please specify the volume size in bytes.
```

**Test 4 (FILESYSTEM with volsize):**
```
Error: Invalid Attribute

volsize is not valid for FILESYSTEM type datasets. Remove the volsize attribute or change type to VOLUME.
```

## Notes

- `volsize` is specified in bytes
- 1GB = 1073741824 bytes
- 10GB = 10737418240 bytes
- 100GB = 107374182400 bytes
- VOLUME datasets (zvols) are typically used for VM disks or iSCSI extents
- FILESYSTEM datasets are used for file storage (NFS, SMB shares)

