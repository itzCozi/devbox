# devbox

A CLI tool for creating isolated development environments using Docker boxes on Debian/Ubuntu Linux.

## Overview

`devbox` creates isolated development environments, contained in a project's Docker box. Each project operates in its own disposable environment, while your code remains neatly organized in a simple, flat folder on the host machine (`~/devbox/<project>`).

## Features

- ✅ **Isolated environments**: Each project runs in its own Docker box
- ✅ **Host file access**: Your code stays on the host filesystem for easy editing
- ✅ **Persistent boxes**: boxes restart automatically and persist between reboots
- ✅ **Simple commands**: Easy-to-use CLI with intuitive commands
- ✅ **Safety checks**: Validates Docker installation and prevents accidental overwrites
- ✅ **Configuration files**: Project-specific `devbox.json` configuration with templates
- ✅ **Environment templates**: Built-in templates for Python, Node.js, Go, and web development
- ✅ **Advanced Docker features**: Port mapping, volume mounting, environment variables, resource limits

## Requirements

- **Operating System**: Debian/Ubuntu Linux only
- **Docker**: Must be installed and running
- **Go**: 1.21+ (for building from source)

## Installation

### Option 1: One-Line Installation (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/itzCozi/devbox/main/install.sh | bash
```

This script will automatically:
- Check system compatibility (Debian/Ubuntu only)
- Install Go, Docker, make, and git if needed
- Clone the repository and build devbox
- Install devbox to `/usr/local/bin`
- Set up proper permissions

### Option 2: Build from Source

```bash
# Clone or download the source code
git clone https://github.com/itzCozi/devbox.git
cd devbox

# Build and install
make build
make install
```

### Option 3: Development Build

```bash
# Build for current OS/arch (for development)
make dev

# The binary will be in ./build/devbox
```

## Commands

### Core Commands

#### `devbox init <project>`

Create a new devbox project with its own Docker box.

```bash
devbox init myproject
devbox init myproject --template python    # Use Python template
devbox init myproject --template nodejs    # Use Node.js template
devbox init myproject --generate-config    # Create with devbox.json
devbox init myproject --config-only        # Generate config file only
```

Options:
- `--force, -f`: Force initialization, overwriting existing project
- `--template, -t`: Initialize from template (python, nodejs, go, web)
- `--generate-config, -g`: Generate devbox.json configuration file
- `--config-only, -c`: Generate configuration file only (don't create box)

#### `devbox shell <project>`

Open an interactive bash shell in the project's box.

```bash
devbox shell myproject
```

#### `devbox run <project> <command>`

Run an arbitrary command inside the project's box.

```bash
devbox run myproject ls -la
devbox run myproject apt update
devbox run myproject python3 --version
```

#### `devbox destroy <project>`

Stop and remove the project's box.

```bash
devbox destroy myproject
```

#### `devbox list`

Show all managed projects and their box status.

```bash
devbox list            # Basic list
devbox list --verbose  # Detailed list with configuration info
```

### Management Commands

#### `devbox config <command>`

Manage devbox configurations.

```bash
devbox config generate myproject    # Generate devbox.json
devbox config validate myproject    # Validate configuration
devbox config show myproject        # Show configuration details
devbox config templates             # List available templates
devbox config global                # Show global settings
```

#### `devbox cleanup [flags]`

Clean up Docker resources and devbox artifacts.

```bash
devbox cleanup                      # Interactive cleanup menu
devbox cleanup --orphaned           # Remove orphaned boxes only
devbox cleanup --images             # Remove unused images only
devbox cleanup --volumes            # Remove unused volumes only
devbox cleanup --networks           # Remove unused networks only
devbox cleanup --all                # Clean up everything
devbox cleanup --system-prune       # Run docker system prune
devbox cleanup --dry-run            # Show what would be cleaned
```

This command helps maintain a clean system by removing:
- Orphaned devbox boxes (not tracked in config)
- Unused Docker images, volumes, and networks
- Dangling build artifacts

#### `devbox maintenance [flags]`

Perform maintenance tasks on devbox projects and boxes.

```bash
devbox maintenance                  # Interactive maintenance menu
devbox maintenance --status         # Show detailed system status
devbox maintenance --health-check   # Check health of all projects
devbox maintenance --update         # Update all boxes
devbox maintenance --restart        # Restart stopped boxes
devbox maintenance --rebuild        # Rebuild all boxes
devbox maintenance --auto-repair    # Auto-fix common issues
```

This command provides:
- System health checks and status monitoring
- Automated updates for all boxes
- box restart and rebuild capabilities
- Auto-repair for common issues

### `devbox shell <project>`

Open an interactive bash shell in the project's box.

```bash
devbox shell myproject
```

This starts the box if it's not running and attaches an interactive shell.

**To exit the shell**: Use `exit`, `logout`, or press `Ctrl+D` to return to your host system.

### `devbox run <project> <command>`

Run an arbitrary command inside the project's box.

```bash
devbox run myproject ls -la
devbox run myproject apt update
devbox run myproject python3 --version
```

### `devbox destroy <project>`

Stop and remove the project's box.

```bash
devbox destroy myproject
```

Options:
- `--force, -f`: Force destruction without confirmation

**Note**: This preserves your project files in `~/devbox/<project>`. To completely remove everything, manually delete the directory.

### `devbox list`

Show all managed projects and their box status.

```bash
devbox list            # Basic list
devbox list --verbose  # Detailed list with configuration info
```

Example output:
```
DEVBOX PROJECTS
PROJECT              BOX            STATUS          CONFIG       WORKSPACE
--------------------  --------------------  ---------------  ------------  ------------------------------
myproject            devbox_myproject     Up 2 hours      devbox.json  /home/user/devbox/myproject
webapp               devbox_webapp        Exited          none         /home/user/devbox/webapp

