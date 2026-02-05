# Example: VM Scheduling with Affinity Rules
#
# This example demonstrates how to use node affinity, pod affinity,
# and pod anti-affinity rules to control VM scheduling in Harvester.

# Example 1: Node Affinity - Schedule on specific nodes
resource "harvester_virtualmachine" "node_affinity_example" {
  name        = "vm-node-affinity"
  namespace   = "default"
  description = "VM with node affinity scheduling"

  cpu    = 2
  memory = "4Gi"

  run_strategy = "RerunOnFailure"

  # Required: VM will only run on nodes matching these criteria
  node_affinity {
    required {
      node_selector_term {
        match_expressions {
          key      = "kubernetes.io/hostname"
          operator = "In"
          values   = ["node1", "node2"]
        }
      }
    }
    # Preferred: Scheduler will try SSD nodes first (weight 100)
    preferred {
      weight = 100
      preference {
        match_expressions {
          key      = "disktype"
          operator = "In"
          values   = ["ssd"]
        }
      }
    }
  }

  network_interface {
    name         = "nic-1"
    network_name = "default/vlan1"
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "20Gi"
    bus        = "virtio"
    boot_order = 1
    image      = "default/ubuntu-22.04"
  }
}

# Example 2: Pod Anti-Affinity for High Availability
# This ensures VMs with the same label are spread across different nodes
resource "harvester_virtualmachine" "ha_vm_1" {
  name        = "ha-app-1"
  namespace   = "default"
  description = "HA VM instance 1"

  cpu    = 2
  memory = "4Gi"

  run_strategy = "RerunOnFailure"

  tags = {
    app = "my-ha-application"
  }

  # Hard requirement: Never schedule on the same node as other VMs with app=my-ha-application
  pod_anti_affinity {
    required {
      topology_key = "kubernetes.io/hostname"
      label_selector {
        match_labels = {
          "tag.harvesterhci.io/app" = "my-ha-application"
        }
      }
    }
  }

  network_interface {
    name         = "nic-1"
    network_name = "default/vlan1"
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "20Gi"
    bus        = "virtio"
    boot_order = 1
    image      = "default/ubuntu-22.04"
  }
}

resource "harvester_virtualmachine" "ha_vm_2" {
  name        = "ha-app-2"
  namespace   = "default"
  description = "HA VM instance 2"

  cpu    = 2
  memory = "4Gi"

  run_strategy = "RerunOnFailure"

  tags = {
    app = "my-ha-application"
  }

  # Same anti-affinity rule - will be scheduled on a different node than ha_vm_1
  pod_anti_affinity {
    required {
      topology_key = "kubernetes.io/hostname"
      label_selector {
        match_labels = {
          "tag.harvesterhci.io/app" = "my-ha-application"
        }
      }
    }
  }

  network_interface {
    name         = "nic-1"
    network_name = "default/vlan1"
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "20Gi"
    bus        = "virtio"
    boot_order = 1
    image      = "default/ubuntu-22.04"
  }

  depends_on = [harvester_virtualmachine.ha_vm_1]
}

# Example 3: Pod Affinity - Co-locate VMs
# This ensures the database VM runs on the same node as the cache
resource "harvester_virtualmachine" "database_vm" {
  name        = "database"
  namespace   = "default"
  description = "Database VM co-located with cache"

  cpu    = 4
  memory = "8Gi"

  run_strategy = "RerunOnFailure"

  # Prefer (soft) co-location with cache pods for low latency
  pod_affinity {
    preferred {
      weight = 100
      pod_affinity_term {
        topology_key = "kubernetes.io/hostname"
        label_selector {
          match_labels = {
            "app" = "cache"
          }
        }
      }
    }
  }

  network_interface {
    name         = "nic-1"
    network_name = "default/vlan1"
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "100Gi"
    bus        = "virtio"
    boot_order = 1
    image      = "default/ubuntu-22.04"
  }
}

# Example 4: Combined Scheduling Rules
# Complex scheduling with multiple constraints
resource "harvester_virtualmachine" "complex_scheduling" {
  name        = "complex-vm"
  namespace   = "default"
  description = "VM with complex scheduling requirements"

  cpu    = 2
  memory = "4Gi"

  run_strategy = "RerunOnFailure"

  labels = {
    tier = "frontend"
  }

  # Node affinity: Must be on SSD nodes, prefer zone-a
  node_affinity {
    required {
      node_selector_term {
        match_expressions {
          key      = "disktype"
          operator = "In"
          values   = ["ssd"]
        }
      }
    }
    preferred {
      weight = 80
      preference {
        match_expressions {
          key      = "zone"
          operator = "In"
          values   = ["zone-a"]
        }
      }
    }
  }

  # Pod affinity: Try to be near cache pods
  pod_affinity {
    preferred {
      weight = 50
      pod_affinity_term {
        topology_key = "kubernetes.io/hostname"
        label_selector {
          match_labels = {
            "app" = "cache"
          }
        }
      }
    }
  }

  # Pod anti-affinity: Spread frontend VMs across nodes
  pod_anti_affinity {
    preferred {
      weight = 100
      pod_affinity_term {
        topology_key = "kubernetes.io/hostname"
        label_selector {
          match_labels = {
            "tier" = "frontend"
          }
        }
      }
    }
  }

  network_interface {
    name         = "nic-1"
    network_name = "default/vlan1"
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "20Gi"
    bus        = "virtio"
    boot_order = 1
    image      = "default/ubuntu-22.04"
  }
}
