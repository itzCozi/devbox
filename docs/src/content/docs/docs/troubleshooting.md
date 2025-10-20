---
title: Troubleshooting
description: Common issues and solutions for devbox
---

Common issues and quick fixes.

## Installation Issues
---

##### "403 Forbidden" when running install script

**Problem**: Running `curl -fsSL https://devbox.ar0.eu/install.sh | bash` returns 403 (often in managed shells like AWS CloudShell).

**Explanation**: Some environments block or challenge CDN traffic. Our mirror uses a CDN and may be affected. We've added a server-side fallback to redirect to GitHub Raw, but certain environments still enforce restrictions.

**Solutions**:
```bash
# Use the primary GitHub Raw URL
curl -fsSL https://raw.githubusercontent.com/itzcozi/devbox/main/install.sh | bash

# Or download and run locally
curl -fsSL -o install.sh https://raw.githubusercontent.com/itzcozi/devbox/main/install.sh
bash install.sh
```

##### Amazon Linux 2023

**Problem**: The install script reports an unsupported OS on Amazon Linux 2023.

**Explanation**: devbox officially supports Debian/Ubuntu. Amazon Linux 2023 (AL2023) is Fedora-like and uses `dnf` instead of `apt`.

**Workaround (manual)**:
```bash
# Install deps (rough equivalent)
sudo dnf install -y git make golang docker
sudo systemctl enable --now docker
sudo usermod -aG docker $USER

# Build and install devbox
git clone https://github.com/itzcozi/devbox.git
cd devbox
make build
sudo make install
```

> Note: We may add broader distro support in the future; contributions welcome.

##### "Command not found: devbox"

**Problem**: After installation, `devbox` command is not recognized.

**Solutions**:
```bash
# Check if devbox is in PATH
which devbox

# Add to PATH if needed
export PATH="/usr/local/bin:$PATH"

# Make permanent (add to ~/.bashrc or ~/.zshrc)
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Verify installation
devbox --help
```

##### "Docker is not installed or not running"

**Problem**: Devbox can't connect to Docker daemon.

**Solutions**:
```bash
# Check Docker status
sudo systemctl status docker

# Start Docker if stopped
sudo systemctl start docker
sudo systemctl enable docker

# Check if user is in docker group
groups $USER

# Add user to docker group
sudo usermod -aG docker $USER
# Note: You must log out and back in for this to take effect

# Test Docker access
docker ps
```

##### "Permission denied while trying to connect to Docker"

**Problem**: User doesn't have permission to access Docker socket.

**Solutions**:
```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Restart terminal session or logout/login

# Alternatively, run with sudo (not recommended)
sudo devbox init myproject
```

## Box Issues
---

##### "Box not found" or "No such box"

**Problem**: Box was manually deleted or doesn't exist.

**Solutions**:
```bash
# Check what boxes exist
docker ps -a --filter "name=devbox_"

# List devbox projects
devbox list

# Recreate missing box
devbox destroy myproject  # Clean up tracking
devbox init myproject     # Recreate

# Or force recreate
devbox init myproject --force
```

##### "Box won't start"

**Problem**: Box fails to start or immediately exits.

**Diagnosis**:
```bash
# Check box status
docker ps -a --filter "name=devbox_myproject"

# Check box logs
docker logs devbox_myproject

# Inspect box configuration
docker inspect devbox_myproject
```

**Solutions**:
```bash
# Try restarting box
docker start devbox_myproject

# If still fails, recreate box
devbox destroy myproject
devbox init myproject

# Check Docker daemon
sudo systemctl restart docker
```

##### "Box stops immediately after starting"

**Problem**: Box keeps exiting instead of staying running.

**Solutions**:
```bash
# Check what command box is running
docker inspect devbox_myproject | grep -A 5 '"Cmd"'

# Box should run 'sleep infinity'
# If not, recreate:
devbox destroy myproject
devbox init myproject

# Check for resource constraints
docker stats --no-stream
```

## File Access Issues
---

##### "Files not showing up in box"

**Problem**: Files created on host don't appear in `/workspace/` inside box.

**Diagnosis**:
```bash
# Check mount point
docker inspect devbox_myproject | grep -A 10 '"Mounts"'

# Should show: ~/devbox/myproject -> /workspace
```

**Solutions**:
```bash
# Verify workspace directory exists
ls -la ~/devbox/myproject/

# Create file on host and check in box
echo "test" > ~/devbox/myproject/test.txt
devbox run myproject cat /workspace/test.txt

# If mount is wrong, recreate box
devbox destroy myproject
devbox init myproject
```

##### "Permission denied accessing files"

**Problem**: Can't read/write files in box workspace.

**Solutions**:
```bash
# Check file permissions
ls -la ~/devbox/myproject/

# Fix ownership if needed
sudo chown -R $USER:$USER ~/devbox/myproject/

# Check box user
devbox run myproject whoami
devbox run myproject id

# If running as different user, use sudo inside box
devbox run myproject "sudo chown -R root:root /workspace/"
```

## Network and Port Issues
---

