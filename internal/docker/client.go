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
