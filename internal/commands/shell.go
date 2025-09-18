package commands

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"

	"devbox/internal/docker"
)

var keepRunningFlag bool

var shellCmd = &cobra.Command{
	Use:   "shell <project>",
	Short: "Open an interactive shell in the project box",
	Long:  `Attach an interactive bash shell to the specified project's box.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		if err := validateProjectName(projectName); err != nil {
			return err
		}

		cfg, err := configManager.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		project, exists := cfg.GetProject(projectName)
		if !exists {
			return fmt.Errorf("project '%s' not found. Run 'devbox init %s' first", projectName, projectName)
		}

		exists, err = dockerClient.BoxExists(project.BoxName)
		if err != nil {
			return fmt.Errorf("failed to check box status: %w", err)
		}

		if !exists {
			return fmt.Errorf("box '%s' not found. Run 'devbox init %s' to recreate", project.BoxName, projectName)
		}

		status, err := dockerClient.GetBoxStatus(project.BoxName)
		if err != nil {
			return fmt.Errorf("failed to get box status: %w", err)
		}

		if status != "running" {
			fmt.Printf("Starting box '%s'...\n", project.BoxName)
			if err := dockerClient.StartBox(project.BoxName); err != nil {
				return fmt.Errorf("failed to start box: %w", err)
			}
		}

		checkCmd := exec.Command("docker", "exec", project.BoxName, "test", "-f", "/etc/devbox-initialized")
		if checkCmd.Run() != nil {
			fmt.Printf("Setting up devbox commands in box...\n")
			if err := dockerClient.SetupDevboxInBox(project.BoxName, projectName); err != nil {
				return fmt.Errorf("failed to setup devbox in box: %w", err)
			}
		}

		fmt.Printf("Attaching to box '%s'...\n", project.BoxName)
		if err := docker.AttachShell(project.BoxName); err != nil {
			return fmt.Errorf("failed to attach shell: %w", err)
		}

		// After shell exits, optionally stop the container when not being used
		if !keepRunningFlag {
			cfg, err := configManager.Load()
			if err == nil && cfg.Settings != nil && cfg.Settings.AutoStopOnExit {
				fmt.Printf("Stopping box '%s' (auto-stop enabled)...\n", project.BoxName)
				if err := dockerClient.StopBox(project.BoxName); err != nil {
					fmt.Printf("Warning: failed to stop box: %v\n", err)
				}
			}
		}

		return nil
	},
}

func init() {
	shellCmd.Flags().BoolVar(&keepRunningFlag, "keep-running", false, "Keep the box running after exiting the shell")
}
