---
title: Quick Start Guide
description: Get up and running with devbox in minutes
---

This guide will get you up and running with devbox in just a few minutes. You'll create your first isolated development environment and learn the basic workflow.

## Prerequisites
---

Before starting, make sure you have devbox installed. If you haven't installed it yet, follow the [Installation Guide](/guides/install/) first.

## Step 1: Create Your First Project
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

## Step 2: Enter Your Development Environment
---

```bash
devbox shell my-python-app
```

You're now inside an isolated Ubuntu container! Notice how your prompt changes to indicate you're in the devbox environment.

## Step 3: Explore the Environment
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

## Step 4: Create and Run Code
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

## Step 5: Edit Files from Your Host
---

Open another terminal on your host system. You can edit the files with your favorite editor:

```bash
# Files are in ~/devbox/my-python-app/ on your host
code ~/devbox/my-python-app/app.py   # VS Code
nano ~/devbox/my-python-app/app.py   # Nano
vim ~/devbox/my-python-app/app.py    # Vim
```

Changes are immediately visible inside the container!

## Step 6: Run Commands from Host
---

You don't always need to enter the container. Run commands directly:

```bash
# Exit the container first (Ctrl+D or type 'exit')
exit

# Run commands from your host
devbox run my-python-app python3 --version
devbox run my-python-app "cd /workspace && python3 app.py"
devbox run my-python-app "pip3 list"
```

## Step 7: Manage Your Projects
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

## Step 8: Clean Up
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

1. **Learn about configuration**: [Configuration Guide](/reference/configuration/)
2. **Explore templates**: Try different project types
3. **Advanced features**: Port mapping, volumes, custom environments
4. **Maintenance**: [Cleanup and Maintenance](/reference/cleanup-maintenance/)

## Commonly Used Templates
---

##### Web Development
```bash
# Full-stack project
devbox init webapp --template web
devbox shell webapp

# Now you have Python, Node.js, and nginx available
python3 --version && node --version
```

##### Data Science
```bash
devbox init data-project --template python
devbox shell data-project

# Install data science packages
pip3 install pandas numpy jupyter matplotlib seaborn
jupyter notebook --ip=0.0.0.0 --allow-root
```

##### Go Development
```bash
# Go is the best language!
devbox init learn-go --template go
devbox shell learn-go

go version
go run hello.go
```
