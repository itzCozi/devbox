# devbox Installation and Quick Start Guide

## Quick Installation

### Prerequisites
- Debian/Ubuntu Linux system
- Docker installed and running
- Go 1.21+ (for building from source)

### Install Docker (if not already installed)
```bash
# Update package index
sudo apt update

# Install Docker
sudo apt install docker.io

# Start and enable Docker
sudo systemctl start docker
sudo systemctl enable docker

# Add your user to docker group (requires logout/login)
sudo usermod -aG docker $USER
```

### Build and Install devbox
```bash
# Clone or download the devbox source code
cd devbox

# Build the binary
make build

# Install to system (requires sudo)
make install

# Verify installation
devbox --help
```

## Quick Start

### 1. Create your first project
```bash
devbox init myproject
```

### 2. Enter the box
```bash
devbox shell myproject
```

### 3. Install packages (inside box)
```bash
# You're now inside the Ubuntu box
apt update
apt install python3 python3-pip nodejs npm

# Install Python packages
pip3 install flask requests

# Create a simple app
cd /workspace
echo "print('Hello from devbox!')" > hello.py
python3 hello.py
```

### 4. Run commands from host
```bash
# Exit the box (Ctrl+D or 'exit')
# Run commands from your host system
devbox run myproject python3 hello.py
devbox run myproject "ls -la /workspace"
```

### 5. List your projects
```bash
devbox list
```

### 6. Clean up when done
```bash
devbox destroy myproject
```

### 7. System maintenance
```bash
# Check system health
devbox maintenance --health-check

# Clean up unused resources
devbox cleanup --all

# Update all boxes
devbox maintenance --update
```

## Common Workflows

### Python Development
```bash
# Create project
devbox init python-dev

# Enter box and setup environment
devbox shell python-dev
apt update && apt install python3 python3-pip python3-venv
pip3 install flask django fastapi pytest

# Create virtual environment (optional)
python3 -m venv /workspace/venv
source /workspace/venv/bin/activate
```

### Node.js Development
```bash
# Create project
devbox init node-app

# Setup Node.js environment
devbox run node-app "apt update && apt install nodejs npm"
devbox run node-app "npm init -y"
devbox run node-app "npm install express"
```

### Full Stack Development
```bash
# Create project
devbox init fullstack

# Enter box and install everything
devbox shell fullstack
apt update
apt install python3 python3-pip nodejs npm git curl wget

# Install Python packages
pip3 install flask django fastapi

# Install Node.js packages
npm install -g typescript vue-cli create-react-app
```

## File Locations

- **Project files**: `~/devbox/<project>/` (on host)
- **Box workspace**: `/workspace/` (inside box)
- **Configuration**: `~/.devbox/config.json`

## Troubleshooting

### Command not found
```bash
# If devbox command not found after install
export PATH="/usr/local/bin:$PATH"
# Or restart your terminal
```

### Docker permission denied
```bash
# Add user to docker group
sudo usermod -aG docker $USER
# Logout and login again
```

### Box won't start
```bash
# Check Docker status
sudo systemctl status docker

# Check box logs
docker logs devbox_<project>

# Restart Docker
sudo systemctl restart docker
```

### Remove everything
```bash
# List all projects
devbox list

# Destroy all projects
devbox destroy project1
devbox destroy project2

# Remove project directories
rm -rf ~/devbox/

# Remove configuration
rm -rf ~/.devbox/
```

## Tips

1. **Host file editing**: Edit files with your favorite editor on the host - they're instantly available in the box at `/workspace/`

2. **Persistent installs**: Packages installed with `apt` persist between box restarts

3. **Port forwarding**: Use `docker run -p` syntax in custom boxes if you need exposed ports

4. **Resource limits**: Boxes share host resources - monitor usage with `docker stats`

5. **Backup projects**: Your code in `~/devbox/<project>/` is safe even if boxes are destroyed

6. **Multiple projects**: Each project is completely isolated - install different Python versions, Node versions, etc.

## Advanced Usage

### Custom base images
Edit `~/.devbox/config.json` to change base images for new projects.

### Project templates
Create setup scripts like `examples/setup-python-web.sh` and run them after `devbox init`.

### Sharing projects
Copy the `~/devbox/<project>/` directory to share project files (boxes are recreated on each system).

---

Happy coding with devbox! 🚀