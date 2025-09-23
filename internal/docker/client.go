package docker

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"devbox/internal/parallel"
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
		"-it",
	}

	if projectConfig != nil {
		if config, ok := projectConfig.(map[string]interface{}); ok {
			args = c.applyProjectConfigToArgs(args, config)
		}
	}

	hasRestart := false
	for i := 0; i < len(args); i++ {
		if args[i] == "--restart" {
			hasRestart = true
			break
		}
	}
	if !hasRestart {
		args = append(args, "--restart", "unless-stopped")
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

	if restart, ok := config["restart"].(string); ok && restart != "" {
		args = append(args, "--restart", restart)
	}

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
				if strings.HasPrefix(volumeStr, "~") {
					if home, err := os.UserHomeDir(); err == nil {
						volumeStr = filepath.Join(home, strings.TrimPrefix(volumeStr, "~"))
					}
				}
				args = append(args, "-v", volumeStr)
			}
		}
	}

	if dotfiles, ok := config["dotfiles"].([]interface{}); ok {
		for _, item := range dotfiles {
			pathStr, ok := item.(string)
			if !ok || pathStr == "" {
				continue
			}
			host := pathStr
			if strings.HasPrefix(host, "~") {
				if home, err := os.UserHomeDir(); err == nil {
					host = filepath.Join(home, strings.TrimPrefix(host, "~"))
				}
			}
			args = append(args, "-v", fmt.Sprintf("%s:%s", host, "/dotfiles"))
			break
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

	config := parallel.LoadConfig()
	if config.EnableParallel {

		executor := parallel.NewSetupCommandExecutor(boxName, showOutput, config.SetupCommandWorkers)
		if err := executor.ExecuteParallel(commands); err != nil {

			fmt.Printf("Parallel execution failed, falling back to sequential: %v\n", err)
			return c.ExecuteSetupCommandsSequential(boxName, commands, showOutput)
		}
	} else {

		return c.ExecuteSetupCommandsSequential(boxName, commands, showOutput)
	}

	if showOutput {
		fmt.Printf("Setup commands completed successfully!\n")
	}
	return nil
}

func (c *Client) ExecuteSetupCommandsSequential(boxName string, commands []string, showOutput bool) error {
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

		wrapped := ". /root/.bashrc >/dev/null 2>&1 || true; " + command
		cmd := exec.Command("docker", "exec", boxName, "bash", "-c", wrapped)

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

func (c *Client) QueryPackagesParallel(boxName string) (aptList, pipList, npmList, yarnList, pnpmList []string) {
	config := parallel.LoadConfig()
	if !config.EnableParallel {

		return c.queryPackagesSequential(boxName)
	}

	executor := parallel.NewPackageQueryExecutor(boxName)

	packageLists, err := executor.QueryAllPackages()
	if err != nil {
		fmt.Printf("Warning: parallel package query failed, falling back to sequential: %v\n", err)

		return c.queryPackagesSequential(boxName)
	}

	return packageLists["apt"], packageLists["pip"], packageLists["npm"], packageLists["yarn"], packageLists["pnpm"]
}

func (c *Client) queryPackagesSequential(boxName string) (aptList, pipList, npmList, yarnList, pnpmList []string) {

	return nil, nil, nil, nil, nil
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
	sed -i '/# Devbox package tracking start/,/# Devbox package tracking end/d' /root/.bashrc 2>/dev/null || true

cat >> /root/.bashrc << 'BASHRC_EOF'

if [ -t 1 ]; then
    echo "🚀 Welcome to devbox project: ` + projectName + `"
    echo "📁 Your files are in: /workspace"
    echo "💡 Type 'devbox help' for available commands"
    echo "🚪 Type 'devbox exit' to leave the box"
    echo ""
fi

if [ -d "/dotfiles" ]; then
	if [ -f "/dotfiles/.bashrc" ]; then
		. /dotfiles/.bashrc
	fi
	for f in .gitconfig .vimrc .zshrc .bash_profile; do
		if [ -f "/dotfiles/$f" ]; then
			ln -sf "/dotfiles/$f" "/root/$f"
		fi
	done
	if [ -d "/dotfiles/.config" ]; then
		mkdir -p /root/.config
		for item in /dotfiles/.config/*; do
			base=$(basename "$item")
			if [ ! -e "/root/.config/$base" ]; then
				ln -s "$item" "/root/.config/$base"
			fi
		done
	fi
fi

devbox_exit() {
    echo "👋 Exiting devbox shell for project \"` + projectName + `\""
    exit 0
}

devbox() {
    if [[ "$1" == "exit" || "$1" == "quit" ]]; then
        devbox_exit
        return
    fi
    /usr/local/bin/devbox "$@"
}

export DEVBOX_LOCKFILE="${DEVBOX_LOCKFILE:-/workspace/devbox.lock}"

devbox_record_cmd() {
	local cmd="$1"
	if [ -n "$DEVBOX_LOCKFILE" ] && [ -w "$(dirname "$DEVBOX_LOCKFILE")" ]; then
		if [ ! -f "$DEVBOX_LOCKFILE" ] || ! grep -Fxq "$cmd" "$DEVBOX_LOCKFILE" 2>/dev/null; then
			echo "$cmd" >> "$DEVBOX_LOCKFILE"
		fi
	fi
}

_devbox_wrap_and_record() {
	local bin="$1"; shift
	local name="$1"; shift
	"$bin" "$@"
	local status=$?
	if [ $status -eq 0 ]; then
		case "$name" in
			apt|apt-get)
				# Track install/remove/purge/autoremove
				if printf ' %s ' "$*" | grep -qE '(^| )(install|remove|purge|autoremove)( |$)'; then
					devbox_record_cmd "$name $*"
				fi
				;;
			pip|pip3)
				if [ "$1" = install ] || [ "$1" = uninstall ]; then
					devbox_record_cmd "$name $*"
				fi
				;;
			npm)
				# Track install and uninstall variants
				if [ "$1" = install ] || [ "$1" = i ] || [ "$1" = add ] \
				   || [ "$1" = uninstall ] || [ "$1" = remove ] || [ "$1" = rm ] || [ "$1" = r ] || [ "$1" = un ]; then
					devbox_record_cmd "$name $*"
				fi
				;;
			yarn)
				# Track add/remove and global add/remove
				if [ "$1" = add ] || [ "$1" = remove ] || { [ "$1" = global ] && { [ "$2" = add ] || [ "$2" = remove ]; }; }; then
					devbox_record_cmd "$name $*"
				fi
				;;
			pnpm)
				# Track add/install and remove/uninstall variants
				if [ "$1" = add ] || [ "$1" = install ] || [ "$1" = i ] \
				   || [ "$1" = remove ] || [ "$1" = rm ] || [ "$1" = uninstall ] || [ "$1" = un ]; then
					devbox_record_cmd "$name $*"
				fi
				;;
			corepack)
				# Handle: corepack yarn add ..., corepack yarn global add ...
				#         corepack yarn remove ..., corepack yarn global remove ...
				#         corepack pnpm add/install/i/remove/rm/uninstall/un ...
				subcmd="$1"; shift || true
				if [ "$subcmd" = yarn ]; then
					if [ "$1" = add ] || [ "$1" = remove ] || { [ "$1" = global ] && { [ "$2" = add ] || [ "$2" = remove ]; }; }; then
						devbox_record_cmd "corepack yarn $*"
					fi
				elif [ "$subcmd" = pnpm ]; then
					if [ "$1" = add ] || [ "$1" = install ] || [ "$1" = i ] \
					   || [ "$1" = remove ] || [ "$1" = rm ] || [ "$1" = uninstall ] || [ "$1" = un ]; then
						devbox_record_cmd "corepack pnpm $*"
					fi
				fi
				;;
		esac
	fi
	return $status
}

APT_BIN="$(command -v apt 2>/dev/null || echo /usr/bin/apt)"
APTGET_BIN="$(command -v apt-get 2>/dev/null || echo /usr/bin/apt-get)"
PIP_BIN="$(command -v pip 2>/dev/null || echo /usr/bin/pip)"
PIP3_BIN="$(command -v pip3 2>/dev/null || echo /usr/bin/pip3)"
NPM_BIN="$(command -v npm 2>/dev/null || echo /usr/bin/npm)"
YARN_BIN="$(command -v yarn 2>/dev/null || echo /usr/bin/yarn)"
PNPM_BIN="$(command -v pnpm 2>/dev/null || echo /usr/bin/pnpm)"
COREPACK_BIN="$(command -v corepack 2>/dev/null || echo /usr/bin/corepack)"

apt()      { _devbox_wrap_and_record "$APT_BIN" apt "$@"; }
apt-get()  { _devbox_wrap_and_record "$APTGET_BIN" apt-get "$@"; }
pip()      { _devbox_wrap_and_record "$PIP_BIN" pip "$@"; }
pip3()     { _devbox_wrap_and_record "$PIP3_BIN" pip3 "$@"; }
npm()      { _devbox_wrap_and_record "$NPM_BIN" npm "$@"; }
yarn()     { _devbox_wrap_and_record "$YARN_BIN" yarn "$@"; }
pnpm()     { _devbox_wrap_and_record "$PNPM_BIN" pnpm "$@"; }
corepack(){ _devbox_wrap_and_record "$COREPACK_BIN" corepack "$@"; }
BASHRC_EOF`

	cmd = exec.Command("docker", "exec", boxName, "bash", "-c", welcomeCmd)
	if err := cmd.Run(); err != nil {

		fmt.Printf("Warning: failed to add welcome message: %v\n", err)
	}

	return nil
}

func (c *Client) StopBox(boxName string) error {

	timeoutSec := 2
	if v := strings.TrimSpace(os.Getenv("DEVBOX_STOP_TIMEOUT")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			timeoutSec = n
		}
	}
	cmd := exec.Command("docker", "stop", "--time", fmt.Sprintf("%d", timeoutSec), boxName)
	if err := cmd.Run(); err != nil {

		if killErr := exec.Command("docker", "kill", boxName).Run(); killErr != nil {
			return fmt.Errorf("failed to stop box: %w", err)
		}
		return nil
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
	cmdStr := strings.Join(command, " ")
	wrapped := ". /root/.bashrc >/dev/null 2>&1 || true; " + cmdStr
	args := []string{"exec", "-it", boxName, "bash", "-lc", wrapped}
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
			return fmt.Errorf("failed to get box status: %w", err)
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

func (c *Client) ListBoxes() ([]BoxInfo, error) {
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}\t{{.Status}}\t{{.Image}}")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return nil, fmt.Errorf("failed to list boxes: %s", stderrStr)
		}
		return nil, fmt.Errorf("failed to list boxes: %w", err)
	}

	var boxes []BoxInfo
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
			boxes = append(boxes, BoxInfo{
				Names:  []string{name},
				Status: parts[1],
				Image:  parts[2],
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan containers: %w", err)
	}

	return boxes, nil
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

type ContainerStats struct {
	CPUPercent string
	MemUsage   string
	MemPercent string
	NetIO      string
	BlockIO    string
	PIDs       string
}

func (c *Client) CommitContainer(containerName, imageTag string) (string, error) {
	args := []string{"commit", containerName, imageTag}
	cmd := exec.Command("docker", args...)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker commit failed: %s", strings.TrimSpace(errb.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

func (c *Client) SaveImage(imageRef, tarPath string) error {
	f, err := os.Create(tarPath)
	if err != nil {
		return fmt.Errorf("failed to create tar file: %w", err)
	}
	defer f.Close()
	cmd := exec.Command("docker", "save", imageRef)
	cmd.Stdout = f
	var errb bytes.Buffer
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker save failed: %s", strings.TrimSpace(errb.String()))
	}
	return nil
}

func (c *Client) LoadImage(tarPath string) (string, error) {
	cmd := exec.Command("docker", "load", "-i", tarPath)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker load failed: %s", strings.TrimSpace(errb.String()))
	}

	s := strings.TrimSpace(out.String())
	lines := strings.Split(s, "\n")
	if len(lines) > 0 {
		last := lines[len(lines)-1]
		if i := strings.LastIndex(last, ": "); i != -1 {
			return strings.TrimSpace(last[i+2:]), nil
		}
	}
	return s, nil
}

func (c *Client) GetContainerStats(boxName string) (*ContainerStats, error) {

	format := "{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}\t{{.BlockIO}}\t{{.PIDs}}"
	cmd := exec.Command("docker", "stats", "--no-stream", "--format", format, boxName)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if s := strings.TrimSpace(stderr.String()); s != "" {
			return nil, fmt.Errorf("failed to get stats: %s", s)
		}
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	line := strings.TrimSpace(stdout.String())
	if line == "" {

		return &ContainerStats{}, nil
	}
	parts := strings.Split(line, "\t")

	for len(parts) < 6 {
		parts = append(parts, "")
	}
	return &ContainerStats{
		CPUPercent: strings.TrimSpace(parts[0]),
		MemUsage:   strings.TrimSpace(parts[1]),
		MemPercent: strings.TrimSpace(parts[2]),
		NetIO:      strings.TrimSpace(parts[3]),
		BlockIO:    strings.TrimSpace(parts[4]),
		PIDs:       strings.TrimSpace(parts[5]),
	}, nil
}

func (c *Client) GetContainerID(boxName string) (string, error) {
	cmd := exec.Command("docker", "inspect", "--format", "{{.Id}}", boxName)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get container ID: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *Client) GetUptime(boxName string) (time.Duration, error) {
	cmd := exec.Command("docker", "inspect", "--format", "{{.State.StartedAt}}\t{{.State.Running}}", boxName)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to inspect container: %w", err)
	}
	s := strings.TrimSpace(string(out))
	parts := strings.Split(s, "\t")
	if len(parts) < 2 {
		return 0, nil
	}
	startedAt := strings.TrimSpace(parts[0])
	running := strings.TrimSpace(parts[1])
	if running != "true" {
		return 0, nil
	}

	t, parseErr := time.Parse(time.RFC3339Nano, startedAt)
	if parseErr != nil {

		if t2, err2 := time.Parse(time.RFC3339, startedAt); err2 == nil {
			return time.Since(t2), nil
		}
		return 0, nil
	}
	return time.Since(t), nil
}

func (c *Client) GetPortMappings(boxName string) ([]string, error) {
	cmd := exec.Command("docker", "port", boxName)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {

		return []string{}, nil
	}
	var ports []string
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			ports = append(ports, line)
		}
	}
	return ports, nil
}

func (c *Client) GetMounts(boxName string) ([]string, error) {
	template := `{{range .Mounts}}{{.Type}} {{.Source}} -> {{.Destination}} (rw={{.RW}})
{{end}}`
	cmd := exec.Command("docker", "inspect", "--format", template, boxName)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if s := strings.TrimSpace(stderr.String()); s != "" {
			return nil, fmt.Errorf("failed to get mounts: %s", s)
		}
		return nil, fmt.Errorf("failed to get mounts: %w", err)
	}
	var mounts []string
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			mounts = append(mounts, line)
		}
	}
	return mounts, nil
}

func (c *Client) IsContainerIdle(boxName string) (bool, error) {
	stats, err := c.GetContainerStats(boxName)
	if err != nil {
		return false, err
	}
	ports, err := c.GetPortMappings(boxName)
	if err != nil {
		return false, err
	}
	pids := 0
	if stats != nil && strings.TrimSpace(stats.PIDs) != "" {
		fmt.Sscanf(stats.PIDs, "%d", &pids)
	}
	return len(ports) == 0 && pids <= 1, nil
}

func (c *Client) ExecCapture(boxName, command string) (string, string, error) {
	wrapped := ". /root/.bashrc >/dev/null 2>&1 || true; set -o pipefail; " + command
	cmd := exec.Command("docker", "exec", boxName, "bash", "-lc", wrapped)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stdout.String(), stderr.String(), fmt.Errorf("exec failed: %w", err)
	}
	return stdout.String(), stderr.String(), nil
}

func (c *Client) GetAptSources(boxName string) (snapshotURL string, sources []string, release string) {

	out, _, err := c.ExecCapture(boxName, "cat /etc/apt/sources.list 2>/dev/null; echo; cat /etc/apt/sources.list.d/*.list 2>/dev/null || true")
	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(out))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			sources = append(sources, line)
			if strings.Contains(line, "snapshot.debian.org") || strings.Contains(line, "snapshot.ubuntu.com") {

				parts := strings.Fields(line)
				for _, p := range parts {
					if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
						snapshotURL = p
						break
					}
				}
			}
		}
	}

	if relOut, _, err2 := c.ExecCapture(boxName, ". /etc/os-release 2>/dev/null; echo $VERSION_CODENAME"); err2 == nil {
		release = strings.TrimSpace(relOut)
	}
	return
}

func (c *Client) GetPipRegistries(boxName string) (indexURL string, extra []string) {

	out, _, err := c.ExecCapture(boxName, "(pip3 config debug || pip config debug) 2>/dev/null | sed -n 's/^ *index-url *= *//p; s/^ *extra-index-url *= *//p')")
	if err == nil && strings.TrimSpace(out) != "" {

		lines := strings.Split(strings.TrimSpace(out), "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" {
				continue
			}
			if indexURL == "" && (strings.Contains(l, "://") || strings.HasPrefix(l, "file:")) {
				indexURL = l
			} else {
				extra = append(extra, l)
			}
		}
	}
	if indexURL == "" {

		if conf, _, err2 := c.ExecCapture(boxName, "grep -hE '^(index-url|extra-index-url)' /etc/pip.conf ~/.pip/pip.conf 2>/dev/null || true"); err2 == nil {
			for _, line := range strings.Split(conf, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "index-url") && indexURL == "" {
					if i := strings.Index(line, "="); i != -1 {
						indexURL = strings.TrimSpace(line[i+1:])
					}
				}
				if strings.HasPrefix(line, "extra-index-url") {
					if i := strings.Index(line, "="); i != -1 {
						extra = append(extra, strings.TrimSpace(line[i+1:]))
					}
				}
			}
		}
	}
	return
}

func (c *Client) GetNodeRegistries(boxName string) (npmReg, yarnReg, pnpmReg string) {
	if out, _, err := c.ExecCapture(boxName, "npm config get registry 2>/dev/null || true"); err == nil {
		npmReg = strings.TrimSpace(out)
	}
	if out, _, err := c.ExecCapture(boxName, "yarn config get npmRegistryServer 2>/dev/null || true"); err == nil {
		yarnReg = strings.TrimSpace(out)
	}
	if out, _, err := c.ExecCapture(boxName, "pnpm config get registry 2>/dev/null || true"); err == nil {
		pnpmReg = strings.TrimSpace(out)
	}
	return
}

func (c *Client) GetImageDigestInfo(ref string) (string, string, error) {
	cmd := exec.Command("docker", "inspect", "--type=image", "--format", "{{join .RepoDigests \",\"}}|{{.Id}}", ref)
	var out bytes.Buffer
	var errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err == nil {
		parts := strings.Split(strings.TrimSpace(out.String()), "|")
		digest := ""
		id := ""
		if len(parts) > 0 {
			ds := strings.Split(parts[0], ",")
			if len(ds) > 0 {
				digest = strings.TrimSpace(ds[0])
			}
		}
		if len(parts) > 1 {
			id = strings.TrimSpace(parts[1])
		}
		return digest, id, nil
	}

	cmd = exec.Command("docker", "inspect", "--type=container", "--format", "{{.Image}}", ref)
	out.Reset()
	errb.Reset()
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("inspect failed: %s", strings.TrimSpace(errb.String()))
	}
	imageID := strings.TrimSpace(out.String())
	if imageID == "" {
		return "", "", nil
	}
	cmd = exec.Command("docker", "inspect", "--type=image", "--format", "{{join .RepoDigests \",\"}}|{{.Id}}", imageID)
	out.Reset()
	errb.Reset()
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return "", imageID, nil
	}
	parts := strings.Split(strings.TrimSpace(out.String()), "|")
	digest := ""
	id := imageID
	if len(parts) > 0 {
		ds := strings.Split(parts[0], ",")
		if len(ds) > 0 {
			digest = strings.TrimSpace(ds[0])
		}
	}
	if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
		id = strings.TrimSpace(parts[1])
	}
	return digest, id, nil
}

