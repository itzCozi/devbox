package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"devbox/internal/docker"
)

var keepRunningRunFlag bool

var runCmd = &cobra.Command{
	Use:   "run <project> <command> [args...]",
	Short: "Run a command in the project box",
	Long:  `Execute an arbitrary command inside the specified project's box.`,
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		command := args[1:]

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

		if err := docker.RunCommand(project.BoxName, command); err != nil {
			return fmt.Errorf("failed to run command: %w", err)
		}

		if !keepRunningRunFlag {
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
	runCmd.Flags().BoolVar(&keepRunningRunFlag, "keep-running", false, "Keep the box running after the command finishes")
}
