#!/bin/bash
# Script d'aide pour l'environnement de développement du Terraform Provider Harvester

# Configuration des variables d'environnement Go
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Aller dans le répertoire du projet
cd "$(dirname "$0")"

# Fonctions utiles
function dev-help() {
    echo "=== Commandes disponibles pour le développement ==="
    echo ""
    echo "  ./dev-env.sh build          - Construire le provider"
    echo "  ./dev-env.sh test           - Exécuter les tests"
    echo "  ./dev-env.sh validate       - Valider le code"
    echo "  ./dev-env.sh generate       - Générer le code (docs, etc.)"
    echo "  ./dev-env.sh install        - Installer les dépendances"
    echo "  ./dev-env.sh clean          - Nettoyer les fichiers de build"
    echo "  ./dev-env.sh env            - Afficher les variables d'environnement"
    echo ""
}

function dev-build() {
    echo "🔨 Construction du provider..."
    mkdir -p bin
    go build -o bin/terraform-provider-harvester .
    echo "✅ Build terminé: bin/terraform-provider-harvester"
}

function dev-test() {
    echo "🧪 Exécution des tests..."
    go test -v ./...
}

function dev-validate() {
    echo "✔️  Validation du code..."
    go fmt ./...
    go vet ./...
    echo "✅ Validation terminée"
}

function dev-generate() {
    echo "📝 Génération du code..."
    go generate ./...
    echo "✅ Génération terminée"
}

function dev-install() {
    echo "📦 Installation des dépendances..."
    go mod download
    go mod tidy
    echo "✅ Dépendances installées"
}

function dev-clean() {
    echo "🧹 Nettoyage..."
    rm -rf bin/
    rm -f coverage.out coverage.html
    go clean -cache
    echo "✅ Nettoyage terminé"
}

function dev-env() {
    echo "=== Variables d'environnement ==="
    echo "GOPATH: $GOPATH"
    echo "GOROOT: $(go env GOROOT)"
    echo "Go version: $(go version)"
    echo "PATH: $PATH"
}

# Gestion des commandes
case "${1:-help}" in
    build)
        dev-build
        ;;
    test)
        dev-test
        ;;
    validate)
        dev-validate
        ;;
    generate)
        dev-generate
        ;;
    install)
        dev-install
        ;;
    clean)
        dev-clean
        ;;
    env)
        dev-env
        ;;
    help|--help|-h)
        dev-help
        ;;
    *)
        echo "Commande inconnue: $1"
        dev-help
        exit 1
        ;;
esac

