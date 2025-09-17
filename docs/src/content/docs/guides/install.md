---
title: Installation Guide
description: How to install devbox on your Debian/Ubuntu system
---

```bash
curl -fsSL https://raw.githubusercontent.com/itzCozi/devbox/main/install.sh | bash
```

This script will automatically:
- Check system compatibility (Debian/Ubuntu only)
- Install Go, Docker, make, and git if needed
- Clone the repository and build devbox
- Install devbox to `/usr/local/bin`
- Set up proper permissions

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

## Troubleshooting
---

##### Command not found
```bash
# If devbox command not found after install
export PATH="/usr/local/bin:$PATH"
# Or restart your terminal
```

##### Docker permission denied
```bash
# Add user to docker group
sudo usermod -aG docker $USER
# Logout and login again
```

##### Container won't start
```bash
# Check Docker status
sudo systemctl status docker

# Check container logs
docker logs devbox_<project>

# Restart Docker
sudo systemctl restart docker
```

##### Remove everything
```bash
# List all projects
devbox list

# Destroy all projects
devbox destroy project1
devbox destroy project2

# Remove project directories
rm -rf ~/devbox/

# Remove configuration
rm -rf ~/.devbox/
```

## Tips
---

1. **Host file editing**: Edit files with your favorite editor on the host - they're instantly available in the container at `/workspace/`
2. **Persistent installs**: Packages installed with `apt` persist between container restarts
3. **Port forwarding**: Use `docker run -p` syntax in custom containers if you need exposed ports
4. **Resource limits**: Containers share host resources - monitor usage with `docker stats`
5. **Backup projects**: Your code in `~/devbox/<project>/` is safe even if containers are destroyed
6. **Multiple projects**: Each project is completely isolated - install different Python versions, Node versions, etc.

## Advanced Usage
---

##### Custom base images
Edit `~/.devbox/config.json` to change base images for new projects.

##### Sharing projects
Copy the `~/devbox/<project>/` directory to share project files (containers are recreated on each system).
