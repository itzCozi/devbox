---
title: Configuration Files
description: Comprehensive configuration management with devbox.json and global settings
---

Devbox supports comprehensive configuration management through both global settings and project-specific configuration files.

## Overview
---

- **Global Configuration**: Stored in `~/.devbox/config.json` - manages projects and global settings
- **Project Configuration**: Stored as `devbox.json` in each project workspace - defines development environment
- **Templates**: Built-in templates for common development environments (Python, Node.js, Go, Web)

## Project Configuration File (devbox.json)
---

Each project can have a `devbox.json` file in its workspace directory that defines the development environment configuration.

##### Basic Structure

```json
{
  "name": "my-project",
  "base_image": "ubuntu:22.04",
  "setup_commands": [
    "apt install -y python3 python3-pip"
  ],
  "environment": {
    "PYTHONPATH": "/workspace"
  },
  "ports": ["5000:5000"],
  "volumes": ["/workspace/data:/data"]
}
```

##### Complete Configuration Options

```json
{
  "name": "example-project",
  "base_image": "ubuntu:22.04",
  "setup_commands": [
    "apt install -y python3 python3-pip nodejs npm"
  ],
  "environment": {
    "PYTHONPATH": "/workspace",
    "NODE_ENV": "development",
    "PYTHONUNBUFFERED": "1"
  },
  "ports": [
    "3000:3000",
    "5000:5000",
    "8080:8080"
  ],
  "volumes": [
    "/workspace/data:/data",
    "/workspace/logs:/var/log/app"
  ],
  "working_dir": "/workspace",
  "shell": "/bin/bash",
  "user": "root",
  "capabilities": ["SYS_PTRACE"],
  "labels": {
    "devbox.project": "example-project",
    "devbox.type": "development"
  },
  "network": "bridge",
  "restart": "unless-stopped",
  "resources": {
    "cpus": "2.0",
    "memory": "2g"
  },
  "health_check": {
    "test": ["CMD", "curl", "-f", "http://localhost:5000/health"],
    "interval": "30s",
    "timeout": "10s",
    "retries": 3
  }
}
```

##### Configuration Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Project name (required) |
| `base_image` | string | Docker base image |
| `setup_commands` | array | Commands to run during initialization (after system update) |
| `environment` | object | Environment variables |
| `ports` | array | Port mappings (host:container) |
| `volumes` | array | Volume mappings |
| `working_dir` | string | Working directory in container |
| `shell` | string | Default shell |
| `user` | string | User to run as |
| `capabilities` | array | Linux capabilities to add |
| `labels` | object | Docker labels |
| `network` | string | Docker network to use |
| `restart` | string | Restart policy |
| `resources` | object | Resource constraints |
| `health_check` | object | Health check configuration |

:::note
Regardless of configuration, devbox always runs `apt update -y && apt full-upgrade -y` first when initializing any container to ensure the system is up to date. Your `setup_commands` will run after this system update.
:::

## System Updates
---

Devbox automatically ensures that all containers start with up-to-date system packages. When you initialize any project (regardless of configuration), devbox will:

1. **System Update**: Always run `apt update -y && apt full-upgrade -y` first
2. **Setup Commands**: Then execute any commands defined in your `setup_commands` array
3. **Devbox Setup**: Finally configure the devbox environment and shell

This ensures:
- Security patches are applied
- Package repositories are current
- Base system is consistent
- Your setup commands work with the latest packages

##### Execution Order

```bash
# 1. System update (automatic, always runs)
apt update -y
apt full-upgrade -y

# 2. Your setup commands (from devbox.json)
apt install -y python3 python3-pip
pip3 install flask

# 3. Devbox environment setup (automatic)
# - Install devbox wrapper script
# - Configure shell environment
# - Set up project-specific settings
```

You don't need to include `apt update` in your `setup_commands` - it's handled automatically.

## Initialize with Configuration
---

