#!/usr/bin/env python3
"""
DevBootstrap - Point d'entrée principal

Affiche un menu unifié avec toutes les applications disponibles
et permet de sélectionner ce que vous souhaitez installer.

Usage:
    python3 install.py
    python3 install.py --dry-run
"""

import sys
import os

# Add the project directory to the path
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from bootstrap.app import main

if __name__ == "__main__":
    main()
