---
title: Quick Start Guide
description: Get up and running with devbox in minutes
---

This guide will get you up and running with devbox in just a few minutes. You'll create your first isolated development environment and learn the basic workflow.

## Prerequisites
---

Before starting, make sure you have devbox installed. If you haven't installed it yet, follow the [Installation Guide](/docs//install/) first.

## Create Your First Project
---

Let's create a Python development environment:

```bash
devbox init my-python-app --template python
```

This command:
- Creates a new project called `my-python-app`
- Uses the Python template (includes Python 3, pip, and common tools)
- Sets up a Docker container with Ubuntu 22.04
- Creates a workspace directory at `~/devbox/my-python-app/`

## Enter Your Development Environment
---

```bash
devbox shell my-python-app
```

You're now inside an isolated Ubuntu container! Notice how your prompt changes to indicate you're in the devbox environment.

## Explore the Environment
---

Inside the container, you can:

```bash
# Check what's available
python3 --version
pip3 --version
which python3

# Your workspace is mounted at /workspace
cd /workspace
ls -la

# Install additional packages
apt update
apt install tree htop

# Install Python packages
pip3 install requests flask
```

## Create and Run Code
---

Create a simple Python application:

```bash
# Create a simple web app
cat > /workspace/app.py << 'EOF'
from flask import Flask

app = Flask(__name__)

@app.route('/')
def hello():
    return 'Hello from devbox! 🚀'

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)
EOF

# Run your app
python3 app.py
```

## Manage Your Projects
---

```bash
# List all your projects
devbox list

# Create more projects
devbox init node-app --template nodejs
devbox init go-service --template go

# Each project is completely isolated
devbox shell node-app    # Node.js environment
devbox shell go-service  # Go environment
```

## Clean Up
---

When you're done with a project:

```bash
# Stop and remove the container (keeps your files)
devbox destroy my-python-app

# Your files are still in ~/devbox/my-python-app/
ls ~/devbox/my-python-app/

# To recreate the environment later:
devbox init my-python-app --template python
```

## Understanding the Workflow
---

Here's the typical devbox workflow:

1. **Create**: `devbox init <project>` - Creates isolated environment
2. **Develop**: `devbox shell <project>` - Enter the environment
3. **Code**: Edit files on host, run inside container
4. **Execute**: `devbox run <project> <command>` - Run commands without entering
5. **Manage**: `devbox list` - See all projects
6. **Cleanup**: `devbox destroy <project>` - Remove when done

## Next Steps
---

Now that you understand the basics:

1. **Learn about configuration**: [Configuration Guide](/docs/configuration/)
2. **Explore templates**: Try different [project templates](/docs/templates/)
3. **Customize**: Create a custom `devbox.json` config file
4. **Maintenance**: [Cleanup and Maintenance](/docs/cleanup-maintenance/)
