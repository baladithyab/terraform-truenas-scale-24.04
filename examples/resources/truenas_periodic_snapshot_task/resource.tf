# Hourly snapshots kept for 1 week
resource "truenas_periodic_snapshot_task" "hourly" {
  dataset        = "tank/mydata"
  recursive      = true
  enabled        = true
  naming_schema  = "auto-%Y-%m-%d_%H-%M"
  lifetime_value = 1
  lifetime_unit  = "WEEK"
  
  # Every hour at minute 0
  schedule = jsonencode({
    minute = "0"
    hour   = "*"
    dom    = "*"
    month  = "*"
    dow    = "*"
  })
}

# Daily snapshots kept for 1 month
resource "truenas_periodic_snapshot_task" "daily" {
  dataset        = "tank/important"
  recursive      = true
  enabled        = true
  naming_schema  = "daily-%Y-%m-%d"
  lifetime_value = 1
  lifetime_unit  = "MONTH"
  
  # Every day at 2:00 AM
  schedule = jsonencode({
    minute = "0"
    hour   = "2"
    dom    = "*"
    month  = "*"
    dow    = "*"
  })
}

# Weekly snapshots kept for 3 months
resource "truenas_periodic_snapshot_task" "weekly" {
  dataset        = "tank/archive"
  recursive      = true
  enabled        = true
  naming_schema  = "weekly-%Y-W%W"
  lifetime_value = 3
  lifetime_unit  = "MONTH"
  
  # Every Sunday at 3:00 AM
  schedule = jsonencode({
    minute = "0"
    hour   = "3"
    dom    = "*"
    month  = "*"
    dow    = "0"
  })
}

# Snapshot with exclusions
resource "truenas_periodic_snapshot_task" "selective" {
  dataset        = "tank/data"
  recursive      = true
  enabled        = true
  naming_schema  = "auto-%Y%m%d-%H%M"
  lifetime_value = 2
  lifetime_unit  = "WEEK"
  
  # Exclude temporary and cache directories
  exclude = [
    "tank/data/temp",
    "tank/data/cache"
  ]
  
  # Every 6 hours
  schedule = jsonencode({
    minute = "0"
    hour   = "*/6"
    dom    = "*"
    month  = "*"
    dow    = "*"
  })
}

# Monthly snapshots kept for 1 year
resource "truenas_periodic_snapshot_task" "monthly" {
  dataset        = "tank/longterm"
  recursive      = false
  enabled        = true
  naming_schema  = "monthly-%Y-%m"
  lifetime_value = 1
  lifetime_unit  = "YEAR"
  allow_empty    = true
  
  # First day of every month at 4:00 AM
  schedule = jsonencode({
    minute = "0"
    hour   = "4"
    dom    = "1"
    month  = "*"
    dow    = "*"
  })
}

# Import an existing periodic snapshot task
# terraform import truenas_periodic_snapshot_task.existing 1

