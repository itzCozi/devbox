package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"devbox/internal/docker"
)

var runCmd = &cobra.Command{
	Use:   "run <project> <command> [args...]",
	Short: "Run a command in the project container",
	Long:  `Execute an arbitrary command inside the specified project's container.`,
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		command := args[1:]

		// Validate project name
		if err := validateProjectName(projectName); err != nil {
			return err
		}

		// Load configuration
		cfg, err := configManager.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Check if project exists
		project, exists := cfg.GetProject(projectName)
		if !exists {
			return fmt.Errorf("project '%s' not found. Run 'devbox init %s' first", projectName, projectName)
		}

		// Check if container exists and is running
		exists, err = dockerClient.ContainerExists(project.ContainerName)
		if err != nil {
			return fmt.Errorf("failed to check container status: %w", err)
		}

		if !exists {
			return fmt.Errorf("container '%s' not found. Run 'devbox init %s' to recreate", project.ContainerName, projectName)
		}

		status, err := dockerClient.GetContainerStatus(project.ContainerName)
		if err != nil {
			return fmt.Errorf("failed to get container status: %w", err)
		}

		if status != "running" {
			fmt.Printf("Starting container '%s'...\n", project.ContainerName)
			if err := dockerClient.StartContainer(project.ContainerName); err != nil {
				return fmt.Errorf("failed to start container: %w", err)
			}
		}

		// Run command
		if err := docker.RunCommand(project.ContainerName, command); err != nil {
			return fmt.Errorf("failed to run command: %w", err)
		}

		return nil
	},
}