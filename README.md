# devbox

**Isolated development environments for anything**

devbox creates isolated development environments, contained in a project's Docker box. Each project operates in its own disposable environment, while your code remains neatly organized in a simple, flat folder on the host machine.

## Features

- 🚀 **Instant Setup** - Create isolated development environments in seconds
- 🐳 **Docker-based** - Leverage the power of containers for consistent environments
- 📁 **Clean Organization** - Keep your code organized in simple, flat folders
- 🔧 **Configurable** - Define your environment with simple JSON configuration
- 🗑️ **Disposable** - Easily destroy and recreate environments as needed
- 🛡️ **Isolated** - Each project runs in its own container, preventing conflicts

## Requirements

- Linux (Debian/Ubuntu)
- Docker

## Installation

```bash
# Using the install script
curl -fsSL https://raw.githubusercontent.com/itzCozi/devbox/main/install.sh | bash

# OR

# Clone the repository
git clone https://github.com/itzCozi/devbox.git
cd devbox

# Build the binary
make build

# Install (optional)
sudo make install
```

## Quick Start

1. **Initialize a new project**
   ```bash
   devbox init my-project
   ```

2. **Enter the development environment**
   ```bash
   devbox shell my-project
   ```

3. **Run commands in the environment**
   ```bash
   devbox run my-project "python --version"
   ```

4. **List your environments**
   ```bash
   devbox list
   ```

5. **Clean up when done**
   ```bash
   devbox destroy my-project
   ```

## Configuration

devbox uses a simple `devbox.json` file to configure your environment:

```json
{
  "name": "my-project",
  "base_image": "ubuntu:22.04",
  "setup_commands": [
    "apt update && apt install -y python3 python3-pip",
    "pip3 install flask requests"
  ],
  "environment": {
    "PYTHON_ENV": "development"
  }
}
```

## Commands

- `devbox init <name>` - Initialize a new development environment
- `devbox shell <name>` - Enter an interactive shell in the environment
- `devbox run <name> <command>` - Run a command in the environment
- `devbox list` - List all environments
- `devbox destroy <name>` - Remove an environment
- `devbox config` - Manage configuration
- `devbox cleanup` - Clean up unused Docker resources
- `devbox maintenance` - Perform maintenance tasks

## Documentation

For detailed documentation, guides, and examples, visit:

**📖 [devbox.ar0.eu](https://devbox.ar0.eu)**

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is open source. See the LICENSE file for details.

---

**Created by BadDeveloper with 💙**