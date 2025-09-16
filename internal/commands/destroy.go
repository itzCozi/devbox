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
	Short: "Stop and remove a project box",
	Long: `Stop and remove the Docker box for the specified project.
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
			fmt.Printf("This will destroy the box '%s' for project '%s'.\n", project.BoxName, projectName)
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

		// Check if box exists
		exists, err = dockerClient.BoxExists(project.BoxName)
		if err != nil {
			return fmt.Errorf("failed to check box status: %w", err)
		}

		if exists {
			// Stop box
			fmt.Printf("Stopping box '%s'...\n", project.BoxName)
			if err := dockerClient.StopBox(project.BoxName); err != nil {
				fmt.Printf("Warning: failed to stop box: %v\n", err)
			}

			// Remove box
			fmt.Printf("Removing box '%s'...\n", project.BoxName)
			if err := dockerClient.RemoveBox(project.BoxName); err != nil {
				return fmt.Errorf("failed to remove box: %w", err)
			}
		} else {
			fmt.Printf("Box '%s' not found (already removed)\n", project.BoxName)
		}

		// Remove project from configuration
		cfg.RemoveProject(projectName)
		if err := configManager.Save(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("✅ Project '%s' destroyed successfully!\n", projectName)

		// Check if project directory exists before suggesting removal
		if _, err := os.Stat(project.WorkspacePath); err == nil {
			fmt.Printf("📁 Project files preserved in: %s\n", project.WorkspacePath)
			fmt.Printf("\nTo completely remove the project:\n")
			fmt.Printf("  rm -rf %s\n", project.WorkspacePath)
		}

		return nil
	},
}
