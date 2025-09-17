---
title: Templates & Setup
description: Built-in templates and custom setup configurations for devbox
---

Devbox provides built-in templates for common development environments and supports extensive customization through configuration files.

## Built-in Templates
---

#### Python Template

**Use Case**: Python development with popular packages and tools

**Includes**:
- Python 3.10+ with pip
- Virtual environment support (venv)
- Development tools (build-essential, git, curl, wget)
- Common Python packages
- Pre-configured environment variables

**Configuration**:
```json
{
  "name": "python-project",
  "base_image": "ubuntu:22.04",
  "setup_commands": [
    "apt install -y python3 python3-pip python3-venv python3-dev",
    "apt install -y build-essential git curl wget",
    "pip3 install --upgrade pip setuptools wheel",
    "pip3 install flask django fastapi requests pytest black flake8"
  ],
  "environment": {
    "PYTHONPATH": "/workspace",
    "PYTHONUNBUFFERED": "1",
    "PYTHONDONTWRITEBYTECODE": "1"
  },
  "ports": ["5000:5000", "8000:8000"],
  "working_dir": "/workspace"
}
```

**Usage**:
```bash
devbox init myapp --template python
devbox shell myapp

# Python is ready to use
python3 --version
pip3 list
python3 -m venv venv
source venv/bin/activate
```

---

#### Node.js Template

**Use Case**: Node.js development with npm and modern tooling

**Includes**:
- Node.js 18+ with npm
- Build tools and dependencies
- TypeScript support
- Common development utilities

**Configuration**:
```json
{
  "name": "nodejs-project",
  "base_image": "ubuntu:22.04",
  "setup_commands": [
    "curl -fsSL https://deb.nodesource.com/setup_18.x | bash -",
    "apt install -y nodejs build-essential git curl wget",
    "npm install -g npm@latest",
    "npm install -g typescript ts-node nodemon",
    "npm install -g @vue/cli create-react-app"
  ],
  "environment": {
    "NODE_ENV": "development",
    "NPM_CONFIG_PREFIX": "/workspace/.npm-global"
  },
  "ports": ["3000:3000", "8080:8080"],
  "working_dir": "/workspace"
}
```

**Usage**:
```bash
devbox init webapp --template nodejs
devbox shell webapp

# Node.js is ready
node --version
npm --version
npm init -y
npm install express
```

---

#### Go Template

**Use Case**: Go development with latest toolchain

**Includes**:
- Go 1.21+ with standard tools
- Git and build utilities
- Properly configured GOPATH and environment

**Configuration**:
```json
{
  "name": "go-project",
  "base_image": "ubuntu:22.04",
  "setup_commands": [
    "apt install -y wget git build-essential",
    "wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz",
    "tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz",
    "rm go1.21.0.linux-amd64.tar.gz"
  ],
  "environment": {
    "GOROOT": "/usr/local/go",
    "GOPATH": "/workspace/go",
    "PATH": "/usr/local/go/bin:/workspace/go/bin:$PATH"
  },
  "ports": ["8080:8080"],
  "working_dir": "/workspace"
}
```

**Usage**:
```bash
devbox init service --template go
devbox shell service

# Go is ready
go version
go mod init myproject
go run main.go
```

---

#### Web Template

**Use Case**: Full-stack web development with multiple languages and tools

**Includes**:
- Python 3 with web frameworks
- Node.js with frontend tools
- nginx for serving static files
- Database clients and tools

**Configuration**:
```json
{
  "name": "web-project",
  "base_image": "ubuntu:22.04",
  "setup_commands": [
    "curl -fsSL https://deb.nodesource.com/setup_18.x | bash -",
    "apt install -y python3 python3-pip nodejs nginx git curl wget",
    "apt install -y build-essential postgresql-client redis-tools",
    "pip3 install flask django fastapi gunicorn",
    "npm install -g typescript vue-cli create-react-app pm2"
  ],
  "environment": {
    "PYTHONPATH": "/workspace",
    "NODE_ENV": "development",
    "FLASK_ENV": "development"
  },
  "ports": [
    "80:80",
    "3000:3000", 
    "5000:5000",
    "8000:8000",
    "8080:8080"
  ],
  "working_dir": "/workspace"
}
```

**Usage**:
```bash
devbox init fullstack --template web
devbox shell fullstack

# Multiple languages available
python3 --version
node --version
nginx -v

# Create full-stack project
mkdir frontend backend
cd frontend && npx create-react-app .
cd ../backend && python3 -m flask --app app run
```

## Template Usage
---

##### Creating from Templates

```bash
# List available templates
devbox config templates

# Create project from template
devbox init myproject --template <template-name>

# Available templates: python, nodejs, go, web
devbox init api --template nodejs
devbox init ml-project --template python
devbox init microservice --template go
devbox init webapp --template web
```

##### Template Customization

You can customize templates by generating configuration first:

```bash
# Generate config from template
devbox init myproject --config-only --template python

# Edit the generated configuration
nano ~/devbox/myproject/devbox.json

# Then create the container
devbox init myproject
```

## Custom Configurations
---

##### Creating Custom Templates

Create your own reusable configurations:

```json
{
  "name": "data-science",
  "base_image": "ubuntu:22.04",
  "setup_commands": [
    "apt install -y python3 python3-pip jupyter",
    "pip3 install pandas numpy matplotlib seaborn scikit-learn",
    "pip3 install jupyter notebook jupyterlab",
    "pip3 install plotly dash streamlit"
  ],
  "environment": {
    "PYTHONPATH": "/workspace",
    "JUPYTER_ENABLE_LAB": "yes"
  },
  "ports": ["8888:8888", "8501:8501"],
  "working_dir": "/workspace"
}
```

