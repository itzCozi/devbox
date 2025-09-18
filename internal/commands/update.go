package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [project]",
	Short: "Pull latest base image(s) and rebuild box(es)",
	Long:  "Update environments by pulling the latest base images and rebuilding the project boxes using current configuration.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			projectName := args[0]
			if err := validateProjectName(projectName); err != nil {
				return err
			}
			return updateSingleProject(projectName)
		}

		return updateAllProjects()
	},
}

func updateSingleProject(projectName string) error {
	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	project, exists := cfg.GetProject(projectName)
	if !exists {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	projectConfig, _ := configManager.LoadProjectConfig(project.WorkspacePath)
	baseImage := cfg.GetEffectiveBaseImage(project, projectConfig)

	fmt.Printf("📥 Pulling latest base image for '%s': %s\n", projectName, baseImage)
	if err := dockerClient.RunDockerCommand([]string{"pull", baseImage}); err != nil {
		return fmt.Errorf("failed to pull base image %s: %w", baseImage, err)
	}

	existsBox, err := dockerClient.BoxExists(project.BoxName)
	if err != nil {
		return fmt.Errorf("failed to check box existence: %w", err)
	}
	if existsBox {
		fmt.Printf("🛑 Stopping and removing existing box '%s'...\n", project.BoxName)

		_ = dockerClient.StopBox(project.BoxName)
		if err := dockerClient.RemoveBox(project.BoxName); err != nil {
			return fmt.Errorf("failed to remove existing box: %w", err)
		}
	}

	workspaceBox := "/workspace"
	if projectConfig != nil && projectConfig.WorkingDir != "" {
		workspaceBox = projectConfig.WorkingDir
	}

	var configMap map[string]interface{}
	if projectConfig != nil {
		if data, err := json.Marshal(projectConfig); err == nil {
			_ = json.Unmarshal(data, &configMap)
		}
	}

	fmt.Printf("🚀 Recreating box '%s' with image '%s'...\n", project.BoxName, baseImage)
	boxID, err := dockerClient.CreateBoxWithConfig(project.BoxName, baseImage, project.WorkspacePath, workspaceBox, configMap)
	if err != nil {
		return fmt.Errorf("failed to create box: %w", err)
	}

	if err := dockerClient.StartBox(boxID); err != nil {
		return fmt.Errorf("failed to start box: %w", err)
	}

	if err := dockerClient.WaitForBox(project.BoxName, 30*time.Second); err != nil {
		return fmt.Errorf("box failed to become ready: %w", err)
	}

	updateCommands := []string{
		"apt update -y",
		"apt full-upgrade -y",
	}
	if err := dockerClient.ExecuteSetupCommandsWithOutput(project.BoxName, updateCommands, false); err != nil {
		fmt.Printf("⚠️  Failed to update system packages: %v\n", err)
	}

	if projectConfig != nil && len(projectConfig.SetupCommands) > 0 {
		if err := dockerClient.ExecuteSetupCommandsWithOutput(project.BoxName, projectConfig.SetupCommands, false); err != nil {
			fmt.Printf("⚠️  Failed to execute setup commands: %v\n", err)
		}
	}

	if err := dockerClient.SetupDevboxInBoxWithUpdate(project.BoxName, projectName); err != nil {
		fmt.Printf("⚠️  Failed to setup devbox environment: %v\n", err)
	}

	if project.BaseImage != baseImage {
		project.BaseImage = baseImage
		if err := configManager.Save(cfg); err != nil {
			return fmt.Errorf("failed to save updated config: %w", err)
		}
	}

	fmt.Printf("✅ Updated '%s' successfully\n", projectName)
	return nil
}

func updateAllProjects() error {
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
	for projectName := range projects {
		if err := updateSingleProject(projectName); err != nil {
			fmt.Printf("❌ Failed to update %s: %v\n", projectName, err)
			failed++
		} else {
			updated++
		}
	}

	fmt.Printf("\nUpdate Summary: %d updated, %d failed\n", updated, failed)
	if failed > 0 {
		return fmt.Errorf("failed to update %d project(s)", failed)
	}
	return nil
}

func init() {

}
