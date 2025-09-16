package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all devbox projects and their status",
	Long:  `Display all managed devbox projects along with their box status.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := configManager.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		projects := cfg.GetProjects()
		if len(projects) == 0 {
			fmt.Println("No devbox projects found.")
			fmt.Println("Create a new project with: devbox init <project-name>")
			return nil
		}

		// Get box statuses
		boxs, err := dockerClient.ListBoxs()
		if err != nil {
			return fmt.Errorf("failed to list boxs: %w", err)
		}

		// Create a map of box names to their status
		boxStatus := make(map[string]string)
		for _, box := range boxs {
			for _, name := range box.Names {
				// Remove leading slash from box name
				cleanName := strings.TrimPrefix(name, "/")
				boxStatus[cleanName] = box.Status
			}
		}

		// Display projects
		fmt.Printf("DEVBOX PROJECTS\n")
		fmt.Printf("%-20s %-20s %-15s %s\n", "PROJECT", "BOX", "STATUS", "WORKSPACE")
		fmt.Printf("%-20s %-20s %-15s %s\n", strings.Repeat("-", 20), strings.Repeat("-", 20), strings.Repeat("-", 15), strings.Repeat("-", 30))

		for _, project := range projects {
			status := "not found"
			if boxStatus[project.BoxName] != "" {
				status = boxStatus[project.BoxName]
			}

			fmt.Printf("%-20s %-20s %-15s %s\n",
				project.Name,
				project.BoxName,
				status,
				project.WorkspacePath)
		}

		fmt.Printf("\nTotal projects: %d\n", len(projects))
		return nil
	},
}
