/*
 * Copyright 2023 National Library of Norway.
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

package config

import (
	"fmt"
	"github.com/kirsle/configdir"
	"github.com/klauspost/compress/gzip"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/nlnwa/warchaeology/internal/warcwriterconfig"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"strings"
	"time"
)

func InitConfig(cmd *cobra.Command) {
	sc := strings.Join(configdir.SystemConfig("warc"), ", ")
	lc := configdir.LocalConfig("warc")
	ch := fmt.Sprintf("config file. If not set, %s, %s and the current directory will be searched for a file named 'config.yaml'", sc, lc)
	cmd.Root().PersistentFlags().String("config", "", ch)
	if err := viper.BindPFlag("config", cmd.Root().PersistentFlags().Lookup("config")); err != nil {
		panic(err)
	}
	updateHelp(cmd.Root())
	cmd.Root().PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		loadConfig(cmd)
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			panic(err)
		}
		return nil
	}
}

func loadConfig(cmd *cobra.Command) {
	viper.SetTypeByDefaultValue(true)
	viper.SetDefault(flag.WarcVersion, "1.1")
	viper.SetDefault(flag.CompressionLevel, gzip.DefaultCompression)
	viper.SetDefault(flag.DefaultDate, time.Now().Format(warcwriterconfig.DefaultDateFormat))

	viper.AutomaticEnv() // read in environment variables that match

	if viper.IsSet("config") {
		// Use config file from 'config' flag.
		viper.SetConfigFile(viper.GetString("config"))

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Config file not found; ignore error
				return
			} else {
				// Config file was found but another error was produced
				log.Fatalf("error reading config file: %v", err)
			}
		}
		conf := consolidateConfig(viper.AllSettings(), cmd)
		if err := viper.MergeConfigMap(conf); err != nil {
			log.Fatalf("error reading config file: %v", err)
		}
	} else {
		viper.SetConfigName("config") // name of config file (without extension)
		viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name

		scd := configdir.SystemConfig("warc")
		ucd := configdir.LocalConfig("warc")
		wd, _ := os.Getwd()
		for _, d := range scd {
			viper.AddConfigPath(d)
		}

		c := readConfigFile(scd...)
		consolidateConfig(c, cmd)
		if err := viper.MergeConfigMap(c); err != nil {
			log.Fatalf("error reading config file: %v", err)
		}

		c = readConfigFile(ucd)
		consolidateConfig(c, cmd)
		if err := viper.MergeConfigMap(c); err != nil {
			log.Fatalf("error reading config file: %v", err)
		}

		c = readConfigFile(wd)
		consolidateConfig(c, cmd)
		if err := viper.MergeConfigMap(c); err != nil {
			log.Fatalf("error reading config file: %v", err)
		}
	}
}

func readConfigFile(path ...string) map[string]interface{} {
	v := viper.New()
	v.SetConfigName("config") // name of config file (without extension)
	v.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	for _, p := range path {
		v.AddConfigPath(p)
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error
			return nil
		} else {
			// Config file was found but another error was produced
			log.Fatalf("error reading config file: %v", err)
		}
	}
	return v.AllSettings()
}

func updateHelp(root *cobra.Command) {
	root.InitDefaultHelpCmd()
	h := root.HelpFunc()
	root.SetHelpFunc(func(cmd *cobra.Command, strings []string) {
		loadConfig(cmd)
		setCurrentValueFromConfig(cmd)
		h(cmd, strings)
	})
}

func setCurrentValueFromConfig(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Name != "help" && viper.IsSet(f.Name) {
			if f.Value.Type() == "string" {
				f.Usage += fmt.Sprintf(" (current %q)", viper.GetString(f.Name))
			} else {
				f.Usage += fmt.Sprintf(" (current %s)", viper.GetString(f.Name))
			}
		}
	})
}

func consolidateConfig(conf map[string]interface{}, cmd *cobra.Command) map[string]interface{} {
	var commands []string
	c := cmd
	for c.HasParent() {
		commands = append(commands, c.Name())
		c = c.Parent()
	}
	cv := conf
	for i := len(commands) - 1; i >= 0; i-- {
		var ok bool
		cv, ok = cv[commands[i]].(map[string]interface{})
		if ok && cv != nil {
			for k, v := range cv {
				f := cmd.Flag(k)
				_, isSub := v.(map[string]interface{})
				if (f == nil || !f.Changed) && !isSub {
					conf[k] = v
				}
			}
		}
	}
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		panic(err)
	}
	return conf
}
