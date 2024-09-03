package flag

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	LogFileNameFlag = "log-file"
	LogFileNameHelp = `log to file`

	LogFormatFlag = "log-format"
	LogFormatHelp = `log format. Valid values: text, json`

	LogLevelFlag = "log-level"
	LogLevelHelp = `log level. Valid values: debug, info, warn, error`

	ConfigFlag = "config"
	ConfigHelp = `config file. If not set, $XDG_CONFIG_DIRS, /etc/xdg/warc $XDG_CONFIG_HOME/warc and the current directory will be searched for a file named 'config.yaml'`
)

type PersistentFlags struct {
}

func (f PersistentFlags) AddFlags(cmd *cobra.Command) {
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
