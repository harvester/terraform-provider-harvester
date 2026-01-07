# Instructions de compilation pour la release

## Emplacement du binaire actuel

Le binaire compilé se trouve dans :
```
/root/terraform-provider-harvester/terraform-provider-harvester
```

## Compilation simple (plateforme actuelle)

```bash
cd /root/terraform-provider-harvester
go build -o terraform-provider-harvester .
```

## Cross-compilation pour plusieurs plateformes

Pour créer des binaires pour différentes plateformes (utile pour une release GitHub) :

### Linux (amd64)

```bash
GOOS=linux GOARCH=amd64 go build -o terraform-provider-harvester_linux_amd64 .
```

### Linux (arm64)

```bash
GOOS=linux GOARCH=arm64 go build -o terraform-provider-harvester_linux_arm64 .
```

### macOS (amd64)

```bash
GOOS=darwin GOARCH=amd64 go build -o terraform-provider-harvester_darwin_amd64 .
```

### macOS (arm64 - Apple Silicon)

```bash
GOOS=darwin GOARCH=arm64 go build -o terraform-provider-harvester_darwin_arm64 .
```

### Windows (amd64)

```bash
GOOS=windows GOARCH=amd64 go build -o terraform-provider-harvester_windows_amd64.exe .
```

## Script de compilation pour toutes les plateformes

Créer un script `build-release.sh` :

```bash
#!/bin/bash
set -e

VERSION=${1:-"v0.7.0"}
OUTPUT_DIR="bin"

mkdir -p ${OUTPUT_DIR}

echo "Building for Linux amd64..."
GOOS=linux GOARCH=amd64 go build -o ${OUTPUT_DIR}/terraform-provider-harvester_${VERSION}_linux_amd64 .

echo "Building for Linux arm64..."
GOOS=linux GOARCH=arm64 go build -o ${OUTPUT_DIR}/terraform-provider-harvester_${VERSION}_linux_arm64 .

echo "Building for macOS amd64..."
GOOS=darwin GOARCH=amd64 go build -o ${OUTPUT_DIR}/terraform-provider-harvester_${VERSION}_darwin_amd64 .

echo "Building for macOS arm64..."
GOOS=darwin GOARCH=arm64 go build -o ${OUTPUT_DIR}/terraform-provider-harvester_${VERSION}_darwin_arm64 .

echo "Building for Windows amd64..."
GOOS=windows GOARCH=amd64 go build -o ${OUTPUT_DIR}/terraform-provider-harvester_${VERSION}_windows_amd64.exe .

echo "Build complete! Binaries are in ${OUTPUT_DIR}/"
ls -lh ${OUTPUT_DIR}/
```

## Réduction de la taille du binaire

Pour réduire la taille du binaire (optionnel) :

```bash
# Stripper les symboles de debug
go build -ldflags="-s -w" -o terraform-provider-harvester .
```

## Vérification du binaire

```bash
# Vérifier le type de fichier
file terraform-provider-harvester

# Vérifier les informations
ls -lh terraform-provider-harvester

# Tester l'exécution (affiche l'aide)
./terraform-provider-harvester --help
```

## Attacher les binaires à la release GitHub

Une fois les binaires créés, vous pouvez les attacher à la release GitHub :

```bash
# Via GitHub CLI
gh release create v0.7.0 \
  --title "v0.7.0" \
  --notes-file RELEASE/RELEASE_TEMPLATE.md \
  bin/terraform-provider-harvester_v0.7.0_linux_amd64 \
  bin/terraform-provider-harvester_v0.7.0_linux_arm64 \
  bin/terraform-provider-harvester_v0.7.0_darwin_amd64 \
  bin/terraform-provider-harvester_v0.7.0_darwin_arm64 \
  bin/terraform-provider-harvester_v0.7.0_windows_amd64.exe
```

Ou via l'interface web GitHub lors de la création de la release.

## Notes importantes

- Les binaires doivent être compilés sur une machine propre ou dans un conteneur
- Vérifier que tous les tests passent avant de compiler
- Les binaires doivent être signés si votre organisation le requiert
- Conserver les binaires dans un endroit sûr jusqu'à la publication de la release

