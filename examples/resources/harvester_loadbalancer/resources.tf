resource "harvester_loadbalancer" "service_loadbalancer" {
  name = "service-loadbalancer"

  # This ensures correct ordering for the creation of the resources.
  # The loadbalancer resource will be rejected by the admission webhook, if not
  # at least one virtual machine with labels matching the backend_selector(s)
  # already exists. This dependency ordering can be used to create that virtual
  # machine with the same Terraform file.
  depends_on = [
    harvester_virtualmachine.name
  ]

  listener {
    # Each listener must have a unique name
    name         = "https"
    port         = 443
    protocol     = "tcp"
    backend_port = 8443
  }

  listener {
    name         = "http"
    port         = 80
    protocol     = "tcp"
    backend_port = 8080
  }

  # Can be "pool" or "dhcp"
  ipam = "pool"

  # Only applicable if ipam="pool"
  ippool = "service-ips"

  # Can be "vm" or "cluster"
  workload_type = "vm"

  # This must be a label on the VirtualMachineInstance
  backend_selector {
    key    = "harvesterhci.io/vmName"
    values = ["testVM"]
  }

  healthcheck {
    # Must be the same as one of the listener backend ports
    port = 8080

    success_threshold = 1
    failure_threshold = 3
    period_seconds    = 10
    timeout_seconds   = 5
  }
}
