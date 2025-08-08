resource "harvester_storageclass" "any-replicas-1" {
  name = "any-replicas-1"

  parameters = {
    "migratable"          = "true"
    "numberOfReplicas"    = "1"
    "staleReplicaTimeout" = "30"
  }
}

resource "harvester_storageclass" "ssd-replicas-3" {
  name = "ssd-replicas-3"

  parameters = {
    "migratable"          = "true"
    "numberOfReplicas"    = "3"
    "staleReplicaTimeout" = "30"
    "diskSelector"        = "ssd,nvme"
  }
}

resource "harvester_storageclass" "lvm" {
  name = "lvm"

  volume_provisioner = "lvm.driver.harvesterhci.io"
  parameters = {
    "type"   = "striped"
    "vgName" = "test-vol"
  }

  allowed_topologies {
    match_label_expressions {
      key    = "topology.lvm.csi/node"
      values = ["harvester-node-0"]
    }
  }
}
