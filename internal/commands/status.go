package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [project]",
	Short: "Show detailed status for a devbox project",
	Long:  "Displays container state, resource usage, uptime, ports, mounts, and other diagnostics for the project's box.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var projectName string
		if len(args) == 1 {
			projectName = args[0]
		} else {

			boxes, err := dockerClient.ListBoxes()
			if err != nil {
				return fmt.Errorf("failed to list boxes: %w", err)
			}
			if len(boxes) == 0 {
				fmt.Println("No devbox containers found.")
				return nil
			}
			fmt.Println("Devbox containers:")
			for _, b := range boxes {
				name := ""
				if len(b.Names) > 0 {
					name = b.Names[0]
				}
				fmt.Printf("- %s\t%s\t%s\n", name, b.Status, b.Image)
			}
			fmt.Println("\nTip: devbox status <project> for detailed view.")
			return nil
		}

		if err := validateProjectName(projectName); err != nil {
			return fmt.Errorf("invalid project name: %w", err)
		}

		cfg, err := configManager.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		project, ok := cfg.GetProject(projectName)
		if !ok {
			return fmt.Errorf("project '%s' not found", projectName)
		}

		box := project.BoxName
		if box == "" {
			box = fmt.Sprintf("devbox_%s", projectName)
		}

		exists, err := dockerClient.BoxExists(box)
		if err != nil {
			return fmt.Errorf("failed to check if box exists: %w", err)
		}
		if !exists {
			fmt.Printf("Project: %s\nBox: %s (not found)\n", projectName, box)
			return nil
		}

		status, err := dockerClient.GetBoxStatus(box)
		if err != nil {
			return fmt.Errorf("failed to get box status: %w", err)
		}
		stats, _ := dockerClient.GetContainerStats(box)
		uptime, _ := dockerClient.GetUptime(box)
		ports, _ := dockerClient.GetPortMappings(box)
		mounts, _ := dockerClient.GetMounts(box)

		fmt.Printf("Devbox status\n")
		fmt.Printf("Project: %s\n", projectName)
		fmt.Printf("Box: %s\n", box)
		fmt.Printf("Image: %s\n", project.BaseImage)
		fmt.Printf("State: %s\n", status)
		if uptime > 0 {
			fmt.Printf("Uptime: %s\n", humanizeDuration(uptime))
		} else {
			fmt.Printf("Uptime: -\n")
		}
		if stats != nil {
			fmt.Printf("CPU: %s\n", stats.CPUPercent)
			fmt.Printf("Memory: %s (%s)\n", stats.MemUsage, stats.MemPercent)
			if stats.NetIO != "" {
				fmt.Printf("Net I/O: %s\n", stats.NetIO)
			}
			if stats.BlockIO != "" {
				fmt.Printf("Block I/O: %s\n", stats.BlockIO)
			}
			if stats.PIDs != "" {
				fmt.Printf("PIDs: %s\n", stats.PIDs)
			}
		}
		if len(ports) > 0 {
			fmt.Printf("Ports:\n  %s\n", strings.Join(ports, "\n  "))
		} else {
			fmt.Println("Ports: -")
		}
		if len(mounts) > 0 {
			fmt.Printf("Mounts:\n  %s\n", strings.Join(mounts, "\n  "))
		}

		return nil
	},
}

func humanizeDuration(d time.Duration) string {
	d = d.Round(time.Second)
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, mins, secs)
	}
	if mins > 0 {
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
