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

type Client struct{}

func NewClient() (*Client, error) {
	return &Client{}, nil
}

func (c *Client) Close() error {
	return nil
}

func IsDockerAvailable() error {
	cmd := exec.Command("docker", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker is not installed or not running. Please ensure Docker is installed and the Docker daemon is running")
	}
	return nil
}

func (c *Client) PullImage(image string) error {

	cmd := exec.Command("docker", "images", "-q", image)
	output, err := cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {

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

func (c *Client) CreateBox(name, image, workspaceHost, workspaceBox string) (string, error) {
	return c.CreateBoxWithConfig(name, image, workspaceHost, workspaceBox, nil)
}

func (c *Client) CreateBoxWithConfig(name, image, workspaceHost, workspaceBox string, projectConfig interface{}) (string, error) {
	args := []string{
		"create",
		"--name", name,
		"--mount", fmt.Sprintf("type=bind,source=%s,target=%s", workspaceHost, workspaceBox),
		"--workdir", workspaceBox,
		"--restart", "unless-stopped",
		"-it",
	}

	if projectConfig != nil {
		if config, ok := projectConfig.(map[string]interface{}); ok {
			args = c.applyProjectConfigToArgs(args, config)
		}
	}

	args = append(args, image, "sleep", "infinity")

	cmd := exec.Command("docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return "", fmt.Errorf("failed to create box: %s", stderrStr)
		}
		return "", fmt.Errorf("failed to create box: %w", err)
	}

	boxID := strings.TrimSpace(stdout.String())
	return boxID, nil
}

func (c *Client) applyProjectConfigToArgs(args []string, config map[string]interface{}) []string {

	if env, ok := config["environment"].(map[string]interface{}); ok {
		for key, value := range env {
			if valueStr, ok := value.(string); ok {
				args = append(args, "-e", fmt.Sprintf("%s=%s", key, valueStr))
			}
		}
	}

	if ports, ok := config["ports"].([]interface{}); ok {
		for _, port := range ports {
			if portStr, ok := port.(string); ok {
				args = append(args, "-p", portStr)
			}
		}
	}

	if volumes, ok := config["volumes"].([]interface{}); ok {
		for _, volume := range volumes {
			if volumeStr, ok := volume.(string); ok {
				args = append(args, "-v", volumeStr)
			}
		}
	}

	if workingDir, ok := config["working_dir"].(string); ok && workingDir != "" {
		args = append(args, "--workdir", workingDir)
	}

	if user, ok := config["user"].(string); ok && user != "" {
		args = append(args, "--user", user)
	}

	if capabilities, ok := config["capabilities"].([]interface{}); ok {
		for _, cap := range capabilities {
			if capStr, ok := cap.(string); ok {
				args = append(args, "--cap-add", capStr)
			}
		}
	}

	if labels, ok := config["labels"].(map[string]interface{}); ok {
		for key, value := range labels {
			if valueStr, ok := value.(string); ok {
				args = append(args, "--label", fmt.Sprintf("%s=%s", key, valueStr))
			}
		}
	}

	if network, ok := config["network"].(string); ok && network != "" {
		args = append(args, "--network", network)
	}

	if resources, ok := config["resources"].(map[string]interface{}); ok {
		if cpus, ok := resources["cpus"].(string); ok && cpus != "" {
			args = append(args, "--cpus", cpus)
		}
		if memory, ok := resources["memory"].(string); ok && memory != "" {
			args = append(args, "--memory", memory)
		}
	}

	if healthCheck, ok := config["health_check"].(map[string]interface{}); ok {
		if test, ok := healthCheck["test"].([]interface{}); ok && len(test) > 0 {
			var testArgs []string
			for _, t := range test {
				if testStr, ok := t.(string); ok {
					testArgs = append(testArgs, testStr)
				}
			}
			if len(testArgs) > 0 {
				args = append(args, "--health-cmd", strings.Join(testArgs, " "))
			}
		}
		if interval, ok := healthCheck["interval"].(string); ok && interval != "" {
			args = append(args, "--health-interval", interval)
		}
		if timeout, ok := healthCheck["timeout"].(string); ok && timeout != "" {
			args = append(args, "--health-timeout", timeout)
		}
		if retries, ok := healthCheck["retries"].(float64); ok && retries > 0 {
			args = append(args, "--health-retries", fmt.Sprintf("%.0f", retries))
		}
	}

	return args
}

func (c *Client) ExecuteSetupCommands(boxName string, commands []string) error {
	return c.ExecuteSetupCommandsWithOutput(boxName, commands, true)
}

func (c *Client) ExecuteSetupCommandsWithOutput(boxName string, commands []string, showOutput bool) error {
	if len(commands) == 0 {
		return nil
	}

	if showOutput {
		fmt.Printf("Executing setup commands in box '%s'...\n", boxName)
	}

	for i, command := range commands {
		if showOutput {
			fmt.Printf("Step %d/%d: %s\n", i+1, len(commands), command)
		}

		cmd := exec.Command("docker", "exec", boxName, "bash", "-c", command)

		if showOutput {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("setup command failed: %s: %w", command, err)
			}
		} else {

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {

				fmt.Printf("Command failed: %s\n", command)
				if stderr.Len() > 0 {
					fmt.Printf("Error output: %s\n", stderr.String())
				}
				if stdout.Len() > 0 {
					fmt.Printf("Standard output: %s\n", stdout.String())
				}
				return fmt.Errorf("setup command failed: %s: %w", command, err)
			}
		}
	}

	if showOutput {
		fmt.Printf("Setup commands completed successfully!\n")
	}
	return nil
}

