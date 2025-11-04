terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.0"
    }
  }
}

provider "truenas" {
  base_url = "http://10.0.0.83:81"
  api_key  = var.truenas_api_key
}

variable "truenas_api_key" {
  description = "TrueNAS API key"
  type        = string
  sensitive   = true
}

variable "pool_name" {
  description = "Name of the storage pool"
  type        = string
  default     = "Loki"
}

# Test 1: Create a FILESYSTEM dataset (should work without volsize)
resource "truenas_dataset" "test_filesystem" {
  name        = "${var.pool_name}/test-filesystem"
  type        = "FILESYSTEM"
  compression = "LZ4"
  comments    = "Test filesystem dataset"
}

# Test 2: Create a VOLUME dataset with volsize (should work)
resource "truenas_dataset" "test_volume" {
  name        = "${var.pool_name}/test-volume"
  type        = "VOLUME"
  volsize     = 1073741824 # 1GB in bytes
  compression = "LZ4"
  comments    = "Test volume dataset (zvol)"
}

# Test 3: This should FAIL - VOLUME without volsize
# Uncomment to test validation
# resource "truenas_dataset" "test_volume_no_size" {
#   name        = "${var.pool_name}/test-volume-fail"
#   type        = "VOLUME"
#   compression = "LZ4"
#   comments    = "This should fail - no volsize"
# }

# Test 4: This should FAIL - FILESYSTEM with volsize
# Uncomment to test validation
# resource "truenas_dataset" "test_filesystem_with_size" {
#   name        = "${var.pool_name}/test-filesystem-fail"
#   type        = "FILESYSTEM"
#   volsize     = 1073741824
#   compression = "LZ4"
#   comments    = "This should fail - filesystem with volsize"
# }

output "filesystem_dataset" {
  value = {
    name = truenas_dataset.test_filesystem.name
    type = truenas_dataset.test_filesystem.type
    id   = truenas_dataset.test_filesystem.id
  }
}

output "volume_dataset" {
  value = {
    name    = truenas_dataset.test_volume.name
    type    = truenas_dataset.test_volume.type
    volsize = truenas_dataset.test_volume.volsize
    id      = truenas_dataset.test_volume.id
  }
}

