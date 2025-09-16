package commands

import (
	"fmt"
	"os"
	"regexp"
	"runtime"

	"github.com/spf13/cobra"

	"devbox/internal/config"
	"devbox/internal/docker"
)

var (
	configManager *config.ConfigManager
	dockerClient  *docker.Client
	forceFlag     bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "devbox",
	Short: "Isolated development environments using Docker containers",
	Long: `devbox isolates development environments so that when you install system 
packages with apt they live only inside the project container and don't affect 
your host system. Each project has its own Docker container, while your code 
stays in a flat folder on the host.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Check if running on Debian/Ubuntu
		if runtime.GOOS != "linux" {
			return fmt.Errorf("devbox only runs on Debian/Ubuntu Linux")
		}

		// Initialize config manager
		var err error
		configManager, err = config.NewConfigManager()
		if err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}

		// Check Docker availability
		if err := docker.IsDockerAvailable(); err != nil {
			return err
		}

		// Initialize Docker client
		dockerClient, err = docker.NewClient()
		if err != nil {
			return fmt.Errorf("failed to initialize Docker client: %w", err)
		}

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if dockerClient != nil {
			dockerClient.Close()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(shellCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(destroyCmd)
	rootCmd.AddCommand(listCmd)

	// Global flags
	destroyCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force operation without confirmation")
}

// validateProjectName validates that a project name is safe to use
func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// Allow alphanumeric, hyphens, and underscores
	matched, err := regexp.MatchString("^[a-zA-Z0-9_-]+$", name)
	if err != nil {
		return fmt.Errorf("error validating project name: %w", err)
	}

	if !matched {
		return fmt.Errorf("project name can only contain alphanumeric characters, hyphens, and underscores")
	}

	return nil
}

// getWorkspacePath returns the workspace path for a project
func getWorkspacePath(projectName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return fmt.Sprintf("%s/devbox/%s", homeDir, projectName), nil
}