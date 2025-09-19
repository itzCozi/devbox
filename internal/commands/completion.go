package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

  $ source <(devbox completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ devbox completion bash > /etc/bash_completion.d/devbox

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ devbox completion zsh > "${fpath[1]}/_devbox"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ devbox completion fish | source

  # To load completions for each session, execute once:
  $ devbox completion fish > ~/.config/fish/completions/devbox.fish

`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		}
	},
}

func getProjectNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if configManager == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	cfg, err := configManager.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	projects := cfg.GetProjects()
	var projectNames []string
	for _, project := range projects {
		projectNames = append(projectNames, project.Name)
	}

	return projectNames, cobra.ShellCompDirectiveNoFileComp
}

func getTemplateNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if configManager == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	templates := configManager.GetAvailableTemplates()
	return templates, cobra.ShellCompDirectiveNoFileComp
}

func init() {

	shellCmd.ValidArgsFunction = getProjectNames
	runCmd.ValidArgsFunction = getProjectNames
	stopCmd.ValidArgsFunction = getProjectNames
	destroyCmd.ValidArgsFunction = getProjectNames

	templatesShowCmd.ValidArgsFunction = getTemplateNames
	templatesDeleteCmd.ValidArgsFunction = getTemplateNames

	initCmd.RegisterFlagCompletionFunc("template", getTemplateNames)
}
