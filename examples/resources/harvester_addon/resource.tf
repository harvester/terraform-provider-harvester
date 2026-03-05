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
