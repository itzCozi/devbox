package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devbox/internal/config"
	"devbox/internal/parallel"
)

type OptimizedSetup struct {
	dockerClient  DockerClientInterface
	configManager *config.ConfigManager
}

type DockerClientInterface interface {
	PullImage(image string) error
	CreateBoxWithConfig(name, image, workspaceHost, workspaceBox string, projectConfig interface{}) (string, error)
	StartBox(boxID string) error
	WaitForBox(boxName string, timeout time.Duration) error
	SetupDevboxInBoxWithUpdate(boxName, projectName string) error
	ExecuteSetupCommandsWithOutput(boxName string, commands []string, showOutput bool) error
	QueryPackagesParallel(boxName string) (aptList, pipList, npmList, yarnList, pnpmList []string)
}

func NewOptimizedSetup(dockerClient DockerClientInterface, configManager *config.ConfigManager) *OptimizedSetup {
	return &OptimizedSetup{
		dockerClient:  dockerClient,
		configManager: configManager,
	}
}

func (optSetup *OptimizedSetup) OptimizedSystemUpdate(boxName string) error {
	fmt.Printf("Performing optimized system update...\n")

	executor := parallel.NewSetupCommandExecutor(boxName, false, 2)

	groups := []parallel.CommandGroup{
		{
			Name: "System Update",
			Commands: []string{
				"apt update -y",
				"apt full-upgrade -y",
			},
			Parallel: false,
		},
		{
			Name: "System Optimization",
			Commands: []string{
				"apt autoremove -y",
				"apt autoclean",
			},
			Parallel: true,
		},
	}

	return executor.ExecuteCommandGroups(groups)
}

func (optSetup *OptimizedSetup) FastInit(projectName string, projectConfig *config.ProjectConfig, cfg *config.Config, workspacePath string, forceFlag bool) error {
	boxName := fmt.Sprintf("devbox_%s", projectName)
	baseImage := cfg.GetEffectiveBaseImage(&config.Project{
		Name:      projectName,
		BaseImage: "ubuntu:22.04",
	}, projectConfig)

	workspaceBox := "/workspace"
	if projectConfig != nil && projectConfig.WorkingDir != "" {
		workspaceBox = projectConfig.WorkingDir
	}

	fmt.Printf("Fast initialization of '%s'...\n", boxName)

	fmt.Printf("Pulling image '%s'...\n", baseImage)
	if err := optSetup.dockerClient.PullImage(baseImage); err != nil {
		return fmt.Errorf("failed to pull base image: %w", err)
	}

	if forceFlag {

		fmt.Printf("Force flag detected, recreating box...\n")
	}

	fmt.Printf("Creating box...\n")
	configMap := make(map[string]interface{})
	if projectConfig != nil {

	}

	boxID, err := optSetup.dockerClient.CreateBoxWithConfig(boxName, baseImage, workspacePath, workspaceBox, configMap)
	if err != nil {
		return fmt.Errorf("failed to create box: %w", err)
	}

	fmt.Printf("Starting box...\n")
	if err := optSetup.dockerClient.StartBox(boxID); err != nil {
		return fmt.Errorf("failed to start box: %w", err)
	}

	fmt.Printf("Waiting for box to be ready...\n")
	if err := optSetup.dockerClient.WaitForBox(boxName, 30*time.Second); err != nil {
		return fmt.Errorf("box failed to start: %w", err)
	}

	fmt.Printf("Running parallel setup operations...\n")

	setupTasks := []parallel.Task{

		func() error {
			return optSetup.OptimizedSystemUpdate(boxName)
		},

		func() error {
			fmt.Printf("Setting up devbox commands...\n")
			return optSetup.dockerClient.SetupDevboxInBoxWithUpdate(boxName, projectName)
		},
	}

	workerPool := parallel.NewWorkerPool(2, 10*time.Minute)
	results := workerPool.Execute(setupTasks)

	for i, err := range results {
		if err != nil {
			return fmt.Errorf("parallel setup task %d failed: %w", i+1, err)
		}
	}

	if projectConfig != nil && len(projectConfig.SetupCommands) > 0 {
		fmt.Printf("Installing packages (%d commands)...\n", len(projectConfig.SetupCommands))
		if err := optSetup.dockerClient.ExecuteSetupCommandsWithOutput(boxName, projectConfig.SetupCommands, false); err != nil {
			return fmt.Errorf("failed to execute setup commands: %w", err)
		}

		_ = WriteLockFileForBox(boxName, projectName, workspacePath, baseImage, "")
	}

	return nil
}

