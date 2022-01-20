resource "harvester_ssh_key" "mysshkey" {
  name      = "mysshkey"
  namespace = "default"

  public_key = "your ssh public key"
}