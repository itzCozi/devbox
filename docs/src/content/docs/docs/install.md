---
title: Installation Guide
description: How to install devbox on your Debian/Ubuntu system
---

```bash
# Primary (recommended)
curl -fsSL https://raw.githubusercontent.com/itzcozi/devbox/main/install.sh | bash

# Mirror (CDN)
curl -fsSL https://devbox.ar0.eu/install.sh | bash
```

If you encounter a 403 when using the mirror (common in some managed shells like AWS CloudShell), use the primary GitHub Raw URL instead.

This script will automatically:
- Check system compatibility (Debian/Ubuntu only)
- Install Go, Docker, make, and git if needed
- Clone the repository and build devbox
- Install devbox to `/usr/local/bin`
- Set up proper permissions

<sub>Already done here? Head over to the [Quick Start Guide](/docs/start/) to learn how to use devbox.</sub>

## Manual Build from Source
---

If you prefer to build devbox manually or the automatic script doesn't work for your system:

### Install Dependencies
```bash
sudo apt update \
	&& sudo apt install -y docker.io golang-go make git \
	&& sudo systemctl enable --now docker \
	&& sudo usermod -aG docker $USER
# Note: log out/in (or run `newgrp docker`) for group changes to take effect.
```

### Build and Install
```bash
# Clone the repository
git clone https://github.com/itzcozi/devbox.git
cd devbox

# Build the binary
make build

# Install to system (requires sudo)
sudo make install
```

## File Locations
---

- **Project files**: `~/devbox/<project>/` (on host)
- **Box workspace**: `/workspace/` (inside box)
- **Configuration**: `~/.devbox/config.json`

## Next Steps
---

Now that you have devbox installed, quickly get started by following the [Quick Start Guide](/docs/start/).
