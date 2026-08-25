# Built-in addons already exist on the cluster: the resource adopts them.
# Destroying an adopted addon only disables it, the addon object is kept.
# Note that enabled is declarative: adopting an addon without enabled = true
# disables it.
resource "harvester_addon" "pcidevices_controller" {
  name    = "pcidevices-controller"
  enabled = true
}

resource "harvester_addon" "vm_import_controller" {
  name    = "vm-import-controller"
  enabled = true

  values_content = <<-EOF
    resources:
      limits:
        memory: 256Mi
  EOF
}

# Heavy addons can take several minutes to converge; raise the timeouts when
# needed (create/update/delete default to 10 minutes).
resource "harvester_addon" "rancher_monitoring" {
  name    = "rancher-monitoring"
  enabled = true

  timeouts {
    create = "20m"
    update = "20m"
    delete = "25m"
  }
}

# When no addon with the given name exists, a custom experimental addon is
# created from repo/chart/version (all three are then required). The chart of
# an existing addon cannot be changed afterwards. Destroying an addon created
# by Terraform deletes the addon object itself.
resource "harvester_addon" "custom" {
  name    = "my-custom-addon"
  enabled = true

  repo    = "https://charts.example.com"
  chart   = "my-chart"
  version = "1.2.3"
}
