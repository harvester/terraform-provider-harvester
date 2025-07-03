resource "harvester_image" "k3os" {
  name      = "k3os"
  namespace = "harvester-public"

  display_name = "k3os"
  source_type  = "download"
  url          = "https://github.com/rancher/k3os/releases/download/v0.20.6-k3s1r0/k3os-amd64.iso"
}

resource "harvester_image" "ubuntu20" {
  name      = "ubuntu20"
  namespace = "harvester-public"

  display_name = "ubuntu-20.04-server-cloudimg-amd64.img"
  source_type  = "download"
  url          = "http://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.img"
}

resource "harvester_image" "ubuntu20-any-1" {
  name      = "ubuntu20-any-1"
  namespace = "harvester-public"

  storage_class_name = harvester_storageclass.any-replicas-1.name

  display_name = "ubuntu20-any-1"
  source_type  = "download"
  url          = "http://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.img"
}

resource "harvester_image" "opensuse154" {
  name      = "opensuse154"
  namespace = "harvester-public"

  display_name = "openSUSE-Leap-15.4.x86_64-NoCloud.qcow2"
  source_type  = "download"
  url          = "https://downloadcontent-us1.opensuse.org/repositories/Cloud:/Images:/Leap_15.4/images/openSUSE-Leap-15.4.x86_64-NoCloud.qcow2"
}

resource "harvester_image" "opensuse154-ssd-3" {
  name      = "opensuse154-ssd-3"
  namespace = "harvester-public"

  storage_class_name = harvester_storageclass.ssd-replicas-3.name

  display_name = "opensuse154-ssd-3"
  source_type  = "download"
  url          = "https://downloadcontent-us1.opensuse.org/repositories/Cloud:/Images:/Leap_15.4/images/openSUSE-Leap-15.4.x86_64-NoCloud.qcow2"
}

resource "kubernetes_secret_v1" "crypto_default" {
  metadata {
    name      = "crypto"
    namespace = "default"
  }

  type = "Opaque"

  data = {
    CRYPTO_KEY_VALUE    = "your-encryption-passphrase-here"
    CRYPTO_KEY_CIPHER   = "aes-xts-plain64"
    CRYPTO_KEY_HASH     = "sha256"
    CRYPTO_KEY_PROVIDER = "secret"
    CRYPTO_KEY_SIZE     = 256
    CRYPTO_PBKDF        = "argon2i"
  }
}

resource "harvester_storageclass" "encryption" {
  name = "encryption"

  parameters = {
    "migratable"                                       = "true"
    "numberOfReplicas"                                 = "1"
    "staleReplicaTimeout"                              = "30"
    "encrypted"                                        = "true"
    "csi.storage.k8s.io/node-publish-secret-name"      = kubernetes_secret_v1.crypto_default.metadata[0].name
    "csi.storage.k8s.io/node-publish-secret-namespace" = kubernetes_secret_v1.crypto_default.metadata[0].namespace
    "csi.storage.k8s.io/node-stage-secret-name"        = kubernetes_secret_v1.crypto_default.metadata[0].name
    "csi.storage.k8s.io/node-stage-secret-namespace"   = kubernetes_secret_v1.crypto_default.metadata[0].namespace
    "csi.storage.k8s.io/provisioner-secret-name"       = kubernetes_secret_v1.crypto_default.metadata[0].name
    "csi.storage.k8s.io/provisioner-secret-namespace"  = kubernetes_secret_v1.crypto_default.metadata[0].namespace
  }
}

resource "harvester_image" "encrypted_image" {
  namespace          = "default"
  name               = "encrypted-ubuntu"
  display_name       = "encrypted-ubuntu"
  source_type        = "clone"
  storage_class_name = harvester_storageclass.encryption.name

  security_parameters = {
    crypto_operation       = "encrypt"
    source_image_name      = harvester_image.ubuntu20.name
    source_image_namespace = harvester_image.ubuntu20.namespace
  }
}

# Example: Image decryption
resource "harvester_image" "decrypted_image" {
  namespace    = "default"
  name         = "decrypted-ubuntu"
  display_name = "decrypted-ubuntu"
  source_type  = "clone"

  security_parameters = {
    crypto_operation       = "decrypt"
    source_image_name      = harvester_image.encrypted_image.name
    source_image_namespace = harvester_image.encrypted_image.namespace
  }
}