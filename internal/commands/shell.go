package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"devbox/internal/docker"
)

var shellCmd = &cobra.Command{
	Use:   "shell <project>",
	Short: "Open an interactive shell in the project box",
	Long:  `Attach an interactive bash shell to the specified project's box.`,
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

		// Check if box exists and is running
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

		// Always ensure devbox wrapper is up to date in the box
		fmt.Printf("Setting up devbox commands in box...\n")
		if err := dockerClient.SetupDevboxInBox(project.BoxName, projectName); err != nil {
			return fmt.Errorf("failed to setup devbox in box: %w", err)
		}

		// Attach shell
		fmt.Printf("Attaching to box '%s'...\n", project.BoxName)
		if err := docker.AttachShell(project.BoxName); err != nil {
			return fmt.Errorf("failed to attach shell: %w", err)
		}

		return nil
	},
}
