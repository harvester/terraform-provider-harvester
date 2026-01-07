# Instructions pour créer une Release GitHub

Ce document explique comment créer une release GitHub pour le Terraform Provider Harvester.

## 📋 Prérequis

1. Avoir les droits d'écriture sur le repository
2. Avoir Git configuré localement
3. Avoir accès à GitHub (via navigateur ou CLI)

## 🚀 Étapes pour créer la release

### 1. Préparer la version

```bash
cd /root/terraform-provider-harvester

# Vérifier l'état actuel
git status

# S'assurer d'être sur la bonne branche
git checkout main  # ou la branche de release

# Mettre à jour depuis le remote
git pull origin main
```

### 2. Créer un tag Git

```bash
# Créer un tag annoté (recommandé)
git tag -a v0.7.0 -m "Release v0.7.0: Support des backups récurrents de VMs"

# Ou créer un tag avec un message détaillé
git tag -a v0.7.0 -F RELEASE/RELEASE_NOTES.md

# Vérifier le tag
git tag -l "v*"
git show v0.7.0
```

### 3. Pousser le tag vers GitHub

```bash
# Pousser un tag spécifique
git push origin v0.7.0

# Ou pousser tous les tags
git push origin --tags
```

### 4. Créer la release sur GitHub

#### Option A : Via l'interface web GitHub

1. Aller sur : https://github.com/jniedergang/terraform-provider-harvester/releases/new
2. Sélectionner le tag créé (ex: `v0.7.0`)
3. Remplir les champs :
   - **Release title** : `v0.7.0 - Support des backups récurrents de VMs`
   - **Description** : Copier le contenu de `RELEASE/RELEASE_TEMPLATE.md`
4. Attacher des fichiers si nécessaire (binaires, assets)
5. Cocher "Set as the latest release" si c'est la dernière version
6. Cliquer sur "Publish release"

#### Option B : Via GitHub CLI (gh)

```bash
# Installer GitHub CLI si nécessaire
# sudo apt install gh  # ou selon votre distribution

# Se connecter
gh auth login

# Créer la release
gh release create v0.7.0 \
  --title "v0.7.0 - Support des backups récurrents de VMs" \
  --notes-file RELEASE/RELEASE_TEMPLATE.md \
  --target main
```

### 5. Vérifier la release

1. Aller sur : https://github.com/jniedergang/terraform-provider-harvester/releases
2. Vérifier que la release apparaît
3. Vérifier que les notes de release sont correctes
4. Tester le téléchargement si des assets sont attachés

## 📝 Notes importantes

### Versioning

- Suivre le [Semantic Versioning](https://semver.org/) :
  - **MAJOR** (1.0.0) : Changements incompatibles
  - **MINOR** (0.1.0) : Nouvelles fonctionnalités compatibles
  - **PATCH** (0.0.1) : Corrections de bugs

### Tags Git

- Utiliser le format `vX.Y.Z` (ex: `v0.7.0`)
- Les tags doivent pointer vers un commit stable
- Ne jamais modifier un tag après publication

### Assets (optionnel)

Si vous devez attacher des binaires :

```bash
# Créer les binaires (exemple)
make build

# Attacher lors de la création de la release
gh release create v0.7.0 \
  --title "v0.7.0" \
  --notes-file RELEASE/RELEASE_TEMPLATE.md \
  ./bin/terraform-provider-harvester_linux_amd64 \
  ./bin/terraform-provider-harvester_darwin_amd64 \
  ./bin/terraform-provider-harvester_windows_amd64.exe
```

## 🔄 Mettre à jour après la release

1. Mettre à jour le CHANGELOG.md principal (si présent)
2. Mettre à jour la version dans go.mod si nécessaire
3. Créer une branche pour la prochaine version

## ❓ Dépannage

### Le tag n'apparaît pas sur GitHub

```bash
# Vérifier que le tag a été poussé
git ls-remote --tags origin

# Re-pousser si nécessaire
git push origin v0.7.0 --force  # Attention : utiliser avec précaution
```

### Erreur de permissions

- Vérifier que vous avez les droits d'écriture sur le repository
- Vérifier que vous êtes authentifié correctement

### Modifier une release existante

1. Aller sur la page de la release
2. Cliquer sur "Edit release"
3. Modifier les informations
4. Sauvegarder

**Note** : Les tags Git ne peuvent pas être modifiés après publication. Si nécessaire, créer un nouveau tag.

