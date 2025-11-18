variable "truenas_url" {
  description = "The base URL of the TrueNAS Scale instance"
  type        = string
  default     = "http://truenas.local:81"
}

variable "truenas_api_key" {
  description = "The API key for authenticating with TrueNAS"
  type        = string
  sensitive   = true
}

variable "pool_name" {
  description = "The name of the storage pool to use for ISO storage"
  type        = string
  default     = "tank"
}

variable "static_ip" {
  description = "Static IP address for the VM"
  type        = string
  default     = "192.168.1.100"
}

variable "static_gateway" {
  description = "Gateway address for the static IP configuration"
  type        = string
  default     = "192.168.1.1"
}

variable "dns_servers" {
  description = "DNS servers for the static IP configuration"
  type        = list(string)
  default     = ["8.8.8.8", "8.8.4.4"]
}

variable "ssh_public_key" {
  description = "SSH public key for VM access"
  type        = string
  default     = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC3ExamplePublicKeyForDemostrationOnly"
}
