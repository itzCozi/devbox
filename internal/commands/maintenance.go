package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	updateFlag      bool
	healthCheckFlag bool
	rebuildFlag     bool
	restartFlag     bool
	statusCheckFlag bool
	autoRepairFlag  bool
)

var maintenanceCmd = &cobra.Command{
	Use:   "maintenance [flags]",
	Short: "Perform maintenance tasks on devbox projects and containers",
	Long: `Perform various maintenance tasks to keep your devbox environment healthy:

- Update system packages in containers
- Check health status of all projects
- Rebuild containers from latest base images
- Restart stopped or problematic containers
- Auto-repair common issues
- System status checks

Examples:
  devbox maintenance                     # Interactive maintenance menu
  devbox maintenance --update            # Update all containers
  devbox maintenance --health-check      # Check health of all projects
  devbox maintenance --restart           # Restart all stopped containers
  devbox maintenance --rebuild           # Rebuild all containers
  devbox maintenance --status            # Show detailed status
  devbox maintenance --auto-repair       # Auto-fix common issues`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		if !updateFlag && !healthCheckFlag && !rebuildFlag && !restartFlag && !statusCheckFlag && !autoRepairFlag {
			return runInteractiveMaintenance()
		}

		var maintenanceTasks []func() error

		if statusCheckFlag {
			maintenanceTasks = append(maintenanceTasks, performStatusCheck)
		}

		if healthCheckFlag {
			maintenanceTasks = append(maintenanceTasks, performHealthCheck)
		}

		if updateFlag {
			maintenanceTasks = append(maintenanceTasks, updateAllContainers)
		}

		if restartFlag {
			maintenanceTasks = append(maintenanceTasks, restartStoppedContainers)
		}

		if rebuildFlag {
			maintenanceTasks = append(maintenanceTasks, rebuildAllContainers)
		}

		if autoRepairFlag {
			maintenanceTasks = append(maintenanceTasks, autoRepairIssues)
		}

		for _, task := range maintenanceTasks {
			if err := task(); err != nil {
				return err
			}
		}

		if len(maintenanceTasks) > 0 {
			fmt.Printf("\n✅ Maintenance completed successfully!\n")
		}

		return nil
	},
}

func runInteractiveMaintenance() error {
	fmt.Printf("🔧 Devbox Maintenance Menu\n\n")
	fmt.Printf("Available maintenance options:\n")
	fmt.Printf("  1. Check system status\n")
	fmt.Printf("  2. Perform health check on all projects\n")
	fmt.Printf("  3. Update system packages in all containers\n")
	fmt.Printf("  4. Restart stopped containers\n")
	fmt.Printf("  5. Rebuild all containers from latest base images\n")
	fmt.Printf("  6. Auto-repair common issues\n")
	fmt.Printf("  7. Full maintenance (options 2-4)\n")
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
			return performStatusCheck()
		case "2":
			return performHealthCheck()
		case "3":
			return updateAllContainers()
		case "4":
			return restartStoppedContainers()
		case "5":
			return rebuildAllContainers()
		case "6":
			return autoRepairIssues()
		case "7":
			fmt.Printf("\nRunning full maintenance...\n")
			tasks := []func() error{
				performHealthCheck,
				updateAllContainers,
				restartStoppedContainers,
			}
			for _, task := range tasks {
				if err := task(); err != nil {
					return err
				}
			}
			fmt.Printf("\n✅ Full maintenance completed!\n")
			return nil
		case "q", "quit", "exit":
			fmt.Printf("Maintenance cancelled.\n")
			return nil
		default:
			fmt.Printf("Invalid option. Please select 1-7 or q.\n")
		}
	}
}

