# Virtual Machine Scheduling with Affinity Rules

This guide explains how to use node affinity, pod affinity, and pod anti-affinity rules to control VM scheduling in Harvester.

## Overview

Harvester supports Kubernetes-native scheduling mechanisms through the `harvester_virtualmachine` resource:

- **Node Affinity**: Schedule VMs on nodes matching specific label criteria
- **Pod Affinity**: Co-locate VMs with other pods/VMs matching label criteria
- **Pod Anti-Affinity**: Separate VMs from other pods/VMs matching label criteria

Reference: [Harvester VM Node Scheduling Documentation](https://docs.harvesterhci.io/v1.7/vm/index/#node-scheduling)

## Node Affinity

Node affinity allows you to constrain which nodes a VM can be scheduled on based on node labels.

### Required Node Affinity

The VM will only be scheduled on nodes matching ALL specified criteria:

```hcl
resource "harvester_virtualmachine" "example" {
  name      = "vm-with-node-affinity"
  namespace = "default"
  cpu       = 2
  memory    = "4Gi"

  node_affinity {
    required {
      node_selector_term {
        match_expressions {
          key      = "kubernetes.io/hostname"
          operator = "In"
          values   = ["node1", "node2"]
        }
        match_expressions {
          key      = "disktype"
          operator = "In"
          values   = ["ssd"]
        }
      }
    }
  }

  disk { ... }
  network_interface { ... }
}
```

### Preferred Node Affinity

The scheduler will try to schedule on matching nodes, but will use other nodes if none match:

```hcl
resource "harvester_virtualmachine" "example" {
  name      = "vm-with-preferred-affinity"
  namespace = "default"
  cpu       = 2
  memory    = "4Gi"

  node_affinity {
    preferred {
      weight = 100
      preference {
        match_expressions {
          key      = "zone"
          operator = "In"
          values   = ["zone-a"]
        }
      }
    }
    preferred {
      weight = 50
      preference {
        match_expressions {
          key      = "zone"
          operator = "In"
          values   = ["zone-b"]
        }
      }
    }
  }

  disk { ... }
  network_interface { ... }
}
```

### Supported Operators

| Operator | Description |
|----------|-------------|
| `In` | Label value must be in the specified set |
| `NotIn` | Label value must not be in the specified set |
| `Exists` | Label key must exist (values ignored) |
| `DoesNotExist` | Label key must not exist (values ignored) |
| `Gt` | Label value must be greater than specified (numeric comparison) |
| `Lt` | Label value must be less than specified (numeric comparison) |

## Pod Affinity

Pod affinity allows you to co-locate VMs with other pods/VMs based on labels.

### Required Pod Affinity

```hcl
resource "harvester_virtualmachine" "database" {
  name      = "database-vm"
  namespace = "default"
  cpu       = 4
  memory    = "8Gi"

  pod_affinity {
    required {
      topology_key = "kubernetes.io/hostname"
      label_selector {
        match_labels = {
          "app" = "cache"
        }
      }
    }
  }

  disk { ... }
  network_interface { ... }
}
```

### Preferred Pod Affinity

```hcl
resource "harvester_virtualmachine" "app" {
  name      = "app-vm"
  namespace = "default"
  cpu       = 2
  memory    = "4Gi"

  pod_affinity {
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

  disk { ... }
  network_interface { ... }
}
```

## Pod Anti-Affinity

Pod anti-affinity prevents VMs from being scheduled on the same node as other matching pods/VMs.

### High Availability Example

Spread VMs across different nodes for high availability:

```hcl
resource "harvester_virtualmachine" "ha_vm" {
  name      = "ha-vm-1"
  namespace = "default"
  cpu       = 2
  memory    = "4Gi"

  labels = {
    "app" = "my-ha-app"
  }

  pod_anti_affinity {
    required {
      topology_key = "kubernetes.io/hostname"
      label_selector {
        match_labels = {
          "app" = "my-ha-app"
        }
      }
    }
  }

  disk { ... }
  network_interface { ... }
}
```

### Soft Anti-Affinity

Prefer spreading VMs but allow co-location if necessary:

```hcl
resource "harvester_virtualmachine" "spread_vm" {
  name      = "spread-vm"
  namespace = "default"
  cpu       = 2
  memory    = "4Gi"

  pod_anti_affinity {
    preferred {
      weight = 100
      pod_affinity_term {
        topology_key = "kubernetes.io/hostname"
        label_selector {
          match_expressions {
            key      = "harvesterhci.io/vmName"
            operator = "Exists"
          }
        }
      }
    }
  }

  disk { ... }
  network_interface { ... }
}
```

## Combining Multiple Rules

You can combine node affinity with pod affinity/anti-affinity:

```hcl
resource "harvester_virtualmachine" "complex_scheduling" {
  name      = "complex-vm"
  namespace = "default"
  cpu       = 2
  memory    = "4Gi"

  # Run only on SSD nodes in zone-a
  node_affinity {
    required {
      node_selector_term {
        match_expressions {
          key      = "disktype"
          operator = "In"
          values   = ["ssd"]
        }
        match_expressions {
          key      = "zone"
          operator = "In"
          values   = ["zone-a"]
        }
      }
    }
  }

  # Prefer co-location with cache pods
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

  # Spread across nodes from other VMs in the same app
  pod_anti_affinity {
    preferred {
      weight = 100
      pod_affinity_term {
        topology_key = "kubernetes.io/hostname"
        label_selector {
          match_labels = {
            "app" = "my-app"
          }
        }
      }
    }
  }

  disk { ... }
  network_interface { ... }
}
```

## Namespace Selector

For pod affinity/anti-affinity, you can specify namespaces to match pods from:

```hcl
pod_affinity {
  required {
    topology_key = "kubernetes.io/hostname"
    namespaces   = ["default", "production"]
    label_selector {
      match_labels = {
        "app" = "shared-service"
      }
    }
  }
}
```

Or use a namespace selector:

```hcl
pod_affinity {
  required {
    topology_key = "kubernetes.io/hostname"
    namespace_selector {
      match_labels = {
        "environment" = "production"
      }
    }
    label_selector {
      match_labels = {
        "app" = "shared-service"
      }
    }
  }
}
```

## Relationship with node_selector

The `node_selector` attribute provides a simpler way to constrain scheduling:

```hcl
# Simple node selection
node_selector = {
  "kubernetes.io/hostname" = "specific-node"
}
```

For more complex requirements, use `node_affinity`. Both can coexist, but `node_affinity` provides:
- Multiple node selector terms (OR logic)
- Soft preferences with weights
- More operators (Gt, Lt, etc.)

## Harvester-Injected Rules

Harvester automatically injects a node affinity rule to ensure VMs are scheduled on nodes with proper network configuration:

```yaml
matchExpressions:
  - key: network.harvesterhci.io/mgmt
    operator: In
    values:
      - "true"
```

This rule is added by Harvester's mutating webhook and cannot be removed. When defining your own `node_affinity`, you may want to include this rule in your Terraform configuration to avoid perpetual drift detection.

## Importing Existing VMs

When importing existing VMs with affinity rules, Terraform will read and populate all affinity configurations:

```bash
terraform import harvester_virtualmachine.example default/my-existing-vm
```

The imported state will include all node affinity, pod affinity, and pod anti-affinity rules configured on the VM.
