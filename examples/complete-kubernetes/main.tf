terraform {
  required_providers {
    truenas = {
      source  = "registry.terraform.io/baladithyab/truenas"
      version = "~> 0.2.14"
    }
  }
}

provider "truenas" {
  base_url = var.truenas_base_url
  api_key  = var.truenas_api_key
}

# ============================================================================
# Media Server Stack
# ============================================================================

# Plex Media Server
resource "truenas_chart_release" "plex" {
  release_name = "plex"
  catalog      = "TRUENAS"
  train        = "charts"
  item         = "plex"
  version      = var.plex_version

  values = jsonencode({
    hostNetwork = true

    environmentVariables = [
      {
        name  = "TZ"
        value = var.timezone
      },
      {
        name  = "PLEX_CLAIM"
        value = var.plex_claim_token
      }
    ]

    storage = {
      config = {
        type        = "ixVolume"
        datasetName = "plex-config"
      }
      media = {
        type     = "hostPath"
        hostPath = "/mnt/${var.pool_name}/media"
      }
      transcode = {
        type        = "ixVolume"
        datasetName = "plex-transcode"
      }
    }

    resources = {
      limits = {
        cpu    = "4000m"
        memory = "8Gi"
      }
      requests = {
        cpu    = "1000m"
        memory = "2Gi"
      }
    }

    gpuConfiguration = {
      nvidia = {
        enabled = var.enable_gpu
      }
    }
  })
}

# Sonarr - TV Show Management
resource "truenas_chart_release" "sonarr" {
  release_name = "sonarr"
  catalog      = "TRUENAS"
  train        = "charts"
  item         = "sonarr"
  version      = "1.0.0"

  values = jsonencode({
    environmentVariables = [
      {
        name  = "TZ"
        value = var.timezone
      }
    ]

    storage = {
      config = {
        type        = "ixVolume"
        datasetName = "sonarr-config"
      }
      media = {
        type     = "hostPath"
        hostPath = "/mnt/${var.pool_name}/media"
      }
      downloads = {
        type     = "hostPath"
        hostPath = "/mnt/${var.pool_name}/downloads"
      }
    }

    service = {
      main = {
        ports = {
          main = {
            port       = 8989
            targetPort = 8989
          }
        }
      }
    }
  })
}

# Radarr - Movie Management
resource "truenas_chart_release" "radarr" {
  release_name = "radarr"
  catalog      = "TRUENAS"
  train        = "charts"
  item         = "radarr"
  version      = "1.0.0"

  values = jsonencode({
    environmentVariables = [
      {
        name  = "TZ"
        value = var.timezone
      }
    ]

    storage = {
      config = {
        type        = "ixVolume"
        datasetName = "radarr-config"
      }
      media = {
        type     = "hostPath"
        hostPath = "/mnt/${var.pool_name}/media"
      }
      downloads = {
        type     = "hostPath"
        hostPath = "/mnt/${var.pool_name}/downloads"
      }
    }

    service = {
      main = {
        ports = {
          main = {
            port       = 7878
            targetPort = 7878
          }
        }
      }
    }
  })
}

# ============================================================================
# Productivity Stack
# ============================================================================

# Nextcloud
resource "truenas_chart_release" "nextcloud" {
  release_name = "nextcloud"
  catalog      = "TRUENAS"
  train        = "charts"
  item         = "nextcloud"
  version      = var.nextcloud_version

  values = jsonencode({
    nextcloud = {
      host     = var.nextcloud_domain
      username = "admin"
    }

    postgresql = {
      enabled            = true
      postgresqlUsername = "nextcloud"
      postgresqlDatabase = "nextcloud"
      persistence = {
        enabled      = true
        storageClass = "ix-storage-class-nextcloud-postgres"
      }
    }

    redis = {
      enabled = true
    }

    storage = {
      data = {
        type        = "ixVolume"
        datasetName = "nextcloud-data"
      }
      config = {
        type        = "ixVolume"
        datasetName = "nextcloud-config"
      }
    }

    resources = {
      limits = {
        cpu    = "2000m"
        memory = "4Gi"
      }
      requests = {
        cpu    = "500m"
        memory = "1Gi"
      }
    }
  })
}

# ============================================================================
# Home Automation Stack
# ============================================================================

# Home Assistant
resource "truenas_chart_release" "homeassistant" {
  release_name = "homeassistant"
  catalog      = "TRUENAS"
  train        = "charts"
  item         = "home-assistant"
  version      = "1.0.0"

  values = jsonencode({
    hostNetwork = true

    environmentVariables = [
      {
        name  = "TZ"
        value = var.timezone
      }
    ]

    storage = {
      config = {
        type        = "ixVolume"
        datasetName = "homeassistant-config"
      }
    }

    resources = {
      limits = {
        cpu    = "1000m"
        memory = "2Gi"
      }
    }
  })
}

# ============================================================================
# Backup & Snapshot Strategy
# ============================================================================

