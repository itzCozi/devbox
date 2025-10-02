package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	verboseFlag bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all devbox projects and their status",
	Long:  `Display all managed devbox projects along with their box status.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

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

		boxes, err := dockerClient.ListBoxes()
		if err != nil {
			return fmt.Errorf("failed to list boxes: %w", err)
		}

		boxStatus := make(map[string]string)
		for _, box := range boxes {
			for _, name := range box.Names {

				cleanName := strings.TrimPrefix(name, "/")
				boxStatus[cleanName] = box.Status
			}
		}

		fmt.Printf("DEVBOX PROJECTS\n")
		if verboseFlag {
			fmt.Printf("%-20s %-20s %-15s %-12s %s\n", "PROJECT", "BOX", "STATUS", "CONFIG", "WORKSPACE")
			fmt.Printf("%-20s %-20s %-15s %-12s %s\n",
				strings.Repeat("-", 20),
				strings.Repeat("-", 20),
				strings.Repeat("-", 15),
				strings.Repeat("-", 12),
				strings.Repeat("-", 30))
		} else {
			fmt.Printf("%-20s %-20s %-15s %s\n", "PROJECT", "BOX", "STATUS", "WORKSPACE")
			fmt.Printf("%-20s %-20s %-15s %s\n",
				strings.Repeat("-", 20),
				strings.Repeat("-", 20),
				strings.Repeat("-", 15),
				strings.Repeat("-", 30))
		}

		for _, project := range projects {
			status := "not found"
			if boxStatus[project.BoxName] != "" {
				status = boxStatus[project.BoxName]
			}

			configStatus := "none"
			if project.ConfigFile != "" {
				configStatus = "devbox.json"
			} else {

				projectConfig, err := configManager.LoadProjectConfig(project.WorkspacePath)
				if err == nil && projectConfig != nil {
					configStatus = "devbox.json"
				}
			}

			if verboseFlag {
				fmt.Printf("%-20s %-20s %-15s %-12s %s\n",
					project.Name,
					project.BoxName,
					status,
					configStatus,
					project.WorkspacePath)
			} else {
				fmt.Printf("%-20s %-20s %-15s %s\n",
					project.Name,
					project.BoxName,
					status,
					project.WorkspacePath)
			}

			if verboseFlag {
				projectConfig, err := configManager.LoadProjectConfig(project.WorkspacePath)
				if err == nil && projectConfig != nil {
					if projectConfig.BaseImage != "" && projectConfig.BaseImage != project.BaseImage {
						fmt.Printf("  - Base image: %s (override)\n", projectConfig.BaseImage)
					}
					if len(projectConfig.Ports) > 0 {
						fmt.Printf("  - Ports: %s\n", strings.Join(projectConfig.Ports, ", "))
					}
					if len(projectConfig.SetupCommands) > 0 {
						fmt.Printf("  - Setup commands: %d\n", len(projectConfig.SetupCommands))
					}
				}
			}
		}

		fmt.Printf("\nTotal projects: %d\n", len(projects))

		if verboseFlag {

			if cfg.Settings != nil {
				fmt.Printf("\nGlobal settings:\n")
				fmt.Printf("  Default base image: %s\n", cfg.Settings.DefaultBaseImage)
				fmt.Printf("  Auto update: %t\n", cfg.Settings.AutoUpdate)
			}
		} else {
			fmt.Printf("\nUse --verbose for detailed information including configurations.\n")
		}

		return nil
	},
}

func init() {
	listCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Show detailed information including configuration details")
}