Total projects: 2
```

## Usage Examples

### Creating a Python Development Environment

```bash
# Create a Python project with template
devbox init python-app --template python

# Enter the box (Python tools already installed)
devbox shell python-app

# Inside the box - Python environment is ready
cd /workspace
python3 --version  # Python 3 is installed
pip3 list          # Common packages are available

# Create a simple app
echo "print('Hello from devbox!')" > hello.py
python3 hello.py

# Exit the box (back to host)
exit  # or press Ctrl+D
```

### Using Configuration Files

```bash
# Create project with custom configuration
devbox init web-app --generate-config

# Edit the generated devbox.json
cat > ~/devbox/web-app/devbox.json << 'EOF'
{
  "name": "web-app",
  "base_image": "ubuntu:22.04",
  "setup_commands": [
    "apt update",
    "apt install -y python3 python3-pip nodejs npm nginx",
    "pip3 install flask gunicorn",
    "npm install -g pm2"
  ],
  "environment": {
    "FLASK_ENV": "development",
    "PORT": "5000"
  },
  "ports": ["80:80", "5000:5000", "3000:3000"]
}
EOF

# Recreate with new configuration
devbox destroy web-app
devbox init web-app
```

### Template-based Development

```bash
# Node.js project
devbox init node-api --template nodejs
devbox shell node-api
# Node.js and npm are ready to use

# Go project  
devbox init go-service --template go
devbox shell go-service
# Go toolchain is installed and configured

# Full-stack web development
devbox init fullstack --template web
devbox shell fullstack
# Python, Node.js, and nginx are all available
```

### Running Commands from Host

```bash
# Exit the box (Ctrl+D or 'exit')
# Run commands from your host system
devbox run python-app python3 hello.py
devbox run python-app apt list --installed
devbox run python-app "cd /workspace && python3 -m http.server 8000"
```

### Managing Projects

```bash
# List all projects
devbox list

# Clean up a project
devbox destroy python-app

