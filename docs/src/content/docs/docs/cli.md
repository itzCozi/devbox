---
title: CLI Reference
description: Comprehensive reference for all devbox commands and options
tableOfContents:
  minHeadingLevel: 2
  maxHeadingLevel: 4
---

Complete reference for all devbox commands, options, and usage patterns.

## Global Options

---

All commands support these global options:

- `--help, -h`: Show help information

## Core Commands

---

### `devbox status`

Show detailed container status and resource usage for a project. With no project specified, prints a quick overview of all devbox containers.

**Syntax:**
```bash
devbox status [project]
```

**Behavior:**
- With a project: shows state, uptime, CPU%, memory usage/%, network I/O, block I/O, PIDs, ports, and mounts
- Without a project: lists all devbox containers with status and image

**Examples:**
```bash
# Overview of all devbox containers
devbox status

# Detailed status for a specific project
devbox status myproject
```

---

### `devbox up`

Start a devbox environment from a shared devbox.json in the current directory. Perfect for onboarding: clone the repo and run `devbox up`.

**Syntax:**
```bash
devbox up [--dotfiles <path>]
```

**Options:**
- `--dotfiles <path>`: Mount a local dotfiles directory into common locations inside the box

**Behavior:**
- Reads `./devbox.json`
- Creates/starts a box named `devbox_<name>` where `<name>` comes from `devbox.json`'s `name` (or the folder name)
- Applies ports, env, and volumes from configuration
- Runs a system update, then `setup_commands`
- Installs the devbox wrapper for nice shell UX
 - Records package installations you perform inside the box to `devbox.lock` (apt/pip/npm/yarn/pnpm). On rebuilds, these commands are replayed to reproduce the environment.

**Examples:**
```bash
# Start from current folder's devbox.json
devbox up

# Mount your dotfiles
devbox up --dotfiles ~/.dotfiles
```

---

### `devbox init`

Create a new devbox project with its own Docker box (container).

**Syntax:**
```bash
devbox init <project> [flags]
```

