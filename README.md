Terraform Provider for Harvester
==================================

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
- [Go](https://golang.org/doc/install) 1.16 to build the provider plugin

## Install The Provider

copy and paste this code into your Terraform configuration. Then, run `terraform init` to initialize it.
```hcl
terraform {
  required_providers {
    harvester = {
      source = "harvester/harvester"
      version = "<replace to the latest release version>"
    }
  }
}

provider "harvester" {
  # Configuration options
}
```

## Using the provider

Documentation about the provider specific configuration options can be found on the [docs directory](docs).