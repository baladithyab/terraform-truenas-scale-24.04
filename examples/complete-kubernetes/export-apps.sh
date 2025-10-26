#!/bin/bash
#
# Export TrueNAS Kubernetes apps for migration
# This script exports app configurations and generates migration commands
#

set -e

# Configuration
TRUENAS_URL="${TRUENAS_BASE_URL:-http://10.0.0.213:81}"
API_KEY="${TRUENAS_API_KEY}"
OUTPUT_DIR="${OUTPUT_DIR:-./migration-export}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_requirements() {
    log_info "Checking requirements..."
    
    if [ -z "$API_KEY" ]; then
        log_error "TRUENAS_API_KEY environment variable not set"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_error "jq is required but not installed"
        exit 1
    fi
    
    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed"
        exit 1
    fi
    
    log_info "All requirements met"
}

create_output_dir() {
    log_info "Creating output directory: $OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR"/{configs,manifests,scripts,data-maps}
}

list_chart_releases() {
    log_info "Fetching chart releases from TrueNAS..."
    
    curl -s -H "Authorization: Bearer $API_KEY" \
        "$TRUENAS_URL/api/v2.0/chart/release" | \
        jq -r '.[].name' > "$OUTPUT_DIR/app-list.txt"
    
    local count=$(wc -l < "$OUTPUT_DIR/app-list.txt")
    log_info "Found $count chart releases"
}

export_app_config() {
    local app_name=$1
    log_info "Exporting configuration for: $app_name"
    
    # Get full app details
    curl -s -H "Authorization: Bearer $API_KEY" \
        "$TRUENAS_URL/api/v2.0/chart/release/id/$app_name" \
        > "$OUTPUT_DIR/configs/${app_name}-full.json"
    
    # Extract just the values
    jq '.config' "$OUTPUT_DIR/configs/${app_name}-full.json" \
        > "$OUTPUT_DIR/configs/${app_name}-values.json"
    
    # Extract metadata
    jq '{
        name: .name,
        chart: .chart_metadata.name,
        version: .chart_metadata.version,
        catalog: .catalog,
        namespace: .namespace,
        status: .status
    }' "$OUTPUT_DIR/configs/${app_name}-full.json" \
        > "$OUTPUT_DIR/configs/${app_name}-metadata.json"
}

generate_pvc_map() {
    local app_name=$1
    log_info "Generating PVC map for: $app_name"
    
    # Extract storage configuration
    jq -r '.config.storage // {} | to_entries[] | 
        select(.value.type == "ixVolume") | 
        "\(.key):\(.value.datasetName)"' \
        "$OUTPUT_DIR/configs/${app_name}-full.json" \
        > "$OUTPUT_DIR/data-maps/${app_name}-pvcs.txt"
    
    # Generate full paths
    while IFS=: read -r mount_name dataset_name; do
        echo "$mount_name=/mnt/tank/ix-applications/releases/$app_name/volumes/ix-$dataset_name"
    done < "$OUTPUT_DIR/data-maps/${app_name}-pvcs.txt" \
        > "$OUTPUT_DIR/data-maps/${app_name}-paths.txt"
}

generate_k8s_manifest() {
    local app_name=$1
    log_info "Generating Kubernetes manifest for: $app_name"
    
    local chart=$(jq -r '.chart_metadata.name' "$OUTPUT_DIR/configs/${app_name}-full.json")
    local version=$(jq -r '.chart_metadata.version' "$OUTPUT_DIR/configs/${app_name}-full.json")
    
    cat > "$OUTPUT_DIR/manifests/${app_name}-helm.yaml" <<EOF
# Helm deployment for $app_name
# Original chart: $chart version $version

# 1. Create namespace
apiVersion: v1
kind: Namespace
metadata:
  name: $app_name

---
# 2. Create PVCs (adjust storage class and size as needed)
EOF
    
    # Add PVC definitions
    if [ -f "$OUTPUT_DIR/data-maps/${app_name}-pvcs.txt" ]; then
        while IFS=: read -r mount_name dataset_name; do
            cat >> "$OUTPUT_DIR/manifests/${app_name}-helm.yaml" <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: $app_name-$mount_name
  namespace: $app_name
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi  # Adjust as needed
  storageClassName: standard  # Adjust to your storage class

---
EOF
        done < "$OUTPUT_DIR/data-maps/${app_name}-pvcs.txt"
    fi
    
    cat >> "$OUTPUT_DIR/manifests/${app_name}-helm.yaml" <<EOF
# 3. Deploy with Helm
# helm install $app_name <chart-repo>/$chart \\
#   --namespace $app_name \\
#   --values ../configs/${app_name}-values.json
EOF
}

