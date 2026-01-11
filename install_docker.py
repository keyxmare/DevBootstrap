#!/usr/bin/env python3
"""Docker Installer - Entry point script."""

import sys
import os

# Add the directory containing this script to the path
script_dir = os.path.dirname(os.path.abspath(__file__))
if script_dir not in sys.path:
    sys.path.insert(0, script_dir)

from docker_installer.app import main

if __name__ == "__main__":
    main()
