/*
 * Copyright 2021 National Library of Norway.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package flag

import (
	"github.com/spf13/cobra"
	"strings"
)

// SuffixCompletionFn can be added set for commands which want to restrict file completion to suffixes set by flag
func SuffixCompletionFn(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if suf, err := cmd.Flags().GetStringSlice(Suffixes); err != nil {
		return nil, cobra.ShellCompDirectiveError
	} else {
		for i := 0; i < len(suf); i++ {
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