func performStatusCheck() error {
	fmt.Printf("📊 Devbox System Status Check\n")
	fmt.Printf("=====================================\n\n")

	fmt.Printf("🐳 Docker Status: ")
	if err := dockerClient.RunDockerCommand([]string{"version", "--format", "Server: {{.Server.Version}}"}); err != nil {
		fmt.Printf("❌ Docker not available: %v\n", err)
		return err
	}

	cfg, err := configManager.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		return err
	}

	projects := cfg.GetProjects()
	fmt.Printf("\n📁 Projects: %d total\n", len(projects))

	containers, err := dockerClient.ListBoxs()
	if err != nil {
		fmt.Printf("❌ Failed to list containers: %v\n", err)
		return err
	}

	containerStatus := make(map[string]string)
	for _, container := range containers {
		for _, name := range container.Names {
			cleanName := strings.TrimPrefix(name, "/")
			containerStatus[cleanName] = container.Status
		}
	}

	var running, stopped, missing int
	fmt.Printf("\n🔍 Container Status:\n")
	for projectName, project := range projects {
		status := containerStatus[project.BoxName]
		if status == "" {
			fmt.Printf("  ❌ %s -> %s (MISSING)\n", projectName, project.BoxName)
			missing++
		} else if strings.Contains(status, "Up") {
			fmt.Printf("  ✅ %s -> %s (RUNNING)\n", projectName, project.BoxName)
			running++
		} else {
			fmt.Printf("  ⚠️  %s -> %s (STOPPED)\n", projectName, project.BoxName)
			stopped++
		}
	}

	fmt.Printf("\nSummary: %d running, %d stopped, %d missing\n", running, stopped, missing)

	fmt.Printf("\n💾 Docker Disk Usage:\n")
	if err := dockerClient.RunDockerCommand([]string{"system", "df"}); err != nil {
		fmt.Printf("❌ Failed to get disk usage: %v\n", err)
	}

	return nil
}

func performHealthCheck() error {
	fmt.Printf("🏥 Health Check: Scanning all devbox projects...\n")

	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	projects := cfg.GetProjects()
	if len(projects) == 0 {
		fmt.Printf("✅ No projects to check.\n")
		return nil
	}

	containers, err := dockerClient.ListBoxs()
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	containerStatus := make(map[string]string)
	for _, container := range containers {
		for _, name := range container.Names {
			cleanName := strings.TrimPrefix(name, "/")
			containerStatus[cleanName] = container.Status
		}
	}

	var healthy, unhealthy, missing int

	fmt.Printf("\nProject Health Report:\n")
	fmt.Printf("----------------------\n")

	for projectName, project := range projects {
		fmt.Printf("📦 %s: ", projectName)

		status := containerStatus[project.BoxName]
		if status == "" {
			fmt.Printf("❌ Container missing\n")
			missing++
			continue
		}

		if !strings.Contains(status, "Up") {
			fmt.Printf("⚠️  Container stopped (%s)\n", status)
			unhealthy++
			continue
		}

		if _, err := os.Stat(project.WorkspacePath); os.IsNotExist(err) {
			fmt.Printf("❌ Workspace directory missing\n")
			unhealthy++
			continue
		}

		if err := dockerClient.RunDockerCommand([]string{"exec", project.BoxName, "echo", "health-check"}); err != nil {
			fmt.Printf("❌ Container not responsive\n")
			unhealthy++
			continue
		}

		fmt.Printf("✅ Healthy\n")
		healthy++
	}

	fmt.Printf("\nHealth Summary:\n")
	fmt.Printf("  ✅ Healthy: %d\n", healthy)
	fmt.Printf("  ⚠️  Unhealthy: %d\n", unhealthy)
	fmt.Printf("  ❌ Missing: %d\n", missing)

	if unhealthy > 0 || missing > 0 {
		fmt.Printf("\n💡 Tip: Use 'devbox maintenance --auto-repair' to fix common issues\n")
	}

	return nil
}

func updateAllContainers() error {
	fmt.Printf("📦 Updating system packages in all devbox containers...\n")

	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	projects := cfg.GetProjects()
	if len(projects) == 0 {
		fmt.Printf("✅ No projects to update.\n")
		return nil
	}

	var updated, failed int

	for projectName, project := range projects {
		fmt.Printf("\n🔄 Updating %s...\n", projectName)

		status, err := dockerClient.GetBoxStatus(project.BoxName)
		if err != nil {
			fmt.Printf("❌ Failed to check status for %s: %v\n", projectName, err)
			failed++
			continue
		}

		if status == "not found" {
			fmt.Printf("⚠️  Container %s not found, skipping\n", project.BoxName)
			continue
		}

		if status != "running" {
			fmt.Printf("🚀 Starting %s...\n", project.BoxName)
			if err := dockerClient.StartBox(project.BoxName); err != nil {
				fmt.Printf("❌ Failed to start %s: %v\n", project.BoxName, err)
				failed++
				continue
			}

			time.Sleep(2 * time.Second)
		}

		updateCommands := []string{
			"apt update -y",
			"apt full-upgrade -y",
			"apt autoremove -y",
			"apt autoclean",
		}

		if err := dockerClient.ExecuteSetupCommands(project.BoxName, updateCommands); err != nil {
			fmt.Printf("❌ Failed to update %s: %v\n", projectName, err)
			failed++
		} else {
			fmt.Printf("✅ Updated %s successfully\n", projectName)
			updated++
		}
	}

	fmt.Printf("\nUpdate Summary: %d updated, %d failed\n", updated, failed)
	if failed > 0 {
		return fmt.Errorf("failed to update %d container(s)", failed)
	}

	return nil
}

