---
title: Installation Guide
description: How to install devbox on your Debian/Ubuntu system
---

```bash
curl -fsSL https://devbox.ar0.eu/install.sh | bash
```

This script will automatically:
- Check system compatibility (Debian/Ubuntu only)
- Install Go, Docker, make, and git if needed
- Clone the repository and build devbox
- Install devbox to `/usr/local/bin`
- Set up proper permissions

<sub>Don't have curl? Read this [quick guide](https://www.cyberciti.biz/faq/howto-install-curl-command-on-debian-linux-using-apt-get/) to install it.</sub>

## Manual Build from Source
---

If you prefer to build devbox manually or the automatic script doesn't work for your system:

### Prerequisites
- Debian/Ubuntu Linux system
- Docker installed and running
- Go 1.21+ installed
- make and git installed

### Install Dependencies
```bash
# Update package index
sudo apt update

# Install Docker (if not already installed)
sudo apt install docker.io

# Install Go (if not already installed)
sudo apt install golang-go

# Install build tools
sudo apt install make git

# Start and enable Docker
sudo systemctl start docker
sudo systemctl enable docker

# Add your user to docker group (requires logout/login)
sudo usermod -aG docker $USER
```

### Build and Install
```bash
# Clone the repository
git clone https://github.com/itzCozi/devbox.git
cd devbox

# Build the binary
make build

# Install to system (requires sudo)
sudo make install

# Verify installation
devbox --help
```

## System Maintenance
---

```bash
# Check system health
devbox maintenance --health-check

# Clean up unused resources
devbox cleanup --all

# Update all containers
devbox maintenance --update
```

## File Locations
---

- **Project files**: `~/devbox/<project>/` (on host)
- **Container workspace**: `/workspace/` (inside container)
- **Configuration**: `~/.devbox/config.json`

## Docker Access
---

By default, all devbox environments have access to the host's Docker daemon, allowing you to:

- Build and manage Docker containers from within your devbox environment
- Run Docker commands without additional configuration
- Execute Docker Compose for multi-container applications

This works by mounting the host's Docker socket (`/var/run/docker.sock`) in your devbox container and installing the Docker CLI tools automatically.
