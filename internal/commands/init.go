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
	Long: `Create a new devbox project with its own Docker box.
This will create a project directory and a corresponding Docker box.`,
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

		// Define box settings
		boxName := fmt.Sprintf("devbox_%s", projectName)
		baseImage := "ubuntu:22.04" // Default to Ubuntu 22.04
		workspaceBox := "/workspace"

		// Pull the base image
		fmt.Printf("Setting up box '%s'...\n", boxName)
		if err := dockerClient.PullImage(baseImage); err != nil {
			return fmt.Errorf("failed to pull base image: %w", err)
		}

		// Remove existing box if force flag is set
		if forceFlag {
			exists, err := dockerClient.BoxExists(boxName)
			if err != nil {
				return fmt.Errorf("failed to check box existence: %w", err)
			}
			if exists {
				fmt.Printf("Removing existing box '%s'...\n", boxName)
				dockerClient.StopBox(boxName)
				if err := dockerClient.RemoveBox(boxName); err != nil {
					return fmt.Errorf("failed to remove existing box: %w", err)
				}
			}
		}

		// Create box
		boxID, err := dockerClient.CreateBox(boxName, baseImage, workspacePath, workspaceBox)
		if err != nil {
			return fmt.Errorf("failed to create box: %w", err)
		}

		// Start box
		if err := dockerClient.StartBox(boxID); err != nil {
			return fmt.Errorf("failed to start box: %w", err)
		}

		// Wait for box to be ready
		fmt.Printf("Starting box...\n")
		if err := dockerClient.WaitForBox(boxName, 30*time.Second); err != nil {
			return fmt.Errorf("box failed to start: %w", err)
		}

		// Setup devbox commands inside the box
		fmt.Printf("Setting up devbox commands in box...\n")
		if err := dockerClient.SetupDevboxInBox(boxName, projectName); err != nil {
			return fmt.Errorf("failed to setup devbox in box: %w", err)
		}

		// Create project configuration
		project := &config.Project{
			Name:          projectName,
			BoxName:       boxName,
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
		fmt.Printf("🐳 Box: %s\n", boxName)
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  devbox shell %s    # Open interactive shell\n", projectName)
		fmt.Printf("  devbox run %s <cmd> # Run a command\n", projectName)

		return nil
	},
}

func init() {
	initCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force initialization, overwriting existing project")
}