##### Database Development

```json
{
  "name": "database-dev",
  "base_image": "ubuntu:22.04",
  "setup_commands": [
    "apt install -y postgresql-client mysql-client redis-tools",
    "apt install -y python3 python3-pip",
    "pip3 install psycopg2-binary pymongo redis sqlalchemy"
  ],
  "environment": {
    "PGHOST": "localhost",
    "REDIS_URL": "redis://localhost:6379"
  },
  "ports": ["5432:5432", "6379:6379", "27017:27017"],
  "working_dir": "/workspace"
}
```

##### DevOps/Infrastructure

```json
{
  "name": "devops",
  "base_image": "ubuntu:22.04",
  "setup_commands": [
    "apt install -y curl wget git jq unzip",
    "curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl",
    "chmod +x kubectl && mv kubectl /usr/local/bin/",
    "curl -fsSL https://get.docker.com | sh",
    "pip3 install awscli ansible terraform"
  ],
  "environment": {
    "KUBECONFIG": "/workspace/.kube/config"
  },
  "volumes": ["/var/run/docker.sock:/var/run/docker.sock"],
  "working_dir": "/workspace"
}
```

##### Mobile Development

```json
{
  "name": "mobile-dev",
  "base_image": "ubuntu:22.04",
  "setup_commands": [
    "curl -fsSL https://deb.nodesource.com/setup_18.x | bash -",
    "apt install -y nodejs default-jdk android-tools-adb",
    "npm install -g @ionic/cli cordova react-native-cli",
    "npm install -g expo-cli @react-native-community/cli"
  ],
  "environment": {
    "ANDROID_HOME": "/workspace/android-sdk",
    "JAVA_HOME": "/usr/lib/jvm/default-java"
  },
  "ports": ["19000:19000", "19001:19001", "19002:19002"],
  "working_dir": "/workspace"
}
```

## Advanced Setup Patterns
---

##### Multi-Stage Setup

For complex environments, use multi-stage setup commands:

```json
{
  "setup_commands": [
    "# Stage 1: System packages",
    "apt install -y curl wget git build-essential",
    
    "# Stage 2: Programming languages",
    "curl -fsSL https://deb.nodesource.com/setup_18.x | bash -",
    "apt install -y nodejs python3 python3-pip",
    
    "# Stage 3: Language-specific packages",
    "pip3 install flask requests",
    "npm install -g typescript express",
    
    "# Stage 4: Project-specific setup",
    "mkdir -p /workspace/{src,tests,docs}",
    "chown -R root:root /workspace"
  ]
}
```

##### Conditional Setup

Use shell conditionals for flexible setup:

```json
{
  "setup_commands": [
    "# Install Python if requirements.txt exists",
    "[ -f /workspace/requirements.txt ] && pip3 install -r /workspace/requirements.txt || true",
    
    "# Install Node.js dependencies if package.json exists", 
    "[ -f /workspace/package.json ] && cd /workspace && npm install || true",
    
    "# Setup git config if not exists",
    "git config --global user.name 'Developer' || true",
    "git config --global user.email 'dev@example.com' || true"
  ]
}
```

##### Environment-Specific Setup

```json
{
  "setup_commands": [
    "# Development tools for dev environment",
    "if [ \"$NODE_ENV\" = \"development\" ]; then",
    "  npm install -g nodemon ts-node",
    "  pip3 install black flake8 pytest",
    "fi",
    
    "# Production optimizations",
    "if [ \"$NODE_ENV\" = \"production\" ]; then",
    "  npm install -g pm2",
    "  pip3 install gunicorn",
    "fi"
  ]
}
```

## Best Practices
---

##### Setup Command Guidelines

1. **Always use -y flag** for apt commands
2. **Group related commands** for better readability
3. **Use full paths** for downloaded files
4. **Clean up temporary files** after installation
5. **Test commands individually** before adding to config

##### Environment Variables

1. **Use meaningful names** that describe purpose
2. **Set PATH modifications** in environment section
3. **Include language-specific variables** (PYTHONPATH, GOPATH, etc.)
4. **Use /workspace** as base for project-specific paths

##### Port Configuration

1. **Map common development ports** (3000, 5000, 8000, 8080)
2. **Include language-specific ports** (Python: 5000, 8000; Node.js: 3000, 8080)
3. **Reserve ports for databases** if needed (5432, 3306, 6379, 27017)
4. **Use consistent port mapping** across similar projects

##### Volume Management

1. **Mount data directories** outside /workspace for persistence
2. **Use absolute paths** for volume mappings
3. **Consider cache directories** for package managers
4. **Mount configuration files** if shared across projects

## Template Development
---

##### Creating New Templates

1. Start with working configuration:
```bash
devbox init test-project --generate-config
# Edit ~/devbox/test-project/devbox.json
devbox destroy test-project && devbox init test-project
```

2. Test thoroughly:
```bash
devbox shell test-project
# Test all tools and commands
# Verify environment variables
# Check port access
```

3. Document your template:
```json
{
  "name": "my-template",
  "description": "Custom template for X development",
  "base_image": "ubuntu:22.04",
  "setup_commands": [
    "# Add descriptive comments",
    "apt install -y tool1 tool2"
  ]
}
```

##### Sharing Templates

Save template configurations in your project repository:

```
project/
├── devbox.json          # Main configuration
├── templates/           # Additional templates
│   ├── development.json
│   ├── testing.json
│   └── production.json
└── scripts/
    └── setup.sh         # Additional setup scripts
```

Templates and custom configurations make devbox incredibly flexible for any development workflow!