# Recreate if needed
devbox init python-app --force
```

## File Structure

```
devbox/
├── cmd/devbox/          # CLI entrypoint
│   └── main.go
├── internal/
│   ├── commands/        # CLI command implementations
│   │   ├── root.go
│   │   ├── init.go
│   │   ├── shell.go
│   │   ├── run.go
│   │   ├── destroy.go
│   │   └── list.go
│   ├── config/          # Configuration management
│   │   └── config.go
│   └── docker/          # Docker wrapper functions
│       └── client.go
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

## Configuration

devbox supports comprehensive configuration through project-specific `devbox.json` files and global settings.

### Project Configuration (devbox.json)

Create a `devbox.json` file in your project workspace to define the development environment:

```json
{
  "name": "my-project",
  "base_image": "ubuntu:22.04",
  "setup_commands": [
    "apt update",
    "apt install -y python3 python3-pip nodejs npm"
  ],
  "environment": {
    "PYTHONPATH": "/workspace",
    "NODE_ENV": "development"
  },
  "ports": ["3000:3000", "5000:5000"],
  "volumes": ["/workspace/data:/data"]
}
```

### Built-in Templates

- **python**: Python development with pip, venv, and common packages
- **nodejs**: Node.js development with npm and build tools  
- **go**: Go development with latest Go toolchain
- **web**: Full-stack web development (Python + Node.js + nginx)

### Configuration Management

```bash
devbox config generate myproject    # Generate devbox.json
devbox config validate myproject    # Validate configuration  
devbox config show myproject        # Show configuration details
devbox config templates             # List available templates
```

See [CONFIGURATION.md](CONFIGURATION.md) for complete documentation.

## Docker Integration

devbox uses Docker CLI commands under the hood:

- **Box Creation**: Uses `docker create` with bind mounts
- **Box Management**: Uses `docker start/stop/rm`
- **Command Execution**: Uses `docker exec` for shells and commands
- **Status Checking**: Uses `docker inspect` and `docker ps`

Each box:
- Is based on Ubuntu 22.04
- Has the restart policy set to `unless-stopped`
- Mounts `~/devbox/<project>` to `/workspace`
- Runs `sleep infinity` to stay alive
- Has working directory set to `/workspace`

## Safety Features

- **OS Validation**: Only runs on Linux systems
- **Docker Validation**: Checks if Docker is installed and running
- **Project Name Validation**: Ensures safe characters only (alphanumeric, hyphens, underscores)
- **Overwrite Protection**: Requires `--force` flag to overwrite existing projects
- **Confirmation Prompts**: Asks for confirmation before destructive operations

## Troubleshooting

### "Docker is not installed or not running"
Ensure Docker is installed and the Docker daemon is running:
```bash
sudo systemctl status docker
sudo systemctl start docker
```

### "Box not found"
If a box was manually deleted, recreate the project:
```bash
devbox destroy myproject
devbox init myproject
```

### "Permission denied"
Ensure your user is in the docker group:
```bash
sudo usermod -aG docker $USER
# Log out and back in
```

### Box won't start
Check Docker logs:
```bash
docker logs devbox_<project>
```

## Development

### Building

```bash
# Install dependencies
make deps

# Build for Linux
make build

# Build for current OS (development)
make dev

# Run tests
make test

# Format code
make fmt
```

### Code Organization

- `cmd/devbox/`: Main entrypoint
- `internal/commands/`: Cobra command implementations
- `internal/config/`: Configuration file management
- `internal/docker/`: Docker CLI wrapper functions

## Future Enhancements

Potential features for future versions:

- **Multi-OS Support**: Support for different Linux distributions beyond Ubuntu
- **Volume Management**: Better handling of persistent data and bind mounts
- **Network Management**: Custom Docker networks for multi-box projects
- **Resource Monitoring**: Real-time monitoring of box resource usage
- **Backup/Restore**: Backup and restore project configurations and data
- **Remote boxes**: Support for remote Docker hosts
- **Plugin System**: Extensible plugin system for custom functionality

## License

This project is open source. Feel free to use, modify, and distribute.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

---

**Note**: This tool is designed specifically for Debian/Ubuntu development workflows and requires Docker to be installed and running.
