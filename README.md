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
- 🔄 **Docker-in-Docker** - Use Docker within your devbox environments by default

## Requirements

- Linux (Debian/Ubuntu)
- Docker

## Installation

```bash
# Using the install script
curl -fsSL https://devbox.ar0.eu/install.sh | bash
# Or manually: https://devbox.ar0.eu/docs/install/#manual-build-from-source
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

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

**Created by BadDeveloper with 💚**