func (optSetup *OptimizedSetup) FastUp(projectConfig *config.ProjectConfig, projectName, boxName, baseImage, cwd, workspaceBox string) error {
	fmt.Printf("Fast startup of environment...\n")

	configMap := make(map[string]interface{})
	if projectConfig != nil {

	}

	fmt.Printf("Creating optimized box...\n")
	boxID, err := optSetup.dockerClient.CreateBoxWithConfig(boxName, baseImage, cwd, workspaceBox, configMap)
	if err != nil {
		return fmt.Errorf("failed to create box: %w", err)
	}

	if err := optSetup.dockerClient.StartBox(boxID); err != nil {
		return fmt.Errorf("failed to start box: %w", err)
	}

	fmt.Printf("Waiting for box startup...\n")
	if err := optSetup.dockerClient.WaitForBox(boxName, 30*time.Second); err != nil {
		return fmt.Errorf("box failed to start: %w", err)
	}

	fmt.Printf("Running parallel initialization...\n")

	setupTasks := []parallel.Task{

		func() error {
			return optSetup.dockerClient.SetupDevboxInBoxWithUpdate(boxName, projectName)
		},

		func() error {
			return optSetup.OptimizedSystemUpdate(boxName)
		},
	}

	workerPool := parallel.NewWorkerPool(2, 10*time.Minute)
	results := workerPool.Execute(setupTasks)

	for i, err := range results {
		if err != nil {
			return fmt.Errorf("parallel setup task %d failed: %w", i+1, err)
		}
	}

	lockfilePath := filepath.Join(cwd, "devbox.lock")
	if _, err := os.Stat(lockfilePath); err == nil {
		fmt.Printf("Processing lock file...\n")
		if err := optSetup.processLockFile(boxName, lockfilePath); err != nil {
			return fmt.Errorf("failed to process lock file: %w", err)
		}
	}

	if projectConfig != nil && len(projectConfig.SetupCommands) > 0 {
		fmt.Printf("Installing packages (%d commands)...\n", len(projectConfig.SetupCommands))
		if err := optSetup.dockerClient.ExecuteSetupCommandsWithOutput(boxName, projectConfig.SetupCommands, false); err != nil {
			return fmt.Errorf("failed to execute setup commands: %w", err)
		}

		_ = WriteLockFileForBox(boxName, projectName, cwd, baseImage, "")
	}

	return nil
}

func (optSetup *OptimizedSetup) processLockFile(boxName, lockfilePath string) error {
	data, err := os.ReadFile(lockfilePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var cmds []string
	for _, line := range lines {
		cmd := strings.TrimSpace(line)
		if cmd == "" || strings.HasPrefix(cmd, "#") {
			continue
		}
		cmds = append(cmds, cmd)
	}

	if len(cmds) > 0 {
		fmt.Printf("Replaying %d commands from lock file...\n", len(cmds))
		return optSetup.dockerClient.ExecuteSetupCommandsWithOutput(boxName, cmds, false)
	}

	return nil
}

func (optSetup *OptimizedSetup) PrewarmImage(image string) error {
	fmt.Printf("Prewarming image %s...\n", image)
	return optSetup.dockerClient.PullImage(image)
}

func (optSetup *OptimizedSetup) OptimizeEnvironment(boxName string) error {
	fmt.Printf("Optimizing environment...\n")

	executor := parallel.NewSetupCommandExecutor(boxName, false, 3)

	optimizationGroups := []parallel.CommandGroup{
		{
			Name: "Package Manager Optimization",
			Commands: []string{
				"apt-get clean",
				"pip cache purge || true",
				"npm cache clean --force || true",
			},
			Parallel: true,
		},
		{
			Name: "System Optimization",
			Commands: []string{
				"updatedb || true",
				"ldconfig",
			},
			Parallel: true,
		},
	}

	return executor.ExecuteCommandGroups(optimizationGroups)
}
