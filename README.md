# DevBootstrap

Collection d'installateurs et désinstallateurs automatiques pour **macOS** (Apple Silicon M1/M2/M3 et Intel) et **Ubuntu/Debian**.

## Applications disponibles

- **[Neovim](#neovim)** - Editeur de texte moderne avec plugins
- **[Docker](#docker)** - Plateforme de conteneurisation
- **[VS Code](#vs-code)** - Editeur de code source
- **[Zsh & Oh My Zsh](#zsh--oh-my-zsh)** - Shell moderne avec framework
- **[Nerd Fonts](#nerd-fonts)** - Polices avec icônes pour terminal
- **[Commande devbootstrap](#commande-devbootstrap)** - Alias global pour l'outil

---

## Installation rapide

### Option 1: Script en une ligne

```bash
curl -fsSL https://raw.githubusercontent.com/keyxmare/DevBootstrap/main/install.sh | bash
```

### Option 2: Clone et lancement

```bash
git clone https://github.com/keyxmare/DevBootstrap.git
cd DevBootstrap
python3 -m bootstrap
```

---

## Utilisation

### Menu interactif

Lancez simplement :

```bash
python3 -m bootstrap
```

Le menu vous permet de :
1. **Installer** - Sélectionner les applications à installer
2. **Désinstaller** - Supprimer les applications installées

### Mode ligne de commande

```bash
# Installation (mode par défaut)
python3 -m bootstrap

# Désinstallation directe
python3 -m bootstrap --uninstall

# Mode simulation (sans changements)
python3 -m bootstrap --dry-run

# Mode non-interactif (installer tout)
python3 -m bootstrap --no-interaction
```

---

## Désinstallation

DevBootstrap inclut des désinstallateurs complets pour toutes les applications. Ils permettent de :

- Supprimer proprement les applications installées
- Nettoyer les fichiers de configuration
- Supprimer le cache et les données
- Restaurer les paramètres par défaut

### Via le menu interactif

```bash
python3 -m bootstrap
# Puis choisir "2. Désinstaller"
```

### Via ligne de commande

```bash
python3 -m bootstrap --uninstall
```

---

# Neovim

Installation automatique de Neovim avec toutes ses dépendances.

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

## Désinstallation Neovim

La désinstallation de Neovim supprime :
- Le binaire Neovim
- La configuration (`~/.config/nvim`)
- Les données (`~/.local/share/nvim`)
- Le cache (`~/.cache/nvim`)
- Les providers Python et Node.js

---

# Docker

Installation automatique de Docker et Docker Compose.

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
  - Kubernetes (optionnel)

### Ubuntu/Debian
- **Docker Engine** (depuis le repository officiel Docker)
  - docker-ce
  - docker-ce-cli
  - containerd.io
  - docker-buildx-plugin
  - docker-compose-plugin

## Commandes utiles

```bash
# Vérifier l'installation
docker --version
docker compose version

# Tester Docker
docker run hello-world

# Lister les conteneurs
docker ps

# Démarrer un projet
docker compose up -d
```

## Désinstallation Docker

La désinstallation de Docker :
- Arrête tous les conteneurs en cours
- Supprime tous les conteneurs
- Supprime toutes les images
- Supprime les volumes (optionnel)
- Désinstalle Docker via le gestionnaire de paquets
- Supprime les données Docker (`/var/lib/docker`)

---

# VS Code

Installation automatique de Visual Studio Code.

## Fonctionnalités

- **Multi-plateforme**: macOS et Ubuntu/Debian
- **Extensions recommandées**: Python, ESLint, Prettier, etc.
- **Configuration automatique**

## Extensions disponibles

| Extension | Description |
|-----------|-------------|
| Python | Support Python |
| Prettier | Formatage de code |
| ESLint | Linting JavaScript |
| TypeScript | Support TypeScript |
| Tailwind CSS | IntelliSense Tailwind |
| GitLens | Intégration Git avancée |
| Material Icon Theme | Icônes |

## Désinstallation VS Code

La désinstallation de VS Code supprime :
- L'application VS Code
- Les extensions (`~/.vscode/extensions`)
- Les paramètres (`~/.config/Code` ou `~/Library/Application Support/Code`)
- Le cache

---

# Zsh & Oh My Zsh

Installation de Zsh et du framework Oh My Zsh.

## Fonctionnalités

- **Zsh**: Shell moderne
- **Oh My Zsh**: Framework de configuration
- **Plugins inclus**:
  - zsh-autosuggestions
  - zsh-syntax-highlighting
  - zsh-completions
- **Thèmes**: robbyrussell, agnoster, powerlevel10k

## Désinstallation Zsh

La désinstallation supprime :
- Oh My Zsh (`~/.oh-my-zsh`)
- Les plugins personnalisés
- Le fichier `.zshrc`
- L'historique et le cache Zsh
- Restaure bash comme shell par défaut (optionnel)

---

# Nerd Fonts

Installation de polices avec icônes pour terminal.

## Polices disponibles

| Police | Description |
|--------|-------------|
| MesloLG | Recommandée pour agnoster |
| FiraCode | Avec ligatures |
| JetBrains Mono | Police JetBrains |
| Hack | Police Hack |

## Désinstallation Nerd Fonts

Supprime les polices Nerd Font installées et met à jour le cache.

---

# Commande devbootstrap

Installe un alias global `devbootstrap` pour lancer l'outil depuis n'importe où.

## Fonctionnalités

- Script exécutable dans `~/.local/bin`
- Alias/fonction dans les fichiers RC shell
- Synchronisation automatique avec GitHub

## Utilisation

```bash
# Après installation, lancer depuis n'importe où
devbootstrap
```

## Désinstallation devbootstrap

Supprime :
- Le script `~/.local/bin/devbootstrap`
- Les alias/fonctions des fichiers RC
- Le répertoire `~/.devbootstrap`

---

# Informations générales

## Prérequis

- **Python 3.9+** (installé automatiquement si absent)
- **Connexion internet** (pour télécharger les paquets)

## Systèmes supportés

| OS | Architecture | Support |
|----|--------------|---------|
| macOS | Apple Silicon (M1/M2/M3) | ✓ |
| macOS | Intel | ✓ |
| Ubuntu 20.04+ | x86_64 | ✓ |
| Ubuntu 20.04+ | ARM64 | ✓ |
| Debian 11+ | x86_64 | ✓ |
| Debian 11+ | ARM64 | ✓ |

## Mode simulation (Dry Run)

Pour tester sans effectuer de changements:

```bash
python3 -m bootstrap --dry-run
```

## Structure du projet

```
DevBootstrap/
├── install.sh              # Script d'installation (bash)
├── install.py              # Point d'entrée Neovim
├── install_docker.sh       # Script Docker (bash)
├── install_docker.py       # Point d'entrée Docker
├── bootstrap/              # Menu principal
│   ├── app.py             # Application principale
│   └── apps.py            # Registre des applications
├── nvim_installer/         # Installateur/Désinstallateur Neovim
│   ├── installers/
│   └── uninstallers/
├── docker_installer/       # Installateur/Désinstallateur Docker
│   ├── installers/
│   └── uninstallers/
├── vscode_installer/       # Installateur/Désinstallateur VS Code
│   ├── installers/
│   └── uninstallers/
├── zsh_installer/          # Installateur/Désinstallateur Zsh
│   ├── installers/
│   └── uninstallers/
├── font_installer/         # Installateur/Désinstallateur Fonts
│   ├── installers/
│   └── uninstallers/
└── alias_installer/        # Installateur/Désinstallateur Alias
    ├── app.py
    └── uninstaller.py
```

## Développement

### Ajouter un nouvel installateur

1. Créer un nouveau package dans le répertoire racine
2. Implémenter les classes dans `installers/` (hériter de BaseInstaller)
3. Implémenter les classes dans `uninstallers/` (hériter de BaseUninstaller)
4. Ajouter l'application dans `bootstrap/apps.py`
5. Intégrer dans `bootstrap/app.py`

### Ajouter un nouveau système d'exploitation

1. Créer une nouvelle classe dans `<package>/installers/`
2. Hériter de `BaseInstaller`
3. Implémenter les méthodes abstraites
4. Faire de même pour `<package>/uninstallers/`
5. Ajouter la détection dans `utils/system.py`

## Licence

MIT

## Contribution

Les contributions sont les bienvenues ! N'hésitez pas à ouvrir une issue ou une pull request.