##### "Port already in use"

**Problem**: Can't bind to port specified in configuration.

**Solutions**:
```bash
# Check what's using the port
sudo netstat -tlnp | grep :5000
# or
sudo ss -tlnp | grep :5000

# Kill process using port
sudo kill -9 <PID>

# Or use different port in devbox.json
# Change "5000:5000" to "5001:5000"

# Recreate box with new config
devbox destroy myproject
devbox init myproject
```

##### "Can't access web application from host"

**Problem**: Web app running in box but not accessible from host.

**Solutions**:
```bash
# Ensure app binds to 0.0.0.0, not localhost
# In your app: app.run(host='0.0.0.0', port=5000)

# Check port mapping in box
docker port devbox_myproject

# Verify ports in devbox.json
cat ~/devbox/myproject/devbox.json

# Test from inside box
devbox run myproject "curl http://localhost:5000"

# Test from host
curl http://localhost:5000
```

## Configuration Issues
---

##### "Invalid JSON in devbox.json"

**Problem**: Configuration file has syntax errors.

**Solutions**:
```bash
# Validate JSON syntax
cat ~/devbox/myproject/devbox.json | python3 -m json.tool

# Or use devbox validation
devbox config validate myproject

# Fix common JSON errors:
# - Missing commas between elements
# - Trailing commas
# - Unquoted strings
# - Mismatched brackets/braces
```

##### "Setup commands fail during initialization"

**Problem**: Commands in `setup_commands` array fail.

**Diagnosis**:
```bash
# Check box logs during init
docker logs devbox_myproject

# Test commands manually
devbox shell myproject
# Run each setup command individually
```

**Solutions**:
```bash
# Common fixes:
# 1. Add 'apt update' before package installs (though devbox does this automatically)
# 2. Use full package names
# 3. Add '-y' flag to apt commands
# 4. Check command syntax

# Example working setup_commands:
{
  "setup_commands": [
    "apt install -y python3-pip nodejs npm",
    "pip3 install flask requests",
    "npm install -g typescript"
  ]
}

# Test commands step by step
devbox shell myproject
apt install -y python3-pip  # Should work
pip3 install flask          # Should work
```

## Performance Issues
---

##### "Box startup is slow"

**Problem**: Takes a long time to start boxes or run commands.

**Solutions**:
```bash
# Check Docker performance
docker system df
docker system prune  # Clean up unused resources

# Monitor during startup
time devbox shell myproject

# Check system resources
docker stats --no-stream
top
```

##### "High disk usage"

**Problem**: Docker/devbox using too much disk space.

**Solutions**:
```bash
# Check disk usage
devbox cleanup --dry-run --all
docker system df -v

# Clean up unused resources
devbox cleanup --all
docker system prune -a

# Check individual boxes
docker exec devbox_myproject du -sh /var/cache/apt
devbox run myproject "apt autoclean"
```

## Recovery Procedures
---

##### "Complete reset of devbox"

If everything is broken, start fresh:

```bash
# Stop all devbox boxes
docker stop $(docker ps -q --filter "name=devbox_")

# Remove all devbox boxes
docker rm $(docker ps -aq --filter "name=devbox_")

# Clean up Docker resources
docker system prune -a

# Remove devbox configuration
rm -rf ~/.devbox/

# Keep or remove project files (your choice)
# rm -rf ~/devbox/  # This deletes your code!

# Reinstall devbox if needed
curl -fsSL https://raw.githubusercontent.com/itzcozi/devbox/main/install.sh | bash
```

##### "Recover project after box deletion"

If box was deleted but files remain:

```bash
# Check if files exist
ls ~/devbox/myproject/

# Recreate box
devbox init myproject

# If you had custom configuration
# Edit ~/devbox/myproject/devbox.json
# Then recreate:
devbox destroy myproject
devbox init myproject
```

##### "Fix corrupted configuration"

If global configuration is corrupted:

```bash
# Backup existing config
cp ~/.devbox/config.json ~/.devbox/config.json.backup

# Reset configuration
rm ~/.devbox/config.json

# Recreate projects
devbox init project1
devbox init project2
# etc.
```

## Getting Help
---

##### Debug Information

When reporting issues, include:

```bash
# System information
uname -a
cat /etc/os-release

# Docker information
docker --version
docker info

# Devbox information
devbox --version
devbox list --verbose

# Box information (if applicable)
docker logs devbox_myproject
docker inspect devbox_myproject

# Configuration
cat ~/.devbox/config.json
cat ~/devbox/myproject/devbox.json
```

##### Log Files

Useful log locations:
- Docker daemon: `journalctl -u docker.service`
- Box logs: `docker logs devbox_<project>`
- System messages: `/var/log/syslog`

##### Common Commands for Diagnosis

```bash
# Check Docker daemon
sudo systemctl status docker

# List all boxes
docker ps -a

# Check Docker disk usage
docker system df

# Test Docker functionality
docker run hello-world

# Check devbox projects
devbox list
devbox maintenance --health-check

# Check system resources
df -h
free -h
```
