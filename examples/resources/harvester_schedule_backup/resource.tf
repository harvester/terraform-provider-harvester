# Example: Create a volume with a recurring backup

# First, create a volume
resource "harvester_volume" "example" {
  name      = "example-volume"
  namespace = "default"

  size = "10Gi"
}

# Configure a recurring backup for the volume
resource "harvester_schedule_backup" "example" {
  name      = "example-volume-backup"
  namespace = "default"

  vm_name     = "default/example-vm"
  schedule    = "0 2 * * *" # Daily at 2 AM
  retain      = 7           # Keep 7 backups
  concurrency = 1
  enabled     = true
}

# Example: Backup with labels and groups
resource "harvester_volume_backup" "example_with_labels" {
  name      = "example-volume-backup-labeled"
  namespace = "default"

  vm_name     = "default/example-vm"
  schedule    = "0 */6 * * *" # Every 6 hours
  retain      = 5
  concurrency = 2

  labels = {
    environment = "production"
    team        = "devops"
  }

  groups  = ["backup-group-1"]
  enabled = true
}

