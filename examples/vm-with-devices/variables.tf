variable "truenas_base_url" {
  description = "TrueNAS API base URL"
  type        = string
}

variable "truenas_api_key" {
  description = "TrueNAS API key"
  type        = string
  sensitive   = true
}

