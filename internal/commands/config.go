package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config <command>",
	Short: "Manage devbox configurations",
	Long: `Manage devbox configurations including project-specific settings and global options.

Available commands:
  generate <project>    Generate devbox.json for project
  validate <project>    Validate project configuration
  show <project>        Show project configuration
  templates             List available templates
  global               Show global configuration`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		subCommand := args[0]

		switch subCommand {
		case "generate":
			if len(args) < 2 {
				return fmt.Errorf("project name required for generate command")
			}
			return generateProjectConfig(args[1])
		case "validate":
			if len(args) < 2 {
				return fmt.Errorf("project name required for validate command")
			}
			return validateProjectConfig(args[1])
		case "show":
			if len(args) < 2 {
				return fmt.Errorf("project name required for show command")
			}
			return showProjectConfig(args[1])
		case "templates":
			return showTemplates()
		case "global":
			return showGlobalConfig()
		default:
			return fmt.Errorf("unknown config command: %s", subCommand)
		}
	},
}

func generateProjectConfig(projectName string) error {
	if err := validateProjectName(projectName); err != nil {
		return err
	}

	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	project, exists := cfg.GetProject(projectName)
	if !exists {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	configPath := filepath.Join(project.WorkspacePath, "devbox.json")
	if _, err := os.Stat(configPath); err == nil && !forceFlag {
		return fmt.Errorf("devbox.json already exists. Use --force to overwrite")
	}

	projectConfig := configManager.GetDefaultProjectConfig(projectName)
	projectConfig.BaseImage = project.BaseImage

	if err := configManager.SaveProjectConfig(project.WorkspacePath, projectConfig); err != nil {
		return fmt.Errorf("failed to save project configuration: %w", err)
	}

	fmt.Printf("✅ Generated devbox.json for project '%s'\n", projectName)
	fmt.Printf("📄 Configuration file: %s\n", configPath)
	fmt.Printf("\nEdit the file to customize your development environment.\n")
	fmt.Printf("Available templates: %s\n", strings.Join(configManager.GetAvailableTemplates(), ", "))

	return nil
}

func validateProjectConfig(projectName string) error {
	if err := validateProjectName(projectName); err != nil {
		return err
	}

	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	project, exists := cfg.GetProject(projectName)
	if !exists {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	projectConfig, err := configManager.LoadProjectConfig(project.WorkspacePath)
	if err != nil {
		return fmt.Errorf("failed to load project configuration: %w", err)
	}

	if projectConfig == nil {
		fmt.Printf("No devbox.json found for project '%s'\n", projectName)
		fmt.Printf("Generate one with: devbox config generate %s\n", projectName)
		return nil
	}

	if err := configManager.ValidateProjectConfig(projectConfig); err != nil {
		fmt.Printf("❌ Configuration validation failed:\n")
		fmt.Printf("   %s\n", err.Error())
		return err
	}

	fmt.Printf("✅ Configuration for project '%s' is valid\n", projectName)

	fmt.Printf("\nConfiguration summary:\n")
	fmt.Printf("  Name: %s\n", projectConfig.Name)
	fmt.Printf("  Base image: %s\n", projectConfig.BaseImage)

	if len(projectConfig.SetupCommands) > 0 {
		fmt.Printf("  Setup commands: %d\n", len(projectConfig.SetupCommands))
	}

	if len(projectConfig.Environment) > 0 {
		fmt.Printf("  Environment variables: %d\n", len(projectConfig.Environment))
	}

	if len(projectConfig.Ports) > 0 {
		fmt.Printf("  Port mappings: %v\n", projectConfig.Ports)
	}

	if len(projectConfig.Volumes) > 0 {
		fmt.Printf("  Volume mappings: %v\n", projectConfig.Volumes)
	}

	return nil
}

func showProjectConfig(projectName string) error {
	if err := validateProjectName(projectName); err != nil {
		return err
	}

	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	project, exists := cfg.GetProject(projectName)
	if !exists {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	projectConfig, err := configManager.LoadProjectConfig(project.WorkspacePath)
	if err != nil {
		return fmt.Errorf("failed to load project configuration: %w", err)
	}

	fmt.Printf("Configuration for project '%s':\n\n", projectName)

	fmt.Printf("Global project settings:\n")
	fmt.Printf("  Name: %s\n", project.Name)
	fmt.Printf("  Box name: %s\n", project.BoxName)
	fmt.Printf("  Base image: %s\n", project.BaseImage)
	fmt.Printf("  Workspace: %s\n", project.WorkspacePath)
	fmt.Printf("  Status: %s\n", project.Status)

	if projectConfig == nil {
		fmt.Printf("\nNo devbox.json configuration file found.\n")
		fmt.Printf("Generate one with: devbox config generate %s\n", projectName)
		return nil
	}

	fmt.Printf("\nProject configuration (devbox.json):\n")

	if projectConfig.BaseImage != "" && projectConfig.BaseImage != project.BaseImage {
		fmt.Printf("  Base image override: %s\n", projectConfig.BaseImage)
	}

	if len(projectConfig.SetupCommands) > 0 {
		fmt.Printf("  Setup commands:\n")
		for i, cmd := range projectConfig.SetupCommands {
			fmt.Printf("    %d. %s\n", i+1, cmd)
		}
	}

	if len(projectConfig.Environment) > 0 {
		fmt.Printf("  Environment variables:\n")
		for key, value := range projectConfig.Environment {
			fmt.Printf("    %s=%s\n", key, value)
		}
	}

	if len(projectConfig.Ports) > 0 {
		fmt.Printf("  Port mappings:\n")
		for _, port := range projectConfig.Ports {
			fmt.Printf("    %s\n", port)
		}
	}

	if len(projectConfig.Volumes) > 0 {
		fmt.Printf("  Volume mappings:\n")
		for _, volume := range projectConfig.Volumes {
			fmt.Printf("    %s\n", volume)
		}
	}

	if projectConfig.WorkingDir != "" {
		fmt.Printf("  Working directory: %s\n", projectConfig.WorkingDir)
	}

	if projectConfig.Shell != "" {
		fmt.Printf("  Shell: %s\n", projectConfig.Shell)
	}

	if projectConfig.User != "" {
		fmt.Printf("  User: %s\n", projectConfig.User)
	}

	if len(projectConfig.Capabilities) > 0 {
		fmt.Printf("  Capabilities: %v\n", projectConfig.Capabilities)
	}

	if len(projectConfig.Labels) > 0 {
		fmt.Printf("  Labels:\n")
		for key, value := range projectConfig.Labels {
			fmt.Printf("    %s=%s\n", key, value)
		}
	}

	if projectConfig.Network != "" {
		fmt.Printf("  Network: %s\n", projectConfig.Network)
	}

	if projectConfig.Resources != nil {
		fmt.Printf("  Resource constraints:\n")
		if projectConfig.Resources.CPUs != "" {
			fmt.Printf("    CPUs: %s\n", projectConfig.Resources.CPUs)
		}
		if projectConfig.Resources.Memory != "" {
			fmt.Printf("    Memory: %s\n", projectConfig.Resources.Memory)
		}
	}

	if projectConfig.HealthCheck != nil {
		fmt.Printf("  Health check:\n")
		if len(projectConfig.HealthCheck.Test) > 0 {
			fmt.Printf("    Test: %v\n", projectConfig.HealthCheck.Test)
		}
		if projectConfig.HealthCheck.Interval != "" {
			fmt.Printf("    Interval: %s\n", projectConfig.HealthCheck.Interval)
		}
		if projectConfig.HealthCheck.Timeout != "" {
			fmt.Printf("    Timeout: %s\n", projectConfig.HealthCheck.Timeout)
		}
		if projectConfig.HealthCheck.Retries > 0 {
			fmt.Printf("    Retries: %d\n", projectConfig.HealthCheck.Retries)
		}
	}

	return nil
}

func showTemplates() error {
	templates := configManager.GetAvailableTemplates()

	fmt.Printf("Available configuration templates:\n\n")

	for _, templateName := range templates {
		templateConfig, err := configManager.CreateProjectConfigFromTemplate(templateName, "example")
		if err != nil {
			fmt.Printf("  %s: Error loading template\n", templateName)
			continue
		}

		fmt.Printf("  %s:\n", templateName)
		fmt.Printf("    Base image: %s\n", templateConfig.BaseImage)

		if len(templateConfig.SetupCommands) > 0 {
			fmt.Printf("    Setup commands: %d steps\n", len(templateConfig.SetupCommands))
		}

		if len(templateConfig.Environment) > 0 {
			fmt.Printf("    Environment: %d variables\n", len(templateConfig.Environment))
		}

		if len(templateConfig.Ports) > 0 {
			fmt.Printf("    Ports: %v\n", templateConfig.Ports)
		}

		fmt.Printf("\n")
	}

	fmt.Printf("Usage:\n")
	fmt.Printf("  devbox init myproject --template python\n")
	fmt.Printf("  devbox init myproject --template nodejs\n")

	return nil
}

func showGlobalConfig() error {
	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Printf("Global devbox configuration:\n\n")

	if cfg.Settings != nil {
		fmt.Printf("Settings:\n")
		fmt.Printf("  Default base image: %s\n", cfg.Settings.DefaultBaseImage)
		fmt.Printf("  Auto update: %t\n", cfg.Settings.AutoUpdate)
		fmt.Printf("  Auto stop on exit: %t\n", cfg.Settings.AutoStopOnExit)

		if cfg.Settings.ConfigTemplatesPath != "" {
			fmt.Printf("  Templates path: %s\n", cfg.Settings.ConfigTemplatesPath)
		}

		if len(cfg.Settings.DefaultEnvironment) > 0 {
			fmt.Printf("  Default environment:\n")
			for key, value := range cfg.Settings.DefaultEnvironment {
				fmt.Printf("    %s=%s\n", key, value)
			}
		}
	}

	fmt.Printf("\nProjects: %d total\n", len(cfg.Projects))

	for name, project := range cfg.Projects {
		fmt.Printf("  %s (%s)\n", name, project.Status)
	}

	return nil
}

func init() {
	configCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force operation, overwriting existing files")
}
