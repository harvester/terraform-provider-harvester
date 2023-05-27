terraform {
  required_providers {
    harvester = {
      source = "terraform.local/local/harvester"
      version = "0.0.0-dev"
    }
  }
}
provider "harvester" {
  kubeconfig = "kubeconfig"
}
