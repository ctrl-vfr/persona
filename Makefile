# Makefile pour le projet Persona

.PHONY: test test-verbose test-cover test-race clean build install help

# Variables
GO_VERSION := 1.24
APP_NAME := persona
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Commandes par défaut
help: ## Affiche cette aide
	@echo "Commandes disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Tests
test: ## Lance tous les tests
	@echo "🧪 Lancement des tests..."
	go test ./...

test-verbose: ## Lance les tests en mode verbose
	@echo "🧪 Lancement des tests (verbose)..."
	go test -v ./...

test-cover: ## Lance les tests avec couverture de code
	@echo "🧪 Lancement des tests avec couverture..."
	go test -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "📊 Rapport de couverture généré: $(COVERAGE_HTML)"

test-race: ## Lance les tests avec détection de race conditions
	@echo "🧪 Lancement des tests avec détection de race..."
	go test -race ./...

test-bench: ## Lance les benchmarks
	@echo "⚡ Lancement des benchmarks..."
	go test -bench=. -benchmem ./...

test-pkg: ## Lance les tests pour un package spécifique (usage: make test-pkg PKG=config)
	@if [ -z "$(PKG)" ]; then echo "❌ Utilisez: make test-pkg PKG=nom_du_package"; exit 1; fi
	@echo "🧪 Tests pour le package $(PKG)..."
	go test -v ./internal/$(PKG)/...

# Couverture détaillée
coverage-report: test-cover ## Génère et affiche le rapport de couverture
	go tool cover -func=$(COVERAGE_FILE)
	@echo "🌐 Ouvrez $(COVERAGE_HTML) dans votre navigateur pour voir le rapport détaillé"

coverage-check: ## Vérifie que la couverture est supérieure à un seuil
	@echo "📊 Vérification de la couverture..."
	@go test -coverprofile=$(COVERAGE_FILE) ./... > /dev/null
	@COVERAGE=$$(go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$COVERAGE < 10" | bc -l) -eq 1 ]; then \
		echo "❌ Couverture insuffisante: $$COVERAGE% (minimum 10%)"; \
		exit 1; \
	else \
		echo "✅ Couverture suffisante: $$COVERAGE%"; \
	fi

# Linting et qualité
lint: ## Lance le linter
	@echo "🔍 Lancement du linter..."
	golangci-lint run

lint-fix: ## Lance le linter avec corrections automatiques
	@echo "🔧 Correction automatique avec le linter..."
	golangci-lint run --fix

vet: ## Lance go vet
	@echo "🔍 Vérification avec go vet..."
	go vet ./...

fmt: ## Formate le code
	@echo "✨ Formatage du code..."
	go fmt ./...

# Build et installation
build: ## Compile l'application
	@echo "🔨 Compilation de $(APP_NAME)..."
	go build -o $(APP_NAME) .

build-all: ## Compile pour toutes les plateformes principales
	@echo "🔨 Compilation multi-plateforme..."
	GOOS=windows GOARCH=amd64 go build -o $(APP_NAME)-windows-amd64.exe .
	GOOS=linux GOARCH=amd64 go build -o $(APP_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o $(APP_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o $(APP_NAME)-darwin-arm64 .
	@echo "✅ Binaires créés pour toutes les plateformes"

install: ## Installe l'application
	@echo "📦 Installation de $(APP_NAME)..."
	go install .

# Nettoyage
clean: ## Nettoie les fichiers générés
	@echo "🧹 Nettoyage..."
	rm -f $(APP_NAME)
	rm -f $(APP_NAME)-*
	rm -f $(COVERAGE_FILE)
	rm -f $(COVERAGE_HTML)
	go clean -testcache
	go clean -cache

# Développement
dev-deps: ## Installe les dépendances de développement
	@echo "📦 Installation des dépendances de développement..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go mod tidy

dev-test: ## Lance les tests en mode développement (watch)
	@echo "👀 Tests en mode watch..."
	@which gow >/dev/null 2>&1 || (echo "📦 Installation de gow..." && go install github.com/mitranim/gow@latest)
	gow test ./...

# Validation complète
validate: fmt vet lint test-race test-cover ## Lance toutes les validations
	@echo "✅ Toutes les validations sont passées!"

# Tests par catégorie
test-unit: ## Lance uniquement les tests unitaires
	@echo "🧪 Tests unitaires..."
	go test -short ./...

test-integration: ## Lance les tests d'intégration
	@echo "🧪 Tests d'intégration..."
	go test -run Integration ./...

# Informations
version: ## Affiche les informations de version
	@echo "📋 Informations de version:"
	@go version
	@echo "App: $(APP_NAME)"
	@echo "Go minimum requis: $(GO_VERSION)"

deps: ## Affiche les dépendances
	@echo "📦 Dépendances:"
	go list -m all

deps-update: ## Met à jour les dépendances
	@echo "⬆️  Mise à jour des dépendances..."
	go get -u ./...
	go mod tidy

# Tests spécifiques par package
test-config: ## Tests du package config
	go test -v ./internal/config/...

test-persona: ## Tests du package persona
	go test -v ./internal/persona/...

test-cmd: ## Tests des commandes
	go test -v ./cmd/...

# Documentation des tests
test-doc: ## Génère la documentation des tests
	@echo "📚 Génération de la documentation des tests..."
	@echo "# Tests du projet Persona" > TEST_RESULTS.md
	@echo "" >> TEST_RESULTS.md
	@echo "## Résultats des tests" >> TEST_RESULTS.md
	@echo "\`\`\`" >> TEST_RESULTS.md
	@go test ./... >> TEST_RESULTS.md 2>&1 || true
	@echo "\`\`\`" >> TEST_RESULTS.md
	@echo "" >> TEST_RESULTS.md
	@echo "## Couverture" >> TEST_RESULTS.md
	@echo "\`\`\`" >> TEST_RESULTS.md
	@go test -coverprofile=$(COVERAGE_FILE) ./... > /dev/null 2>&1 || true
	@go tool cover -func=$(COVERAGE_FILE) >> TEST_RESULTS.md 2>&1 || true
	@echo "\`\`\`" >> TEST_RESULTS.md
	@echo "📄 Documentation générée: TEST_RESULTS.md"

# CI/CD helpers
ci-test: ## Tests pour CI/CD
	@echo "🤖 Tests CI/CD..."
	go test -race -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -func=$(COVERAGE_FILE)

# Statistiques
stats: ## Affiche les statistiques du projet
	@echo "📊 Statistiques du projet:"
	@echo "Lignes de code Go:"
	@find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1
	@echo "Nombre de tests:"
	@find . -name "*_test.go" -not -path "./vendor/*" | xargs grep -c "^func Test" | awk -F: '{sum += $$2} END {print sum}'
	@echo "Nombre de benchmarks:"
	@find . -name "*_test.go" -not -path "./vendor/*" | xargs grep -c "^func Benchmark" | awk -F: '{sum += $$2} END {print sum}' 2>/dev/null || echo "0"
	@echo "Packages:"
	@find ./internal ./cmd -name "*.go" -not -name "*_test.go" | xargs dirname | sort -u | wc -l
