package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/kirsle/configdir"
	"github.com/nlnwa/warchaeology/v4/cmd/aart"
	"github.com/nlnwa/warchaeology/v4/cmd/cat"
	"github.com/nlnwa/warchaeology/v4/cmd/console"
	"github.com/nlnwa/warchaeology/v4/cmd/convert"
	"github.com/nlnwa/warchaeology/v4/cmd/dedup"
	"github.com/nlnwa/warchaeology/v4/cmd/internal/flag"
	"github.com/nlnwa/warchaeology/v4/cmd/ls"
	"github.com/nlnwa/warchaeology/v4/cmd/validate"
	"github.com/nlnwa/warchaeology/v4/cmd/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// NewWarcCommand returns a new cobra.Command implementing the root command for warc
func NewWarcCommand() *cobra.Command {
	var logCloser io.Closer

	cmd := &cobra.Command{
		Use:   "warc",
		Short: "A tool for handling warc files",
		Long:  ``,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := loadConfig(cmd.Flags())
			if err != nil {
				return err
			}
			logCloser, err = initLogger(os.Stderr, flag.LogFileName(), flag.LogFormat(), flag.LogLevel())
			return err
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if logCloser != nil {
				_ = logCloser.Close()
			}
		},
	}

	// Add flags
	flag.AddPersistentFlags(cmd)

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

func loadConfig(flags *pflag.FlagSet) error {
	viper.SetEnvPrefix("WARC")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	err := viper.BindPFlags(flags)
	if err != nil {
		return err
	}

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

func initLogger(out io.WriteCloser, file string, format string, level string) (io.Closer, error) {
	if file == "-" {
		return out, nil
	}
	w, err := os.Create(file)
	if err != nil {
		return nil, err
	}

	levelVar := new(slog.LevelVar)
	levelVar.Set(toLogLevel(level))

	opts := &slog.HandlerOptions{Level: levelVar}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}

	slog.SetDefault(slog.New(handler))

	return w, nil
}

func toLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
