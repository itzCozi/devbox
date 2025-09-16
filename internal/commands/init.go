package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"devbox/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init <project>",
	Short: "Initialize a new devbox project",
	Long: `Create a new devbox project with its own Docker container.
This will create a project directory and a corresponding Docker container.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		// Validate project name
		if err := validateProjectName(projectName); err != nil {
			return err
		}

		// Load configuration
		cfg, err := configManager.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Check if project already exists
		if _, exists := cfg.GetProject(projectName); exists && !forceFlag {
			return fmt.Errorf("project '%s' already exists. Use --force to overwrite", projectName)
		}

		// Get workspace path
		workspacePath, err := getWorkspacePath(projectName)
		if err != nil {
			return err
		}

		// Create workspace directory
		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			return fmt.Errorf("failed to create workspace directory: %w", err)
		}

		fmt.Printf("Created workspace directory: %s\n", workspacePath)

		// Define container settings
		containerName := fmt.Sprintf("devbox_%s", projectName)
		baseImage := "ubuntu:22.04" // Default to Ubuntu 22.04
		workspaceContainer := "/workspace"

		// Pull the base image
		fmt.Printf("Setting up container '%s'...\n", containerName)
		if err := dockerClient.PullImage(baseImage); err != nil {
			return fmt.Errorf("failed to pull base image: %w", err)
		}

		// Remove existing container if force flag is set
		if forceFlag {
			exists, err := dockerClient.ContainerExists(containerName)
			if err != nil {
				return fmt.Errorf("failed to check container existence: %w", err)
			}
			if exists {
				fmt.Printf("Removing existing container '%s'...\n", containerName)
				dockerClient.StopContainer(containerName)
				if err := dockerClient.RemoveContainer(containerName); err != nil {
					return fmt.Errorf("failed to remove existing container: %w", err)
				}
			}
		}

		// Create container
		containerID, err := dockerClient.CreateContainer(containerName, baseImage, workspacePath, workspaceContainer)
		if err != nil {
			return fmt.Errorf("failed to create container: %w", err)
		}

		// Start container
		if err := dockerClient.StartContainer(containerID); err != nil {
			return fmt.Errorf("failed to start container: %w", err)
		}

		// Wait for container to be ready
		fmt.Printf("Starting container...\n")
		if err := dockerClient.WaitForContainer(containerName, 30*time.Second); err != nil {
			return fmt.Errorf("container failed to start: %w", err)
		}

		// Setup devbox commands inside the container
		fmt.Printf("Setting up devbox commands in container...\n")
		if err := dockerClient.SetupDevboxInContainer(containerName, projectName); err != nil {
			return fmt.Errorf("failed to setup devbox in container: %w", err)
		}

		// Create project configuration
		project := &config.Project{
			Name:          projectName,
			ContainerName: containerName,
			BaseImage:     baseImage,
			WorkspacePath: workspacePath,
			Status:        "running",
		}

		// Add project to configuration and save
		cfg.AddProject(project)
		if err := configManager.Save(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("✅ Project '%s' initialized successfully!\n", projectName)
		fmt.Printf("📁 Workspace: %s\n", workspacePath)
		fmt.Printf("🐳 Container: %s\n", containerName)
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  devbox shell %s    # Open interactive shell\n", projectName)
		fmt.Printf("  devbox run %s <cmd> # Run a command\n", projectName)

		return nil
	},
}

func init() {
	initCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force initialization, overwriting existing project")
}
