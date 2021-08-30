terraform {
  required_version = ">= 0.13"
  required_providers {
    harvester = {
      source  = "registry.terraform.io/harvester/harvester"
      version = "~> 0.3.0"
    }
  }
}

provider "harvester" {
}
