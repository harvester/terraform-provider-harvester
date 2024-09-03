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
    port = 443
    protocol = "tcp"
    backend_port = 8080
  }

  listener {
    port = 80
    protocol = "tcp"
    backend_port = 8080
  }

  ipam = "ippool"
  ippool = "service-ips"

  workload_type = "vm"

  backend_selector {
    key = "app"
    values = [ "test" ]
  }

  backend_selector {
    key = "component"
    values = [ "frontend", "ui" ]
  }

  healthcheck {
    port = 443
    success_threshold = 1
    failure_threshold = 3
    period_seconds = 10
    timeout_seconds = 5
  }
}
