package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	dryRunFlag      bool
	allFlag         bool
	orphanedFlag    bool
	imagesFlag      bool
	volumesFlag     bool
	networksFlag    bool
	systemPruneFlag bool
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup [flags]",
	Short: "Clean up Docker resources and devbox artifacts",
	Long: `Clean up various Docker resources and devbox-related artifacts.
This command helps maintain a clean system by removing:

- Orphaned devbox containers (not tracked in config)
- Unused Docker images
- Unused Docker volumes  
- Unused Docker networks
- Dangling build artifacts

Examples:
  devbox cleanup                    # Interactive cleanup menu
  devbox cleanup --orphaned         # Remove orphaned containers only
  devbox cleanup --images           # Remove unused images only
  devbox cleanup --all              # Clean up everything
  devbox cleanup --system-prune     # Run docker system prune
  devbox cleanup --dry-run          # Show what would be cleaned`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		if !orphanedFlag && !imagesFlag && !volumesFlag && !networksFlag && !systemPruneFlag && !allFlag {
			return runInteractiveCleanup()
		}

		if allFlag {
			orphanedFlag = true
			imagesFlag = true
			volumesFlag = true
			networksFlag = true
		}

		var cleanupTasks []func() error

		if orphanedFlag {
			cleanupTasks = append(cleanupTasks, cleanupOrphanedFromCleanup)
		}

		if imagesFlag {
			cleanupTasks = append(cleanupTasks, cleanupUnusedImages)
		}

		if volumesFlag {
			cleanupTasks = append(cleanupTasks, cleanupUnusedVolumes)
		}

		if networksFlag {
			cleanupTasks = append(cleanupTasks, cleanupUnusedNetworks)
		}

		if systemPruneFlag {
			cleanupTasks = append(cleanupTasks, runSystemPrune)
		}

		for _, task := range cleanupTasks {
			if err := task(); err != nil {
				return err
			}
		}

		if len(cleanupTasks) > 0 {
			fmt.Printf("\n✅ Cleanup completed successfully!\n")
		}

		return nil
	},
}

func runInteractiveCleanup() error {
	fmt.Printf("🧹 Devbox Cleanup Menu\n\n")
	fmt.Printf("Available cleanup options:\n")
	fmt.Printf("  1. Clean up orphaned devbox containers\n")
	fmt.Printf("  2. Remove unused Docker images\n")
	fmt.Printf("  3. Remove unused Docker volumes\n")
	fmt.Printf("  4. Remove unused Docker networks\n")
	fmt.Printf("  5. Run Docker system prune (comprehensive cleanup)\n")
	fmt.Printf("  6. Clean up everything (options 1-4)\n")
	fmt.Printf("  7. Show system status (disk usage, containers, images)\n")
	fmt.Printf("  q. Quit\n\n")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Select an option [1-7, q]: ")
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		switch response {
		case "1":
			return cleanupOrphanedFromCleanup()
		case "2":
			return cleanupUnusedImages()
		case "3":
			return cleanupUnusedVolumes()
		case "4":
			return cleanupUnusedNetworks()
		case "5":
			return runSystemPrune()
		case "6":
			fmt.Printf("\nRunning comprehensive cleanup...\n")
			tasks := []func() error{
				cleanupOrphanedFromCleanup,
				cleanupUnusedImages,
				cleanupUnusedVolumes,
				cleanupUnusedNetworks,
			}
			for _, task := range tasks {
				if err := task(); err != nil {
					return err
				}
			}
			fmt.Printf("\n✅ Comprehensive cleanup completed!\n")
			return nil
		case "7":
			return showSystemStatus()
		case "q", "quit", "exit":
			fmt.Printf("Cleanup cancelled.\n")
			return nil
		default:
			fmt.Printf("Invalid option. Please select 1-7 or q.\n")
		}
	}
}

