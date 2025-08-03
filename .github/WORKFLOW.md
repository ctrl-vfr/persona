# GitHub Actions - Persona

Ce dossier contient les workflows GitHub Actions pour automatiser la construction, les tests, et les releases du projet Persona.

## 🔄 Workflows disponibles

### 1. **release.yml** - Build et Release automatique

- **Déclencheur** : Tags version (v\*), releases, ou manuel
- **Fonctions** :
  - Compile pour 9 architectures différentes (Windows, Linux, macOS)
  - Crée des artefacts pour chaque build
  - Génère automatiquement une release GitHub avec tous les binaires
  - Calcule les checksums pour vérification d'intégrité

### 2. **build.yml** - Tests et builds de développement

- **Déclencheur** : Push sur main/develop, Pull Requests
- **Fonctions** :
  - Lance les tests unitaires avec coverage
  - Compile les binaires de développement
  - Vérifie le code avec golangci-lint
  - Upload des artefacts de dev (rétention 7 jours)

### 3. **auto-tag.yml** - Versioning automatique

- **Déclencheur** : Push sur main
- **Fonctions** :
  - Analyse les messages de commit pour déterminer le type de version
  - Crée automatiquement les tags de version
  - Support des versions : major, minor, patch

### 4. **cleanup.yml** - Nettoyage automatique

- **Déclencheur** : Hebdomadaire (dimanche 2h) ou manuel
- **Fonctions** :
  - Supprime les anciens artefacts de développement (>7 jours)
  - Nettoie les artefacts très anciens (>30 jours)
  - Supprime les pre-releases de développement expirées

## 🏗️ Architectures supportées

| OS          | Architecture | Fichier                      |
| ----------- | ------------ | ---------------------------- |
| **Windows** | amd64        | persona-\*-windows-amd64.exe |
| Windows     | 386          | persona-\*-windows-386.exe   |
| Windows     | arm64        | persona-\*-windows-arm64.exe |
| **Linux**   | amd64        | persona-\*-linux-amd64       |
| Linux       | 386          | persona-\*-linux-386         |
| Linux       | arm64        | persona-\*-linux-arm64       |
| Linux       | arm          | persona-\*-linux-arm         |
| **macOS**   | amd64        | persona-\*-darwin-amd64      |
| macOS       | arm64        | persona-\*-darwin-arm64      |

## 🚀 Comment déclencher une release

### Méthode 1 : Tags manuels

```bash
# Créer un tag de version
git tag v1.0.0
git push origin v1.0.0
```

### Méthode 2 : Versioning automatique par commit

```bash
# Version patch (défaut)
git commit -m "fix: correction bug audio"

# Version minor
git commit -m "feat: nouvelle interface TUI [minor]"

# Version major
git commit -m "refactor: nouvelle architecture [major]"

# Ignorer le versioning
git commit -m "docs: mise à jour README [skip-tag]"
```

### Méthode 3 : Release manuelle

1. Aller dans l'onglet "Actions" de votre repo GitHub
2. Sélectionner "Build and Release"
3. Cliquer "Run workflow"
4. Choisir la branche et lancer

## 📦 Artefacts générés

### Pour les releases

- **Archives** : `.tar.gz` (Linux/macOS), `.zip` (Windows)
- **Checksums** : `checksums.txt` pour vérification d'intégrité
- **Info files** : Métadonnées de build pour chaque binaire

### Pour le développement

- **Binaires bruts** : Rétention 7 jours
- **Nommage** : `persona-dev-{sha}-{os}-{arch}`

## 🔧 Configuration

### Variables d'environnement requises

- `GITHUB_TOKEN` : Automatiquement fourni par GitHub Actions

### Secrets optionnels

- `CODECOV_TOKEN` : Pour les rapports de couverture de tests

### Paramètres de build

- **Go version** : 1.24
- **CGO** : Désactivé pour la portabilité
- **Optimisations** : `-ldflags="-w -s"` pour réduire la taille

## 📋 Exemple de release automatique

Lorsqu'un tag `v1.2.3` est créé, la release contiendra :

```
📦 Release v1.2.3
├── 📁 persona-v1.2.3-windows-amd64.zip
├── 📁 persona-v1.2.3-windows-386.zip
├── 📁 persona-v1.2.3-windows-arm64.zip
├── 📁 persona-v1.2.3-linux-amd64.tar.gz
├── 📁 persona-v1.2.3-linux-386.tar.gz
├── 📁 persona-v1.2.3-linux-arm64.tar.gz
├── 📁 persona-v1.2.3-linux-arm.tar.gz
├── 📁 persona-v1.2.3-darwin-amd64.tar.gz
├── 📁 persona-v1.2.3-darwin-arm64.tar.gz
└── 📄 checksums.txt
```

## 🛠️ Maintenance

### Modifier les architectures

Éditer la matrix dans `release.yml` section `strategy.matrix.include`

### Ajuster la rétention des artefacts

- **Release** : `retention-days: 30`
- **Development** : `retention-days: 7`

### Personnaliser le versioning

Modifier les règles dans `auto-tag.yml` section "Determine next version"
