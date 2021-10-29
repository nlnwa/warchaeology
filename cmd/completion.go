package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Output bash completion code",
	Long: `Output bash completion code. The shell code must be evaluated to provide
interactive completion of veidemannctl commands.  This can be done by sourcing it from the .bash_profile.

Example:
  ## Load the kubectl completion code for bash into the current shell
  source <(veidemannctl completion bash)
`,
}

var bashCmd = &cobra.Command{
	Use:   "bash",
	Short: "bash completion.",
	Long:  "Generate command completion for bash.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Root().GenBashCompletion(os.Stdout)
	},
}

var zshCmd = &cobra.Command{
	Use:   "zsh",
	Short: "zsh completion.",
	Long:  "Generate command completion for zsh.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Root().GenZshCompletion(os.Stdout)
	},
}

func init() {
	completionCmd.AddCommand(bashCmd)
	completionCmd.AddCommand(zshCmd)
}