```bash
# Basic initialization
devbox init myproject

# Initialize with template
devbox init myproject --template python
devbox init myproject --template nodejs
devbox init myproject --template go
devbox init myproject --template web

# Generate config file only
devbox init myproject --config-only --template python

# Initialize and generate config
devbox init myproject --generate-config
```

## Configuration Management
---

```bash
# Generate devbox.json for existing project
devbox config generate myproject

# Validate project configuration
devbox config validate myproject

# Show project configuration
devbox config show myproject

# List available templates
devbox config templates

# Show global configuration
devbox config global
```

## Built-in Templates
---

##### Python Template
- Ubuntu 22.04 base
- Python 3, pip, venv, development tools
- Common Python packages
- PYTHONPATH and PYTHONUNBUFFERED environment
- Ports 5000, 8000

##### Node.js Template  
- Ubuntu 22.04 base
- Node.js 18, npm, build tools
- Latest npm version
- NODE_ENV development environment
- Ports 3000, 8080

##### Go Template
- Ubuntu 22.04 base
- Go 1.21, git, build tools
- GOPATH environment setup
- Port 8080

##### Web Template
- Ubuntu 22.04 base
- Python, Node.js, nginx
- Flask, Django, FastAPI
- TypeScript, Vue CLI, Create React App
- Multiple ports for different services

## Usage Examples
---

##### Python Development Project

```bash
# Create Python project
devbox init python-app --template python

# The generated devbox.json includes:
# - Python 3 with pip and venv
# - Development tools
# - PYTHONPATH configuration
# - Common ports
```

##### Custom Configuration

1. Initialize project:
```bash
devbox init custom-project --generate-config
```

2. Edit `devbox.json`:
```json
{
  "name": "custom-project",
  "base_image": "ubuntu:22.04",
  "setup_commands": [
    "apt install -y postgresql-client redis-tools"
  ],
  "environment": {
    "DATABASE_URL": "postgresql://localhost/mydb",
    "REDIS_URL": "redis://localhost:6379"
  },
  "ports": ["5432:5432", "6379:6379"]
}
```

3. Recreate with new configuration:
```bash
devbox destroy custom-project
devbox init custom-project
```

## Global Configuration
---

Global settings are stored in `~/.devbox/config.json`:

```json
{
  "projects": {
    "my-project": {
      "name": "my-project",
      "container_name": "devbox_my-project",
      "base_image": "ubuntu:22.04",
      "workspace_path": "/home/user/devbox/my-project",
      "config_file": "/home/user/devbox/my-project/devbox.json"
    }
  },
  "settings": {
    "default_base_image": "ubuntu:22.04",
    "auto_update": true,
    "default_environment": {
      "TZ": "UTC"
    }
  }
}
```

## Migration
---

Existing projects continue to work without configuration files. You can:

1. Generate configuration for existing projects:
```bash
devbox config generate existing-project
```

2. Apply templates to existing projects:
```bash
devbox config generate existing-project --template python
```

3. Recreate projects with new configuration:
```bash
devbox destroy old-project
devbox init old-project --template nodejs
```

## Best Practices
---

1. **Use Templates**: Start with built-in templates for common environments
2. **Version Control**: Include `devbox.json` in your project repository
3. **Environment Variables**: Store non-sensitive config in environment section
4. **Port Management**: Define all needed ports in configuration
5. **Setup Commands**: Use for environment setup, package installation
6. **Resource Limits**: Set appropriate CPU and memory limits
7. **Health Checks**: Define health checks for long-running services

## Error Handling
---

- Invalid JSON in `devbox.json` will show parsing errors
- Missing required fields are validated before container creation
- Invalid port/volume formats are caught during validation
- Failed setup commands stop initialization with clear error messages

## Configuration Precedence
---

1. Project `devbox.json` configuration (highest priority)
2. Global project settings in `~/.devbox/config.json`
3. Global default settings
4. Built-in defaults (lowest priority)

This allows for flexible configuration management while maintaining backward compatibility.