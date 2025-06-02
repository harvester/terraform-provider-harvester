data "harvester_ippool" "service_ips" {
  name = "service_ips"

  range {
    start = "192.168.0.1"
    end = "192.168.0.254"
    subnet = "192.168.0.1/24"
    gateway = "192.168.0.1"
  }
}