func cleanupOrphanedFromCleanup() error {
	fmt.Printf("🔍 Scanning for orphaned devbox containers...\n")

	if dryRunFlag {
		fmt.Printf("DRY RUN - No containers will be removed\n")
	}

	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	containers, err := dockerClient.ListBoxs()
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	trackedContainers := make(map[string]bool)
	for _, project := range cfg.GetProjects() {
		trackedContainers[project.BoxName] = true
	}

	var orphanedContainers []string
	for _, container := range containers {
		for _, name := range container.Names {
			cleanName := strings.TrimPrefix(name, "/")
			if strings.HasPrefix(cleanName, "devbox_") && !trackedContainers[cleanName] {
				orphanedContainers = append(orphanedContainers, cleanName)
			}
		}
	}

	if len(orphanedContainers) == 0 {
		fmt.Printf("✅ No orphaned containers found.\n")
		return nil
	}

	fmt.Printf("Found %d orphaned devbox container(s):\n", len(orphanedContainers))
	for _, containerName := range orphanedContainers {
		fmt.Printf("  • %s\n", containerName)
	}

	if dryRunFlag {
		fmt.Printf("\nDRY RUN: Would remove %d orphaned containers\n", len(orphanedContainers))
		return nil
	}

	if !forceFlag {
		fmt.Print("\nRemove these orphaned containers? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Printf("Cleanup cancelled.\n")
			return nil
		}
	}

	var removed, failed int
	for _, containerName := range orphanedContainers {
		fmt.Printf("Removing %s...\n", containerName)
		if err := dockerClient.RemoveBox(containerName); err != nil {
			fmt.Printf("❌ Failed to remove %s: %v\n", containerName, err)
			failed++
		} else {
			fmt.Printf("✅ Removed %s\n", containerName)
			removed++
		}
	}

	fmt.Printf("\nOrphaned containers cleanup complete: %d removed, %d failed\n", removed, failed)
	if failed > 0 {
		return fmt.Errorf("failed to remove %d container(s)", failed)
	}

	return nil
}

func cleanupUnusedImages() error {
	fmt.Printf("🔍 Scanning for unused Docker images...\n")

	if dryRunFlag {
		fmt.Printf("DRY RUN - No images will be removed\n")
		if err := dockerClient.RunDockerCommand([]string{"image", "prune", "--dry-run"}); err != nil {
			return fmt.Errorf("failed to show unused images: %w", err)
		}
	} else {
		if !forceFlag {
			fmt.Print("Remove unused Docker images? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read confirmation: %w", err)
			}

			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Printf("Image cleanup cancelled.\n")
				return nil
			}
		}

		fmt.Printf("Removing unused images...\n")
		if err := dockerClient.RunDockerCommand([]string{"image", "prune", "-f"}); err != nil {
			return fmt.Errorf("failed to prune images: %w", err)
		}
		fmt.Printf("✅ Unused images removed.\n")
	}

	return nil
}

func cleanupUnusedVolumes() error {
	fmt.Printf("🔍 Scanning for unused Docker volumes...\n")

	if dryRunFlag {
		fmt.Printf("DRY RUN - No volumes will be removed\n")
		if err := dockerClient.RunDockerCommand([]string{"volume", "prune", "--dry-run"}); err != nil {
			return fmt.Errorf("failed to show unused volumes: %w", err)
		}
	} else {
		if !forceFlag {
			fmt.Print("Remove unused Docker volumes? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read confirmation: %w", err)
			}

			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Printf("Volume cleanup cancelled.\n")
				return nil
			}
		}

		fmt.Printf("Removing unused volumes...\n")
		if err := dockerClient.RunDockerCommand([]string{"volume", "prune", "-f"}); err != nil {
			return fmt.Errorf("failed to prune volumes: %w", err)
		}
		fmt.Printf("✅ Unused volumes removed.\n")
	}

	return nil
}

