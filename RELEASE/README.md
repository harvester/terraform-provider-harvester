# 📦 Répertoire RELEASE

Ce répertoire contient tous les fichiers et informations nécessaires pour créer une release GitHub du Terraform Provider Harvester.

## 📁 Contenu

- **RELEASE_NOTES.md** : Notes de release détaillées avec toutes les fonctionnalités et changements
- **RELEASE_TEMPLATE.md** : Template prêt à copier-coller dans l'interface GitHub
- **INSTRUCTIONS.md** : Instructions pas à pas pour créer la release
- **CHANGELOG_SUMMARY.md** : Résumé des commits et changements inclus
- **README.md** : Ce fichier

## 🚀 Démarrage rapide

1. Lire `INSTRUCTIONS.md` pour les étapes détaillées
2. Utiliser `RELEASE_TEMPLATE.md` comme description de la release GitHub
3. Vérifier `CHANGELOG_SUMMARY.md` pour la liste des changements
4. Créer le tag Git et la release selon les instructions

## 📝 Workflow recommandé

```bash
# 1. Vérifier l'état
cd /root/terraform-provider-harvester
git status

# 2. Créer le tag
git tag -a v0.7.0 -F RELEASE/RELEASE_NOTES.md

# 3. Pousser le tag
git push origin v0.7.0

# 4. Créer la release sur GitHub
# Aller sur : https://github.com/jniedergang/terraform-provider-harvester/releases/new
# Copier le contenu de RELEASE_TEMPLATE.md dans la description
```

## 🔗 Liens utiles

- [Créer une release](https://github.com/jniedergang/terraform-provider-harvester/releases/new)
- [Liste des releases](https://github.com/jniedergang/terraform-provider-harvester/releases)
- [Documentation GitHub Releases](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository)

## 📌 Notes

- Les versions suivent le [Semantic Versioning](https://semver.org/)
- Les tags doivent être au format `vX.Y.Z`
- Toujours vérifier que tous les tests passent avant de créer une release
- Les notes de release doivent être claires et complètes

