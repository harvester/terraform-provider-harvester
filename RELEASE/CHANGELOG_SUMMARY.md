# Résumé des changements pour la release

## Commits inclus dans cette release

### Feature: harvester_schedule_backup resource

- **Commit** : `2426239` - feat: Add harvester_schedule_backup resource for VM-level recurring backups
  - Ajout de la ressource Terraform pour gérer les backups récurrents
  - Support du CRD ScheduleVMBackup de Harvester
  - Gestion de la compatibilité arrière avec volume_name

### Fixes et améliorations

- **Commit** : `ced6f26` - fix: Format code and organize imports
  - Formatage du code selon les standards Go
  - Réorganisation des imports

- **Commit** : `e49dc9c` - refactor: Reduce method complexity for CodeFactor compliance
  - Extraction de fonctions helper pour réduire la complexité
  - Amélioration de la lisibilité du code
  - Conformité avec CodeFactor

- **Commit** : `f4f27e7` - fix: Use correct client type in helper functions
  - Correction du type client dans les fonctions helper
  - Ajout de l'import manquant

## Fichiers modifiés

### Nouveaux fichiers

- `internal/provider/volumebackup/resource_volumebackup.go`
- `internal/provider/volumebackup/schema_volumebackup.go`
- `internal/provider/volumebackup/resource_volumebackup_constructor.go`
- `internal/provider/volumebackup/types.go`
- `pkg/constants/constants_volumebackup.go`
- `pkg/importer/resource_volumebackup_importer.go`

### Fichiers modifiés

- `internal/provider/provider.go` - Ajout de la ressource volumebackup

## Tests effectués

- ✅ Compilation réussie
- ✅ Formatage du code (gofmt)
- ✅ Vérification des imports (go mod tidy)
- ✅ Tests de création de backup récurrent
- ✅ Tests de restauration depuis backup
- ✅ Vérification de la compatibilité arrière

## Conformité

- ✅ DCO (Developer Certificate of Origin) - Tous les commits signés
- ✅ CodeFactor - Complexité des méthodes réduite
- ✅ Standards Go - Formatage et organisation des imports

## Prochaines étapes

1. Créer le tag Git
2. Pousser le tag vers GitHub
3. Créer la release sur GitHub
4. Mettre à jour la documentation principale
5. Annoncer la release

