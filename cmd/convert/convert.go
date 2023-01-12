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
 */

package convert

import (
	"github.com/nlnwa/warchaeology/cmd/convert/arc"
	"github.com/nlnwa/warchaeology/cmd/convert/nedlib"
	"github.com/nlnwa/warchaeology/cmd/convert/warc"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "convert",
		Short: "Convert web archives to warc files. Use subcommands for the supported formats",
		Long:  ``,
	}

	// Subcommands
	cmd.AddCommand(nedlib.NewCommand())
	cmd.AddCommand(arc.NewCommand())
	cmd.AddCommand(warc.NewCommand())

	return cmd
}
