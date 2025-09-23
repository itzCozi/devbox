package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type applyLockFile struct {
	Version    int            `json:"version"`
	Project    string         `json:"project"`
	BoxName    string         `json:"box_name"`
	Packages   lockPackages   `json:"packages"`
	Registries lockRegistries `json:"registries"`
	AptSources lockAptSources `json:"apt_sources"`
}

var applyCmd = &cobra.Command{
	Use:   "apply <project>",
	Short: "Apply devbox.lock.json: set registries and apt sources, then reconcile packages",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		cfg, err := configManager.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		proj, ok := cfg.GetProject(projectName)
		if !ok {
			return fmt.Errorf("project '%s' not found", projectName)
		}

		lockPath := filepath.Join(proj.WorkspacePath, "devbox.lock.json")
		data, err := os.ReadFile(lockPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", lockPath, err)
		}

		var lf applyLockFile
		if err := json.Unmarshal(data, &lf); err != nil {
			return fmt.Errorf("invalid lockfile: %w", err)
		}

		exists, err := dockerClient.BoxExists(proj.BoxName)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("box '%s' not found; run 'devbox up %s' first", proj.BoxName, projectName)
		}
		status, err := dockerClient.GetBoxStatus(proj.BoxName)
		if err != nil {
			return err
		}
		if status != "running" {
			if err := dockerClient.StartBox(proj.BoxName); err != nil {
				return fmt.Errorf("failed to start box: %w", err)
			}
		}

		var applyCmds []string

		if len(lf.AptSources.SourcesLists) > 0 {

			heredoc := "cat > /etc/apt/sources.list <<'EOF'\n" + strings.Join(lf.AptSources.SourcesLists, "\n") + "\nEOF"
			applyCmds = append(applyCmds,
				"cp /etc/apt/sources.list /etc/apt/sources.list.bak 2>/dev/null || true",
				"rm -f /etc/apt/sources.list.d/*.list 2>/dev/null || true",
				heredoc,
			)
		}
		if lf.AptSources.PinnedRelease != "" {
			applyCmds = append(applyCmds, fmt.Sprintf("bash -lc 'echo APT::Default-Release \"%s\"; > /etc/apt/apt.conf.d/99defaultrelease'", escapeBash(lf.AptSources.PinnedRelease)))
		}
		if len(lf.AptSources.SourcesLists) > 0 {
			applyCmds = append(applyCmds, "apt update -y")
		}

		if lf.Registries.PipIndexURL != "" || len(lf.Registries.PipExtraIndex) > 0 {
			var b strings.Builder
			b.WriteString("cat > /etc/pip.conf <<'EOF'\n[global]\n")
			if lf.Registries.PipIndexURL != "" {
				b.WriteString("index-url = ")
				b.WriteString(lf.Registries.PipIndexURL)
				b.WriteString("\n")
			}
			for _, u := range lf.Registries.PipExtraIndex {
				if strings.TrimSpace(u) == "" {
					continue
				}
				b.WriteString("extra-index-url = ")
				b.WriteString(u)
				b.WriteString("\n")
			}
			b.WriteString("EOF")
			applyCmds = append(applyCmds, b.String())
		}

		if lf.Registries.NpmRegistry != "" {
			applyCmds = append(applyCmds, fmt.Sprintf("npm config set registry %s -g", lf.Registries.NpmRegistry))
		}
		if lf.Registries.YarnRegistry != "" {
			applyCmds = append(applyCmds, fmt.Sprintf("yarn config set npmRegistryServer %s -g", lf.Registries.YarnRegistry))
		}
		if lf.Registries.PnpmRegistry != "" {
			applyCmds = append(applyCmds, fmt.Sprintf("pnpm config set registry %s -g", lf.Registries.PnpmRegistry))
		}

		if err := dockerClient.ExecuteSetupCommandsWithOutput(proj.BoxName, applyCmds, false); err != nil {
			return fmt.Errorf("failed applying registries/sources: %w", err)
		}

		curApt, curPip, curNpm, curYarn, curPnpm := dockerClient.QueryPackagesParallel(proj.BoxName)

		actions := buildReconcileActions(lf.Packages, curApt, curPip, curNpm, curYarn, curPnpm)
		if len(actions) > 0 {
			if err := dockerClient.ExecuteSetupCommandsWithOutput(proj.BoxName, actions, true); err != nil {
				return fmt.Errorf("failed to reconcile packages: %w", err)
			}
		}

		fmt.Println("✅ Applied lockfile: registries/sources configured and packages reconciled")
		return nil
	},
}

