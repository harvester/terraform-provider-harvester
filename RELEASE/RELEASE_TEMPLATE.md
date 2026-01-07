# Template de Release GitHub

## Titre de la release

```
v0.7.0 - Support des backups récurrents de VMs
```

## Description de la release

```markdown
## 🎉 Nouvelle fonctionnalité : harvester_schedule_backup

Cette release introduit le support des backups récurrents de machines virtuelles dans Harvester via Terraform.

### Fonctionnalités principales

- **Resource `harvester_schedule_backup`** : Nouvelle ressource Terraform pour gérer les backups récurrents de VMs
  - Configuration de backups récurrents au niveau VM (tous les disques)
  - Support des schedules cron en UTC
  - Gestion de la rétention des backups
  - Support des labels personnalisés
  - Activation/désactivation des backups

### Améliorations techniques

- Refactorisation du code pour réduire la complexité cyclomatique
- Amélioration de la lisibilité et de la maintenabilité du code
- Conformité avec les standards de qualité CodeFactor
- Support de la compatibilité arrière avec `volume_name` (déprécié)

### Détails techniques

- Utilise le CRD `ScheduleVMBackup` de Harvester
- Support d'un seul schedule par VM (limitation Harvester)
- Gestion automatique des schedules existants (mise à jour)
- Import de ressources existantes via `terraform import`

---

## 📝 Changements

### Ajouts

- ✨ Nouvelle ressource `harvester_schedule_backup`
- 📚 Documentation complète dans le schéma Terraform
- 🔄 Support de l'import de ressources existantes

### Corrections

- 🐛 Correction du type client dans les fonctions helper
- 🐛 Formatage et organisation des imports selon les conventions Go
- 🐛 Réduction de la complexité des méthodes pour CodeFactor

### Refactorisations

- ♻️ Extraction de fonctions helper pour réduire la complexité
- ♻️ Amélioration de la structure du code
- ♻️ Optimisation de la gestion des erreurs

---

## 📦 Installation

```bash
# Via Terraform Registry (recommandé)
terraform {
  required_providers {
    harvester = {
      source  = "harvester/harvester"
      version = "~> 0.7.0"
    }
  }
}
```

---

## 🔗 Liens

- [Documentation de la ressource](https://github.com/harvester/terraform-provider-harvester/blob/main/docs/resources/schedule_backup.md)
- [Issue GitHub](https://github.com/harvester/terraform-provider-harvester/issues)
- [Pull Request](https://github.com/harvester/terraform-provider-harvester/pull/150)

---

## 👥 Contributeurs

- @jniedergang
```

