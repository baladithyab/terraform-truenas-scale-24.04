# Deploy Plex Media Server
resource "truenas_chart_release" "plex" {
  release_name = "plex"
  catalog      = "TRUENAS"
  train        = "charts"
  item         = "plex"
  version      = "1.0.0"
  
  values = jsonencode({
    hostNetwork = true
    environmentVariables = [
      {
        name  = "TZ"
        value = "America/New_York"
      }
    ]
    storage = {
      config = {
        type = "hostPath"
        hostPath = "/mnt/tank/plex/config"
      }
      media = {
        type = "hostPath"
        hostPath = "/mnt/tank/media"
      }
    }
  })
}

# Deploy Nextcloud
resource "truenas_chart_release" "nextcloud" {
  release_name = "nextcloud"
  catalog      = "TRUENAS"
  train        = "charts"
  item         = "nextcloud"
  version      = "2.0.0"
  
  values = jsonencode({
    nextcloud = {
      host = "nextcloud.example.com"
    }
    postgresql = {
      enabled = true
    }
    storage = {
      data = {
        type = "hostPath"
        hostPath = "/mnt/tank/nextcloud/data"
      }
    }
  })
}

# Deploy Minecraft Server
resource "truenas_chart_release" "minecraft" {
  release_name = "minecraft"
  catalog      = "TRUENAS"
  train        = "community"
  item         = "minecraft-java"
  version      = "1.0.0"
  
  values = jsonencode({
    env = {
      EULA = "TRUE"
      TYPE = "PAPER"
      VERSION = "1.20.1"
    }
    service = {
      main = {
        ports = {
          main = {
            port = 25565
          }
        }
      }
    }
  })
}

# Import an existing chart release
# terraform import truenas_chart_release.existing plex

