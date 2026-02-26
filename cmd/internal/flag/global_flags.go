package flag

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	LogFileNameFlag = "log-file"
	LogFileNameHelp = `log output destination ('-' for stderr)`

	LogFormatFlag = "log-format"
	LogFormatHelp = `log output format (text or json)`

	LogLevelFlag = "log-level"
	LogLevelHelp = `minimum log level (debug, info, warn, error)`

	ConfigFlag = "config"
	ConfigHelp = `path to config file; if unset, searches standard XDG config locations and the current directory for config.yaml`
)

type PersistentFlags struct {
}

func AddPersistentFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String(ConfigFlag, "", ConfigHelp)
	cmd.PersistentFlags().String(LogLevelFlag, "info", LogLevelHelp)
	cmd.PersistentFlags().StringP(LogFileNameFlag, "O", "-", LogFileNameHelp)
	cmd.PersistentFlags().String(LogFormatFlag, "text", LogFormatHelp)
	_ = cmd.RegisterFlagCompletionFunc(LogFormatFlag, SliceCompletion{
		"text\tHuman readable text output",
		"json\tJSON output",
	}.CompletionFn)
	_ = cmd.RegisterFlagCompletionFunc(LogLevelFlag, SliceCompletion{
		"debug\tLog debug messages",
		"info\tLog informational messages",
		"warn\tLog warnings",
		"error\tLog errors",
	}.CompletionFn)
}

func LogFormat() string {
	return viper.GetString(LogFormatFlag)
}

func LogFileName() string {
	return viper.GetString(LogFileNameFlag)
}

func LogLevel() string {
	return viper.GetString(LogLevelFlag)
}

func Config() string {
	return viper.GetString(ConfigFlag)
}
