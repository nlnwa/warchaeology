package flag

import (
	"strings"

	"github.com/spf13/cobra"
)

// SuffixCompletionFn can be used by commands that want to restrict file completion to suffixes set by flag
func SuffixCompletionFn(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if suf, err := cmd.Flags().GetStringSlice(Suffixes); err != nil {
		return nil, cobra.ShellCompDirectiveError
	} else {
		for i := range suf {
			suf[i] = strings.TrimLeft(suf[i], ".")
		}
		return suf, cobra.ShellCompDirectiveFilterFileExt | cobra.ShellCompDirectiveNoFileComp
	}
}

type SliceCompletion []string

func (validCompletions SliceCompletion) CompletionFn(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if toComplete == "" {
		return validCompletions, cobra.ShellCompDirectiveNoFileComp
	}

	var completedTokens []string
	idxLastToken := strings.LastIndex(toComplete, ",")
	if idxLastToken > -1 {
		completedTokens = strings.Split(toComplete[:idxLastToken], ",")
		// Add completed options as prefix
		p := toComplete[:idxLastToken+1]
		for i := 0; i < len(validCompletions); i++ {
			validCompletions[i] = p + validCompletions[i]
		}
	}

	// Remove already used options
	for _, t := range completedTokens {
		for i, v := range validCompletions {
			if strings.HasPrefix(v, t) {
				copy(validCompletions[i:], validCompletions[i+1:])            // Shift a[i+1:] left one index.
				validCompletions[len(validCompletions)-1] = ""                // Erase last element (write zero value).
				validCompletions = validCompletions[:len(validCompletions)-1] // Truncate slice.
				break
			}
		}
	}

	return validCompletions, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
}
