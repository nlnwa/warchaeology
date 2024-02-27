package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

    $ source <(warc completion bash)

    # To load completions for each session, execute once:
    # Linux:
    $ warc completion bash > /etc/bash_completion.d/warc
    # macOS:
    $ warc completion bash > /usr/local/etc/bash_completion.d/warc

Zsh:

    # If shell completion is not already enabled in your environment,
    # you will need to enable it.  You can execute the following once:

    $ echo "autoload -U compinit; compinit" >> ~/.zshrc

    # To load completions for each session, execute once:
    $ warc completion zsh > "${fpath[1]}/_warc"

    # You will need to start a new shell for this setup to take effect.

fish:

    $ warc completion fish | source

    # To load completions for each session, execute once:
    $ warc completion fish > ~/.config/fish/completions/warc.fish

PowerShell:

    PS> warc completion powershell | Out-String | Invoke-Expression

    # To load completions for every new session, run:
    PS> warc completion powershell > warc.ps1
    # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			_ = cmd.Root().GenBashCompletionV2(os.Stdout, true)
		case "zsh":
			_ = cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			_ = cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			_ = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}
