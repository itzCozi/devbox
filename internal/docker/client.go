package docker

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Client wraps Docker CLI functionality
type Client struct{}

// NewClient creates a new Docker client
func NewClient() (*Client, error) {
	return &Client{}, nil
}

// Close closes the Docker client (no-op for CLI client)
func (c *Client) Close() error {
	return nil
}

// IsDockerAvailable checks if Docker is installed and running
func IsDockerAvailable() error {
	cmd := exec.Command("docker", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker is not installed or not running. Please ensure Docker is installed and the Docker daemon is running")
	}
	return nil
}

// PullImage pulls a Docker image if not already present
func (c *Client) PullImage(image string) error {
	// Check if image exists locally
	cmd := exec.Command("docker", "images", "-q", image)
	output, err := cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		// Image already exists
		return nil
	}

	fmt.Printf("Pulling image %s...\n", image)
	cmd = exec.Command("docker", "pull", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull image %s: %w", image, err)
	}

	return nil
}

// CreateBox creates a new Docker box
func (c *Client) CreateBox(name, image, workspaceHost, workspaceBox string) (string, error) {
	args := []string{
		"create",
		"--name", name,
		"--mount", fmt.Sprintf("type=bind,source=%s,target=%s", workspaceHost, workspaceBox),
		"--workdir", workspaceBox,
		"--restart", "unless-stopped",
		"-it",
		image,
		"sleep", "infinity",
	}

	cmd := exec.Command("docker", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to create box: %w", err)
	}

	boxID := strings.TrimSpace(string(output))
	return boxID, nil
}

// StartBox starts a Docker box
func (c *Client) StartBox(boxID string) error {
	cmd := exec.Command("docker", "start", boxID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start box: %w", err)
	}
	return nil
}

// SetupDevboxInBox installs the devbox wrapper script inside the box
func (c *Client) SetupDevboxInBox(boxName, projectName string) error {
	// First, run system updates to ensure packages are up to date
	fmt.Printf("Updating system packages in box...\n")
	updateCmd := exec.Command("docker", "exec", boxName, "bash", "-c", "apt update -y && apt full-upgrade -y")
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr
	if err := updateCmd.Run(); err != nil {
		fmt.Printf("Warning: failed to update packages in box: %v\n", err)
		// Don't fail the whole setup if package update fails
	}

	// Create the wrapper script content with proper escaping
	wrapperScript := `#!/bin/bash

# devbox-wrapper.sh
# This script provides devbox commands inside the box

BOX_NAME="` + boxName + `"
PROJECT_NAME="` + projectName + `"

case "$1" in
    "status"|"info")
        echo "📊 Devbox Box Status"
        echo "Project: $PROJECT_NAME"
        echo "Box: $BOX_NAME"
        echo "Workspace: /workspace"
        echo "Host: $(cat /etc/hostname)"
        echo "User: $(whoami)"
        echo "Working Directory: $(pwd)"
        echo ""
        echo "💡 Available devbox commands inside box:"
        echo "  devbox exit     - Exit the shell"
        echo "  devbox status   - Show box information"
        echo "  devbox help     - Show this help"
        echo "  devbox host     - Run command on host (experimental)"
        ;;
    "help"|"--help"|"-h")
        echo "🚀 Devbox Box Commands"
        echo ""
        echo "Available commands inside the box:"
        echo "  devbox exit         - Exit the devbox shell"
        echo "  devbox status       - Show box and project information"
        echo "  devbox help         - Show this help message"
        echo "  devbox host <cmd>   - Execute command on host (experimental)"
        echo ""
        echo "📁 Your project files are in: /workspace"
        echo "🐧 You are in an Ubuntu box with full package management"
        echo ""
        echo "Examples:"
        echo "  devbox exit                    # Exit to host"
        echo "  devbox status                  # Check box info"
        echo "  devbox host \"devbox list\"     # Run host command"
        echo ""
        echo "💡 Tip: Files in /workspace are shared with your host system"
        ;;
    "host")
        if [ -z "$2" ]; then
            echo "❌ Usage: devbox host <command>"
            echo "Example: devbox host \"devbox list\""
            exit 1
        fi
        echo "🔄 Executing on host: $2"
        echo "⚠️  Note: This is experimental and may not work in all environments"
        # This is a placeholder - we cannot easily execute on host from box
        # without additional setup like Docker socket mounting
        echo "❌ Host command execution not yet implemented"
        echo "💡 Exit the box and run commands on the host instead"
        ;;
    "version")
        echo "devbox box wrapper v1.0"
        echo "Box: $BOX_NAME"
        echo "Project: $PROJECT_NAME"
        ;;
    "")
        echo "❌ Missing command. Use \"devbox help\" for available commands."
        exit 1
        ;;
    *)
        echo "❌ Unknown devbox command: $1"
        echo "💡 Use \"devbox help\" to see available commands inside the box"
        echo ""
        echo "Available commands:"
        echo "  exit, status, help, host, version"
        echo ""
        echo "Note: 'devbox exit' is handled by the shell function for proper exit behavior"
        exit 1
        ;;
esac`

	// Install the wrapper script in the box using a here-document to avoid quoting issues
	// First remove any existing wrapper to ensure we get the new version
	installCmd := `rm -f /usr/local/bin/devbox && cat > /usr/local/bin/devbox << 'DEVBOX_WRAPPER_EOF'
` + wrapperScript + `
DEVBOX_WRAPPER_EOF
chmod +x /usr/local/bin/devbox`

	cmd := exec.Command("docker", "exec", boxName, "bash", "-c", installCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install devbox wrapper in box: %w", err)
	}

	// Clean up any existing devbox configurations in .bashrc and add new ones
	welcomeCmd := `# Remove any existing devbox configurations
sed -i '/# Devbox welcome message/,/^$/d' /root/.bashrc 2>/dev/null || true
sed -i '/devbox_exit()/,/^}$/d' /root/.bashrc 2>/dev/null || true
sed -i '/devbox() {/,/^}$/d' /root/.bashrc 2>/dev/null || true

# Add new devbox configuration
cat >> /root/.bashrc << 'BASHRC_EOF'

# Devbox welcome message  
if [ -t 1 ]; then
    echo "🚀 Welcome to devbox project: ` + projectName + `"
    echo "📁 Your files are in: /workspace"
    echo "💡 Type 'devbox help' for available commands"
    echo "🚪 Type 'devbox exit' to leave the box"
    echo ""
fi

# Define exit function for devbox
devbox_exit() {
    echo "👋 Exiting devbox shell for project \"` + projectName + `\""
    exit 0
}

# Override devbox command when it's called with exit
devbox() {
    if [[ "$1" == "exit" || "$1" == "quit" ]]; then
        devbox_exit
        return
    fi
    # Call the actual devbox script for other commands
    /usr/local/bin/devbox "$@"
}
BASHRC_EOF`

	cmd = exec.Command("docker", "exec", boxName, "bash", "-c", welcomeCmd)
	if err := cmd.Run(); err != nil {
		// Don't fail the whole setup if welcome message fails
		fmt.Printf("Warning: failed to add welcome message: %v\n", err)
	}

	return nil
}

