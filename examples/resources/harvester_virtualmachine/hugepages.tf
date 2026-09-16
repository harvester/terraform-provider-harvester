# VM backed by 2Mi huge pages for memory-intensive workloads.
#
# The Harvester node must have huge pages pre-allocated (for example with
# vm.nr_hugepages on the host). The memory request must also be a multiple of
# the page size once the Harvester overcommit ratio has been applied: with the
# default 150% memory overcommit, "3Gi" requests 2Gi from the node and works,
# while "2Gi" would request 1365Mi and be rejected by KubeVirt.
resource "harvester_virtualmachine" "hugepages" {
  name        = "hugepages-vm"
  namespace   = "default"
  description = "VM using 2Mi huge pages"

  cpu    = 2
  memory = "3Gi"

  hugepages = "2Mi"

  run_strategy = "RerunOnFailure"
  machine_type = "q35"

  network_interface {
    name = "nic-1"
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "20Gi"
    bus        = "virtio"
    boot_order = 1
    image      = "default/ubuntu-24.04"
  }
}
