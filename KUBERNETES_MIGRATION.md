# Kubernetes App Migration Guide

This guide explains how to use Terraform to manage TrueNAS Kubernetes applications and migrate them to another Kubernetes installation, including PVCs and all data.

## Overview

TrueNAS Scale includes a built-in Kubernetes cluster that runs applications as Helm charts. This provider allows you to:

1. **Manage apps as code** - Define all your Kubernetes apps in Terraform
2. **Backup apps with data** - Create backups that include PVCs and configuration
3. **Migrate to external K8s** - Export apps and data to standard Kubernetes clusters
4. **Version control** - Track all app configurations in Git

## Architecture

### TrueNAS Kubernetes Structure

```
TrueNAS Scale
├── Kubernetes Cluster (k3s)
├── ix-applications dataset (stores all app data)
│   ├── releases/
│   │   ├── plex/
│   │   │   ├── volumes/ (PVCs)
│   │   │   └── config/
│   │   └── nextcloud/
│   │       ├── volumes/ (PVCs)
│   │       └── config/
│   └── backups/ (chart release backups)
└── Helm Charts (from catalogs)
```

### Migration Targets

- **Another TrueNAS** - Direct backup/restore
- **Standard Kubernetes** - Export to YAML + data migration
- **Cloud Kubernetes** - EKS, GKE, AKS with data sync

## Workflow 1: Manage Apps with Terraform

### Step 1: Define Your Apps

```hcl
# main.tf
terraform {
  required_providers {
    truenas = {
      source = "github.com/baladithyab/truenas"
    }
  }
}

provider "truenas" {
  base_url = "http://10.0.0.83:81"
  api_key  = var.truenas_api_key
}

# Plex Media Server
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
      },
      {
        name  = "PLEX_CLAIM"
        value = var.plex_claim_token
      }
    ]
    
    storage = {
      config = {
        type     = "ixVolume"
        datasetName = "plex-config"
      }
      media = {
        type     = "hostPath"
        hostPath = "/mnt/tank/media"
      }
      transcode = {
        type     = "ixVolume"
        datasetName = "plex-transcode"
      }
    }
    
    resources = {
      limits = {
        cpu    = "4000m"
        memory = "8Gi"
      }
    }
  })
}

# Nextcloud
resource "truenas_chart_release" "nextcloud" {
  release_name = "nextcloud"
  catalog      = "TRUENAS"
  train        = "charts"
  item         = "nextcloud"
  version      = "2.0.0"
  
  values = jsonencode({
    nextcloud = {
      host     = "nextcloud.example.com"
      username = "admin"
    }
    
    postgresql = {
      enabled = true
      postgresqlUsername = "nextcloud"
      postgresqlDatabase = "nextcloud"
      persistence = {
        enabled = true
        storageClass = "ix-storage-class-nextcloud-postgres"
      }
    }
    
    storage = {
      data = {
        type     = "ixVolume"
        datasetName = "nextcloud-data"
      }
      config = {
        type     = "ixVolume"
        datasetName = "nextcloud-config"
      }
    }
  })
}

# Automated snapshots for app data
resource "truenas_periodic_snapshot_task" "apps_backup" {
  dataset        = "ix-applications"
  recursive      = true
  enabled        = true
  naming_schema  = "apps-backup-%Y-%m-%d_%H-%M"
  lifetime_value = 7
  lifetime_unit  = "DAY"
  
  # Daily at 2 AM
  schedule = jsonencode({
    minute = "0"
    hour   = "2"
    dom    = "*"
    month  = "*"
    dow    = "*"
  })
}
```

### Step 2: Apply Configuration

```bash
terraform init
terraform plan
terraform apply
```

### Step 3: Track in Git

```bash
git add main.tf
git commit -m "Add Plex and Nextcloud apps"
git push
```

## Workflow 2: Backup Apps for Migration

### Method A: Using TrueNAS Snapshots (Recommended)

The `ix-applications` dataset contains all app data. Create snapshots for point-in-time backups:

```hcl
# Create a snapshot before migration
resource "truenas_snapshot" "apps_pre_migration" {
  dataset   = "ix-applications"
  name      = "pre-migration-${formatdate("YYYY-MM-DD", timestamp())}"
  recursive = true
}
```

