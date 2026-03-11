resource "harvester_kubeovn_switch_lb_rule" "example" {
  name = "example-slr"
  vip  = "example-vip"

  ports {
    name     = "http"
    port     = 80
    protocol = "TCP"
  }
}
