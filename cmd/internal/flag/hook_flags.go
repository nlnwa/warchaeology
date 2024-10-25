package flag

import (
	"github.com/nlnwa/warchaeology/v3/internal/hooks"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	OpenInputFileHook     = "open-input-file-hook"
	OpenInputFileHookHelp = `a command to run before opening each input file. The command has access to data as environment variables.
	WARC_COMMAND contains the subcommand name
	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
	WARC_FILE_NAME contains the file name of the input file`

	CloseInputFileHook     = "close-input-file-hook"
	CloseInputFileHookHelp = `a command to run after closing each input file. The command has access to data as environment variables.
	WARC_COMMAND contains the subcommand name
	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
	WARC_FILE_NAME contains the file name of the input file
	WARC_ERROR_COUNT contains the number of errors found if the file was validated and the validation failed`

	OpenOutputFileHook     = "open-output-file-hook"
	OpenOutputFileHookHelp = `a command to run before opening each output file. The command has access to data as environment variables.
	WARC_COMMAND contains the subcommand name
	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
	WARC_FILE_NAME contains the file name of the output file
	WARC_SRC_FILE_NAME contains the file name of the input file if the output file is generated from an input file`

	CloseOutputFileHook     = "close-output-file-hook"
	CloseOutputFileHookHelp = `a command to run after closing each output file. The command has access to data as environment variables.
	WARC_COMMAND contains the subcommand name
	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
	WARC_FILE_NAME contains the file name of the output file
	WARC_SIZE contains the size of the output file
	WARC_INFO_ID contains the ID of the output file's WARCInfo-record if created
	WARC_SRC_FILE_NAME contains the file name of the input file if the output file is generated from an input file
	WARC_HASH contains the hash of the output file if computed
	WARC_ERROR_COUNT contains the number of errors found if the file was validated and the validation failed`
)

type InputHookFlags struct {
	name string
}

func (f *InputHookFlags) AddFlags(cmd *cobra.Command, opts ...func(*InputHookFlags)) {
	f.name = cmd.Name()
	flags := cmd.Flags()
	flags.String(OpenInputFileHook, "", OpenInputFileHookHelp)
	flags.String(CloseInputFileHook, "", CloseInputFileHookHelp)
}

func (f InputHookFlags) OpenInputFileHook() string {
	return viper.GetString(OpenInputFileHook)
}

func (f InputHookFlags) CloseInputFileHook() string {
	return viper.GetString(CloseInputFileHook)
}

func (f InputHookFlags) ToOpenInputFileHook() (hooks.OpenInputFileHook, error) {
	return hooks.NewOpenInputFileHook(f.name, f.OpenInputFileHook())
}

func (f InputHookFlags) ToCloseInputFileHook() (hooks.CloseInputFileHook, error) {
	return hooks.NewCloseInputFileHook(f.name, f.CloseInputFileHook())
}

type OutputHookFlags struct {
	name string
}

func (f *OutputHookFlags) AddFlags(cmd *cobra.Command, opts ...func(*OutputHookFlags)) {
	f.name = cmd.Name()
	flags := cmd.Flags()
	flags.String(OpenOutputFileHook, "", OpenOutputFileHookHelp)
	flags.String(CloseOutputFileHook, "", CloseOutputFileHookHelp)
}

func (f OutputHookFlags) OpenOutputFileHook() string {
	return viper.GetString(OpenOutputFileHook)
}

func (f OutputHookFlags) CloseOutputFileHook() string {
	return viper.GetString(CloseOutputFileHook)
}

func (f OutputHookFlags) ToOpenOutputFileHook() (hooks.OpenOutputFileHook, error) {
	return hooks.NewOpenOutputFileHook(f.name, f.OpenOutputFileHook())
}

func (f OutputHookFlags) ToCloseOutputFileHook() (hooks.CloseOutputFileHook, error) {
	return hooks.NewCloseOutputFileHook(f.name, f.CloseOutputFileHook())
}