func (c *Client) GetContainerMeta(boxName string) (map[string]string, string, string, string, map[string]string, []string, map[string]string, string) {
	type inspectType struct {
		Config struct {
			Env        []string          `json:"Env"`
			WorkingDir string            `json:"WorkingDir"`
			User       string            `json:"User"`
			Labels     map[string]string `json:"Labels"`
		} `json:"Config"`
		HostConfig struct {
			RestartPolicy struct {
				Name string `json:"Name"`
			} `json:"RestartPolicy"`
			CapAdd      []string `json:"CapAdd"`
			NanoCpus    int64    `json:"NanoCpus"`
			Memory      int64    `json:"Memory"`
			NetworkMode string   `json:"NetworkMode"`
		} `json:"HostConfig"`
	}
	cmd := exec.Command("docker", "inspect", boxName)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return map[string]string{}, "", "", "", map[string]string{}, []string{}, map[string]string{}, ""
	}
	var arr []inspectType
	if err := json.Unmarshal(out.Bytes(), &arr); err != nil || len(arr) == 0 {
		return map[string]string{}, "", "", "", map[string]string{}, []string{}, map[string]string{}, ""
	}
	ins := arr[0]
	env := map[string]string{}
	for _, e := range ins.Config.Env {
		if kv := strings.SplitN(e, "=", 2); len(kv) == 2 {
			env[kv[0]] = kv[1]
		}
	}
	resources := map[string]string{}
	if ins.HostConfig.NanoCpus > 0 {

		cpu := float64(ins.HostConfig.NanoCpus) / 1e9
		resources["cpus"] = strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", cpu), "0"), ".")
	}
	if ins.HostConfig.Memory > 0 {

		mb := float64(ins.HostConfig.Memory) / (1024 * 1024)
		resources["memory"] = fmt.Sprintf("%.0fMB", mb)
	}
	return env, ins.Config.WorkingDir, ins.Config.User, ins.HostConfig.RestartPolicy.Name, ins.Config.Labels, ins.HostConfig.CapAdd, resources, ins.HostConfig.NetworkMode
}