### Method B: Using Kubernetes Backup API

TrueNAS provides a backup API that creates Helm backups + snapshots:

```bash
# Create backup via API
curl -X POST \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  http://10.0.0.83:81/api/v2.0/kubernetes/backup_chart_releases

# List backups
curl -H "Authorization: Bearer $API_KEY" \
  http://10.0.0.83:81/api/v2.0/kubernetes/list_backups
```

### Method C: Export to Standard Kubernetes YAML

For migration to external Kubernetes, export the Helm values and create standard K8s manifests:

```bash
# Get current chart release configuration
curl -H "Authorization: Bearer $API_KEY" \
  http://10.0.0.83:81/api/v2.0/chart/release/id/plex | \
  jq '.config' > plex-values.json

# Convert to standard Kubernetes manifests
helm template plex truecharts/plex -f plex-values.json > plex-k8s.yaml
```

## Workflow 3: Migrate to Another TrueNAS

### Step 1: Backup Source System

```bash
# Create snapshot
terraform apply -target=truenas_snapshot.apps_pre_migration

# Or use API
curl -X POST \
  -H "Authorization: Bearer $API_KEY" \
  http://source-truenas/api/v2.0/kubernetes/backup_chart_releases
```

### Step 2: Transfer Data

```bash
# Option A: ZFS send/receive
zfs send -R tank/ix-applications@pre-migration-2024-01-15 | \
  ssh target-truenas zfs receive tank/ix-applications

# Option B: Rsync
rsync -avz --progress \
  /mnt/tank/ix-applications/ \
  target-truenas:/mnt/tank/ix-applications/
```

### Step 3: Restore on Target

```bash
# Restore backup via API
curl -X POST \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"backup_name": "backup-2024-01-15"}' \
  http://target-truenas/api/v2.0/kubernetes/restore_backup
```

### Step 4: Apply Terraform on Target

```bash
# Point to new TrueNAS
export TRUENAS_BASE_URL="http://target-truenas:81"

# Import existing apps
terraform import truenas_chart_release.plex plex
terraform import truenas_chart_release.nextcloud nextcloud

# Verify
terraform plan  # Should show no changes
```

## Workflow 4: Migrate to External Kubernetes

### Step 1: Export App Configurations

```bash
# Export all chart releases
for app in plex nextcloud; do
  curl -H "Authorization: Bearer $API_KEY" \
    http://truenas/api/v2.0/chart/release/id/$app | \
    jq '.config' > ${app}-values.json
done
```

### Step 2: Create Kubernetes Manifests

```hcl
# Create a migration helper script
locals {
  apps = {
    plex = {
      chart = "truecharts/plex"
      namespace = "media"
    }
    nextcloud = {
      chart = "nextcloud/nextcloud"
      namespace = "productivity"
    }
  }
}

# Generate Helm commands
output "migration_commands" {
  value = {
    for name, config in local.apps :
    name => "helm install ${name} ${config.chart} -n ${config.namespace} -f ${name}-values.json"
  }
}
```

### Step 3: Migrate PVC Data

```bash
# Create PVCs in target cluster
kubectl apply -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: plex-config
  namespace: media
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi
  storageClassName: fast-ssd
EOF

# Copy data from TrueNAS to target PVC
# Method A: Using a temporary pod
kubectl run -n media data-copy --image=alpine --restart=Never -- sleep 3600
kubectl cp /mnt/tank/ix-applications/releases/plex/volumes/config \
  media/data-copy:/data

# Method B: Using rsync
kubectl run -n media rsync --image=instrumentisto/rsync-ssh
kubectl exec -n media rsync -- \
  rsync -avz truenas:/mnt/tank/ix-applications/releases/plex/volumes/config/ /data/
```

### Step 4: Deploy to Target Cluster

```bash
# Deploy apps
helm install plex truecharts/plex -n media -f plex-values.json
helm install nextcloud nextcloud/nextcloud -n productivity -f nextcloud-values.json

# Verify
kubectl get pods -A
kubectl get pvc -A
```

## Workflow 5: Continuous Sync for Zero-Downtime Migration

### Step 1: Set Up Replication

