# Terraform Provider Harvester - Instructions d'installation

## Version: 1.7.1

Ce package contient le binaire compilé du Terraform Provider Harvester pour Linux amd64.

## 📦 Contenu

- `terraform-provider-harvester` : Binaire exécutable du provider
- `README_BINARY.md` : Ce fichier d'instructions

## 🚀 Installation

### Méthode 1 : Installation manuelle (recommandée)

1. **Créer le répertoire de plugins Terraform** (si nécessaire) :
   ```bash
   mkdir -p ~/.terraform.d/plugins/registry.terraform.io/harvester/harvester/1.7.1/linux_amd64
   ```

2. **Copier le binaire** :
   ```bash
   cp terraform-provider-harvester ~/.terraform.d/plugins/registry.terraform.io/harvester/harvester/1.7.1/linux_amd64/
   ```

3. **Rendre le binaire exécutable** :
   ```bash
   chmod +x ~/.terraform.d/plugins/registry.terraform.io/harvester/harvester/1.7.1/linux_amd64/terraform-provider-harvester
   ```

### Méthode 2 : Installation dans le répertoire de travail Terraform

1. **Créer le répertoire** :
   ```bash
   mkdir -p .terraform/plugins/registry.terraform.io/harvester/harvester/1.7.1/linux_amd64
   ```

2. **Copier le binaire** :
   ```bash
   cp terraform-provider-harvester .terraform/plugins/registry.terraform.io/harvester/harvester/1.7.1/linux_amd64/
   ```

3. **Rendre le binaire exécutable** :
   ```bash
   chmod +x .terraform/plugins/registry.terraform.io/harvester/harvester/1.7.1/linux_amd64/terraform-provider-harvester
   ```

### Méthode 3 : Installation système (optionnel)

```bash
# Copier dans un répertoire système (nécessite les droits root)
sudo cp terraform-provider-harvester /usr/local/bin/
sudo chmod +x /usr/local/bin/terraform-provider-harvester
```

## 📝 Configuration Terraform

Dans votre fichier `main.tf` ou `versions.tf`, spécifiez la version du provider :

```hcl
terraform {
  required_providers {
    harvester = {
      source  = "harvester/harvester"
      version = "1.7.1"
    }
  }
}
```

## 🔧 Configuration du provider

Exemple de configuration basique :

```hcl
provider "harvester" {
  kubeconfig = "~/.kube/config"
  # ou
  # kubeconfig_base64 = "base64_encoded_kubeconfig"
}
```

## ✅ Vérification de l'installation

1. **Initialiser Terraform** :
   ```bash
   terraform init
   ```

2. **Vérifier que le provider est reconnu** :
   ```bash
   terraform providers
   ```

Vous devriez voir :
```
Providers required by configuration:
.
└── provider[registry.terraform.io/harvester/harvester] 1.7.1
```

## 🎯 Utilisation

### Exemple : Créer une VM avec backup récurrent

```hcl
resource "harvester_virtualmachine" "test-vm" {
  name        = "test-vm"
  namespace   = "default"
  description = "Test VM with backup"
  
  cpu    = 2
  memory = "4Gi"
  
  disk {
    name       = "disk-1"
    type       = "disk"
    size       = "20Gi"
    bus        = "virtio"
    boot_order = 1
    image      = "harvester-public/image-ubuntu20.04"
    auto_delete = true
  }
  
  network_interface {
    name         = "nic-1"
    network_name = "vlan1"
  }
}

resource "harvester_schedule_backup" "vm_backup" {
  name        = "test-vm-backup"
  namespace   = "default"
  vm_name     = "${harvester_virtualmachine.test-vm.namespace}/${harvester_virtualmachine.test-vm.name}"
  schedule    = "0 2 * * *"  # Tous les jours à 2h UTC
  retain      = 5
  enabled     = true
  
  labels = {
    environment = "production"
    managed-by  = "terraform"
  }
}
```

## 🔍 Dépannage

### Le provider n'est pas trouvé

1. Vérifier que le binaire est au bon emplacement :
   ```bash
   ls -la ~/.terraform.d/plugins/registry.terraform.io/harvester/harvester/1.7.1/linux_amd64/
   ```

2. Vérifier les permissions :
   ```bash
   chmod +x ~/.terraform.d/plugins/registry.terraform.io/harvester/harvester/1.7.1/linux_amd64/terraform-provider-harvester
   ```

3. Nettoyer le cache Terraform :
   ```bash
   rm -rf .terraform
   terraform init
   ```

### Erreur de connexion à Harvester

1. Vérifier que `kubeconfig` est correctement configuré
2. Vérifier que vous avez accès au cluster Harvester :
   ```bash
   kubectl get nodes
   ```

### Erreur de version

Si Terraform ne trouve pas la bonne version, vérifier que :
- Le répertoire correspond à la version (1.7.1)
- Le nom du binaire est exactement `terraform-provider-harvester`
- Le répertoire `linux_amd64` correspond à votre architecture

## 📚 Documentation

Pour plus d'informations :
- [Documentation complète](https://github.com/harvester/terraform-provider-harvester)
- [Exemples d'utilisation](https://github.com/harvester/terraform-provider-harvester/tree/main/examples)
- [Issues GitHub](https://github.com/harvester/terraform-provider-harvester/issues)

## 🔗 Liens utiles

- Repository GitHub : https://github.com/harvester/terraform-provider-harvester
- Releases : https://github.com/harvester/terraform-provider-harvester/releases
- Documentation Harvester : https://harvesterhci.io/docs

## 📄 Licence

Voir le fichier LICENSE dans le repository GitHub pour les détails de licence.

## 🆘 Support

Pour obtenir de l'aide :
- Ouvrir une issue sur GitHub
- Consulter la documentation
- Rejoindre la communauté Harvester

---

**Note** : Ce binaire est compilé pour Linux amd64. Pour d'autres plateformes, consultez les releases GitHub.

