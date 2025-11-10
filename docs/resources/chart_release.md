---
page_title: "truenas_chart_release Resource - terraform-provider-truenas"
subcategory: "Kubernetes"
description: |-
  Manages a Kubernetes application (chart release) on TrueNAS.
---

# truenas_chart_release (Resource)

Manages a Kubernetes application (chart release) on TrueNAS. This resource allows you to deploy, configure, and manage Helm charts from the TrueNAS application catalog.

## Example Usage

### Basic Chart Release

```terraform
resource "truenas_chart_release" "plex" {
  release_name = "plex"
  catalog      = "TRUENAS"
  train        = "charts"
  item         = "plex"
  version      = "1.0.0"
}
```

### Chart Release with Custom Values

```terraform
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
```

### Chart Release with Environment Variables and Storage

```terraform
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
    persistence = {
      data = {
        enabled = true
        type = "hostPath"
        hostPath = "/mnt/tank/minecraft/data"
      }
    }
  })
}
```

### Chart Release with Host Networking

```terraform
resource "truenas_chart_release" "media_server" {
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
      },
      {
        name  = "PLEX_CLAIM"
        value = "https://plex.example.com/claim"
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
```

### Chart Release with Resource Limits

```terraform
resource "truenas_chart_release" "resource_intensive_app" {
  release_name = "heavy-app"
  catalog      = "TRUENAS"
  train        = "charts"
  item         = "example-app"
  version      = "1.5.0"
  
  values = jsonencode({
    resources = {
      limits = {
        cpu = "2000m"
        memory = "4Gi"
      }
      requests = {
        cpu = "1000m"
        memory = "2Gi"
      }
    }
    nodeSelector = {
      "kubernetes.io/arch" = "amd64"
    }
  })
}
```

## Schema

### Required

- `release_name` (String) Name of the chart release. Must be unique within the namespace.
- `catalog` (String) Catalog name (e.g., TRUENAS, COMMUNITY).
- `train` (String) Catalog train (e.g., charts, community, enterprise).
- `item` (String) Chart item name (e.g., plex, nextcloud, minecraft).
- `version` (String) Chart version to deploy.

### Optional

- `values` (String) Chart values in JSON format. Use `jsonencode()` to convert HCL to JSON.

### Read-Only

- `id` (String) Chart release identifier (same as release_name).
- `status` (String) Current status of the chart release.

## Import

Chart releases can be imported using the release name:

```shell
terraform import truenas_chart_release.existing plex
```

## Notes

### Chart Values Configuration

The `values` parameter accepts JSON configuration for the Helm chart. Use Terraform's `jsonencode()` function to convert HCL objects to JSON:

```terraform
values = jsonencode({
  # Your chart configuration here
  service = {
    type = "LoadBalancer"
    port = 80
  }
})
```

### Catalog and Train Selection

- **Catalog**: The source catalog for the chart
  - `TRUENAS`: Official TrueNAS catalog
  - `COMMUNITY`: Community-maintained charts
  - Custom catalogs can be added in TrueNAS web interface

- **Train**: The category within the catalog
  - `charts`: Stable, well-maintained applications
  - `community`: Community-contributed applications
  - `enterprise`: Enterprise-grade applications

### Version Management

- Always specify exact versions for production deployments
- Use version constraints for development environments
- Check TrueNAS web interface for available chart versions

### Storage Configuration

Common storage patterns in chart values:

```terraform
values = jsonencode({
  persistence = {
    data = {
      enabled = true
      type = "hostPath"
      hostPath = "/mnt/tank/app/data"
      size = "100Gi"
    }
    config = {
      enabled = true
      type = "hostPath"
      hostPath = "/mnt/tank/app/config"
      size = "1Gi"
    }
  }
})
```

### Networking Configuration

Common networking patterns:

```terraform
values = jsonencode({
  service = {
    type = "ClusterIP"  # ClusterIP, NodePort, LoadBalancer
    port = 8080
  }
  ingress = {
    enabled = true
    hosts = ["app.example.com"]
    tls = true
  }
})
```

### Environment Variables

Set environment variables for containers:

```terraform
values = jsonencode({
  env = {
    TZ = "America/New_York"
    DATABASE_URL = "postgresql://user:pass@db:5432/app"
  }
  envFrom = [
    {
      secretRef = {
        name = "app-secrets"
      }
    }
  ]
})
```

### Resource Management

Control resource allocation:

```terraform
values = jsonencode({
  resources = {
    limits = {
      cpu = "2000m"      # 2 CPU cores
      memory = "4Gi"      # 4 GB RAM
    }
    requests = {
      cpu = "1000m"      # 1 CPU core
      memory = "2Gi"      # 2 GB RAM
    }
  }
})
```

## Best Practices

### Security

1. **Use Secrets**: Store sensitive data in Kubernetes secrets, not in chart values
2. **Network Policies**: Implement network policies when possible
3. **Resource Limits**: Always set resource limits to prevent resource exhaustion
4. **Image Security**: Use specific image tags, not `latest`

### Performance

1. **Resource Requests**: Set appropriate resource requests for scheduling
2. **Storage Classes**: Use appropriate storage classes for performance needs
3. **Node Selection**: Use node selectors for specialized workloads

### Maintenance

1. **Version Pinning**: Pin chart versions for production stability
2. **Backup Values**: Keep chart values in version control
3. **Monitoring**: Monitor application health and resource usage

## Troubleshooting

### Chart Not Found

- Verify catalog and train names are correct
- Check if the chart exists in the TrueNAS web interface
- Ensure the specified version is available

### Values Validation Errors

- Validate JSON syntax using online validators
- Check chart documentation for required values
- Use `jsonencode()` to avoid syntax errors

### Deployment Failures

- Check TrueNAS system logs
- Verify sufficient storage space
- Ensure resource limits are not exceeded
- Validate network connectivity

### Import Issues

- Use the exact release name as shown in TrueNAS
- Check if the release is in the correct catalog/train
- Verify the release is not in a failed state

## See Also

- [TrueNAS Applications Documentation](https://www.truenas.com/docs/scale/apps/) - Official TrueNAS application management guide
- [Helm Chart Values](https://helm.sh/docs/topics/charts/#values-files) - Helm chart configuration reference
- [Kubernetes Resources](https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/) - Kubernetes object management