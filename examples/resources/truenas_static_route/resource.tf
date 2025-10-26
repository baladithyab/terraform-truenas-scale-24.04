# Default route
resource "truenas_static_route" "default" {
  destination = "default"  # 0.0.0.0/0
  gateway     = "192.168.1.1"
  description = "Default gateway"
}

# Specific network route
resource "truenas_static_route" "private_network" {
  destination = "10.0.0.0/8"
  gateway     = "192.168.1.254"
  description = "Route to private network"
}

# Another subnet route
resource "truenas_static_route" "subnet" {
  destination = "192.168.2.0/24"
  gateway     = "192.168.1.10"
  description = "Route to subnet 2"
}

# Import an existing route
# terraform import truenas_static_route.existing 1