func restartStoppedContainers() error {
	fmt.Printf("🔄 Restarting stopped devbox containers...\n")

	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	projects := cfg.GetProjects()
	if len(projects) == 0 {
		fmt.Printf("✅ No projects to restart.\n")
		return nil
	}

	var restarted, failed int

	for projectName, project := range projects {
		status, err := dockerClient.GetBoxStatus(project.BoxName)
		if err != nil {
			fmt.Printf("❌ Failed to check status for %s: %v\n", projectName, err)
			failed++
			continue
		}

		if status == "not found" {
			fmt.Printf("⚠️  Container %s not found, skipping\n", project.BoxName)
			continue
		}

		if status != "running" {
			fmt.Printf("🚀 Starting %s...\n", projectName)
			if err := dockerClient.StartBox(project.BoxName); err != nil {
				fmt.Printf("❌ Failed to start %s: %v\n", projectName, err)
				failed++
			} else {
				fmt.Printf("✅ Started %s\n", projectName)
				restarted++
			}
		} else {
			fmt.Printf("✅ %s already running\n", projectName)
		}
	}

	fmt.Printf("\nRestart Summary: %d restarted, %d failed\n", restarted, failed)
	if failed > 0 {
		return fmt.Errorf("failed to restart %d container(s)", failed)
	}

	return nil
}

func rebuildAllContainers() error {
	fmt.Printf("🔨 Rebuilding all devbox containers from latest base images...\n")

	if !forceFlag {
		fmt.Print("This will destroy and recreate all containers. Continue? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Printf("Rebuild cancelled.\n")
			return nil
		}
	}

	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	projects := cfg.GetProjects()
	if len(projects) == 0 {
		fmt.Printf("✅ No projects to rebuild.\n")
		return nil
	}

	var rebuilt, failed int

	for projectName, project := range projects {
		fmt.Printf("\n🔨 Rebuilding %s...\n", projectName)

		if exists, err := dockerClient.BoxExists(project.BoxName); err != nil {
			fmt.Printf("❌ Failed to check if %s exists: %v\n", project.BoxName, err)
			failed++
			continue
		} else if exists {
			fmt.Printf("🛑 Stopping and removing existing container...\n")
			dockerClient.StopBox(project.BoxName)
			if err := dockerClient.RemoveBox(project.BoxName); err != nil {
				fmt.Printf("❌ Failed to remove %s: %v\n", project.BoxName, err)
				failed++
				continue
			}
		}

		fmt.Printf("🚀 Recreating container...\n")

		projectConfig, err := configManager.LoadProjectConfig(project.WorkspacePath)
		if err != nil {
			fmt.Printf("⚠️  Could not load project config: %v\n", err)
		}

		baseImage := cfg.GetEffectiveBaseImage(project, projectConfig)
		if err := dockerClient.PullImage(baseImage); err != nil {
			fmt.Printf("❌ Failed to pull %s: %v\n", baseImage, err)
			failed++
			continue
		}

		workspaceBox := "/workspace"
		if projectConfig != nil && projectConfig.WorkingDir != "" {
			workspaceBox = projectConfig.WorkingDir
		}

		boxID, err := dockerClient.CreateBox(project.BoxName, baseImage, project.WorkspacePath, workspaceBox)
		if err != nil {
			fmt.Printf("❌ Failed to create %s: %v\n", project.BoxName, err)
			failed++
			continue
		}

		if err := dockerClient.StartBox(boxID); err != nil {
			fmt.Printf("❌ Failed to start %s: %v\n", project.BoxName, err)
			failed++
			continue
		}

		if err := dockerClient.WaitForBox(project.BoxName, 30*time.Second); err != nil {
			fmt.Printf("❌ Container %s failed to start: %v\n", project.BoxName, err)
			failed++
			continue
		}

		updateCommands := []string{
			"apt update -y",
			"apt full-upgrade -y",
		}
		if err := dockerClient.ExecuteSetupCommands(project.BoxName, updateCommands); err != nil {
			fmt.Printf("⚠️  Failed to update system packages: %v\n", err)
		}

		if projectConfig != nil && len(projectConfig.SetupCommands) > 0 {
			if err := dockerClient.ExecuteSetupCommands(project.BoxName, projectConfig.SetupCommands); err != nil {
				fmt.Printf("⚠️  Failed to execute setup commands: %v\n", err)
			}
		}

		if err := dockerClient.SetupDevboxInBoxWithUpdate(project.BoxName, projectName); err != nil {
			fmt.Printf("⚠️  Failed to setup devbox environment: %v\n", err)
		}

		fmt.Printf("✅ Rebuilt %s successfully\n", projectName)
		rebuilt++
	}

	fmt.Printf("\nRebuild Summary: %d rebuilt, %d failed\n", rebuilt, failed)
	if failed > 0 {
		return fmt.Errorf("failed to rebuild %d container(s)", failed)
	}

	return nil
}