# Hourly snapshots for critical apps (kept for 1 day)
resource "truenas_periodic_snapshot_task" "apps_hourly" {
  dataset        = "ix-applications"
  recursive      = true
  enabled        = true
  naming_schema  = "hourly-%Y-%m-%d_%H-%M"
  lifetime_value = 1
  lifetime_unit  = "DAY"

  # Every hour
  schedule = jsonencode({
    minute = "0"
    hour   = "*"
    dom    = "*"
    month  = "*"
    dow    = "*"
  })
}

# Daily snapshots (kept for 1 week)
resource "truenas_periodic_snapshot_task" "apps_daily" {
  dataset        = "ix-applications"
  recursive      = true
  enabled        = true
  naming_schema  = "daily-%Y-%m-%d"
  lifetime_value = 1
  lifetime_unit  = "WEEK"

  # Daily at 2 AM
  schedule = jsonencode({
    minute = "0"
    hour   = "2"
    dom    = "*"
    month  = "*"
    dow    = "*"
  })
}

# Weekly snapshots (kept for 1 month)
resource "truenas_periodic_snapshot_task" "apps_weekly" {
  dataset        = "ix-applications"
  recursive      = true
  enabled        = true
  naming_schema  = "weekly-%Y-W%W"
  lifetime_value = 1
  lifetime_unit  = "MONTH"

  # Every Sunday at 3 AM
  schedule = jsonencode({
    minute = "0"
    hour   = "3"
    dom    = "*"
    month  = "*"
    dow    = "0"
  })
}

# Monthly snapshots (kept for 1 year)
resource "truenas_periodic_snapshot_task" "apps_monthly" {
  dataset        = "ix-applications"
  recursive      = true
  enabled        = true
  naming_schema  = "monthly-%Y-%m"
  lifetime_value = 1
  lifetime_unit  = "YEAR"

  # First day of month at 4 AM
  schedule = jsonencode({
    minute = "0"
    hour   = "4"
    dom    = "1"
    month  = "*"
    dow    = "*"
  })
}

# Pre-migration snapshot (manual trigger)
resource "truenas_snapshot" "apps_pre_migration" {
  count = var.create_migration_snapshot ? 1 : 0

  dataset   = "ix-applications"
  name      = "pre-migration-${formatdate("YYYY-MM-DD-hhmm", timestamp())}"
  recursive = true
}

# ============================================================================
# Outputs for Migration
# ============================================================================

output "app_summary" {
  description = "Summary of deployed applications"
  value = {
    media = {
      plex = {
        release_name = truenas_chart_release.plex.release_name
        version      = truenas_chart_release.plex.version
        status       = truenas_chart_release.plex.status
      }
      sonarr = {
        release_name = truenas_chart_release.sonarr.release_name
        version      = truenas_chart_release.sonarr.version
        status       = truenas_chart_release.sonarr.status
      }
      radarr = {
        release_name = truenas_chart_release.radarr.release_name
        version      = truenas_chart_release.radarr.version
        status       = truenas_chart_release.radarr.status
      }
    }
    productivity = {
      nextcloud = {
        release_name = truenas_chart_release.nextcloud.release_name
        version      = truenas_chart_release.nextcloud.version
        status       = truenas_chart_release.nextcloud.status
      }
    }
    home_automation = {
      homeassistant = {
        release_name = truenas_chart_release.homeassistant.release_name
        version      = truenas_chart_release.homeassistant.version
        status       = truenas_chart_release.homeassistant.status
      }
    }
  }
}

output "pvc_migration_paths" {
  description = "Paths to PVC data for migration"
  value = {
    plex = {
      config    = "/mnt/${var.pool_name}/ix-applications/releases/plex/volumes/ix-plex-config"
      transcode = "/mnt/${var.pool_name}/ix-applications/releases/plex/volumes/ix-plex-transcode"
    }
    sonarr = {
      config = "/mnt/${var.pool_name}/ix-applications/releases/sonarr/volumes/ix-sonarr-config"
    }
    radarr = {
      config = "/mnt/${var.pool_name}/ix-applications/releases/radarr/volumes/ix-radarr-config"
    }
    nextcloud = {
      data   = "/mnt/${var.pool_name}/ix-applications/releases/nextcloud/volumes/ix-nextcloud-data"
      config = "/mnt/${var.pool_name}/ix-applications/releases/nextcloud/volumes/ix-nextcloud-config"
    }
    homeassistant = {
      config = "/mnt/${var.pool_name}/ix-applications/releases/homeassistant/volumes/ix-homeassistant-config"
    }
  }
}

output "migration_commands" {
  description = "Commands for migrating to external Kubernetes"
  value = {
    export_configs = "See KUBERNETES_MIGRATION.md for detailed steps"
    backup_data    = "zfs send -R ${var.pool_name}/ix-applications@pre-migration | gzip > apps-backup.gz"
    import_data    = "gunzip -c apps-backup.gz | ssh target-k8s 'zfs receive backup-pool/apps'"
  }
}

