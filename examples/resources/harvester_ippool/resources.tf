resource "harvester_ippool" "service_ips" {
  name = "service-ips"

  range {
    start   = "192.168.1.10"
    end     = "192.168.1.80"
    subnet  = "192.168.1.1/24"
    gateway = "192.168.1.1"
  }

  range {
    start   = "192.168.2.10"
    end     = "192.168.2.80"
    subnet  = "192.168.2.1/24"
    gateway = "192.168.2.1"
  }

  selector {
    priority = 100
    network  = "default/vm-network"
    scope {
      project       = "services"
      namespace     = "default"
      guest_cluster = "prod-services"
    }
  }
}
