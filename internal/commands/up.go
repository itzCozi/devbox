package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"devbox/internal/config"
)

var (
	upDotfilesPath string
)

var keepRunningUpFlag bool

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start a devbox environment from the current folder's devbox.json",
	Long:  "Reads devbox.json in the current directory and boots the environment so new teammates can simply run 'devbox up'.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		projectConfig, err := configManager.LoadProjectConfig(cwd)
		if err != nil {
			return fmt.Errorf("failed to load devbox.json: %w", err)
		}
		if projectConfig == nil {
			return fmt.Errorf("no devbox.json found in %s", cwd)
		}

		if err := configManager.ValidateProjectConfig(projectConfig); err != nil {
			return fmt.Errorf("invalid devbox.json: %w", err)
		}

		projectName := projectConfig.Name
		if projectName == "" {

			projectName = filepath.Base(cwd)
		}

		cfg, err := configManager.Load()
		if err != nil {
			return fmt.Errorf("failed to load global config: %w", err)
		}

		boxName := fmt.Sprintf("devbox_%s", projectName)
		baseImage := cfg.GetEffectiveBaseImage(&config.Project{Name: projectName, BaseImage: projectConfig.BaseImage}, projectConfig)

		workspaceBox := "/workspace"
		if projectConfig.WorkingDir != "" {
			workspaceBox = projectConfig.WorkingDir
		}

		exists, err := dockerClient.BoxExists(boxName)
		if err != nil {
			return fmt.Errorf("failed to check box existence: %w", err)
		}

		if exists {
			status, err := dockerClient.GetBoxStatus(boxName)
			if err != nil {
				return fmt.Errorf("failed to get box status: %w", err)
			}
			if status != "running" {
				if err := dockerClient.StartBox(boxName); err != nil {
					return fmt.Errorf("failed to start existing box: %w", err)
				}
			}

			checkCmd := exec.Command("docker", "exec", boxName, "test", "-f", "/etc/devbox-initialized")
			if checkCmd.Run() != nil {
				if err := dockerClient.SetupDevboxInBox(boxName, projectName); err != nil {
					return fmt.Errorf("failed to setup devbox in existing box: %w", err)
				}
			}
			fmt.Printf("✅ Environment is up!\n")
			fmt.Printf("📁 Workspace: %s\n", cwd)
			fmt.Printf("🐳 Box: %s\n", boxName)
			fmt.Printf("🖼️  Image: %s\n", baseImage)
			fmt.Printf("Tip: run 'devbox shell %s' to enter the environment.\n", projectName)

			// Auto-stop if configured, not explicitly kept running, and idle
			if cfg.Settings != nil && cfg.Settings.AutoStopOnExit && !keepRunningUpFlag {
				stats, _ := dockerClient.GetContainerStats(boxName)
				ports, _ := dockerClient.GetPortMappings(boxName)
				pids := 0
				if stats != nil && stats.PIDs != "" {
					// best-effort parse
					fmt.Sscanf(stats.PIDs, "%d", &pids)
				}
				if len(ports) == 0 && pids <= 1 {
					fmt.Printf("Stopping box '%s' (auto-stop: idle)...\n", boxName)
					if err := dockerClient.StopBox(boxName); err != nil {
						fmt.Printf("Warning: failed to stop box: %v\n", err)
					}
				}
			}
			return nil
		}

		fmt.Printf("Setting up box '%s' with image '%s'...\n", boxName, baseImage)
		if err := dockerClient.PullImage(baseImage); err != nil {
			return fmt.Errorf("failed to pull base image: %w", err)
		}

		var configMap map[string]interface{}
		if projectConfig != nil {
			data, _ := json.Marshal(projectConfig)
			_ = json.Unmarshal(data, &configMap)
		}

		// If auto-stop is enabled globally and restart policy isn't explicitly set,
		// prefer a non-restarting policy so manual stops remain stopped.
		if cfg.Settings != nil && cfg.Settings.AutoStopOnExit {
			if configMap == nil {
				configMap = map[string]interface{}{}
			}
			if _, ok := configMap["restart"]; !ok {
				configMap["restart"] = "no"
			}
		}

		var dotfiles []string
		if len(projectConfig.Dotfiles) > 0 {
			dotfiles = append(dotfiles, projectConfig.Dotfiles...)
		}
		if upDotfilesPath != "" {
			dotfiles = append(dotfiles, upDotfilesPath)
		}
		if len(dotfiles) > 0 {
			arr := make([]interface{}, 0, len(dotfiles))
			for _, s := range dotfiles {
				arr = append(arr, s)
			}
			if configMap == nil {
				configMap = map[string]interface{}{}
			}
			configMap["dotfiles"] = arr
		}

		boxID, err := dockerClient.CreateBoxWithConfig(boxName, baseImage, cwd, workspaceBox, configMap)
		if err != nil {
			return fmt.Errorf("failed to create box: %w", err)
		}

		if err := dockerClient.StartBox(boxID); err != nil {
			return fmt.Errorf("failed to start box: %w", err)
		}

		fmt.Printf("Starting box...\n")
		if err := dockerClient.WaitForBox(boxName, 30*time.Second); err != nil {
			return fmt.Errorf("box failed to start: %w", err)
		}

		fmt.Printf("Setting up devbox commands in box...\n")
		if err := dockerClient.SetupDevboxInBoxWithUpdate(boxName, projectName); err != nil {
			return fmt.Errorf("failed to setup devbox in box: %w", err)
		}

		fmt.Printf("Updating system packages...\n")
		systemUpdateCommands := []string{"apt update -y", "apt full-upgrade -y"}
		if err := dockerClient.ExecuteSetupCommandsWithOutput(boxName, systemUpdateCommands, false); err != nil {
			return fmt.Errorf("failed to update system packages: %w", err)
		}

		lockfilePath := filepath.Join(cwd, "devbox.lock")
		if _, err := os.Stat(lockfilePath); err == nil {
			fmt.Printf("Replaying recorded package installs from devbox.lock...\n")

			data, readErr := os.ReadFile(lockfilePath)
			if readErr == nil {
				lines := strings.Split(string(data), "\n")
				var cmds []string
				for _, line := range lines {
					cmd := strings.TrimSpace(line)
					if cmd == "" || strings.HasPrefix(cmd, "#") {
						continue
					}
					cmds = append(cmds, cmd)
				}
				if len(cmds) > 0 {
					if err := dockerClient.ExecuteSetupCommandsWithOutput(boxName, cmds, false); err != nil {
						return fmt.Errorf("failed to replay devbox.lock commands: %w", err)
					}
				}
			}
		}

		if len(projectConfig.SetupCommands) > 0 {
			fmt.Printf("Installing setup packages (%d commands)...\n", len(projectConfig.SetupCommands))
			if err := dockerClient.ExecuteSetupCommandsWithOutput(boxName, projectConfig.SetupCommands, false); err != nil {
				return fmt.Errorf("failed to execute setup commands: %w", err)
			}
		}

		fmt.Printf("✅ Environment is up!\n")
		fmt.Printf("📁 Workspace: %s\n", cwd)
		fmt.Printf("🐳 Box: %s\n", boxName)
		fmt.Printf("🖼️  Image: %s\n", baseImage)
		fmt.Printf("Tip: run 'devbox shell %s' to enter the environment.\n", projectName)

		// Auto-stop if configured, not explicitly kept running, and idle
		if cfg.Settings != nil && cfg.Settings.AutoStopOnExit && !keepRunningUpFlag {
			stats, _ := dockerClient.GetContainerStats(boxName)
			ports, _ := dockerClient.GetPortMappings(boxName)
			pids := 0
			if stats != nil && stats.PIDs != "" {
				fmt.Sscanf(stats.PIDs, "%d", &pids)
			}
			if len(ports) == 0 && pids <= 1 {
				fmt.Printf("Stopping box '%s' (auto-stop: idle)...\n", boxName)
				if err := dockerClient.StopBox(boxName); err != nil {
					fmt.Printf("Warning: failed to stop box: %v\n", err)
				}
			}
		}
		return nil
	},
}

func init() {
	upCmd.Flags().StringVar(&upDotfilesPath, "dotfiles", "", "Path to local dotfiles directory to mount into the box")
	upCmd.Flags().BoolVar(&keepRunningUpFlag, "keep-running", false, "Keep the box running after 'up' finishes")
}