**Options:**
- `--force, -f`: Force initialization, overwriting existing project
- `--template, -t <template>`: Initialize from template (python, nodejs, go, web)
- `--generate-config, -g`: Generate devbox.json configuration file
- `--config-only, -c`: Generate configuration file only (don't create box)

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

### `devbox shell`

Open an interactive bash shell in the project's box.

**Syntax:**
```bash
devbox shell <project> [--keep-running]
```

**Examples:**
```bash
# Enter project environment
devbox shell myproject

# Start stopped box and enter shell
devbox shell python-app
```

**Notes:**
- Automatically starts the box if stopped
- Sets working directory to `/workspace`
- Your project files are available at `/workspace`
- Exit with `exit`, `logout`, or `Ctrl+D`
- By default, the box stops automatically after you exit the shell when global setting `auto_stop_on_exit` is enabled (default)
- Use `--keep-running` to keep the box running after you exit the shell

---

### `devbox run`

Run an arbitrary command inside the project's box.

**Syntax:**
```bash
devbox run <project> <command> [args...] [--keep-running]
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
- Box starts automatically if stopped
- By default, the box stops automatically after the command finishes when global setting `auto_stop_on_exit` is enabled (default)
- Use `--keep-running` to keep the box running after the command finishes

---

### `devbox stop`

Stop a project's box if it's running.

**Syntax:**
```bash
devbox stop <project>
```

**Examples:**
```bash
# Stop a running box
devbox stop myproject

# Stop another project's box
devbox stop webapp
```

**Notes:**
- Safe to run if the box is already stopped (no-op)
- Complements the default auto-stop behavior after `shell` and `run`

---

### `devbox destroy`

Stop and remove the project's box.

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
- Box can be recreated with `devbox init`
- Use `rm -rf ~/devbox/<project>/` to remove files

---

### `devbox list`

Show all managed projects and their box status.

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
PROJECT              BOX                  STATUS          CONFIG       WORKSPACE
--------------------  --------------------  ---------------  ------------  ------------------------------
myproject            devbox_myproject     Up 2 hours      devbox.json  /home/user/devbox/myproject
webapp               devbox_webapp        Exited          none         /home/user/devbox/webapp

Total projects: 2
```

## Configuration Commands

---

### `devbox templates`

Manage devbox project templates (built-in and user-defined).

**Subcommands:**

#### `devbox templates list`
List available templates (built-in + user templates in `~/.devbox/templates`).

**Syntax:**
```bash
devbox templates list
```

#### `devbox templates show`
Show a template’s JSON (name, description, and config).

**Syntax:**
```bash
devbox templates show <name>
```

#### `devbox templates create`
Create `devbox.json` in the current directory from a template.

**Syntax:**
```bash
devbox templates create <name> [project]
```

**Examples:**
```bash
cd ~/devbox/myapp
devbox templates create python MyApp

# If project name omitted, folder name is used
devbox templates create nodejs
```

#### `devbox templates save`
Save the current folder’s `devbox.json` as a reusable user template in `~/.devbox/templates/<name>.json`.

**Syntax:**
```bash
devbox templates save <name>
```

#### `devbox templates delete`
Delete a user template by name.

**Syntax:**
```bash
devbox templates delete <name>
```

---

### `devbox config`

Manage devbox configurations.

**Subcommands:**

#### `devbox config generate`
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

#### `devbox config validate`
Validate project configuration file.

**Syntax:**
```bash
devbox config validate <project>
```

#### `devbox config show`
Display project configuration details.

**Syntax:**
```bash
devbox config show <project>
```

Note: Template listing and management has moved to the top-level `devbox templates` command.

#### `devbox config global`
Show global devbox configuration.

**Syntax:**
```bash
devbox config global
```

## Maintenance Commands

---

### `devbox version`

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

---

### `devbox cleanup`

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

### `devbox maintenance`

Perform maintenance tasks on devbox projects and boxes.

**Syntax:**
```bash
devbox maintenance [flags]
```

**Options:**
- `--status`: Show detailed system status
- `--health-check`: Check health of all projects
- `--update`: Update all boxes
- `--restart`: Restart stopped boxes
- `--rebuild`: Rebuild all boxes
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

---

### `devbox update`

Pull the latest base image(s) and rebuild environment box(es).

This command replaces boxes to ensure they are based on the newest upstream images, while preserving your workspace files on the host.

**Syntax:**
```bash
devbox update [project]
```

**Behavior:**
- When a project is specified, only that environment is updated
- With no project, all registered projects are updated
- Pulls the latest base image, recreates the box with current devbox.json config, and re-runs setup commands
 - Replays package install commands from `devbox.lock` to restore your previously installed packages

**Options:**
- None currently. Uses your existing configuration in `devbox.json` if present.

**Examples:**
```bash
# Update a single project
devbox update myproject

# Update all projects
devbox update
```

**Notes:**
- Your files remain in ~/devbox/<project>/ and are re-mounted into the new box
- If the project has a devbox.json, its settings (ports, env, volumes, etc.) are applied on rebuild
- System packages inside the box are updated as part of the rebuild
 - If the box exists, it will be stopped and replaced; if missing, it will be created

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

**Inside Box:**
```
/workspace/                 # Mounted from ~/devbox/<project>/
├── devbox.json            # Same files as host
├── your-files...
└── ...
```

## Shell Completion

---

### `devbox completion`

Generate completion scripts for your shell to enable tab autocompletion for devbox commands, flags, project names, and template names.

**Syntax:**
```bash
devbox completion [bash|zsh|fish]
```

**Supported Shells:**
- **Bash**: Autocompletion for commands, flags, project names, and templates (Linux)
- **Zsh**: Full autocompletion with descriptions (Linux)
- **Fish**: Intelligent completion with suggestions (Linux)

**Setup Instructions:**

**Bash:**
```bash
# Load completion for current session
source <(devbox completion bash)

# Install for all sessions (Linux)
sudo devbox completion bash > /etc/bash_completion.d/devbox


```

**Zsh:**
```bash
# Enable completion if not already enabled
echo "autoload -U compinit; compinit" >> ~/.zshrc

# Install for all sessions
devbox completion zsh > "${fpath[1]}/_devbox"

# Restart your shell or source ~/.zshrc
```

**Fish:**
```bash
# Load completion for current session
devbox completion fish | source

# Install for all sessions
devbox completion fish > ~/.config/fish/completions/devbox.fish
```



**What Gets Completed:**
- Command names (`init`, `shell`, `run`, `list`, etc.)
- Command flags (`--template`, `--force`, `--keep-running`)
- Project names for commands like `shell`, `run`, `stop`, `destroy`
- Template names for `--template` flag and `templates show/delete`

**Examples:**
```bash
# Tab completion examples (press TAB after typing)
devbox <TAB>                    # Shows: init, shell, run, list, etc.
devbox shell <TAB>              # Shows: your-project-names
devbox init myapp --template <TAB>  # Shows: python, nodejs, go, web
devbox templates show <TAB>     # Shows: available-template-names
```

## Docker Integration

---

Devbox creates boxes (Docker containers) with these characteristics:

- **Name**: `devbox_<project>`
- **Base Image**: `ubuntu:22.04` (configurable)
- **Working Directory**: `/workspace`
- **Mount**: `~/devbox/<project>` → `/workspace`
- **Restart Policy**: `unless-stopped`
- **Command**: `sleep infinity` (keeps box alive)

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
