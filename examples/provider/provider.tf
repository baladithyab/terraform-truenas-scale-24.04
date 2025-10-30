terraform {
  required_providers {
    truenas = {
      source = "terraform-providers/truenas"
    }
  }
}

provider "truenas" {
  base_url = "http://10.0.0.83:81"
  api_key  = "your-api-key-here"
}

# Or use environment variables:
# export TRUENAS_BASE_URL="http://10.0.0.83:81"
# export TRUENAS_API_KEY="your-api-key-here"

