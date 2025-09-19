package parallel

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type SetupCommandExecutor struct {
	boxName    string
	workerPool *WorkerPool
	showOutput bool
}

func NewSetupCommandExecutor(boxName string, showOutput bool, maxWorkers int) *SetupCommandExecutor {
	if maxWorkers <= 0 {
		maxWorkers = 3
	}

	return &SetupCommandExecutor{
		boxName:    boxName,
		workerPool: NewWorkerPool(maxWorkers, 10*time.Minute),
		showOutput: showOutput,
	}
}

type CommandGroup struct {
	Name     string
	Commands []string
	Parallel bool
}

func (sce *SetupCommandExecutor) ExecuteCommandGroups(groups []CommandGroup) error {
	if len(groups) == 0 {
		return nil
	}

	var parallelBatches []Batch
	var sequentialGroups []CommandGroup

	for _, group := range groups {
		if group.Parallel {

			tasks := make([]Task, len(group.Commands))
			for i, cmd := range group.Commands {
				tasks[i] = sce.createCommandTask(cmd, i+1, len(group.Commands), group.Name)
			}
			parallelBatches = append(parallelBatches, Batch{Name: group.Name, Tasks: tasks})
		} else {
			sequentialGroups = append(sequentialGroups, group)
		}
	}

	if len(parallelBatches) > 0 {
		if sce.showOutput {
			fmt.Printf("Executing %d parallel command groups...\n", len(parallelBatches))
		}

		batchResults := sce.workerPool.ExecuteBatches(parallelBatches)

		for batchName, results := range batchResults {
			for i, err := range results {
				if err != nil {
					return fmt.Errorf("parallel command group '%s', command %d failed: %w", batchName, i+1, err)
				}
			}
		}

		if sce.showOutput {
			fmt.Printf("All parallel command groups completed successfully!\n")
		}
	}

	for _, group := range sequentialGroups {
		if sce.showOutput {
			fmt.Printf("Executing sequential group: %s\n", group.Name)
		}

		for i, cmd := range group.Commands {
			if err := sce.executeCommand(cmd, i+1, len(group.Commands), group.Name); err != nil {
				return fmt.Errorf("sequential command group '%s', command %d failed: %w", group.Name, i+1, err)
			}
		}

		if sce.showOutput {
			fmt.Printf("Sequential group '%s' completed successfully!\n", group.Name)
		}
	}

	return nil
}

func (sce *SetupCommandExecutor) ExecuteParallel(commands []string) error {
	if len(commands) == 0 {
		return nil
	}

	groups := sce.categorizeCommands(commands)
	return sce.ExecuteCommandGroups(groups)
}

func (sce *SetupCommandExecutor) categorizeCommands(commands []string) []CommandGroup {
	var groups []CommandGroup

	var aptCommands []string
	var pipCommands []string
	var npmCommands []string
	var yarnCommands []string
	var pnpmCommands []string
	var systemCommands []string
	var otherCommands []string

	for _, cmd := range commands {
		cmdLower := strings.ToLower(strings.TrimSpace(cmd))

		switch {
		case strings.HasPrefix(cmdLower, "apt ") || strings.HasPrefix(cmdLower, "apt-get "):
			aptCommands = append(aptCommands, cmd)
		case strings.HasPrefix(cmdLower, "pip ") || strings.HasPrefix(cmdLower, "pip3 "):
			pipCommands = append(pipCommands, cmd)
		case strings.HasPrefix(cmdLower, "npm "):
			npmCommands = append(npmCommands, cmd)
		case strings.HasPrefix(cmdLower, "yarn "):
			yarnCommands = append(yarnCommands, cmd)
		case strings.HasPrefix(cmdLower, "pnpm "):
			pnpmCommands = append(pnpmCommands, cmd)
		case strings.HasPrefix(cmdLower, "systemctl ") || strings.HasPrefix(cmdLower, "service ") ||
			strings.HasPrefix(cmdLower, "update-alternatives ") || strings.HasPrefix(cmdLower, "adduser ") ||
			strings.HasPrefix(cmdLower, "usermod "):
			systemCommands = append(systemCommands, cmd)
		default:
			otherCommands = append(otherCommands, cmd)
		}
	}

	if len(systemCommands) > 0 {
		groups = append(groups, CommandGroup{Name: "System Commands", Commands: systemCommands, Parallel: false})
	}

	if len(aptCommands) > 0 {
		groups = append(groups, CommandGroup{Name: "APT Packages", Commands: aptCommands, Parallel: false})
	}

	var packageGroups []CommandGroup
	if len(pipCommands) > 0 {
		packageGroups = append(packageGroups, CommandGroup{Name: "Python Packages", Commands: pipCommands, Parallel: true})
	}
	if len(npmCommands) > 0 {
		packageGroups = append(packageGroups, CommandGroup{Name: "NPM Packages", Commands: npmCommands, Parallel: true})
	}
	if len(yarnCommands) > 0 {
		packageGroups = append(packageGroups, CommandGroup{Name: "Yarn Packages", Commands: yarnCommands, Parallel: true})
	}
	if len(pnpmCommands) > 0 {
		packageGroups = append(packageGroups, CommandGroup{Name: "PNPM Packages", Commands: pnpmCommands, Parallel: true})
	}

	groups = append(groups, packageGroups...)

	if len(otherCommands) > 0 {
		groups = append(groups, CommandGroup{Name: "Other Commands", Commands: otherCommands, Parallel: false})
	}

	return groups
}

