#!/usr/bin/env python3
"""
Neovim Installer - Point d'entrée direct

Usage:
    python3 install.py
    python3 install.py --dry-run
"""

import sys
import os

# Add the project directory to the path
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from nvim_installer.app import main

if __name__ == "__main__":
    main()
