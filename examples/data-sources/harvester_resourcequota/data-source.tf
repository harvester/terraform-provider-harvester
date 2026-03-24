data "harvester_resourcequota" "dev_limits" {
  name      = "dev-limits"
  namespace = "dev-environment"
}
