---
title: Troubleshooting
description: Common issues and solutions for devbox
---

This guide covers common issues you might encounter with devbox and how to resolve them.

## Installation Issues
---

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

## Container Issues
---

##### "Container not found" or "No such container"

**Problem**: Container was manually deleted or doesn't exist.

**Solutions**:
```bash
# Check what containers exist
docker ps -a --filter "name=devbox_"

# List devbox projects
devbox list

# Recreate missing container
devbox destroy myproject  # Clean up tracking
devbox init myproject     # Recreate

# Or force recreate
devbox init myproject --force
```

##### "Container won't start"

**Problem**: Container fails to start or immediately exits.

**Diagnosis**:
```bash
# Check container status
docker ps -a --filter "name=devbox_myproject"

# Check container logs
docker logs devbox_myproject

# Inspect container configuration
docker inspect devbox_myproject
```

**Solutions**:
```bash
# Try restarting container
docker start devbox_myproject

# If still fails, recreate container
devbox destroy myproject
devbox init myproject

# Check Docker daemon
sudo systemctl restart docker
```

##### "Container stops immediately after starting"

**Problem**: Container keeps exiting instead of staying running.

**Solutions**:
```bash
# Check what command container is running
docker inspect devbox_myproject | grep -A 5 '"Cmd"'

# Container should run 'sleep infinity'
# If not, recreate:
devbox destroy myproject
devbox init myproject

# Check for resource constraints
docker stats --no-stream
```

## File Access Issues
---

##### "Files not showing up in container"

**Problem**: Files created on host don't appear in `/workspace/` inside container.

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

# Create file on host and check in container
echo "test" > ~/devbox/myproject/test.txt
devbox run myproject cat /workspace/test.txt

# If mount is wrong, recreate container
devbox destroy myproject
devbox init myproject
```

##### "Permission denied accessing files"

**Problem**: Can't read/write files in container workspace.

**Solutions**:
```bash
# Check file permissions
ls -la ~/devbox/myproject/

# Fix ownership if needed
sudo chown -R $USER:$USER ~/devbox/myproject/

# Check container user
devbox run myproject whoami
devbox run myproject id

# If running as different user, use sudo inside container
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

# Recreate container with new config
devbox destroy myproject
devbox init myproject
```

##### "Can't access web application from host"

**Problem**: Web app running in container but not accessible from host.

**Solutions**:
```bash
# Ensure app binds to 0.0.0.0, not localhost
# In your app: app.run(host='0.0.0.0', port=5000)

# Check port mapping in container
docker port devbox_myproject

# Verify ports in devbox.json
cat ~/devbox/myproject/devbox.json

# Test from inside container
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
# Check container logs during init
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

##### "Container startup is slow"

**Problem**: Takes a long time to start containers or run commands.

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

# Check individual containers
docker exec devbox_myproject du -sh /var/cache/apt
devbox run myproject "apt autoclean"
```

## Recovery Procedures
---

##### "Complete reset of devbox"

If everything is broken, start fresh:

```bash
# Stop all devbox containers
docker stop $(docker ps -q --filter "name=devbox_")

# Remove all devbox containers
docker rm $(docker ps -aq --filter "name=devbox_")

# Clean up Docker resources
docker system prune -a

# Remove devbox configuration
rm -rf ~/.devbox/

# Keep or remove project files (your choice)
# rm -rf ~/devbox/  # This deletes your code!

# Reinstall devbox if needed
curl -fsSL https://raw.githubusercontent.com/itzCozi/devbox/main/install.sh | bash
```

##### "Recover project after container deletion"

If container was deleted but files remain:

```bash
# Check if files exist
ls ~/devbox/myproject/

# Recreate container
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

# Container information (if applicable)
docker logs devbox_myproject
docker inspect devbox_myproject

# Configuration
cat ~/.devbox/config.json
cat ~/devbox/myproject/devbox.json
```

##### Log Files

Useful log locations:
- Docker daemon: `journalctl -u docker.service`
- Container logs: `docker logs devbox_<project>`
- System messages: `/var/log/syslog`

##### Common Commands for Diagnosis

```bash
# Check Docker daemon
sudo systemctl status docker

# List all containers
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

This troubleshooting guide should help resolve most common issues with devbox. If you encounter problems not covered here, check the container logs and Docker status for more specific error messages.