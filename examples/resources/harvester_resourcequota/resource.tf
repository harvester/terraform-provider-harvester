# Limit resource usage in a namespace
resource "harvester_resourcequota" "dev_limits" {
  name      = "dev-limits"
  namespace = "dev-environment"

  description = "Resource limits for the development namespace"

  # Total CPU and memory that can be requested by all VMs
  hard = {
    "requests.cpu"    = "16"
    "requests.memory" = "32Gi"
    "limits.cpu"      = "32"
    "limits.memory"   = "64Gi"
  }
}

# Restrict number of PVCs and total storage in a namespace
resource "harvester_resourcequota" "storage_limits" {
  name      = "storage-limits"
  namespace = "dev-environment"

  hard = {
    "persistentvolumeclaims" = "20"
    "requests.storage"       = "500Gi"
  }
}
