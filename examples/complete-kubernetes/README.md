# Complete Kubernetes Application Stack

This example demonstrates a complete Kubernetes application stack on TrueNAS Scale with migration capabilities.

## What's Included

### Media Server Stack
- **Plex** - Media server with GPU transcoding support
- **Sonarr** - TV show management
- **Radarr** - Movie management

### Productivity Stack
- **Nextcloud** - File sync and collaboration with PostgreSQL

### Home Automation Stack
- **Home Assistant** - Home automation platform

### Backup Strategy
- Hourly snapshots (kept for 1 day)
- Daily snapshots (kept for 1 week)
- Weekly snapshots (kept for 1 month)
- Monthly snapshots (kept for 1 year)
- Pre-migration snapshot (on-demand)

## Prerequisites

1. TrueNAS Scale 24.04 with Kubernetes enabled
2. Sufficient storage pool (recommended: 500GB+)
3. TrueNAS API key
4. (Optional) NVIDIA GPU for Plex transcoding

## Quick Start

### 1. Configure Variables

Create `terraform.tfvars`:

```hcl
truenas_base_url = "http://10.0.0.83:81"
truenas_api_key  = "your-api-key-here"
pool_name        = "tank"
timezone         = "America/New_York"

# Plex configuration
plex_claim_token = "claim-xxxxxxxxxxxx"  # Get from https://www.plex.tv/claim/
enable_gpu       = false

# Nextcloud configuration
nextcloud_domain = "nextcloud.example.com"
```

### 2. Initialize and Apply

```bash
terraform init
terraform plan
terraform apply
```

### 3. Access Applications

After deployment, access your applications:

- **Plex**: `http://truenas-ip:32400/web`
- **Sonarr**: `http://truenas-ip:8989`
- **Radarr**: `http://truenas-ip:7878`
- **Nextcloud**: `http://nextcloud.example.com` (configure DNS/reverse proxy)
- **Home Assistant**: `http://truenas-ip:8123`

## Migration Scenarios

### Scenario 1: Migrate to Another TrueNAS

```bash
# 1. Create pre-migration snapshot
terraform apply -var="create_migration_snapshot=true"

# 2. Transfer data
zfs send -R tank/ix-applications@pre-migration-2024-01-15 | \
  ssh target-truenas zfs receive tank/ix-applications

# 3. On target TrueNAS, import resources
export TRUENAS_BASE_URL="http://target-truenas:81"
terraform import truenas_chart_release.plex plex
terraform import truenas_chart_release.nextcloud nextcloud
# ... import other apps

# 4. Verify
terraform plan  # Should show no changes
```

### Scenario 2: Migrate to External Kubernetes (EKS/GKE/AKS)

```bash
# 1. Export configurations
terraform output pvc_migration_paths > migration-paths.json

# 2. Create PVCs in target cluster
kubectl apply -f k8s-pvcs.yaml

# 3. Copy data
# See KUBERNETES_MIGRATION.md for detailed steps

# 4. Deploy apps
helm install plex truecharts/plex -f plex-values.yaml
helm install nextcloud nextcloud/nextcloud -f nextcloud-values.yaml
```

### Scenario 3: Backup Before Major Changes

```bash
# Create snapshot before updates
terraform apply -var="create_migration_snapshot=true"

# Update app versions
# Edit main.tf to change versions

# Apply updates
terraform apply

# If something goes wrong, rollback
zfs rollback tank/ix-applications@pre-migration-2024-01-15
```

## Storage Layout

```
/mnt/tank/
├── ix-applications/          # Kubernetes app data
│   ├── releases/
│   │   ├── plex/
│   │   │   └── volumes/
│   │   │       ├── ix-plex-config/
│   │   │       └── ix-plex-transcode/
│   │   ├── sonarr/
│   │   │   └── volumes/
│   │   │       └── ix-sonarr-config/
│   │   ├── radarr/
│   │   │   └── volumes/
│   │   │       └── ix-radarr-config/
│   │   ├── nextcloud/
│   │   │   └── volumes/
│   │   │       ├── ix-nextcloud-data/
│   │   │       ├── ix-nextcloud-config/
│   │   │       └── ix-nextcloud-postgres/
│   │   └── homeassistant/
│   │       └── volumes/
│   │           └── ix-homeassistant-config/
│   └── backups/              # Kubernetes backups
├── media/                    # Media files (shared)
│   ├── movies/
│   ├── tv/
│   └── music/
└── downloads/                # Download directory (shared)
```

