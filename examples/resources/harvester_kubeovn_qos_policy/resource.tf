resource "harvester_kubeovn_qos_policy" "example" {
  name         = "example-qos"
  shared       = true
  binding_type = "EIP"

  bandwidth_limit_rules {
    name      = "ingress-limit"
    rate_max  = "100M"
    burst_max = "200M"
    priority  = 1
    direction = "ingress"
  }
}
