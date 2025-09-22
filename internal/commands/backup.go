package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"devbox/internal/config"
)

type backupManifest struct {
	Version      int                   `json:"version"`
	Project      string                `json:"project"`
	BoxName      string                `json:"box_name"`
	CreatedAt    string                `json:"created_at"`
	ImageTag     string                `json:"image_tag"`
	DevboxConfig *config.ProjectConfig `json:"devbox_config,omitempty"`
	LockFileJSON json.RawMessage       `json:"lock_file_json,omitempty"`
}

var (
	backupOutput string
)

var backupCmd = &cobra.Command{
	Use:   "backup <project>",
	Short: "Backup the project's devbox environment (container state + config)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		cfg, err := configManager.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
		proj, ok := cfg.GetProject(projectName)
		if !ok {
			return fmt.Errorf("project '%s' not found", projectName)
		}

		exists, err := dockerClient.BoxExists(proj.BoxName)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("box '%s' does not exist", proj.BoxName)
		}

		ts := time.Now().UTC().Format("20060102-150405")
		defaultDir := filepath.Join(proj.WorkspacePath, ".devbox_backups", ts)
		outDir := backupOutput
		if strings.TrimSpace(outDir) == "" {
			outDir = defaultDir
		}
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create backup directory: %w", err)
		}

		imageTag := fmt.Sprintf("devbox/%s:backup-%s", projectName, ts)
		fmt.Printf("Creating image from box '%s'...\n", proj.BoxName)
		imgID, err := dockerClient.CommitContainer(proj.BoxName, imageTag)
		if err != nil {
			return fmt.Errorf("failed to commit container: %w", err)
		}
		_ = imgID

		imageTar := filepath.Join(outDir, "image.tar")
		fmt.Printf("Saving image '%s' to %s...\n", imageTag, imageTar)
		if err := dockerClient.SaveImage(imageTag, imageTar); err != nil {
			return fmt.Errorf("failed to save image: %w", err)
		}

		var pcfg *config.ProjectConfig
		if c, err := configManager.LoadProjectConfig(proj.WorkspacePath); err == nil {
			pcfg = c
		}
		var lockRaw json.RawMessage
		if b, err := os.ReadFile(filepath.Join(proj.WorkspacePath, "devbox.lock.json")); err == nil {
			lockRaw = json.RawMessage(b)
		}

		manifest := backupManifest{
			Version:      1,
			Project:      proj.Name,
			BoxName:      proj.BoxName,
			CreatedAt:    time.Now().UTC().Format(time.RFC3339),
			ImageTag:     imageTag,
			DevboxConfig: pcfg,
			LockFileJSON: lockRaw,
		}
		manPath := filepath.Join(outDir, "metadata.json")
		b, _ := json.MarshalIndent(manifest, "", "  ")
		if err := os.WriteFile(manPath, b, 0644); err != nil {
			return fmt.Errorf("failed to write metadata: %w", err)
		}

		fmt.Printf("✅ Backup complete\n")
		fmt.Printf("📦 Directory: %s\n", outDir)
		fmt.Printf("🖼️  Image tag: %s\n", imageTag)
		fmt.Printf("📄 Files: image.tar, metadata.json\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().StringVarP(&backupOutput, "output", "o", "", "Output directory for backup (default: <workspace>/.devbox_backups/<timestamp>)")
}
