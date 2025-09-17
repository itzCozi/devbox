---
title: CLI Reference
description: Comprehensive reference for all devbox commands and options
---

Complete reference for all devbox commands, options, and usage patterns.

## Global Options

---

All commands support these global options:

- `--help, -h`: Show help information

## Core Commands

---

##### `devbox init`

Create a new devbox project with its own Docker container.

**Syntax:**
```bash
devbox init <project> [flags]
```

**Options:**
- `--force, -f`: Force initialization, overwriting existing project
- `--template, -t <template>`: Initialize from template (python, nodejs, go, web)
- `--generate-config, -g`: Generate devbox.json configuration file
- `--config-only, -c`: Generate configuration file only (don't create container)

**Examples:**
```bash
# Basic project
devbox init myproject

# Python project with template
devbox init python-app --template python

# Force overwrite existing project
devbox init myproject --force

# Generate config file only
devbox init myproject --config-only --template nodejs

# Create with custom configuration
devbox init webapp --generate-config
```

**Templates:**
- `python`: Python 3, pip, venv, development tools
- `nodejs`: Node.js 18, npm, build tools
- `go`: Go 1.21, git, build tools
- `web`: Python + Node.js + nginx for full-stack development

---

##### `devbox shell`

Open an interactive bash shell in the project's container.

**Syntax:**
```bash
devbox shell <project>
```

**Examples:**
```bash
# Enter project environment
devbox shell myproject

# Start stopped container and enter shell
devbox shell python-app
```

**Notes:**
- Automatically starts the container if stopped
- Sets working directory to `/workspace`
- Your project files are available at `/workspace`
- Exit with `exit`, `logout`, or `Ctrl+D`

---

##### `devbox run`

Run an arbitrary command inside the project's container.

**Syntax:**
```bash
devbox run <project> <command> [args...]
```

**Examples:**
```bash
# Run single command
devbox run myproject python3 --version

# Run with arguments
devbox run myproject apt install -y htop

# Complex command with pipes
devbox run myproject "cd /workspace && python3 -m http.server 8000"

# Execute script
devbox run myproject bash /workspace/setup.sh
```

**Notes:**
- Commands run in `/workspace` by default
- Use quotes for complex commands with pipes, redirects, etc.
- Container starts automatically if stopped

---

##### `devbox destroy`

Stop and remove the project's container.

**Syntax:**
```bash
devbox destroy <project> [flags]
```

**Options:**
- `--force, -f`: Force destruction without confirmation

**Examples:**
```bash
# Destroy with confirmation
devbox destroy myproject

# Force destroy without prompt
devbox destroy myproject --force
```

**Notes:**
- Preserves project files in `~/devbox/<project>/`
- Container can be recreated with `devbox init`
- Use `rm -rf ~/devbox/<project>/` to remove files

---

##### `devbox list`

Show all managed projects and their container status.

**Syntax:**
```bash
devbox list [flags]
```

**Options:**
- `--verbose, -v`: Show detailed information including configuration

**Examples:**
```bash
# Basic list
devbox list

# Detailed information
devbox list --verbose
```

**Output Format:**
```
DEVBOX PROJECTS
PROJECT              CONTAINER            STATUS          CONFIG       WORKSPACE
--------------------  --------------------  ---------------  ------------  ------------------------------
myproject            devbox_myproject     Up 2 hours      devbox.json  /home/user/devbox/myproject
webapp               devbox_webapp        Exited          none         /home/user/devbox/webapp

Total projects: 2
```

## Configuration Commands

---

##### `devbox config`

Manage devbox configurations.

**Subcommands:**

###### `devbox config generate`
Generate devbox.json configuration file for a project.

**Syntax:**
```bash
devbox config generate <project> [flags]
```

**Options:**
- `--template, -t <template>`: Use template configuration

**Examples:**
```bash
# Generate basic config
devbox config generate myproject

# Generate with template
devbox config generate myproject --template python
```

###### `devbox config validate`
Validate project configuration file.

**Syntax:**
```bash
devbox config validate <project>
```

###### `devbox config show`
Display project configuration details.

**Syntax:**
```bash
devbox config show <project>
```

###### `devbox config templates`
List available configuration templates.

**Syntax:**
```bash
devbox config templates
```

###### `devbox config global`
Show global devbox configuration.

**Syntax:**
```bash
devbox config global
```

## Maintenance Commands

---

##### `devbox version`

Display the version information for devbox.

**Syntax:**
```bash
devbox version
```

**Examples:**
```bash
# Display version information
devbox version
```

**Output Format:**
```
devbox (v1.0)
```

##### `devbox cleanup`

Clean up Docker resources and devbox artifacts.

**Syntax:**
```bash
devbox cleanup [flags]
```

**Options:**
- `--orphaned`: Remove orphaned containers only
- `--images`: Remove unused images only
- `--volumes`: Remove unused volumes only
- `--networks`: Remove unused networks only
- `--system-prune`: Run docker system prune
- `--all`: Clean up everything
- `--dry-run`: Show what would be cleaned (no changes)
- `--force`: Skip confirmation prompts

**Examples:**
```bash
# Interactive cleanup menu
devbox cleanup

# Clean specific resources
devbox cleanup --orphaned
devbox cleanup --images

# Comprehensive cleanup
devbox cleanup --all

# Preview cleanup actions
devbox cleanup --dry-run --all

# Cleanup without prompts
devbox cleanup --all --force
```

---

##### `devbox maintenance`

Perform maintenance tasks on devbox projects and containers.

**Syntax:**
```bash
devbox maintenance [flags]
```

**Options:**
- `--status`: Show detailed system status
- `--health-check`: Check health of all projects
- `--update`: Update all containers
- `--restart`: Restart stopped containers
- `--rebuild`: Rebuild all containers
- `--auto-repair`: Auto-fix common issues
- `--force`: Skip confirmation prompts

**Examples:**
```bash
# Interactive maintenance menu
devbox maintenance

# Individual tasks
devbox maintenance --health-check
devbox maintenance --update
devbox maintenance --restart

# Combined operations
devbox maintenance --health-check --update --restart

# Auto-repair issues
devbox maintenance --auto-repair

# Force operations without prompts
devbox maintenance --force --rebuild
```

## Exit Codes

---

Devbox uses standard exit codes:

- `0`: Success
- `1`: General error
- `2`: Invalid arguments or usage
- `125`: Docker daemon not running
- `126`: Container not executable
- `127`: Container/command not found

## Environment Variables

---

Devbox respects these environment variables:

- `DOCKER_HOST`: Docker daemon socket
- `DEVBOX_HOME`: Override default `~/.devbox` directory
- `DEVBOX_WORKSPACE`: Override default `~/devbox` workspace directory

## Project Structure

---

When you create a project, devbox sets up:

```
~/devbox/<project>/          # Project workspace (host)
├── devbox.json             # Configuration file (optional)
├── your-files...           # Your project files
└── ...

~/.devbox/                  # Global configuration
├── config.json            # Global settings and project registry
└── ...
```

**Inside Container:**
```
/workspace/                 # Mounted from ~/devbox/<project>/
├── devbox.json            # Same files as host
├── your-files...
└── ...
```

## Docker Integration

---

Devbox creates containers with these characteristics:

- **Name**: `devbox_<project>`
- **Base Image**: `ubuntu:22.04` (configurable)
- **Working Directory**: `/workspace`
- **Mount**: `~/devbox/<project>` → `/workspace`
- **Restart Policy**: `unless-stopped`
- **Command**: `sleep infinity` (keeps container alive)

**Docker Commands Equivalent:**
```bash
# devbox init myproject
docker create --name devbox_myproject \
  --restart unless-stopped \
  -v ~/devbox/myproject:/workspace \
  -w /workspace \
  ubuntu:22.04 sleep infinity

# devbox shell myproject
docker start devbox_myproject
docker exec -it devbox_myproject bash

# devbox run myproject <command>
docker exec devbox_myproject <command>

# devbox destroy myproject
docker stop devbox_myproject
docker rm devbox_myproject
```
