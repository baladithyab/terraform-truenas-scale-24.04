variable "truenas_base_url" {
  description = "TrueNAS base URL"
  type        = string
  default     = "http://10.0.0.83:81"
}

variable "truenas_api_key" {
  description = "TrueNAS API key"
  type        = string
  sensitive   = true
}

variable "pool_name" {
  description = "ZFS pool name"
  type        = string
  default     = "tank"
}

variable "timezone" {
  description = "Timezone for applications"
  type        = string
  default     = "America/New_York"
}

# Plex Configuration
variable "plex_version" {
  description = "Plex chart version"
  type        = string
  default     = "1.0.0"
}

variable "plex_claim_token" {
  description = "Plex claim token from https://www.plex.tv/claim/"
  type        = string
  sensitive   = true
  default     = ""
}

variable "enable_gpu" {
  description = "Enable GPU for Plex transcoding"
  type        = bool
  default     = false
}

# Nextcloud Configuration
variable "nextcloud_version" {
  description = "Nextcloud chart version"
  type        = string
  default     = "2.0.0"
}

variable "nextcloud_domain" {
  description = "Nextcloud domain name"
  type        = string
  default     = "nextcloud.local"
}

# Migration Configuration
variable "create_migration_snapshot" {
  description = "Create a pre-migration snapshot"
  type        = bool
  default     = false
}