func cleanupUnusedNetworks() error {
	fmt.Printf("🔍 Scanning for unused Docker networks...\n")

	if dryRunFlag {
		fmt.Printf("DRY RUN - No networks will be removed\n")
		if err := dockerClient.RunDockerCommand([]string{"network", "prune", "--dry-run"}); err != nil {
			return fmt.Errorf("failed to show unused networks: %w", err)
		}
	} else {
		if !forceFlag {
			fmt.Print("Remove unused Docker networks? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read confirmation: %w", err)
			}

			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Printf("Network cleanup cancelled.\n")
				return nil
			}
		}

		fmt.Printf("Removing unused networks...\n")
		if err := dockerClient.RunDockerCommand([]string{"network", "prune", "-f"}); err != nil {
			return fmt.Errorf("failed to prune networks: %w", err)
		}
		fmt.Printf("✅ Unused networks removed.\n")
	}

	return nil
}

func runSystemPrune() error {
	fmt.Printf("🔍 Running comprehensive Docker system cleanup...\n")

	if dryRunFlag {
		fmt.Printf("DRY RUN - No resources will be removed\n")
		if err := dockerClient.RunDockerCommand([]string{"system", "prune", "--dry-run"}); err != nil {
			return fmt.Errorf("failed to show system prune info: %w", err)
		}
	} else {
		if !forceFlag {
			fmt.Print("Run Docker system prune (removes all unused resources)? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read confirmation: %w", err)
			}

			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Printf("System prune cancelled.\n")
				return nil
			}
		}

		fmt.Printf("Running system prune...\n")
		if err := dockerClient.RunDockerCommand([]string{"system", "prune", "-f"}); err != nil {
			return fmt.Errorf("failed to run system prune: %w", err)
		}
		fmt.Printf("✅ System prune completed.\n")
	}

	return nil
}

func showSystemStatus() error {
	fmt.Printf("📊 Docker System Status\n\n")

	fmt.Printf("=== Disk Usage ===\n")
	if err := dockerClient.RunDockerCommand([]string{"system", "df"}); err != nil {
		fmt.Printf("❌ Failed to get disk usage: %v\n", err)
	}

	fmt.Printf("\n=== Devbox Containers ===\n")
	containers, err := dockerClient.ListBoxs()
	if err != nil {
		fmt.Printf("❌ Failed to list containers: %v\n", err)
	} else {
		fmt.Printf("Active devbox containers: %d\n", len(containers))
		for _, container := range containers {
			for _, name := range container.Names {
				fmt.Printf("  • %s (%s)\n", strings.TrimPrefix(name, "/"), container.Status)
			}
		}
	}

	fmt.Printf("\n=== Tracked Projects ===\n")
	cfg, err := configManager.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
	} else {
		projects := cfg.GetProjects()
		fmt.Printf("Tracked projects: %d\n", len(projects))
		for name, project := range projects {
			fmt.Printf("  • %s -> %s\n", name, project.BoxName)
		}
	}

	fmt.Printf("\n=== Docker Version ===\n")
	if err := dockerClient.RunDockerCommand([]string{"version", "--format", "{{.Server.Version}}"}); err != nil {
		fmt.Printf("❌ Failed to get Docker version: %v\n", err)
	}

	return nil
}

func init() {
	cleanupCmd.Flags().BoolVarP(&dryRunFlag, "dry-run", "n", false, "Show what would be cleaned without actually removing anything")
	cleanupCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Clean up all unused resources (containers, images, volumes, networks)")
	cleanupCmd.Flags().BoolVar(&orphanedFlag, "orphaned", false, "Clean up orphaned devbox containers only")
	cleanupCmd.Flags().BoolVar(&imagesFlag, "images", false, "Clean up unused Docker images only")
	cleanupCmd.Flags().BoolVar(&volumesFlag, "volumes", false, "Clean up unused Docker volumes only")
	cleanupCmd.Flags().BoolVar(&networksFlag, "networks", false, "Clean up unused Docker networks only")
	cleanupCmd.Flags().BoolVar(&systemPruneFlag, "system-prune", false, "Run Docker system prune for comprehensive cleanup")
	cleanupCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force cleanup without confirmation prompts")
}
