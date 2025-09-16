package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all devbox projects and their status",
	Long:  `Display all managed devbox projects along with their container status.`,
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

		// Get container statuses
		containers, err := dockerClient.ListContainers()
		if err != nil {
			return fmt.Errorf("failed to list containers: %w", err)
		}

		// Create a map of container names to their status
		containerStatus := make(map[string]string)
		for _, container := range containers {
			for _, name := range container.Names {
				// Remove leading slash from container name
				cleanName := strings.TrimPrefix(name, "/")
				containerStatus[cleanName] = container.Status
			}
		}

		// Display projects
		fmt.Printf("DEVBOX PROJECTS\n")
		fmt.Printf("%-20s %-20s %-15s %s\n", "PROJECT", "CONTAINER", "STATUS", "WORKSPACE")
		fmt.Printf("%-20s %-20s %-15s %s\n", strings.Repeat("-", 20), strings.Repeat("-", 20), strings.Repeat("-", 15), strings.Repeat("-", 30))

		for _, project := range projects {
			status := "not found"
			if containerStatus[project.ContainerName] != "" {
				status = containerStatus[project.ContainerName]
			}

			fmt.Printf("%-20s %-20s %-15s %s\n",
				project.Name,
				project.ContainerName,
				status,
				project.WorkspacePath)
		}

		fmt.Printf("\nTotal projects: %d\n", len(projects))
		return nil
	},
}
