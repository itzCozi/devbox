package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy <project>",
	Short: "Stop and remove a project box",
	Long: `Stop and remove the Docker box for the specified project.
Removes empty project directories automatically.

Special usage:
  devbox destroy --cleanup-orphaned  Remove containers not tracked in config`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		// Handle special cleanup command
		if projectName == "--cleanup-orphaned" {
			return cleanupOrphanedContainers()
		}

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
			fmt.Printf("Empty project directories will be automatically removed.\n")
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
			// Stop and remove box (force remove will stop it automatically)
			fmt.Printf("Stopping and removing box '%s'...\n", project.BoxName)
			if err := dockerClient.RemoveBox(project.BoxName); err != nil {
				fmt.Printf("Warning: failed to remove box: %v\n", err)
				// Continue anyway - we still want to remove from config
			}
		} else {
			fmt.Printf("Box '%s' not found (already removed)\n", project.BoxName)
		}

		// Remove project from configuration
		cfg.RemoveProject(projectName)
		if err := configManager.Save(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("Project '%s' destroyed successfully!\n", projectName)

		// Check if project directory exists and handle removal
		if _, err := os.Stat(project.WorkspacePath); err == nil {
			// Check if directory is empty
			isEmpty, err := isDirEmpty(project.WorkspacePath)
			if err != nil {
				fmt.Printf("Warning: failed to check if directory is empty: %v\n", err)
				fmt.Printf("Project files preserved in: %s\n", project.WorkspacePath)
			} else if isEmpty {
				fmt.Printf("Removing empty project directory: %s\n", project.WorkspacePath)
				if err := os.RemoveAll(project.WorkspacePath); err != nil {
					fmt.Printf("Warning: failed to remove empty directory: %v\n", err)
				} else {
					fmt.Printf("Empty project directory removed!\n")
				}
			} else {
				fmt.Printf("Project files preserved in: %s\n", project.WorkspacePath)
				fmt.Printf("\nTo completely remove the project files:\n")
				fmt.Printf("  rm -rf %s\n", project.WorkspacePath)
			}
		}

		return nil
	},
}

// isDirEmpty checks if a directory is empty
func isDirEmpty(dirPath string) (bool, error) {
	f, err := os.Open(dirPath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Try to read one entry
	if err == io.EOF {
		return true, nil // Directory is empty
	}
	return false, err // Directory is not empty or error occurred
}

// cleanupOrphanedContainers removes devbox containers that are not tracked in config
func cleanupOrphanedContainers() error {
	fmt.Println("Cleaning up orphaned devbox containers...")

	// Load configuration to get tracked projects
	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get all devbox containers
	boxes, err := dockerClient.ListBoxs()
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	// Build set of tracked box names
	trackedBoxes := make(map[string]bool)
	for _, project := range cfg.GetProjects() {
		trackedBoxes[project.BoxName] = true
	}

	// Find orphaned containers
	var orphanedBoxes []string
	for _, box := range boxes {
		for _, name := range box.Names {
			cleanName := strings.TrimPrefix(name, "/")
			if !trackedBoxes[cleanName] {
				orphanedBoxes = append(orphanedBoxes, cleanName)
			}
		}
	}

	if len(orphanedBoxes) == 0 {
		fmt.Println("No orphaned containers found.")
		return nil
	}

	fmt.Printf("Found %d orphaned devbox container(s):\n", len(orphanedBoxes))
	for _, boxName := range orphanedBoxes {
		fmt.Printf("  - %s\n", boxName)
	}

	// Confirm cleanup unless force flag is set
	if !forceFlag {
		fmt.Print("\nRemove these orphaned containers? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cleanup cancelled.")
			return nil
		}
	}

	// Remove orphaned containers
	var removed, failed int
	for _, boxName := range orphanedBoxes {
		fmt.Printf("Removing %s...\n", boxName)
		if err := dockerClient.RemoveBox(boxName); err != nil {
			fmt.Printf("Failed to remove %s: %v\n", boxName, err)
			failed++
		} else {
			fmt.Printf("Removed %s\n", boxName)
			removed++
		}
	}

	fmt.Printf("\nCleanup complete: %d removed, %d failed\n", removed, failed)
	if failed > 0 {
		return fmt.Errorf("failed to remove %d container(s)", failed)
	}

	return nil
}
