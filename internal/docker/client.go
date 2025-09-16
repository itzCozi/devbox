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
		return fmt.Errorf("Docker is not installed or not running. Please ensure Docker is installed and the Docker daemon is running")
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

// CreateContainer creates a new Docker container
func (c *Client) CreateContainer(name, image, workspaceHost, workspaceContainer string) (string, error) {
	args := []string{
		"create",
		"--name", name,
		"--mount", fmt.Sprintf("type=bind,source=%s,target=%s", workspaceHost, workspaceContainer),
		"--workdir", workspaceContainer,
		"--restart", "unless-stopped",
		"-it",
		image,
		"sleep", "infinity",
	}

	cmd := exec.Command("docker", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	containerID := strings.TrimSpace(string(output))
	return containerID, nil
}

// StartContainer starts a Docker container
func (c *Client) StartContainer(containerID string) error {
	cmd := exec.Command("docker", "start", containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}
	return nil
}

// SetupDevboxInContainer installs the devbox wrapper script inside the container
func (c *Client) SetupDevboxInContainer(containerName, projectName string) error {
	// Create the wrapper script content
	wrapperScript := `#!/bin/bash

# devbox-wrapper.sh
# This script provides devbox commands inside the container

CONTAINER_NAME="` + containerName + `"
PROJECT_NAME="` + projectName + `"

case "$1" in
    "exit"|"quit")
        echo "👋 Exiting devbox shell for project '$PROJECT_NAME'"
        exit 0
        ;;
    "status"|"info")
        echo "📊 Devbox Container Status"
        echo "Project: $PROJECT_NAME"
        echo "Container: $CONTAINER_NAME"
        echo "Workspace: /workspace"
        echo "Host: $(cat /etc/hostname)"
        echo "User: $(whoami)"
        echo "Working Directory: $(pwd)"
        echo ""
        echo "💡 Available devbox commands inside container:"
        echo "  devbox exit     - Exit the shell"
        echo "  devbox status   - Show container information"
        echo "  devbox help     - Show this help"
        ;;
    "help"|"--help"|"-h")
        echo "🚀 Devbox Container Commands"
        echo ""
        echo "Available commands inside the container:"
        echo "  devbox exit         - Exit the devbox shell"
        echo "  devbox status       - Show container and project information"  
        echo "  devbox help         - Show this help message"
        echo ""
        echo "📁 Your project files are in: /workspace"
        echo "🐧 You're in an Ubuntu container with full package management"
        echo ""
        echo "Examples:"
        echo "  devbox exit                    # Exit to host"
        echo "  devbox status                  # Check container info"
        echo ""
        echo "💡 Tip: Files in /workspace are shared with your host system"
        ;;
    "version")
        echo "devbox container wrapper v1.0"
        echo "Container: $CONTAINER_NAME"
        echo "Project: $PROJECT_NAME"
        ;;
    "")
        echo "❌ Missing command. Use 'devbox help' for available commands."
        exit 1
        ;;
    *)
        echo "❌ Unknown devbox command: $1"
        echo "💡 Use 'devbox help' to see available commands inside the container"
        echo ""
        echo "Available commands:"
        echo "  exit, status, help, version"
        exit 1
        ;;
esac`

	// Install the wrapper script in the container
	installCmd := fmt.Sprintf(`echo '%s' > /usr/local/bin/devbox && chmod +x /usr/local/bin/devbox`, wrapperScript)

	cmd := exec.Command("docker", "exec", containerName, "bash", "-c", installCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install devbox wrapper in container: %w", err)
	}

	// Add a welcome message to .bashrc
	welcomeCmd := `echo '
# Devbox welcome message  
if [ -t 1 ]; then
    echo "🚀 Welcome to devbox project: ` + projectName + `"
    echo "📁 Your files are in: /workspace"
    echo "💡 Type 'devbox help' for available commands"
    echo "🚪 Type 'devbox exit' to leave the container"
    echo ""
fi' >> /root/.bashrc`

	cmd = exec.Command("docker", "exec", containerName, "bash", "-c", welcomeCmd)
	if err := cmd.Run(); err != nil {
		// Don't fail the whole setup if welcome message fails
		fmt.Printf("Warning: failed to add welcome message: %v\n", err)
	}

	return nil
}

// StopContainer stops a Docker container
func (c *Client) StopContainer(containerName string) error {
	cmd := exec.Command("docker", "stop", containerName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	return nil
}

// RemoveContainer removes a Docker container
func (c *Client) RemoveContainer(containerName string) error {
	cmd := exec.Command("docker", "rm", "-f", containerName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}
	return nil
}

// ContainerExists checks if a container exists
func (c *Client) ContainerExists(containerName string) (bool, error) {
	cmd := exec.Command("docker", "inspect", containerName)
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("failed to inspect container: %w", err)
	}
	return true, nil
}

// GetContainerStatus returns the status of a container
func (c *Client) GetContainerStatus(containerName string) (string, error) {
	cmd := exec.Command("docker", "inspect", "--format", "{{.State.Status}}", containerName)
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return "not found", nil
		}
		return "", fmt.Errorf("failed to inspect container: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// AttachShell attaches an interactive shell to a container using docker exec
func AttachShell(containerName string) error {
	cmd := exec.Command("docker", "exec", "-it", containerName, "/bin/bash")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to attach shell: %w", err)
	}
	return nil
}

// RunCommand runs a command in a container using docker exec
func RunCommand(containerName string, command []string) error {
	args := append([]string{"exec", "-it", containerName}, command...)
	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}
	return nil
}

// WaitForContainer waits for a container to be running
func (c *Client) WaitForContainer(containerName string, timeout time.Duration) error {
	start := time.Now()
	for {
		if time.Since(start) > timeout {
			return fmt.Errorf("timeout waiting for container to be ready")
		}

		status, err := c.GetContainerStatus(containerName)
		if err != nil {
			return err
		}

		if status == "running" {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}
}

// ContainerInfo represents container information
type ContainerInfo struct {
	Names  []string
	Status string
	Image  string
}

// ListContainers lists all containers with the devbox prefix
func (c *Client) ListContainers() ([]ContainerInfo, error) {
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}\t{{.Status}}\t{{.Image}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var containers []ContainerInfo
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
			containers = append(containers, ContainerInfo{
				Names:  []string{name},
				Status: parts[1],
				Image:  parts[2],
			})
		}
	}

	return containers, scanner.Err()
}