func (c *Client) StartBox(boxID string) error {
	cmd := exec.Command("docker", "start", boxID)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return fmt.Errorf("failed to start box: %s", stderrStr)
		}
		return fmt.Errorf("failed to start box: %w", err)
	}
	return nil
}

func (c *Client) SetupDevboxInBox(boxName, projectName string) error {
	return c.setupDevboxInBoxWithOptions(boxName, projectName, false)
}

func (c *Client) SetupDevboxInBoxWithUpdate(boxName, projectName string) error {
	return c.setupDevboxInBoxWithOptions(boxName, projectName, true)
}

func (c *Client) setupDevboxInBoxWithOptions(boxName, projectName string, forceUpdate bool) error {

	checkCmd := exec.Command("docker", "exec", boxName, "test", "-f", "/etc/devbox-initialized")
	isFirstTime := checkCmd.Run() != nil

	if isFirstTime {
		markerCmd := exec.Command("docker", "exec", boxName, "touch", "/etc/devbox-initialized")
		if err := markerCmd.Run(); err != nil {
			fmt.Printf("Warning: failed to create initialization marker: %v\n", err)
		}
	}

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

	installCmd := `rm -f /usr/local/bin/devbox && cat > /usr/local/bin/devbox << 'DEVBOX_WRAPPER_EOF'
` + wrapperScript + `
DEVBOX_WRAPPER_EOF
chmod +x /usr/local/bin/devbox`

	cmd := exec.Command("docker", "exec", boxName, "bash", "-c", installCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install devbox wrapper in box: %w", err)
	}

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

		fmt.Printf("Warning: failed to add welcome message: %v\n", err)
	}

	return nil
}

func (c *Client) StopBox(boxName string) error {
	cmd := exec.Command("docker", "stop", boxName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop box: %w", err)
	}
	return nil
}

func (c *Client) RemoveBox(boxName string) error {

	cmd := exec.Command("docker", "rm", "-f", boxName)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return fmt.Errorf("failed to remove box: %s", stderrStr)
		}
		return fmt.Errorf("failed to remove box: %w", err)
	}
	return nil
}

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

func AttachShell(boxName string) error {

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

type BoxInfo struct {
	Names  []string
	Status string
	Image  string
}

func (c *Client) ListBoxs() ([]BoxInfo, error) {
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}\t{{.Status}}\t{{.Image}}")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return nil, fmt.Errorf("failed to list boxs: %s", stderrStr)
		}
		return nil, fmt.Errorf("failed to list boxs: %w", err)
	}

	var boxs []BoxInfo
	scanner := bufio.NewScanner(&stdout)
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

func (c *Client) RunDockerCommand(args []string) error {
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker command failed: %w", err)
	}
	return nil
}
