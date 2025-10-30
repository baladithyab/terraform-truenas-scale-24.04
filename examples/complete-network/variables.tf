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

