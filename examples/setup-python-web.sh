#!/bin/bash

# Example setup script for a Python web development environment
# This would be run inside the devbox box after initialization

set -e

echo "🚀 Setting up Python web development environment..."

# Update package lists
echo "📦 Updating packages..."
apt update

# Install Python and related tools
echo "🐍 Installing Python development tools..."
apt install -y \
    python3 \
    python3-pip \
    python3-venv \
    python3-dev \
    build-essential \
    git \
    curl \
    wget \
    vim \
    nano

# Install common Python packages
echo "📚 Installing Python packages..."
pip3 install \
    flask \
    django \
    fastapi \
    requests \
    beautifulsoup4 \
    pandas \
    numpy \
    jupyter \
    pytest \
    black \
    flake8

# Install Node.js (for frontend development)
echo "🟢 Installing Node.js..."
curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
apt install -y nodejs

# Install global npm packages
echo "📦 Installing npm packages..."
npm install -g \
    typescript \
    @types/node \
    create-react-app \
    vue-cli \
    eslint \
    prettier

# Create a sample project structure
echo "📁 Creating project structure..."
cd /workspace
mkdir -p {src,tests,docs,scripts}

# Create a sample Flask app
cat > src/app.py << 'EOF'
from flask import Flask, jsonify

app = Flask(__name__)

@app.route('/')
def hello():
    return jsonify({"message": "Hello from devbox!", "status": "running"})

@app.route('/health')
def health():
    return jsonify({"status": "healthy"})

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)
EOF

# Create a requirements.txt
cat > requirements.txt << 'EOF'
flask>=2.0.0
requests>=2.25.0
beautifulsoup4>=4.9.0
pandas>=1.3.0
numpy>=1.21.0
pytest>=6.0.0
black>=21.0.0
flake8>=3.9.0
EOF

# Create a simple test
cat > tests/test_app.py << 'EOF'
import pytest
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src'))

from app import app

def test_hello():
    client = app.test_client()
    response = client.get('/')
    assert response.status_code == 200
    data = response.get_json()
    assert data['message'] == "Hello from devbox!"

def test_health():
    client = app.test_client()
    response = client.get('/health')
    assert response.status_code == 200
    data = response.get_json()
    assert data['status'] == "healthy"
EOF

# Create a README for the project
cat > README.md << 'EOF'
# Sample Devbox Project

This is a sample Python web development project created with devbox.

## Getting Started

1. Start the development server:
   ```bash
   cd /workspace
   python3 src/app.py
   ```

2. Run tests:
   ```bash
   cd /workspace
   python3 -m pytest tests/
   ```

3. Format code:
   ```bash
   cd /workspace
   black src/ tests/
   ```

4. Lint code:
   ```bash
   cd /workspace
   flake8 src/ tests/
   ```

## Project Structure

- `src/` - Source code
- `tests/` - Test files
- `docs/` - Documentation
- `scripts/` - Utility scripts
- `requirements.txt` - Python dependencies

## API Endpoints

- `GET /` - Hello message
- `GET /health` - Health check
EOF

echo "✅ Environment setup complete!"
echo ""
echo "🎯 Next steps:"
echo "  cd /workspace"
echo "  python3 src/app.py"
echo ""
echo "🌐 Your Flask app will be available at http://localhost:5000"