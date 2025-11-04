variable "truenas_base_url" {
  description = "TrueNAS API base URL"
  type        = string
  default     = "http://10.0.0.83"
}

variable "truenas_api_key" {
  description = "TrueNAS API key (set via TF_VAR_truenas_api_key in .envrc)"
  type        = string
  sensitive   = true
}

variable "network_bridge" {
  description = "Network bridge/interface to attach VM to"
  type        = string
  default     = "eno1"
}

variable "disk_zvol_path" {
  description = "Path to zvol for VM disk"
  type        = string
  default     = "/dev/zvol/Loki/vms/talos-minimal-test-disk0"
}

variable "talos_iso_path" {
  description = "Path to Talos ISO file on TrueNAS"
  type        = string
  default     = "/mnt/Loki/isos/talos-v1.11.3-metal-amd64.iso"
}
