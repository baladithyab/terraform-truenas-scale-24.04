# Create a ZFS dataset
resource "truenas_dataset" "example" {
  name        = "tank/mydata"
  type        = "FILESYSTEM"
  compression = "LZ4"
  atime       = "OFF"
  quota       = 107374182400 # 100GB in bytes
  comments    = "Managed by Terraform"
}

# Create a dataset with custom settings
resource "truenas_dataset" "media" {
  name        = "tank/media"
  type        = "FILESYSTEM"
  compression = "LZ4"
  recordsize  = "1M"
  atime       = "OFF"
  sync        = "STANDARD"
  copies      = 1
}

# Import an existing dataset
# terraform import truenas_dataset.existing tank/existing-dataset

