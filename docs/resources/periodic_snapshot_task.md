---
page_title: "truenas_periodic_snapshot_task Resource - terraform-provider-truenas"
subcategory: "Storage"
description: |-
  Manages a periodic snapshot task on TrueNAS.
---

# truenas_periodic_snapshot_task (Resource)

Manages a periodic snapshot task on TrueNAS. This resource automates the creation of ZFS snapshots on a schedule, providing data protection and point-in-time recovery capabilities.

## Example Usage

### Daily Snapshot Task

```terraform
resource "truenas_periodic_snapshot_task" "daily_backup" {
  dataset        = "tank/data"
  recursive      = true
  enabled        = true
  naming_schema  = "auto-%Y-%m-%d_%H-%M"
  schedule       = "0 2 * * *"  # Daily at 2 AM
  lifetime_value = 7
  lifetime_unit  = "DAY"
}
```

### Hourly Snapshot Task

```terraform
resource "truenas_periodic_snapshot_task" "hourly_backup" {
  dataset        = "tank/critical"
  recursive      = true
  enabled        = true
  naming_schema  = "hourly-%Y-%m-%d_%H-%M"
  schedule       = "0 * * * *"  # Every hour
  lifetime_value = 24
  lifetime_unit  = "HOUR"
}
```

### Weekly Snapshot Task

```terraform
resource "truenas_periodic_snapshot_task" "weekly_backup" {
  dataset        = "tank/archives"
  recursive      = true
  enabled        = true
  naming_schema  = "weekly-%Y-%m-%d"
  schedule       = "0 3 * * 0"  # Sunday at 3 AM
  lifetime_value = 4
  lifetime_unit  = "WEEK"
}
```

### Monthly Snapshot Task

```terraform
resource "truenas_periodic_snapshot_task" "monthly_backup" {
  dataset        = "tank/monthly"
  recursive      = true
  enabled        = true
  naming_schema  = "monthly-%Y-%m-%d"
  schedule       = "0 4 1 * *"  # 1st of month at 4 AM
  lifetime_value = 12
  lifetime_unit  = "MONTH"
}
```

### Snapshot Task with Exclusions

```terraform
resource "truenas_periodic_snapshot_task" "backup_with_exclusions" {
  dataset        = "tank/data"
  recursive      = true
  enabled        = true
  naming_schema  = "backup-%Y-%m-%d_%H-%M"
  schedule       = "0 1 * * *"  # Daily at 1 AM
  lifetime_value = 14
  lifetime_unit  = "DAY"
  
  exclude = [
    "tank/data/temp",
    "tank/data/cache",
    "tank/data/logs"
  ]
}
```

### Non-recursive Snapshot Task

```terraform
resource "truenas_periodic_snapshot_task" "single_dataset" {
  dataset        = "tank/important"
  recursive      = false
  enabled        = true
  naming_schema  = "single-%Y-%m-%d_%H-%M"
  schedule       = "*/15 * * * *"  # Every 15 minutes
  lifetime_value = 48
  lifetime_unit  = "HOUR"
}
```

### Snapshot Task with Empty Snapshots

```terraform
resource "truenas_periodic_snapshot_task" "allow_empty" {
  dataset        = "tank/development"
  recursive      = true
  enabled        = true
  naming_schema  = "dev-%Y-%m-%d_%H-%M"
  schedule       = "0 */6 * * *"  # Every 6 hours
  lifetime_value = 7
  lifetime_unit  = "DAY"
  allow_empty    = true
}
```

### Complex Schedule Example

```terraform
resource "truenas_periodic_snapshot_task" "complex_schedule" {
  dataset        = "tank/production"
  recursive      = true
  enabled        = true
  naming_schema  = "prod-%Y-%m-%d_%H-%M"
  schedule       = "0 2,14 * * 1-5"  # Weekdays at 2 AM and 2 PM
  lifetime_value = 30
  lifetime_unit  = "DAY"
}
```

## Schema

### Required

- `dataset` (String) Dataset to snapshot (e.g., tank/mydata).
- `naming_schema` (String) Naming schema for snapshots (e.g., auto-%Y-%m-%d_%H-%M).
- `schedule` (String) Cron schedule format (minute hour day month weekday).
- `lifetime_value` (Number) How long to keep snapshots.
- `lifetime_unit` (String) Lifetime unit. Options: `HOUR`, `DAY`, `WEEK`, `MONTH`, `YEAR`.

### Optional

- `recursive` (Boolean) Create recursive snapshots of all children. Default: true.
- `exclude` (List of String) List of child datasets to exclude from recursive snapshots.
- `enabled` (Boolean) Enable this snapshot task. Default: true.
- `allow_empty` (Boolean) Allow taking empty snapshots. Default: false.

### Read-Only

- `id` (String) Task identifier.

## Import

Periodic snapshot tasks can be imported using task ID:

```shell
terraform import truenas_periodic_snapshot_task.existing 1
```

## Notes

### Cron Schedule Format

The schedule uses standard cron format: `minute hour day month weekday`

#### Fields
- **minute**: 0-59
- **hour**: 0-23
- **day**: 1-31
- **month**: 1-12
- **weekday**: 0-7 (0 and 7 are Sunday)