## Customization

### Add More Applications

```hcl
resource "truenas_chart_release" "myapp" {
  release_name = "myapp"
  catalog      = "TRUENAS"
  train        = "charts"
  item         = "myapp"
  version      = "1.0.0"
  
  values = jsonencode({
    # Your configuration here
  })
}
```

### Adjust Snapshot Retention

```hcl
resource "truenas_periodic_snapshot_task" "apps_daily" {
  # ... other config ...
  lifetime_value = 30  # Keep for 30 days instead of 7
  lifetime_unit  = "DAY"
}
```

### Enable GPU for Plex

```bash
terraform apply -var="enable_gpu=true"
```

## Monitoring

### Check App Status

```bash
# Via Terraform
terraform output app_summary

# Via TrueNAS API
curl -H "Authorization: Bearer $API_KEY" \
  http://truenas-ip/api/v2.0/chart/release
```

### Check Snapshots

```bash
# List snapshots
zfs list -t snapshot | grep ix-applications

# Check snapshot size
zfs list -t snapshot -o name,used,refer | grep ix-applications
```

### Check Storage Usage

```bash
# Via TrueNAS API
curl -H "Authorization: Bearer $API_KEY" \
  http://truenas-ip/api/v2.0/pool/dataset/id/tank%2Fix-applications
```

## Troubleshooting

### Apps Won't Start

```bash
# Check Kubernetes status
kubectl get pods -A

# Check app logs
kubectl logs -n ix-plex plex-xxxxx

# Restart app
kubectl rollout restart deployment -n ix-plex plex
```

### Storage Issues

```bash
# Check pool status
zfs list

# Check available space
df -h /mnt/tank

# Clean up old snapshots
zfs destroy tank/ix-applications@old-snapshot
```

### Migration Issues

See [KUBERNETES_MIGRATION.md](../../KUBERNETES_MIGRATION.md) for detailed troubleshooting.

## Cost Optimization

### For Cloud Migration

When migrating to cloud Kubernetes:

1. **Use appropriate storage classes**
   - EBS gp3 for AWS
   - Persistent Disk SSD for GCP
   - Azure Disk Premium for Azure

2. **Right-size resources**
   ```hcl
   resources = {
     limits = {
       cpu    = "1000m"  # Reduce from 4000m
       memory = "2Gi"    # Reduce from 8Gi
     }
   }
   ```

3. **Use spot instances** for non-critical workloads

4. **Enable autoscaling**

## Security Considerations

1. **Use secrets for sensitive data**
   ```bash
   export TF_VAR_truenas_api_key="your-key"
   export TF_VAR_plex_claim_token="your-token"
   ```

2. **Enable HTTPS** for Nextcloud and other web apps

3. **Configure firewall rules** to restrict access

4. **Regular backups** - Automated snapshots are configured

5. **Update regularly**
   ```bash
   # Update app versions in main.tf
   terraform apply
   ```

## Performance Tuning

### Plex Transcoding

```hcl
# Enable GPU
enable_gpu = true

# Increase transcode storage
storage = {
  transcode = {
    type        = "ixVolume"
    datasetName = "plex-transcode"
    # Add size limit if needed
  }
}
```

### Nextcloud

```hcl
# Enable Redis caching
redis = {
  enabled = true
}

# Increase PHP memory
environmentVariables = [
  {
    name  = "PHP_MEMORY_LIMIT"
    value = "512M"
  }
]
```

## See Also

- [KUBERNETES_MIGRATION.md](../../KUBERNETES_MIGRATION.md) - Detailed migration guide
- [IMPORT_GUIDE.md](../../IMPORT_GUIDE.md) - Importing existing resources
- [TrueNAS Scale Apps Documentation](https://www.truenas.com/docs/scale/scaletutorials/apps/)

