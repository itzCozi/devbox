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
  devbox destroy --cleanup-orphaned  Remove boxes not tracked in config`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		if projectName == "--cleanup-orphaned" {
			return cleanupOrphanedboxes()
		}

		if err := validateProjectName(projectName); err != nil {
			return err
		}

		cfg, err := configManager.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		project, exists := cfg.GetProject(projectName)
		if !exists {
			return fmt.Errorf("project '%s' not found", projectName)
		}

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

		exists, err = dockerClient.BoxExists(project.BoxName)
		if err != nil {
			return fmt.Errorf("failed to check box status: %w", err)
		}

		if exists {

			fmt.Printf("Stopping and removing box '%s'...\n", project.BoxName)
			if err := dockerClient.RemoveBox(project.BoxName); err != nil {
				fmt.Printf("Warning: failed to remove box: %v\n", err)

			}
		} else {
			fmt.Printf("Box '%s' not found (already removed)\n", project.BoxName)
		}

		cfg.RemoveProject(projectName)
		if err := configManager.Save(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("Project '%s' destroyed successfully!\n", projectName)

		if _, err := os.Stat(project.WorkspacePath); err == nil {

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

func isDirEmpty(dirPath string) (bool, error) {
	f, err := os.Open(dirPath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func cleanupOrphanedboxes() error {
	fmt.Println("Cleaning up orphaned devbox boxes...")

	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	boxes, err := dockerClient.ListBoxes()
	if err != nil {
		return fmt.Errorf("failed to list boxes: %w", err)
	}

	trackedBoxes := make(map[string]bool)
	for _, project := range cfg.GetProjects() {
		trackedBoxes[project.BoxName] = true
	}

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
		fmt.Println("No orphaned boxes found.")
		return nil
	}

	fmt.Printf("Found %d orphaned devbox box(s):\n", len(orphanedBoxes))
	for _, boxName := range orphanedBoxes {
		fmt.Printf("  - %s\n", boxName)
	}

	if !forceFlag {
		fmt.Print("\nRemove these orphaned boxes? (y/N): ")
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
		return fmt.Errorf("failed to remove %d box(s)", failed)
	}

	return nil
}
