terraform {
  required_version = ">= 0.13"
  required_providers {
    harvester = {
      source  = "harvester/harvester"
      version = "0.5.3"
    }
  }
}

provider "harvester" {
}
