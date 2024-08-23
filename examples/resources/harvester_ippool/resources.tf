resource "harvester_ippool" "service_ips" {
  name = "service_ips"

  range {
    start = "10.11.0.1"
    end = "10.11.0.254"
    subnet = "10.11.0.1/24"
    gateway = "10.11.0.1"
  }

  range {
    start = "10.12.0.1"
    end = "10.12.0.254"
    subnet = "10.12.0.1/24"
    gateway = "10.12.0.1"
  }

  selector {
    priority = 100
    network = "vm-network"
    scope {
      project = "services"
      namespace = "prod-default"
      guest_cluster = "prod-services"
    }
  }
}
