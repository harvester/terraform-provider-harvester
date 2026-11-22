# kube-ovn names the IP of a VM <vm name>.<vm namespace> on the default
# network, and <vm name>.<vm namespace>.<network name>.<network namespace>.ovn
# on an overlay network.
data "harvester_kubeovn_ip" "example" {
  name = "my-vm.default.my-overlay-network.default.ovn"
}