func autoRepairIssues() error {
	fmt.Printf("🔧 Auto-repairing common devbox issues...\n")

	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	projects := cfg.GetProjects()
	if len(projects) == 0 {
		fmt.Printf("✅ No projects to repair.\n")
		return nil
	}

	var repaired, failed int

	for projectName, project := range projects {
		fmt.Printf("\n🔍 Checking %s...\n", projectName)

		issuesFound := false

		if _, err := os.Stat(project.WorkspacePath); os.IsNotExist(err) {
			fmt.Printf("📁 Creating missing workspace directory...\n")
			if err := os.MkdirAll(project.WorkspacePath, 0755); err != nil {
				fmt.Printf("❌ Failed to create workspace: %v\n", err)
				failed++
				continue
			}
			issuesFound = true
		}

		status, err := dockerClient.GetBoxStatus(project.BoxName)
		if err != nil {
			fmt.Printf("❌ Failed to check container status: %v\n", err)
			failed++
			continue
		}

		if status == "not found" {
			fmt.Printf("🔨 Recreating missing container...\n")

			projectConfig, _ := configManager.LoadProjectConfig(project.WorkspacePath)
			baseImage := cfg.GetEffectiveBaseImage(project, projectConfig)

			workspaceBox := "/workspace"
			if projectConfig != nil && projectConfig.WorkingDir != "" {
				workspaceBox = projectConfig.WorkingDir
			}

			boxID, err := dockerClient.CreateBox(project.BoxName, baseImage, project.WorkspacePath, workspaceBox)
			if err != nil {
				fmt.Printf("❌ Failed to recreate container: %v\n", err)
				failed++
				continue
			}

			if err := dockerClient.StartBox(boxID); err != nil {
				fmt.Printf("❌ Failed to start container: %v\n", err)
				failed++
				continue
			}

			if err := dockerClient.SetupDevboxInBoxWithUpdate(project.BoxName, projectName); err != nil {
				fmt.Printf("⚠️  Failed to setup devbox environment: %v\n", err)
			}

			issuesFound = true
		} else if status != "running" {
			fmt.Printf("🚀 Starting stopped container...\n")
			if err := dockerClient.StartBox(project.BoxName); err != nil {
				fmt.Printf("❌ Failed to start container: %v\n", err)
				failed++
				continue
			}
			issuesFound = true
		}

		if status != "not found" {
			if err := dockerClient.RunDockerCommand([]string{"exec", project.BoxName, "echo", "test"}); err != nil {
				fmt.Printf("🔄 Container unresponsive, restarting...\n")
				dockerClient.StopBox(project.BoxName)
				if err := dockerClient.StartBox(project.BoxName); err != nil {
					fmt.Printf("❌ Failed to restart container: %v\n", err)
					failed++
					continue
				}
				issuesFound = true
			}
		}

		if issuesFound {
			fmt.Printf("✅ Repaired %s\n", projectName)
			repaired++
		} else {
			fmt.Printf("✅ %s is healthy\n", projectName)
		}
	}

	fmt.Printf("\nAuto-repair Summary: %d repaired, %d failed\n", repaired, failed)
	if failed > 0 {
		return fmt.Errorf("failed to repair %d project(s)", failed)
	}

	return nil
}

func init() {
	maintenanceCmd.Flags().BoolVar(&updateFlag, "update", false, "Update system packages in all containers")
	maintenanceCmd.Flags().BoolVar(&healthCheckFlag, "health-check", false, "Perform health check on all projects")
	maintenanceCmd.Flags().BoolVar(&rebuildFlag, "rebuild", false, "Rebuild all containers from latest base images")
	maintenanceCmd.Flags().BoolVar(&restartFlag, "restart", false, "Restart stopped containers")
	maintenanceCmd.Flags().BoolVar(&statusCheckFlag, "status", false, "Show detailed system status")
	maintenanceCmd.Flags().BoolVar(&autoRepairFlag, "auto-repair", false, "Automatically repair common issues")
	maintenanceCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force operations without confirmation prompts")
}
