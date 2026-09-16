resource "harvester_namespace" "example" {
  name = "my-namespace"

  tags = {
    env = "dev"
  }

  # Set to true to allow terraform destroy to delete the namespace.
  # Default is false to prevent accidental deletion.
  # delete_on_destroy = true
}
