# DebBootstrap

Collection d'installateurs automatiques pour **macOS** (Apple Silicon M1/M2/M3 et Intel) et **Ubuntu/Debian**.

## Applications disponibles

- **[Neovim Installer](#neovim-installer)** - Installation automatique de Neovim avec plugins
- **[Docker Installer](#docker-installer)** - Installation automatique de Docker et Docker Compose

---

# Neovim Installer

Installation automatique de Neovim avec toutes ses dépendances.

## Installation rapide

### Option 1: Script en une ligne

```bash
curl -fsSL https://raw.githubusercontent.com/keyxmare/DebBootstrap/main/install.sh | bash
```

### Option 2: Clone et lancement

```bash
git clone https://github.com/keyxmare/DebBootstrap.git
cd DebBootstrap
python3 install.py
```

## Fonctionnalités

- **Multi-plateforme**: macOS (Intel + Apple Silicon) et Ubuntu/Debian
- **Installation automatique**: Détecte l'OS et utilise le gestionnaire de paquets approprié
- **Dépendances incluses**: Git, ripgrep, fzf, Node.js, Python, etc.
- **Configurations prêtes à l'emploi**:
  - **Minimal**: Options de base
  - **Standard**: Plugins essentiels (Telescope, LSP, Treesitter, nvim-cmp)
  - **Full**: Configuration complète avec dashboard, notifications, lazygit, etc.
- **Import personnalisé**: Importez votre propre configuration depuis un chemin local ou un repo Git

## Ce qui est installé

### Neovim
- Version stable par défaut (ou nightly sur demande)
- Installé via Homebrew (macOS) ou AppImage/PPA (Ubuntu)

### Dépendances
| Outil | Description |
|-------|-------------|
| git | Contrôle de version |
| ripgrep (rg) | Recherche ultra-rapide |
| fd | Alternative moderne à find |
| fzf | Fuzzy finder |
| Node.js | Pour LSP et plugins |
| Python 3 | Pour plugins Python |
| lazygit | Interface Git en terminal |

### Plugins (configuration Standard/Full)
| Plugin | Description |
|--------|-------------|
| lazy.nvim | Gestionnaire de plugins |
| catppuccin | Thème de couleurs |
| telescope.nvim | Recherche fuzzy |
| nvim-treesitter | Coloration syntaxique avancée |
| nvim-lspconfig | Configuration LSP |
| mason.nvim | Gestionnaire de serveurs LSP |
| nvim-cmp | Auto-complétion |
| lualine.nvim | Barre de statut |
| gitsigns.nvim | Intégration Git |

## Raccourcis clavier (configuration Standard)

| Raccourci | Action |
|-----------|--------|
| `<Space>` | Leader key |
| `<Space>ff` | Rechercher des fichiers |
| `<Space>fg` | Rechercher du texte (grep) |
| `<Space>e` | Explorateur de fichiers |
| `<Space>w` | Sauvegarder |
| `gd` | Aller à la définition |
| `K` | Documentation au survol |

---

# Docker Installer

Installation automatique de Docker et Docker Compose.

## Installation rapide

### Option 1: Script en une ligne

```bash
curl -fsSL https://raw.githubusercontent.com/keyxmare/DebBootstrap/main/install_docker.sh | bash
```

### Option 2: Clone et lancement

```bash
git clone https://github.com/keyxmare/DebBootstrap.git
cd DebBootstrap
python3 install_docker.py
```

## Fonctionnalités

- **Multi-plateforme**: macOS (Intel + Apple Silicon) et Ubuntu/Debian
- **Docker Engine**: Installation de la dernière version stable
- **Docker Compose**: Inclus automatiquement (v2 plugin)
- **Docker BuildKit**: Inclus pour des builds optimisés
- **Configuration automatique**:
  - Ajout de l'utilisateur au groupe docker (Linux)
  - Démarrage automatique au boot
  - Test d'installation avec hello-world

## Ce qui est installé

### macOS
- **Docker Desktop** via Homebrew
  - Docker Engine
  - Docker CLI
  - Docker Compose
  - Docker BuildKit
  - Kubernetes (optionnel, activable dans les préférences)

### Ubuntu/Debian
- **Docker Engine** (depuis le repository officiel Docker)
  - docker-ce
  - docker-ce-cli
  - containerd.io
  - docker-buildx-plugin
  - docker-compose-plugin

## Systèmes supportés

| OS | Architecture | Méthode d'installation |
|----|--------------|----------------------|
| macOS | Apple Silicon (M1/M2/M3) | Docker Desktop (Homebrew) |
| macOS | Intel | Docker Desktop (Homebrew) |
| Ubuntu 20.04+ | x86_64 | Repository officiel Docker |
| Ubuntu 20.04+ | ARM64 | Repository officiel Docker |
| Debian 11+ | x86_64 | Repository officiel Docker |
| Debian 11+ | ARM64 | Repository officiel Docker |

## Commandes utiles après installation

```bash
# Vérifier l'installation
docker --version
docker compose version

# Tester Docker
docker run hello-world

# Lister les conteneurs
docker ps

# Lister les images
docker images

# Démarrer un projet avec docker-compose
docker compose up -d

# Arrêter un projet
docker compose down
```

## Notes importantes (Linux)

Après l'installation sur Linux, vous devez vous **déconnecter et reconnecter** pour utiliser Docker sans `sudo`.

Alternative rapide (sans déconnexion):
```bash
newgrp docker
```

---

# Informations générales

## Prérequis

- **Python 3.9+** (installé automatiquement si absent)
- **Connexion internet** (pour télécharger les paquets)

## Mode simulation (Dry Run)

Pour tester sans effectuer de changements:

```bash
# Neovim
python3 install.py --dry-run

# Docker
python3 install_docker.py --dry-run
```

## Structure du projet

```
DebBootstrap/
├── install.sh              # Script Neovim (bash)
├── install.py              # Point d'entrée Neovim
├── install_docker.sh       # Script Docker (bash)
├── install_docker.py       # Point d'entrée Docker
├── nvim_installer/         # Package Neovim
│   ├── app.py
│   ├── config_manager.py
│   ├── installers/
│   │   ├── base.py
│   │   ├── macos.py
│   │   └── ubuntu.py
│   └── utils/
│       ├── cli.py
│       ├── runner.py
│       └── system.py
└── docker_installer/       # Package Docker
    ├── app.py
    ├── installers/
    │   ├── base.py
    │   ├── macos.py
    │   └── ubuntu.py
    └── utils/
        ├── cli.py
        ├── runner.py
        └── system.py
```

## Développement

### Ajouter un nouvel OS

1. Créer une nouvelle classe dans `<package>/installers/`
2. Hériter de `BaseInstaller`
3. Implémenter les méthodes abstraites
4. Ajouter la détection dans `utils/system.py`
5. Ajouter l'instanciation dans `app.py`

## Licence

MIT

## Contribution

Les contributions sont les bienvenues ! N'hésitez pas à ouvrir une issue ou une pull request.
