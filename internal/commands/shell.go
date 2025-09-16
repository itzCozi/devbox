package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"devbox/internal/docker"
)

var shellCmd = &cobra.Command{
	Use:   "shell <project>",
	Short: "Open an interactive shell in the project container",
	Long:  `Attach an interactive bash shell to the specified project's container.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

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

		// Always ensure devbox wrapper is up to date in the container
		fmt.Printf("Setting up devbox commands in container...\n")
		if err := dockerClient.SetupDevboxInContainer(project.ContainerName, projectName); err != nil {
			return fmt.Errorf("failed to setup devbox in container: %w", err)
		}

		// Attach shell
		fmt.Printf("Attaching to container '%s'...\n", project.ContainerName)
		if err := docker.AttachShell(project.ContainerName); err != nil {
			return fmt.Errorf("failed to attach shell: %w", err)
		}

		return nil
	},
}
