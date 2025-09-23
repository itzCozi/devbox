---
title: Configuration Files
description: Comprehensive configuration management with devbox.json and global settings
---

Devbox supports configuration via a per-project `devbox.json` and a global `~/.devbox/config.json`.

## Project Configuration
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

##### Common Fields

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
  "dotfiles": ["~/.dotfiles"],
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

Key fields you may use: `name`, `base_image`, `setup_commands`, `environment`, `ports`, `volumes`, `dotfiles`, `working_dir`. Advanced options like `capabilities`, `labels`, `network`, `restart`, `resources`, and `health_check` are supported but optional.

:::note
Regardless of configuration, devbox always runs `apt update -y && apt full-upgrade -y` first when initializing any box to ensure the system is up to date. Your `setup_commands` will run after this system update.
:::

## Reproducible Installs
---

Devbox automatically records package manager installs you run inside the box to `/workspace/devbox.lock` (which is your project folder on the host).

Lock file paths:
- Inside box: `/workspace/devbox.lock`
- On host: `~/devbox/<project>/devbox.lock`

The following commands are tracked when they succeed:

- `apt install ...` and `apt-get install ...`
- `pip install ...` and `pip3 install ...`
- `npm install ...`, `npm i ...`, `npm add ...`
- `yarn add ...` and `yarn global add ...`
- `pnpm add ...`, `pnpm install ...`, and `pnpm i ...`

On `devbox up` and during `devbox update` rebuilds, devbox replays the commands from `devbox.lock` before running `setup_commands`. This makes it easy to reproduce the exact environment or share it with teammates by committing `devbox.lock` to your repo.

Notes:
- Only successful install commands are recorded, and duplicates are de-duplicated line-by-line.
- You can edit `devbox.lock` manually to remove mistakes or add comments (lines starting with `#` are ignored).
- If you prefer explicit configuration, keep using `setup_commands` in `devbox.json`; the lock file complements it for ad-hoc installs.

## Environment Snapshot
---

For a more comprehensive, shareable snapshot similar to Nix-style locks, use:

```bash
devbox lock <project>
``

This writes a JSON snapshot (by default to `<workspace>/devbox.lock.json`) that includes:

- Base image: name, digest (if available), and image ID
- Container configuration: working_dir, user, restart policy, network, ports, volumes, labels, environment, capabilities, resources (cpus/memory)
- Installed packages:
  - apt: manually installed packages pinned as `name=version`
  - pip: `pip freeze`
  - npm/yarn/pnpm: globally installed packages `name@version` (Yarn global versions are read from Yarn's global directory)
- Registries and sources for reproducibility:
  - pip: `index-url` and `extra-index-url`
  - npm/yarn/pnpm: global registry URLs
  - apt: full `sources.list` lines, snapshot base URL if present, and OS release codename
- Any `setup_commands` from your `devbox.json` (for context)

Usage notes:
- Commit `devbox.lock.json` to your repository to share environment details with teammates.
- This file is an authoritative snapshot for auditing/sharing. The current execution path for rebuilds remains `devbox.json` + the simple `devbox.lock` replay file. You can now also use:
  - `devbox verify <project>` to validate a box matches the lock (fails fast on drift)
  - `devbox apply <project>` to configure registries/sources and reconcile package sets to the lock
- Local app dependencies (e.g. non-global Node packages in your repo) are intentionally not included; rely on your project’s own lockfiles (package-lock.json, yarn.lock, pnpm-lock.yaml, requirements.txt/poetry.lock, etc.).

## Initialize with Configuration
---

```bash
# Basic initialization
devbox init myproject

# Initialize with template
devbox init myproject --template python

# Generate config file only
devbox init myproject --config-only --template python

# Initialize and generate config
devbox init myproject --generate-config
```

### Shared Configs

To make onboarding easy, commit your `devbox.json` to the repository. New teammates can simply run:

```bash
devbox up
```

This reads `./devbox.json` and starts the environment without requiring prior project registration.

### Dotfile Injection

You can mount your personal dotfiles into the box to keep your editor/shell preferences:

```bash
# One-off via flag
devbox up --dotfiles ~/.dotfiles

# Or persist via config
{
  "name": "my-project",
  "dotfiles": ["~/.dotfiles"]
}
```

Behavior summary: mount at `/dotfiles` and source/symlink common files on shell init.

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
    "auto_stop_on_exit": true,
    "default_environment": {
      "TZ": "UTC"
    }
  }
}
```

### Global Settings

| Setting | Type | Default | Description |
|--------|------|---------|-------------|
| `default_base_image` | string | `ubuntu:22.04` | Default base image for new projects |
| `auto_update` | boolean | `true` | Whether to run updates during initialization |
| `auto_stop_on_exit` | boolean | `true` | If enabled, devbox stops a project's box automatically after exiting an interactive shell or finishing a one-off `run` command. Override per-invocation with `--keep-running`. |

When `auto_stop_on_exit` is enabled:
- `devbox up` will also stop the container if it is idle right after setup (no ports exposed and only the init process running), unless `--keep-running` is passed.
- If your `devbox.json` does not specify a `restart` policy, devbox will default to `--restart no` so that manual stops persist.

Note: If `auto_stop_on_exit` is missing in older installs, add it under `settings`.

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

## Error Handling
---

- Invalid JSON in `devbox.json` will show parsing errors
- Missing required fields are validated before box creation
- Invalid port/volume formats are caught during validation
- Failed setup commands stop initialization with clear error messages

## Configuration Precedence
---

1. Project `devbox.json` configuration (highest priority)
2. Global project settings in `~/.devbox/config.json`
3. Global default settings
4. Built-in defaults (lowest priority)

This allows for flexible configuration management while maintaining backward compatibility.