#### Special Characters
- **\***: Any value
- **,**: Value list separator
- **-**: Range of values
- **/**: Step values

#### Common Schedules

```terraform
# Every minute
schedule = "* * * * *"

# Every hour at minute 0
schedule = "0 * * * *"

# Every day at 2:30 AM
schedule = "30 2 * * *"

# Every Monday at 3 AM
schedule = "0 3 * * 1"

# Every weekday at 9 AM
schedule = "0 9 * * 1-5"

# Every 15 minutes
schedule = "*/15 * * * *"

# Every 2 hours
schedule = "0 */2 * * *"

# First of month at midnight
schedule = "0 0 1 * *"

# Work hours (9 AM to 5 PM) every hour
schedule = "0 9-17 * * 1-5"
```

### Naming Schema

Use strftime format codes for snapshot names:

#### Common Format Codes
- **%Y**: 4-digit year (2024)
- **%m**: 2-digit month (01-12)
- **%d**: 2-digit day (01-31)
- **%H**: 2-digit hour (00-23)
- **%M**: 2-digit minute (00-59)

#### Examples

```terraform
# Daily snapshots
naming_schema = "daily-%Y-%m-%d"

# Hourly snapshots
naming_schema = "hourly-%Y-%m-%d_%H-%M"

# Weekly snapshots
naming_schema = "weekly-%Y-%U"  # %U = week number

# Custom format
naming_schema = "backup-%Y%m%d-%H%M"
```

### Lifetime Management

Configure how long snapshots are retained:

```terraform
# Keep for 24 hours
lifetime_value = 24
lifetime_unit  = "HOUR"

# Keep for 7 days
lifetime_value = 7
lifetime_unit  = "DAY"

# Keep for 4 weeks
lifetime_value = 4
lifetime_unit  = "WEEK"

# Keep for 6 months
lifetime_value = 6
lifetime_unit  = "MONTH"

# Keep for 2 years
lifetime_value = 2
lifetime_unit  = "YEAR"
```

### Recursive Snapshots

Control whether to snapshot child datasets:

```terraform
# Snapshot dataset and all children
recursive = true

# Snapshot only the specified dataset
recursive = false
```

### Dataset Exclusions

Exclude specific child datasets from recursive snapshots:

```terraform
recursive = true
exclude = [
  "tank/data/temp",           # Exclude temp directory
  "tank/data/cache",          # Exclude cache directory
  "tank/data/logs",           # Exclude logs directory
  "tank/data/scratch"         # Exclude scratch space
]
```

### Empty Snapshots

Control whether to create snapshots when no data changed:

```terraform
# Don't create empty snapshots (default)
allow_empty = false

# Create empty snapshots
allow_empty = true
```

Use cases for empty snapshots:
- Maintain consistent schedule
- Track snapshot creation times
- Ensure backup system continuity

## Best Practices

### Planning

1. **Schedule Design**: Balance frequency with storage requirements
2. **Retention Policy**: Align lifetime with business requirements
3. **Naming Convention**: Use consistent, descriptive naming
4. **Resource Planning**: Consider storage space for snapshots

### Performance

1. **Off-Peak Hours**: Schedule snapshots during low usage periods
2. **Frequency Balance**: Avoid excessive snapshot creation
3. **Storage Monitoring**: Monitor snapshot storage usage
4. **Performance Impact**: Consider impact on system performance

### Data Protection

1. **Multiple Frequencies**: Use different schedules for different data types
2. **Retention Strategy**: Keep snapshots for appropriate time periods
3. **Testing**: Regularly test snapshot restoration
4. **Monitoring**: Monitor snapshot creation and deletion

### Maintenance

1. **Regular Reviews**: Periodically review snapshot policies
2. **Storage Management**: Monitor and manage snapshot storage
3. **Schedule Optimization**: Adjust schedules based on usage patterns
4. **Documentation**: Document snapshot policies and procedures

## Troubleshooting

### Snapshots Not Creating

1. Verify task is enabled
2. Check cron schedule syntax
3. Ensure dataset exists
4. Review system logs

### Schedule Issues

1. Validate cron format
2. Check system time and timezone
3. Test schedule manually
4. Review cron expression

### Storage Issues

1. Monitor available storage space
2. Check snapshot retention settings
3. Review dataset usage
4. Consider storage expansion

### Permission Issues

1. Verify dataset permissions
2. Check user access rights
3. Review system permissions
4. Test manual snapshot creation

### Import Issues

1. Use correct task ID
2. Verify task exists in TrueNAS
3. Check task configuration
4. Ensure proper permissions

### Performance Problems

1. Monitor system resources
2. Check storage performance
3. Review snapshot frequency
4. Consider off-peak scheduling

## See Also

- [truenas_snapshot](snapshot) - Manual snapshot management
- [truenas_dataset](dataset) - Dataset management
- [TrueNAS Snapshot Documentation](https://www.truenas.com/docs/scale/snapshots/) - Official snapshot documentation
- [ZFS Snapshot Best Practices](https://www.truenas.com/docs/scale/snapshots/bestpractices/) - Snapshot optimization and management