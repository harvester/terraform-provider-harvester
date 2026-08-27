# VM with explicit CPU topology for NUMA-aware workloads
resource "harvester_virtualmachine" "cpu_topology_example" {
  name      = "cpu-topology-vm"
  namespace = "default"

  description = "VM with explicit CPU socket/thread topology"

  cpu    = 4
  memory = "8Gi"

  # Expose 2 sockets with 2 cores each (total: 4 vCPUs)
  cpu_sockets = 2
  cpu_threads = 1

  run_strategy = "RerunOnFailure"
  machine_type = "q35"

  network_interface {
    name = "nic-1"
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "40Gi"
    bus        = "virtio"
    boot_order = 1
    image      = "default/ubuntu-24.04"
  }
}

# VM with hyper-threading (2 threads per core)
resource "harvester_virtualmachine" "hyperthreaded" {
  name      = "hyperthreaded-vm"
  namespace = "default"

  description = "VM simulating hyper-threading with 2 threads per core"

  cpu    = 8
  memory = "16Gi"

  # 1 socket, 4 cores, 2 threads = 8 logical CPUs
  cpu_sockets = 1
  cpu_threads = 2

  run_strategy = "RerunOnFailure"
  machine_type = "q35"

  network_interface {
    name = "nic-1"
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "80Gi"
    bus        = "virtio"
    boot_order = 1
    image      = "default/ubuntu-24.04"
  }
}
