package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kirsle/configdir"
	"github.com/nlnwa/warchaeology/cmd/aart"
	"github.com/nlnwa/warchaeology/cmd/cat"
	"github.com/nlnwa/warchaeology/cmd/console"
	"github.com/nlnwa/warchaeology/cmd/convert"
	"github.com/nlnwa/warchaeology/cmd/dedup"
	"github.com/nlnwa/warchaeology/cmd/internal/flag"
	"github.com/nlnwa/warchaeology/cmd/ls"
	"github.com/nlnwa/warchaeology/cmd/validate"
	"github.com/nlnwa/warchaeology/cmd/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewWarcCommand returns a new cobra.Command implementing the root command for warc
func NewWarcCommand() *cobra.Command {
	flags := flag.PersistentFlags{}

	cmd := &cobra.Command{
		Use:   "warc",
		Short: "A tool for handling warc files",
		Long:  ``,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			viper.SetEnvPrefix("WARC")
			viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
			viper.AutomaticEnv()
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return loadConfig()
		},
	}

	// Add flags
	flags.AddFlags(cmd)

	// Add subcommands
	cmd.AddCommand(ls.NewCmdList())           // ls
	cmd.AddCommand(cat.NewCmdCat())           // cat
	cmd.AddCommand(validate.NewCmdValidate()) // validate
	cmd.AddCommand(console.NewCmdConsole())   // console
	cmd.AddCommand(convert.NewCmdConvert())   // convert
	cmd.AddCommand(dedup.NewCmdDedup())       // dedup
	cmd.AddCommand(aart.NewCmdAart())         // aart
	cmd.AddCommand(version.NewCmdVersion())   // version

	return cmd
}

func loadConfig() error {
	if viper.IsSet("config") {
		// Read config file specified by 'config' flag
		viper.SetConfigFile(viper.GetString("config"))
		return viper.ReadInConfig()
	}

	// Read config file from default locations
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name

	// current directory
	workingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}
	// local user config path
	userConfigDir := configdir.LocalConfig("warc")
	// system config path(s)
	systemConfigDirs := configdir.SystemConfig("warc")
	configDirs := []string{workingDirectory, userConfigDir}
	configDirs = append(configDirs, systemConfigDirs...)
	for _, configDir := range configDirs {
		viper.AddConfigPath(configDir)
	}

	return readConfigFile()
}

func readConfigFile() error {
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error
			return nil
		}
		return err
	}
	return nil
}