```hcl
# Create replication task (future resource)
resource "truenas_replication_task" "apps_to_backup" {
  source_dataset = "tank/ix-applications"
  target_dataset = "backup-pool/ix-applications-replica"
  recursive      = true
  enabled        = true
  
  schedule = jsonencode({
    minute = "*/15"  # Every 15 minutes
    hour   = "*"
    dom    = "*"
    month  = "*"
    dow    = "*"
  })
}
```

### Step 2: Sync to External Storage

```bash
# Continuous sync to S3/NFS/etc
while true; do
  rsync -avz --delete \
    /mnt/tank/ix-applications/ \
    /mnt/backup-nfs/ix-applications/
  sleep 900  # 15 minutes
done
```

### Step 3: Cutover

```bash
# 1. Stop apps on TrueNAS
kubectl scale deployment plex --replicas=0 -n ix-plex

# 2. Final sync
rsync -avz --delete /mnt/tank/ix-applications/ /mnt/backup-nfs/

# 3. Deploy to target cluster
helm install plex truecharts/plex -n media -f plex-values.json

# 4. Verify and switch DNS/load balancer
```

## Best Practices

### 1. Version Control Everything

```bash
# Store all Terraform configs in Git
git add *.tf
git commit -m "Update Plex to version 1.0.1"
git tag v1.0.0
git push --tags
```

### 2. Test Migrations

```bash
# Create a test environment
terraform workspace new test
terraform apply

# Test migration process
# Destroy when done
terraform workspace select default
terraform workspace delete test
```

### 3. Automate Backups

```hcl
# Daily snapshots
resource "truenas_periodic_snapshot_task" "apps_daily" {
  dataset        = "ix-applications"
  recursive      = true
  enabled        = true
  naming_schema  = "daily-%Y-%m-%d"
  lifetime_value = 30
  lifetime_unit  = "DAY"
  
  schedule = jsonencode({
    minute = "0"
    hour   = "3"
    dom    = "*"
    month  = "*"
    dow    = "*"
  })
}

# Weekly snapshots
resource "truenas_periodic_snapshot_task" "apps_weekly" {
  dataset        = "ix-applications"
  recursive      = true
  enabled        = true
  naming_schema  = "weekly-%Y-W%W"
  lifetime_value = 12
  lifetime_unit  = "WEEK"
  
  schedule = jsonencode({
    minute = "0"
    hour   = "4"
    dom    = "*"
    month  = "*"
    dow    = "0"
  })
}
```

### 4. Document PVC Mappings

```hcl
# Create a mapping document
locals {
  pvc_mappings = {
    plex = {
      truenas_path = "/mnt/tank/ix-applications/releases/plex/volumes"
      k8s_pvcs = [
        "plex-config",
        "plex-transcode"
      ]
    }
    nextcloud = {
      truenas_path = "/mnt/tank/ix-applications/releases/nextcloud/volumes"
      k8s_pvcs = [
        "nextcloud-data",
        "nextcloud-config",
        "nextcloud-postgres"
      ]
    }
  }
}

output "pvc_migration_guide" {
  value = local.pvc_mappings
}
```

## Troubleshooting

### Issue: Apps won't start after migration

**Cause**: PVC permissions or ownership mismatch

**Solution**:
```bash
# Fix permissions
kubectl exec -n media plex-0 -- chown -R 1000:1000 /config
kubectl exec -n media plex-0 -- chmod -R 755 /config
```

### Issue: Data not syncing

**Cause**: Rsync or ZFS send/receive errors

**Solution**:
```bash
# Check ZFS snapshots
zfs list -t snapshot | grep ix-applications

# Verify rsync
rsync -avz --dry-run /source/ /target/
```

### Issue: Helm values incompatible

**Cause**: Different chart versions or schemas

**Solution**:
```bash
# Compare schemas
helm show values truecharts/plex > truecharts-schema.yaml
helm show values k8s-at-home/plex > k8s-at-home-schema.yaml
diff truecharts-schema.yaml k8s-at-home-schema.yaml
```

## See Also

- [IMPORT_GUIDE.md](IMPORT_GUIDE.md) - Importing existing resources
- [examples/complete-kubernetes/](examples/complete-kubernetes/) - Complete K8s examples
- [TrueNAS Scale Kubernetes Documentation](https://www.truenas.com/docs/scale/scaletutorials/apps/)