// StopBox stops a Docker box
func (c *Client) StopBox(boxName string) error {
	cmd := exec.Command("docker", "stop", boxName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop box: %w", err)
	}
	return nil
}

// RemoveBox removes a Docker box
func (c *Client) RemoveBox(boxName string) error {
	cmd := exec.Command("docker", "rm", "-f", boxName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove box: %w", err)
	}
	return nil
}

// BoxExists checks if a box exists
func (c *Client) BoxExists(boxName string) (bool, error) {
	cmd := exec.Command("docker", "inspect", boxName)
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("failed to inspect box: %w", err)
	}
	return true, nil
}

// GetBoxStatus returns the status of a box
func (c *Client) GetBoxStatus(boxName string) (string, error) {
	cmd := exec.Command("docker", "inspect", "--format", "{{.State.Status}}", boxName)
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return "not found", nil
		}
		return "", fmt.Errorf("failed to inspect box: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// AttachShell attaches an interactive shell to a box using docker exec
func AttachShell(boxName string) error {
	// Set environment variables for the box session
	cmd := exec.Command("docker", "exec", "-it",
		"-e", fmt.Sprintf("DEVBOX_BOX_NAME=%s", boxName),
		boxName, "/bin/bash", "-c",
		"export PS1='devbox(\\$PROJECT_NAME):\\w\\$ '; exec /bin/bash")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to attach shell: %w", err)
	}
	return nil
}

// RunCommand runs a command in a box using docker exec
func RunCommand(boxName string, command []string) error {
	args := append([]string{"exec", "-it", boxName}, command...)
	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}
	return nil
}

// WaitForBox waits for a box to be running
func (c *Client) WaitForBox(boxName string, timeout time.Duration) error {
	start := time.Now()
	for {
		if time.Since(start) > timeout {
			return fmt.Errorf("timeout waiting for box to be ready")
		}

		status, err := c.GetBoxStatus(boxName)
		if err != nil {
			return err
		}

		if status == "running" {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}
}

// BoxInfo represents box information
type BoxInfo struct {
	Names  []string
	Status string
	Image  string
}

// ListBoxs lists all boxs with the devbox prefix
func (c *Client) ListBoxs() ([]BoxInfo, error) {
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}\t{{.Status}}\t{{.Image}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list boxs: %w", err)
	}

	var boxs []BoxInfo
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}

		name := parts[0]
		if strings.HasPrefix(name, "devbox_") {
			boxs = append(boxs, BoxInfo{
				Names:  []string{name},
				Status: parts[1],
				Image:  parts[2],
			})
		}
	}

	return boxs, scanner.Err()
}