generate_migration_script() {
    local app_name=$1
    log_info "Generating migration script for: $app_name"
    
    cat > "$OUTPUT_DIR/scripts/migrate-${app_name}.sh" <<'EOF'
#!/bin/bash
# Migration script for APP_NAME
set -e

APP_NAME="APP_NAME"
SOURCE_TRUENAS="SOURCE_TRUENAS"
TARGET_K8S="TARGET_K8S"

echo "=== Migrating $APP_NAME ==="

# Step 1: Create namespace
echo "Creating namespace..."
kubectl create namespace $APP_NAME --dry-run=client -o yaml | kubectl apply -f -

# Step 2: Create PVCs
echo "Creating PVCs..."
kubectl apply -f ../manifests/${APP_NAME}-helm.yaml

# Step 3: Wait for PVCs to be bound
echo "Waiting for PVCs..."
kubectl wait --for=condition=Bound pvc --all -n $APP_NAME --timeout=300s

# Step 4: Copy data
echo "Copying data from TrueNAS..."
EOF
    
    # Add data copy commands
    if [ -f "$OUTPUT_DIR/data-maps/${app_name}-paths.txt" ]; then
        while IFS= read -r line; do
            mount_name=$(echo "$line" | cut -d= -f1)
            source_path=$(echo "$line" | cut -d= -f2)
            cat >> "$OUTPUT_DIR/scripts/migrate-${app_name}.sh" <<EOF

# Copy $mount_name
kubectl run -n $APP_NAME data-copy-$mount_name --image=alpine --restart=Never -- sleep 3600
kubectl wait --for=condition=Ready pod/data-copy-$mount_name -n $APP_NAME --timeout=60s
kubectl exec -n $APP_NAME data-copy-$mount_name -- mkdir -p /data
# Use rsync or kubectl cp to copy data from $source_path
# Example: kubectl cp $source_path $APP_NAME/data-copy-$mount_name:/data
kubectl delete pod data-copy-$mount_name -n $APP_NAME
EOF
        done < "$OUTPUT_DIR/data-maps/${app_name}-paths.txt"
    fi
    
    cat >> "$OUTPUT_DIR/scripts/migrate-${app_name}.sh" <<'EOF'

# Step 5: Deploy application
echo "Deploying application..."
# helm install $APP_NAME <chart-repo>/<chart> \
#   --namespace $APP_NAME \
#   --values ../configs/${APP_NAME}-values.json

echo "=== Migration complete for $APP_NAME ==="
echo "Verify with: kubectl get pods -n $APP_NAME"
EOF
    
    # Replace placeholders
    sed -i "s/APP_NAME/$app_name/g" "$OUTPUT_DIR/scripts/migrate-${app_name}.sh"
    sed -i "s|SOURCE_TRUENAS|$TRUENAS_URL|g" "$OUTPUT_DIR/scripts/migrate-${app_name}.sh"
    
    chmod +x "$OUTPUT_DIR/scripts/migrate-${app_name}.sh"
}

generate_summary() {
    log_info "Generating migration summary..."
    
    cat > "$OUTPUT_DIR/MIGRATION_SUMMARY.md" <<EOF
# Migration Summary

Generated: $(date)
Source: $TRUENAS_URL

## Applications Exported

EOF
    
    while read -r app_name; do
        local status=$(jq -r '.status' "$OUTPUT_DIR/configs/${app_name}-metadata.json")
        local chart=$(jq -r '.chart' "$OUTPUT_DIR/configs/${app_name}-metadata.json")
        local version=$(jq -r '.version' "$OUTPUT_DIR/configs/${app_name}-metadata.json")
        
        cat >> "$OUTPUT_DIR/MIGRATION_SUMMARY.md" <<EOF
### $app_name
- **Status**: $status
- **Chart**: $chart
- **Version**: $version
- **Config**: \`configs/${app_name}-values.json\`
- **Manifest**: \`manifests/${app_name}-helm.yaml\`
- **Migration Script**: \`scripts/migrate-${app_name}.sh\`

EOF
    done < "$OUTPUT_DIR/app-list.txt"
    
    cat >> "$OUTPUT_DIR/MIGRATION_SUMMARY.md" <<EOF

## Migration Steps

### Option 1: Migrate to Another TrueNAS

\`\`\`bash
# 1. Create snapshot
zfs snapshot -r tank/ix-applications@migration-$(date +%Y%m%d)

# 2. Send to target
zfs send -R tank/ix-applications@migration-$(date +%Y%m%d) | \\
  ssh target-truenas zfs receive tank/ix-applications

# 3. Import with Terraform
cd /path/to/terraform
terraform import truenas_chart_release.APP_NAME APP_NAME
\`\`\`

### Option 2: Migrate to External Kubernetes

\`\`\`bash
# For each app, run the migration script
cd scripts
./migrate-APP_NAME.sh
\`\`\`

### Option 3: Manual Migration

1. Review \`configs/APP_NAME-values.json\` for configuration
2. Review \`data-maps/APP_NAME-paths.txt\` for data locations
3. Create PVCs using \`manifests/APP_NAME-helm.yaml\`
4. Copy data from TrueNAS to PVCs
5. Deploy with Helm using the values file

## Data Backup Command

\`\`\`bash
# Backup all app data
zfs send -R tank/ix-applications@migration-$(date +%Y%m%d) | \\
  gzip > truenas-apps-backup-$(date +%Y%m%d).gz
\`\`\`

## Restore Command

\`\`\`bash
# Restore from backup
gunzip -c truenas-apps-backup-YYYYMMDD.gz | \\
  zfs receive -F tank/ix-applications
\`\`\`
EOF
    
    log_info "Summary written to: $OUTPUT_DIR/MIGRATION_SUMMARY.md"
}

# Main execution
main() {
    log_info "Starting TrueNAS Kubernetes app export..."
    
    check_requirements
    create_output_dir
    list_chart_releases
    
    while read -r app_name; do
        export_app_config "$app_name"
        generate_pvc_map "$app_name"
        generate_k8s_manifest "$app_name"
        generate_migration_script "$app_name"
    done < "$OUTPUT_DIR/app-list.txt"
    
    generate_summary
    
    log_info "Export complete! Output directory: $OUTPUT_DIR"
    log_info "Review MIGRATION_SUMMARY.md for next steps"
}

# Run main
main "$@"

