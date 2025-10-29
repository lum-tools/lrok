#!/usr/bin/env python3

import os
import platform
import sys
import tarfile
import urllib.request
from pathlib import Path

from setuptools import setup
from setuptools.command.install import install

VERSION = "0.1.4"

PLATFORM_MAP = {
    "Darwin": "darwin",
    "Linux": "linux",
    "Windows": "windows"
}

ARCH_MAP = {
    "x86_64": "amd64",
    "AMD64": "amd64",
    "arm64": "arm64",
    "aarch64": "arm64"
}


class PostInstallCommand(install):
    """Post-installation command to download lrok binary."""
    
    def run(self):
        install.run(self)
        
        sys_platform = PLATFORM_MAP.get(platform.system())
        sys_arch = ARCH_MAP.get(platform.machine())
        
        if not sys_platform or not sys_arch:
            print(f"Unsupported platform: {platform.system()}/{platform.machine()}")
            sys.exit(1)
        
        ext = ".exe" if sys_platform == "windows" else ""
        archive_name = f"lrok_{VERSION}_{sys_platform}_{sys_arch}.tar.gz"
        url = f"https://github.com/lum-tools/lrok/releases/download/v{VERSION}/{archive_name}"
        
        print(f"📦 Installing lrok v{VERSION} for {sys_platform}/{sys_arch}...")
        
        # Determine install directory
        bin_dir = Path(self.install_scripts)
        bin_dir.mkdir(parents=True, exist_ok=True)
        
        binary_path = bin_dir / f"lrok{ext}"
        tmp_archive = bin_dir / archive_name
        
        try:
            # Download
            print(f"  → Downloading from GitHub releases...")
            urllib.request.urlretrieve(url, tmp_archive)
            
            # Extract
            print(f"  → Extracting binary...")
            with tarfile.open(tmp_archive, 'r:gz') as tar:
                tar.extractall(bin_dir)
            
            # Clean up
            tmp_archive.unlink()
            
            # Make executable (Unix only)
            if sys_platform != "windows":
                binary_path.chmod(0o755)
            
            print(f"✅ lrok installed successfully!")
            print(f"\nRun: lrok version")
            
        except Exception as e:
            print(f"❌ Installation failed: {e}")
            print(f"\nTry manual installation from: https://github.com/lum-tools/lrok/releases")
            sys.exit(1)


with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setup(
    name="lrok",
    version=VERSION,
    author="lum.tools",
    author_email="ops@lum.tools",
    description="Expose local services with readable tunnel names",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/lum-tools/lrok",
    project_urls={
        "Bug Tracker": "https://github.com/lum-tools/lrok/issues",
        "Platform": "https://platform.lum.tools",
        "Dashboard": "https://platform.lum.tools/tunnels",
        "Blog": "https://blog.lum.tools",
        "Homepage": "https://lum.tools",
    },
    license="MIT",
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "Topic :: Internet",
        "Topic :: System :: Networking",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Operating System :: OS Independent",
    ],
    python_requires=">=3.7",
    cmdclass={
        'install': PostInstallCommand,
    },
)