func escapeBash(s string) string {
	return strings.ReplaceAll(s, "'", "'\\''")
}

func parseMap(list []string, sep string) map[string]string {
	m := map[string]string{}
	for _, line := range list {
		s := strings.TrimSpace(line)
		if s == "" {
			continue
		}
		if sep == "==" {
			if i := strings.Index(s, "=="); i != -1 {
				name := strings.ToLower(strings.TrimSpace(s[:i]))
				ver := strings.TrimSpace(s[i+2:])
				m[name] = ver
			}
			continue
		}
		if sep == "@" {

			idx := strings.LastIndex(s, "@")
			if idx > 0 {
				name := strings.ToLower(strings.TrimSpace(s[:idx]))
				ver := strings.TrimSpace(s[idx+1:])
				m[name] = ver
			}
			continue
		}
		if sep == "=" {
			if i := strings.Index(s, "="); i != -1 {
				name := strings.ToLower(strings.TrimSpace(s[:i]))
				ver := strings.TrimSpace(s[i+1:])
				m[name] = ver
			}
		}
	}
	return m
}

func keysNotIn(a, b map[string]string) []string {
	var out []string
	for k := range a {
		if _, ok := b[k]; !ok {
			out = append(out, k)
		}
	}
	return out
}

func buildReconcileActions(lockPkgs lockPackages, curApt, curPip, curNpm, curYarn, curPnpm []string) []string {
	var cmds []string

	lockA := parseMap(lockPkgs.Apt, "=")
	curA := parseMap(curApt, "=")
	lockP := parseMap(lockPkgs.Pip, "==")
	curP := parseMap(curPip, "==")
	lockN := parseMap(lockPkgs.Npm, "@")
	curN := parseMap(curNpm, "@")
	lockY := parseMap(lockPkgs.Yarn, "@")
	curY := parseMap(curYarn, "@")
	lockQ := parseMap(lockPkgs.Pnpm, "@")
	curQ := parseMap(curPnpm, "@")

	var aptInstall []string
	for name, ver := range lockA {
		if curVer, ok := curA[name]; !ok || curVer != ver {
			aptInstall = append(aptInstall, fmt.Sprintf("%s=%s", name, ver))
		}
	}
	if len(aptInstall) > 0 {
		cmds = append(cmds, "apt update -y", "DEBIAN_FRONTEND=noninteractive apt-get install -y "+strings.Join(aptInstall, " "))
	}

	for _, extra := range keysNotIn(curA, lockA) {
		cmds = append(cmds, fmt.Sprintf("apt-get remove -y %s", extra))
	}
	if len(keysNotIn(curA, lockA)) > 0 {
		cmds = append(cmds, "apt-get autoremove -y")
	}

	for name, ver := range lockP {
		if curVer, ok := curP[name]; !ok || curVer != ver {
			cmds = append(cmds, fmt.Sprintf("python3 -m pip install %s==%s", name, ver))
		}
	}
	for _, extra := range keysNotIn(curP, lockP) {
		cmds = append(cmds, fmt.Sprintf("python3 -m pip uninstall -y %s", extra))
	}

	for name, ver := range lockN {
		if curVer, ok := curN[name]; !ok || curVer != ver {
			cmds = append(cmds, fmt.Sprintf("npm i -g %s@%s", name, ver))
		}
	}
	for _, extra := range keysNotIn(curN, lockN) {
		cmds = append(cmds, fmt.Sprintf("npm rm -g %s", extra))
	}

	for name, ver := range lockY {
		if curVer, ok := curY[name]; !ok || curVer != ver {
			cmds = append(cmds, fmt.Sprintf("yarn global add %s@%s", name, ver))
		}
	}
	for _, extra := range keysNotIn(curY, lockY) {
		cmds = append(cmds, fmt.Sprintf("yarn global remove %s", extra))
	}

	for name, ver := range lockQ {
		if curVer, ok := curQ[name]; !ok || curVer != ver {
			cmds = append(cmds, fmt.Sprintf("pnpm add -g %s@%s", name, ver))
		}
	}
	for _, extra := range keysNotIn(curQ, lockQ) {
		cmds = append(cmds, fmt.Sprintf("pnpm remove -g %s", extra))
	}

	return cmds
}

func init() {
	rootCmd.AddCommand(applyCmd)
}
