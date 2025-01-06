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
  # alternatively the base64 encoded contents of the kubeconfig file:
  # kubeconfig = "YXBpVmVyc2lvb...xvY2FsIgo="

  kubecontext = "mycontext"
}
