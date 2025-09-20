---
title: Frequently Asked Questions
description: Answers to common questions about devbox
---

Below are answers to the most common questions. If you're stuck, also check out the Troubleshooting page or ask for help.

## General
---

##### What is devbox?
Devbox creates isolated development environments using Docker (called "boxes"). Your code stays on the host while tools and dependencies live inside a box.

##### What platforms are supported?
Official support targets Debian/Ubuntu Linux. On Windows, use WSL2 with Ubuntu. macOS is not officially supported.

##### Where are my project files stored?
On the host in a simple folder (e.g., `~/devbox/<project>`). Inside the box, they're mounted at `/workspace` by default.

##### Does devbox require root?
Docker usually requires membership in the `docker` group. Devbox runs its Docker commands on the host; inside the box it often runs as root for setup convenience.

##### Do I need Docker Desktop or just Docker Engine?
You need a working Docker daemon accessible from Linux. On native Linux, install Docker Engine. On Windows, run inside WSL2 (Ubuntu) and either enable Docker Desktop's WSL integration or install Docker in the WSL distro. Devbox checks `docker version` before running.

##### Where does devbox store its own configuration?
Global config lives at `~/.devbox/config.json`. It tracks projects and settings like default base image and auto‑stop. A project‑local config file is `devbox.json` in your workspace (optional but recommended).

##### What project name characters are allowed?
Alphanumeric, hyphen, and underscore only: `^[a-zA-Z0-9_-]+$`. Names become part of the box name (e.g., `devbox_my-project`).

## Usage
---

##### How do I create a project?
Run `devbox init <project>`. Optionally use `--template` (python, nodejs, go, web) or `--generate-config` to produce a `devbox.json`.

##### How do I enter the environment?
Use `devbox shell <project>` for an interactive shell, or `devbox run <project> "<command>"` for a one‑off command.

##### Where is the configuration file?
In your project directory as `devbox.json`. You can generate one with `devbox init <project> --generate-config` or customize templates.

##### Can I keep my box running?
By default, global setting `auto_stop_on_exit` is enabled which prefers not restarting containers automatically. You can specify `"restart": "unless-stopped"` in `devbox.json` to keep it running, or toggle the global setting.

##### How do I expose ports?
Add mappings like `"ports": ["3000:3000", "8000:8000"]` in `devbox.json`. Recreate or restart the box to apply changes.

##### How do I mount extra folders?
Use `"volumes"` in `devbox.json`, for example: `"volumes": ["/var/run/docker.sock:/var/run/docker.sock", "/path/on/host:/path/in/box"]`.

##### What does `devbox up` do?
From a folder that contains `devbox.json`, `devbox up` starts the environment defined by that file so new teammates can run it without `init`. Use `--keep-running` to avoid auto‑stop, and `--dotfiles <path>` to mount local dotfiles.

##### How do I prevent the box from stopping after I exit?
Use `--keep-running` with `devbox shell`, `devbox run`, or `devbox up`. Or set `"restart": "unless-stopped"` in `devbox.json`.

##### How do I share my setup with teammates?
Commit `devbox.json` to your repo. Teammates clone the repo and run `devbox up` (or `devbox init <name> --generate-config` if they want a managed entry) to reproduce the environment.

## Templates & Packages
---

##### What templates are available?
Built-in templates: `python`, `nodejs`, `go`, and `web`. You can also create custom templates in `~/.devbox/templates/` as JSON files.

##### Can I install Docker tools inside the box?
Yes. By default, devbox mounts the Docker socket and many templates install Docker CLI, enabling Docker‑in‑Docker workflows.

##### How do setup commands work?
`setup_commands` run inside the box after creation. Use them to install packages and tools. For example:

```
{
  "setup_commands": [
    "apt install -y python3-pip",
    "pip3 install flask"
  ]
}
```

##### How do I create my own template?
Save a JSON file in `~/.devbox/templates/<name>.json` with a `config` object that mirrors `devbox.json` fields. List available templates with `devbox config templates`, and use it via `devbox init <project> --template <name>`.

##### Are package installs recorded anywhere?
Yes. Inside the box, devbox wraps common package managers (apt, pip/pip3, npm/yarn/pnpm/corepack) and appends successful install/remove commands to `/workspace/devbox.lock`. You can replay them during updates. To change or disable it, set the `DEVBOX_LOCKFILE` env var in `devbox.json` (empty to disable, or set a custom path).

## Management
---

##### How do I list or remove projects?
Use `devbox list` to list, and `devbox destroy <project>` to remove the box and clean up tracking. Your host files remain unless you delete them manually.

##### How do I update the base image or config?
Edit `devbox.json` (e.g., `base_image`, ports, volumes). Then `devbox destroy <project>` and `devbox init <project>` to recreate with the new config.

##### Where is the global config stored?
In `~/.devbox/config.json`. It tracks projects and global settings like default base image and auto‑update/auto‑stop behavior.

##### How do I update all environments to the latest base image?
Run `devbox update` to update all, or `devbox update <project>` for one. This pulls the latest base image, recreates the box, and re-runs recorded/setup steps.

##### How do I rebuild everything from scratch?
`devbox maintenance --rebuild` destroys and recreates all managed boxes using your current configs.

##### How do I clean up orphaned boxes?
If a box exists but isn't tracked, run `devbox destroy --cleanup-orphaned` to remove untracked `devbox_*` containers.

##### How do I see more details in the list output?
Use `devbox list --verbose` to include config presence, base image overrides, ports, and setup command counts.

##### Can I enable shell autocompletion?
Yes. Generate completion scripts with `devbox completion bash|zsh|fish` and follow the printed instructions.

##### How do I change global defaults like base image or auto‑stop?
Edit `~/.devbox/config.json` under `settings` (e.g., `default_base_image`, `auto_stop_on_exit`, `auto_update`).

## Advanced configuration
---

##### How do I mount my dotfiles into the box?
Add a `dotfiles` entry in `devbox.json` or pass `--dotfiles <path>` to `devbox up`. Devbox mounts the directory at `/dotfiles` and symlinks common files like `.gitconfig`, `.vimrc`, `.bashrc`, and `.config/*` into the root user's home.

##### How do I run as a non‑root user or change the shell/working directory?
Use these fields in `devbox.json`:
- `"user"`: e.g., `"1000:1000"`
- `"shell"`: e.g., `"/bin/bash"`
- `"working_dir"`: default `/workspace`

##### Can I set CPU/memory limits?
Yes, via `resources`:

```
{
  "resources": { "cpus": "2", "memory": "4g" }
}
```

##### Do health checks exist?
Yes. Use `health_check` to configure Docker health checks:

```
{
  "health_check": {
    "test": ["CMD-SHELL", "curl -fsS http://localhost:8080/health || exit 1"],
    "interval": "30s",
    "timeout": "5s",
    "retries": 5
  }
}
```

##### How do I attach to a custom network or add capabilities/labels?
`devbox.json` supports `network`, `capabilities` (for `--cap-add`), and `labels`:

```
{
  "network": "my-net",
  "capabilities": ["SYS_ADMIN"],
  "labels": { "com.example.owner": "team-a" }
}
```

## Getting Help
---

- GitHub: https://github.com/itzcozi/devbox
- Telegram: http://t.me/devboxcli
- Website & Docs: https://devbox.ar0.eu
