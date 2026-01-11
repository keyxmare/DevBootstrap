# Neovim Installer

Installation automatique de Neovim avec toutes ses dépendances pour **macOS** (Apple Silicon M1/M2/M3 et Intel) et **Ubuntu/Debian**.

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

### Option 3: Téléchargement direct

```bash
# Télécharger le repo
curl -L https://github.com/keyxmare/DebBootstrap/archive/main.zip -o nvim-installer.zip
unzip nvim-installer.zip
cd DebBootstrap-main
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
- **Interface interactive**: Questions simples avec défauts intelligents

## Prérequis

- **Python 3.9+** (installé automatiquement si absent)
- **Connexion internet** (pour télécharger les paquets)

## Systèmes supportés

| OS | Architecture | Méthode d'installation |
|----|--------------|----------------------|
| macOS | Apple Silicon (M1/M2/M3) | Homebrew |
| macOS | Intel | Homebrew |
| Ubuntu 20.04+ | x86_64 | APT + AppImage |
| Ubuntu 20.04+ | ARM64 | APT + PPA |
| Debian 11+ | x86_64 | APT + AppImage |

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
| which-key.nvim | Aide aux raccourcis |
| nvim-tree.lua | Explorateur de fichiers |

## Configuration

### Presets disponibles

#### Minimal
Configuration de base sans plugins:
- Options modernes (numéros relatifs, clipboard système, etc.)
- Raccourcis de base

#### Standard (recommandé)
Configuration avec plugins essentiels:
- Tout le preset Minimal
- Telescope pour la recherche
- Treesitter pour la coloration
- LSP + Mason pour l'auto-complétion
- Interface utilisateur améliorée

#### Full
Configuration complète:
- Tout le preset Standard
- Dashboard de démarrage
- Notifications
- Gestion de sessions
- Intégration LazyGit
- Terminal intégré
- Todo comments
- Et plus...

### Import de configuration personnalisée

Vous pouvez importer votre propre configuration:

```bash
# Depuis un repo Git
python3 install.py
# Choisir "Personnalisé" et entrer l'URL:
# https://github.com/username/nvim-config

# Depuis un dossier local
# Choisir "Personnalisé" et entrer le chemin:
# /chemin/vers/ma/config
```

## Raccourcis clavier (configuration Standard)

| Raccourci | Action |
|-----------|--------|
| `<Space>` | Leader key |
| `<Space>ff` | Rechercher des fichiers |
| `<Space>fg` | Rechercher du texte (grep) |
| `<Space>fb` | Liste des buffers |
| `<Space>e` | Explorateur de fichiers |
| `<Space>w` | Sauvegarder |
| `<Space>q` | Quitter |
| `gd` | Aller à la définition |
| `gr` | Voir les références |
| `K` | Documentation au survol |
| `<Space>ca` | Actions de code |
| `<Space>rn` | Renommer |

## Structure du projet

```
DebBootstrap/
├── install.sh           # Script de lancement bash
├── install.py           # Point d'entrée Python
├── nvim_installer/
│   ├── __init__.py
│   ├── __main__.py
│   ├── app.py           # Application principale
│   ├── config_manager.py # Gestion de configuration
│   ├── installers/
│   │   ├── base.py      # Classe de base
│   │   ├── macos.py     # Installateur macOS
│   │   └── ubuntu.py    # Installateur Ubuntu
│   └── utils/
│       ├── cli.py       # Interface CLI
│       ├── runner.py    # Exécution de commandes
│       └── system.py    # Détection système
└── configs/
    └── nvim/            # Configurations embarquées
```

## Mode simulation (Dry Run)

Pour tester sans effectuer de changements:

```bash
python3 install.py --dry-run
```

## Développement

### Ajouter un nouvel OS

1. Créer une nouvelle classe dans `nvim_installer/installers/`
2. Hériter de `BaseInstaller`
3. Implémenter les méthodes abstraites
4. Ajouter la détection dans `utils/system.py`
5. Ajouter l'instanciation dans `app.py`

### Ajouter de nouvelles dépendances

Modifier la liste `DEPENDENCIES` dans la classe d'installateur appropriée.

## Licence

MIT

## Contribution

Les contributions sont les bienvenues ! N'hésitez pas à ouvrir une issue ou une pull request.
