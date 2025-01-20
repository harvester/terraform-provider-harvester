terraform {
  required_version = ">= 0.13"
  required_providers {
    harvester = {
      source  = "harvester/harvester"
      version = "0.6.4"
    }
  }
}

provider "harvester" {
  # Path to kubeconfig file
  kubeconfig = "/path/to/kubeconfig.yaml"

  # Alternatively the base64 encoded contents of the kubeconfig file.
  # CAUTION: When supplying the kubeconfig as base64 encoded string, the
  # content will be preserved in the Terraform state files in the clear.
  # Take appropriate measures to avoid leaking sensitive information.
  #
  # kubeconfig = "YXBpVmVyc2lvb...xvY2FsIgo="

  kubecontext = "mycontext"
}
