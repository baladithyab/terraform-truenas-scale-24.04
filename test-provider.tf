terraform {
  required_providers {
    truenas = {
      source  = "baladithyab/truenas"
      version = "0.2.23"
    }
  }
}

provider "truenas" {
  # Will use environment variables
}

# Test data source
data "truenas_pool" "tank" {
  id = "Loki" # Replace with your pool name
}

output "pool_status" {
  value = data.truenas_pool.tank.status
}
