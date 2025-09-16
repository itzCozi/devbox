package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy <project>",
	Short: "Stop and remove a project container",
	Long: `Stop and remove the Docker container for the specified project.
Optionally delete the project folder with --force.`,
	Args: cobra.ExactArgs(1),
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
			return fmt.Errorf("project '%s' not found", projectName)
		}

		// Confirm destruction unless force flag is set
		if !forceFlag {
			fmt.Printf("This will destroy the container '%s' for project '%s'.\n", project.ContainerName, projectName)
			fmt.Printf("The project files in '%s' will be preserved.\n", project.WorkspacePath)
			fmt.Print("Are you sure? (y/N): ")

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read confirmation: %w", err)
			}

			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Destruction cancelled.")
				return nil
			}
		}

		// Check if container exists
		exists, err = dockerClient.ContainerExists(project.ContainerName)
		if err != nil {
			return fmt.Errorf("failed to check container status: %w", err)
		}

		if exists {
			// Stop container
			fmt.Printf("Stopping container '%s'...\n", project.ContainerName)
			if err := dockerClient.StopContainer(project.ContainerName); err != nil {
				fmt.Printf("Warning: failed to stop container: %v\n", err)
			}

			// Remove container
			fmt.Printf("Removing container '%s'...\n", project.ContainerName)
			if err := dockerClient.RemoveContainer(project.ContainerName); err != nil {
				return fmt.Errorf("failed to remove container: %w", err)
			}
		} else {
			fmt.Printf("Container '%s' not found (already removed)\n", project.ContainerName)
		}

		// Remove project from configuration
		cfg.RemoveProject(projectName)
		if err := configManager.Save(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("✅ Project '%s' destroyed successfully!\n", projectName)
		fmt.Printf("📁 Project files preserved in: %s\n", project.WorkspacePath)
		fmt.Printf("\nTo completely remove the project:\n")
		fmt.Printf("  rm -rf %s\n", project.WorkspacePath)

		return nil
	},
}