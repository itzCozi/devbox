#!/bin/bash




set -e


RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'


print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_header() {
    echo -e "${BLUE}"
    echo "=================================================="
    echo "           devbox Installation Script"
    echo "=================================================="
    echo -e "${NC}"
}


check_root() {
    if [[ $EUID -eq 0 ]]; then
        print_error "This script should not be run as root. Please run as a regular user."
        print_info "The script will prompt for sudo when needed."
        exit 1
    fi
}


check_os() {
    print_info "Checking operating system compatibility..."

    if [[ ! -f /etc/os-release ]]; then
        print_error "Cannot determine operating system"
        exit 1
    fi

    . /etc/os-release

    case $ID in
        ubuntu|debian)
            print_success "Compatible OS detected: $PRETTY_NAME"
            ;;
        *)
            print_error "Unsupported operating system: $PRETTY_NAME"
            print_error "devbox requires Debian or Ubuntu Linux"
            exit 1
            ;;
    esac
}


command_exists() {
    command -v "$1" >/dev/null 2>&1
}


install_go() {
    if command_exists go; then
        GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
        MAJOR=$(echo "$GO_VERSION" | cut -d. -f1)
        MINOR=$(echo "$GO_VERSION" | cut -d. -f2)

        if [[ $MAJOR -gt 1 ]] || [[ $MAJOR -eq 1 && $MINOR -ge 21 ]]; then
            print_success "Go $GO_VERSION is already installed and compatible"
            return 0
        else
            print_warning "Go $GO_VERSION is installed but version 1.21+ is required"
        fi
    fi

    print_info "Installing Go..."


    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64)
            ARCH="arm64"
            ;;
        armv6l)
            ARCH="armv6l"
            ;;
        armv7l)
            ARCH="armv6l"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    GO_VERSION="1.22.1"
    GO_TARBALL="go${GO_VERSION}.linux-${ARCH}.tar.gz"


    cd /tmp
    print_info "Downloading Go ${GO_VERSION} for ${ARCH}..."
    curl -fsSL "https://golang.org/dl/${GO_TARBALL}" -o "${GO_TARBALL}"

    print_info "Installing Go to /usr/local/go..."
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "${GO_TARBALL}"


    if ! echo "$PATH" | grep -q "/usr/local/go/bin"; then
        echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.bashrc
        echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.profile
        export PATH=$PATH:/usr/local/go/bin
    fi


    if [[ -z "$GOPATH" ]]; then
        echo "export GOPATH=\$HOME/go" >> ~/.bashrc
        echo "export GOPATH=\$HOME/go" >> ~/.profile
        export GOPATH=$HOME/go
    fi

    rm -f "${GO_TARBALL}"
    print_success "Go installed successfully"
}


install_docker() {
    if command_exists docker; then
        if docker version >/dev/null 2>&1; then
            print_success "Docker is already installed and running"
            return 0
        else
            print_warning "Docker is installed but may not be running"
        fi
    fi

    print_info "Installing Docker..."


    sudo apt update


    sudo apt install -y \
        apt-transport-https \
        ca-certificates \
        curl \
        gnupg \
        lsb-release


    sudo mkdir -p /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg


    echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
        $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null


    sudo apt update


    sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin


    sudo systemctl start docker
    sudo systemctl enable docker


    sudo usermod -aG docker "$USER"

    print_success "Docker installed successfully"
    print_warning "You may need to log out and log back in for Docker group permissions to take effect"
}


install_make() {
    if command_exists make; then
        print_success "make is already installed"
        return 0
    fi

    print_info "Installing make..."
    sudo apt update
    sudo apt install -y build-essential
    print_success "make installed successfully"
}


install_git() {
    if command_exists git; then
        print_success "git is already installed"
        return 0
    fi

    print_info "Installing git..."
    sudo apt update
    sudo apt install -y git
    print_success "git installed successfully"
}


install_devbox() {
    print_info "Cloning devbox repository..."


    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"


    git clone https://github.com/itzCozi/devbox.git
    cd devbox

    print_info "Building devbox..."


    export PATH=$PATH:/usr/local/go/bin


    make build

    print_info "Installing devbox to /usr/local/bin..."


    sudo make install


    cd /
    rm -rf "$TEMP_DIR"

    print_success "devbox installed successfully"
}


verify_installation() {
    print_info "Verifying installation..."

    if command_exists devbox; then
        DEVBOX_VERSION=$(devbox --version 2>/dev/null || echo "unknown")
        print_success "devbox is installed and accessible: $DEVBOX_VERSION"


        if docker ps >/dev/null 2>&1; then
            print_success "Docker is accessible"
        else
            print_warning "Docker may require logout/login for group permissions"
        fi

        return 0
    else
        print_error "devbox installation verification failed"
        return 1
    fi
}


print_next_steps() {
    echo
    print_success "🎉 devbox installation completed successfully!"
    echo
    print_info "Next steps:"
    echo "  1. If Docker group permissions are needed, log out and log back in"
    echo "  2. Create your first project:"
    echo "     devbox init myproject"
    echo "  3. Enter the development environment:"
    echo "     devbox shell myproject"
    echo "  4. Get help anytime:"
    echo "     devbox --help"
    echo
    print_info "For more information, visit: https://github.com/itzCozi/devbox"
    echo
}


main() {
    print_header

    check_root
    check_os

    print_info "Installing dependencies..."
    install_git
    install_make
    install_go
    install_docker

    print_info "Installing devbox..."
    install_devbox

    if verify_installation; then
        print_next_steps
    else
        print_error "Installation completed but verification failed"
        print_info "Try running 'devbox --help' manually to test"
        exit 1
    fi
}


main "$@"