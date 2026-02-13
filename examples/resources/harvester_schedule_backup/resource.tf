# Example: Create a recurring VM backup schedule

resource "harvester_schedule_backup" "daily" {
  name      = "daily-vm-backup"
  namespace = "default"

  vm_name  = "default/my-vm"
  schedule = "0 2 * * *" # Daily at 2 AM UTC
  retain   = 7           # Keep 7 backups
  enabled  = true
}

# Example: Backup with labels
resource "harvester_schedule_backup" "labeled" {
  name      = "labeled-vm-backup"
  namespace = "default"

  vm_name  = "default/my-vm"
  schedule = "0 */6 * * *" # Every 6 hours UTC
  retain   = 5

  labels = {
    environment = "production"
    team        = "devops"
  }
}
