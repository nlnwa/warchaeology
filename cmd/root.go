/*
 * Copyright © 2019 National Library of Norway
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"github.com/klauspost/compress/gzip"
	"github.com/nlnwa/warchaeology/cmd/cat"
	"github.com/nlnwa/warchaeology/cmd/console"
	"github.com/nlnwa/warchaeology/cmd/convert"
	"github.com/nlnwa/warchaeology/cmd/dedup"
	"github.com/nlnwa/warchaeology/cmd/ls"
	"github.com/nlnwa/warchaeology/cmd/validate"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/nlnwa/warchaeology/internal/warcwriterconfig"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"time"
)

type conf struct {
	cfgFile  string
	logLevel string
}

// NewCommand returns a new cobra.Command implementing the root command for warc
func NewCommand() *cobra.Command {
	c := &conf{}
	cmd := &cobra.Command{
		Use:   "warc",
		Short: "A tool for handling warc files",
		Long:  ``,

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Overwrite config values if set in command specific key
			cv := viper.Sub(cmd.Name())
			if cv != nil {
				for _, k := range cv.AllKeys() {
					viper.Set(k, cv.Get(k))
				}
			}

			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				// Hack to let bool flags toggle on or off
				if flag.Value.Type() == "bool" {
					fv, err := cmd.Flags().GetBool(flag.Name)
					if err != nil {
						panic(err)
					}
					vv := viper.GetBool(flag.Name)
					if fv {
						vv = !vv
					}
					if vv {
						err = cmd.Flags().Set(flag.Name, "true")
					} else {
						err = cmd.Flags().Set(flag.Name, "false")
					}
					if err != nil {
						panic(err)
					}
				}
			})

			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				panic(err)
			}

			return nil
		},
	}

	cobra.OnInitialize(func() { c.initConfig() })

	// Flags
	cmd.PersistentFlags().StringVar(&c.cfgFile, "config", "", "config file. If not set, /etc/warc/, $HOME/.warc/ and current working dir will be searched for file config.yaml")
	if err := viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config")); err != nil {
		panic(err)
	}
	cmd.PersistentFlags().StringP(flag.LogFileName, "L", "", flag.LogFileNameHelp)
	cmd.PersistentFlags().StringSlice(flag.LogFile, []string{"info", "error", "summary"}, flag.LogFileHelp)
	cmd.PersistentFlags().StringSlice(flag.LogConsole, []string{"progress", "summary"}, flag.LogConsoleHelp)
	cmd.PersistentFlags().String(flag.TmpDir, os.TempDir(), flag.TmpDirHelp)

	// Subcommands
	cmd.AddCommand(ls.NewCommand())
	cmd.AddCommand(cat.NewCommand())
	cmd.AddCommand(validate.NewCommand())
	cmd.AddCommand(console.NewCommand())
	cmd.AddCommand(convert.NewCommand())
	cmd.AddCommand(dedup.NewCommand())
	cmd.AddCommand(completionCmd)

	return cmd
}

// initConfig reads in config file and ENV variables if set.
func (c *conf) initConfig() {
	viper.SetTypeByDefaultValue(true)
	viper.SetDefault(flag.WarcVersion, "1.1")
	viper.SetDefault(flag.CompressionLevel, gzip.DefaultCompression)
	viper.SetDefault(flag.DefaultDate, time.Now().Format(warcwriterconfig.DefaultDateFormat))

	viper.AutomaticEnv() // read in environment variables that match

	if viper.IsSet("config") {
		// Use config file from the flag.
		viper.SetConfigFile(viper.GetString("config"))
	} else {
		// Search config in home directory with name ".warc" (without extension).
		viper.SetConfigName("config")      // name of config file (without extension)
		viper.SetConfigType("yaml")        // REQUIRED if the config file does not have the extension in the name
		viper.AddConfigPath("/etc/warc/")  // path to look for the config file in
		viper.AddConfigPath("$HOME/.warc") // call multiple times to add many search paths
		viper.AddConfigPath(".")           // optionally look for config in the working directory
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error
			return
		} else {
			// Config file was found but another error was produced
			log.Fatalf("error reading config file: %v", err)
		}
	}
}