func (sce *SetupCommandExecutor) createCommandTask(command string, step, total int, groupName string) Task {
	return func() error {
		return sce.executeCommand(command, step, total, groupName)
	}
}

func (sce *SetupCommandExecutor) executeCommand(command string, step, total int, groupName string) error {
	if sce.showOutput {
		fmt.Printf("[%s] Step %d/%d: %s\n", groupName, step, total, command)
	}

	wrapped := ". /root/.bashrc >/dev/null 2>&1 || true; " + command
	cmd := exec.Command("docker", "exec", sce.boxName, "bash", "-c", wrapped)

	if sce.showOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed: %s: %w", command, err)
		}
	} else {
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("Command failed: %s\n", command)
			if stderr.Len() > 0 {
				fmt.Printf("Error output: %s\n", stderr.String())
			}
			if stdout.Len() > 0 {
				fmt.Printf("Standard output: %s\n", stdout.String())
			}
			return fmt.Errorf("command failed: %s: %w", command, err)
		}
	}

	return nil
}

type PackageQueryExecutor struct {
	boxName    string
	workerPool *WorkerPool
}

func NewPackageQueryExecutor(boxName string) *PackageQueryExecutor {
	return &PackageQueryExecutor{
		boxName:    boxName,
		workerPool: NewWorkerPool(5, 2*time.Minute),
	}
}

type PackageQuery struct {
	Name    string
	Command string
}

func (pqe *PackageQueryExecutor) QueryAllPackages() (map[string][]string, error) {
	queries := []PackageQuery{
		{"apt", "dpkg-query -W -f='${Package}=${Version}\\n' $(apt-mark showmanual 2>/dev/null || true) 2>/dev/null | sort"},
		{"pip", "python3 -m pip freeze 2>/dev/null || pip3 freeze 2>/dev/null || true"},
		{"npm", "npm list -g --depth=0 --json 2>/dev/null || true"},
		{"yarn", "yarn global list --depth=0 2>/dev/null | sed -n 's/^[[:space:]]*├──[[:space:]]*//p' | sed -n 's/^[[:space:]]*└──[[:space:]]*//p' | sed 's/ (.*)//'"},
		{"pnpm", "pnpm ls -g --depth=0 --json 2>/dev/null || true"},
	}

	tasks := make([]StringTask, len(queries))
	for i, query := range queries {
		tasks[i] = pqe.createQueryTask(query.Command)
	}

	results, errors := pqe.workerPool.ExecuteStringTasks(tasks)

	packageLists := make(map[string][]string)
	for i, query := range queries {
		if errors[i] != nil {
			fmt.Printf("Warning: failed to query %s packages: %v\n", query.Name, errors[i])
			packageLists[query.Name] = nil
			continue
		}

		switch query.Name {
		case "apt", "pip":
			packageLists[query.Name] = parseLineList(results[i])
		case "npm", "pnpm":
			packageLists[query.Name] = parseJSONPackageList(results[i])
		case "yarn":
			packageLists[query.Name] = parseLineList(results[i])
		}
	}

	return packageLists, nil
}

func (pqe *PackageQueryExecutor) createQueryTask(command string) StringTask {
	return func() (string, error) {
		cmd := exec.Command("docker", "exec", pqe.boxName, "bash", "-c", command)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("query failed: %w", err)
		}

		return stdout.String(), nil
	}
}

func parseLineList(output string) []string {
	var result []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

func parseJSONPackageList(output string) []string {

	if strings.TrimSpace(output) == "" {
		return nil
	}

	return nil
}
