package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"devbox/internal/config"
)

var (
	templateFlag   string
	generateConfig bool
	configOnlyFlag bool
)

var initCmd = &cobra.Command{
	Use:   "init <project>",
	Short: "Initialize a new devbox project",
	Long: `Create a new devbox project with its own Docker box.
This will create a project directory and a corresponding Docker box.

Examples:
  devbox init myproject                    # Basic project
  devbox init myproject --template python # Python development project  
  devbox init myproject --config-only     # Generate devbox.json only
  devbox init myproject --generate-config # Create box and generate devbox.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		if err := validateProjectName(projectName); err != nil {
			return err
		}

		cfg, err := configManager.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		if _, exists := cfg.GetProject(projectName); exists && !forceFlag {
			return fmt.Errorf("project '%s' already exists. Use --force to overwrite", projectName)
		}

		workspacePath, err := getWorkspacePath(projectName)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			return fmt.Errorf("failed to create workspace directory: %w", err)
		}

		fmt.Printf("Created workspace directory: %s\n", workspacePath)

		var projectConfig *config.ProjectConfig

		if existingConfig, err := configManager.LoadProjectConfig(workspacePath); err == nil && existingConfig != nil {
			fmt.Printf("Found existing devbox.json configuration\n")
			projectConfig = existingConfig
		} else if templateFlag != "" {

			fmt.Printf("Creating project from template: %s\n", templateFlag)
			projectConfig, err = configManager.CreateProjectConfigFromTemplate(templateFlag, projectName)
			if err != nil {
				return fmt.Errorf("failed to create project from template: %w", err)
			}
		} else if generateConfig {

			projectConfig = configManager.GetDefaultProjectConfig(projectName)
		}

		if projectConfig != nil && (generateConfig || templateFlag != "") {
			if err := configManager.SaveProjectConfig(workspacePath, projectConfig); err != nil {
				return fmt.Errorf("failed to save project configuration: %w", err)
			}
			fmt.Printf("Generated devbox.json configuration file\n")
		}

		if configOnlyFlag {
			fmt.Printf("✅ Configuration file generated for project '%s'\n", projectName)
			fmt.Printf("📁 Workspace: %s\n", workspacePath)
			fmt.Printf("📄 Config: %s/devbox.json\n", workspacePath)
			return nil
		}

		if projectConfig != nil {
			if err := configManager.ValidateProjectConfig(projectConfig); err != nil {
				return fmt.Errorf("invalid project configuration: %w", err)
			}
		}

		boxName := fmt.Sprintf("devbox_%s", projectName)

		baseImage := cfg.GetEffectiveBaseImage(&config.Project{
			Name:      projectName,
			BaseImage: "ubuntu:22.04",
		}, projectConfig)

		workspaceBox := "/workspace"
		if projectConfig != nil && projectConfig.WorkingDir != "" {
			workspaceBox = projectConfig.WorkingDir
		}

		fmt.Printf("Setting up box '%s' with image '%s'...\n", boxName, baseImage)
		if err := dockerClient.PullImage(baseImage); err != nil {
			return fmt.Errorf("failed to pull base image: %w", err)
		}

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

		var configMap map[string]interface{}
		if projectConfig != nil {
			configData, _ := json.Marshal(projectConfig)
			json.Unmarshal(configData, &configMap)
		}

		boxID, err := dockerClient.CreateBoxWithConfig(boxName, baseImage, workspacePath, workspaceBox, configMap)
		if err != nil {
			return fmt.Errorf("failed to create box: %w", err)
		}

		if err := dockerClient.StartBox(boxID); err != nil {
			return fmt.Errorf("failed to start box: %w", err)
		}

		fmt.Printf("Starting box...\n")
		if err := dockerClient.WaitForBox(boxName, 30*time.Second); err != nil {
			return fmt.Errorf("box failed to start: %w", err)
		}

		fmt.Printf("Updating system packages...\n")
		systemUpdateCommands := []string{
			"apt update -y",
			"apt full-upgrade -y",
		}
		if err := dockerClient.ExecuteSetupCommands(boxName, systemUpdateCommands); err != nil {
			return fmt.Errorf("failed to update system packages: %w", err)
		}

		if projectConfig != nil && len(projectConfig.SetupCommands) > 0 {
			if err := dockerClient.ExecuteSetupCommands(boxName, projectConfig.SetupCommands); err != nil {
				return fmt.Errorf("failed to execute setup commands: %w", err)
			}
		}

		fmt.Printf("Setting up devbox commands in box...\n")
		if err := dockerClient.SetupDevboxInBoxWithUpdate(boxName, projectName); err != nil {
			return fmt.Errorf("failed to setup devbox in box: %w", err)
		}

		project := &config.Project{
			Name:          projectName,
			BoxName:       boxName,
			BaseImage:     baseImage,
			WorkspacePath: workspacePath,
			Status:        "running",
		}

		cfg.MergeProjectConfig(project, projectConfig)

		cfg.AddProject(project)
		if err := configManager.Save(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("✅ Project '%s' initialized successfully!\n", projectName)
		fmt.Printf("📁 Workspace: %s\n", workspacePath)
		fmt.Printf("🐳 Box: %s\n", boxName)
		fmt.Printf("🖼️  Image: %s\n", baseImage)

		if projectConfig != nil {
			fmt.Printf("⚙️  Configuration: devbox.json\n")
			if len(projectConfig.SetupCommands) > 0 {
				fmt.Printf("🔧 Setup commands: %d executed\n", len(projectConfig.SetupCommands))
			}
			if len(projectConfig.Ports) > 0 {
				fmt.Printf("🌐 Ports: %v\n", projectConfig.Ports)
			}
		}

		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  devbox shell %s       # Open interactive shell\n", projectName)
		fmt.Printf("  devbox run %s <cmd>   # Run a command\n", projectName)
		if projectConfig == nil && !generateConfig {
			fmt.Printf("  devbox config %s      # Generate devbox.json config\n", projectName)
		}

		return nil
	},
}

func init() {
	initCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force initialization, overwriting existing project")
	initCmd.Flags().StringVarP(&templateFlag, "template", "t", "", "Initialize from template (python, nodejs, go, web)")
	initCmd.Flags().BoolVarP(&generateConfig, "generate-config", "g", false, "Generate devbox.json configuration file")
	initCmd.Flags().BoolVarP(&configOnlyFlag, "config-only", "c", false, "Generate configuration file only (don't create box)")
}
