package flag

import (
	"compress/gzip"
	"time"

	"github.com/nlnwa/warchaeology/internal/warcwriterconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ConcurrentWriters     = "concurrent-writers"
	ConcurrentWritersHelp = `maximum concurrent WARC writers. This is the number of WARC-files simultaneously written to.
	A consequence is that at least this many WARC files are created even if there is only one input file.`

	FileSize     = "file-size"
	FileSizeHelp = `The maximum size for WARC files`

	Compress     = "compress"
	CompressHelp = `use gzip compression for WARC files`

	CompressionLevel     = "compression-level"
	CompressionLevelHelp = `the gzip compression level to use (value between 1 and 9)`

	FilePrefix     = "prefix"
	FilePrefixHelp = `filename prefix for WARC files`

	SubdirPattern     = "subdir-pattern"
	SubdirPatternHelp = `a pattern to use for generating subdirectories.
	/ in pattern separates subdirectories on all platforms
	{YYYY} is replaced with a 4 digit year
	{YY} is replaced with a 2 digit year
	{MM} is replaced with a 2 digit month
	{DD} is replaced with a 2 digit day
	The date used is the WARC date of each record. Therefore a input file might be split into 
	WARC files in different subdirectories. If NameGenerator is 'identity' only the first record
	of each file's date is used to keep the file as one.`

	NameGenerator     = "name-generator"
	NameGeneratorHelp = `the name generator to use. By setting this to 'identity', the input filename will also be used as
output file name (prefix and suffix might still change). In this mode exactly one file is generated for every input file`

	Flush     = "flush"
	FlushHelp = `if true, sync WARC file to disk after writing each record`

	WarcVersion     = "warc-version"
	WarcVersionHelp = `the WARC version to use for created files`

	DefaultDate     = "default-date"
	DefaultDateHelp = `fetch date to use for records missing date metadata. Fetchtime is set to 12:00 UTC for the date`

	OutputDir     = "output-dir"
	OutputDirHelp = `output directory for generated warc files. Directory must exist.`

	OneToOne     = "one-to-one"
	OneToOneHelp = `write each input file to a separate output file
The same as --concurrent-writers=1, and --name-generator=identity`
)

type WarcWriterConfigFlags struct {
	filePrefix string
	name       string
}

func WithDefaultFilePrefix(prefix string) func(*WarcWriterConfigFlags) {
	return func(f *WarcWriterConfigFlags) {
		f.filePrefix = prefix
	}
}

func WithCmdName(name string) func(*WarcWriterConfigFlags) {
	return func(f *WarcWriterConfigFlags) {
		f.name = name
	}
}

func (f WarcWriterConfigFlags) AddFlags(cmd *cobra.Command, options ...func(*WarcWriterConfigFlags)) {
	for _, option := range options {
		option(&f)
	}
	flags := cmd.Flags()
	flags.BoolP(Compress, "z", false, CompressHelp)
	flags.Int(CompressionLevel, gzip.DefaultCompression, CompressionLevelHelp)
	flags.IntP(ConcurrentWriters, "C", 16, ConcurrentWritersHelp)
	flags.String(DefaultDate, time.Now().Format(warcwriterconfig.DefaultDateFormat), DefaultDateHelp) // TODO -t --record-type collision
	flags.String(FileSize, "1GB", FileSizeHelp)                                                       // TODO -S --response-code collision
	flags.StringP(FilePrefix, "p", f.filePrefix, FilePrefixHelp)
	flags.Bool(Flush, false, FlushHelp)
	flags.String(NameGenerator, "default", NameGeneratorHelp)
	flags.String(SubdirPattern, "", SubdirPatternHelp)
	flags.StringP(OutputDir, "w", ".", OutputDirHelp)
	flags.String(WarcVersion, "1.1", WarcVersionHelp)
	flags.Bool(OneToOne, false, OneToOneHelp)

	var lastErr error
	if err := cmd.RegisterFlagCompletionFunc(FilePrefix, cobra.NoFileCompletions); err != nil {
		lastErr = err
	}
	if err := cmd.MarkFlagDirname(OutputDir); err != nil {
		lastErr = err
	}
	if err := cmd.RegisterFlagCompletionFunc(SubdirPattern, cobra.NoFileCompletions); err != nil {
		lastErr = err
	}
	if err := cmd.RegisterFlagCompletionFunc(NameGenerator, cobra.NoFileCompletions); err != nil {
		lastErr = err
	}
	if lastErr != nil {
		panic(lastErr)
	}

}

func (f WarcWriterConfigFlags) ConcurrentWriters() int {
	return viper.GetInt(ConcurrentWriters)
}

func (f WarcWriterConfigFlags) FileSize() string {
	return viper.GetString(FileSize)
}

func (f WarcWriterConfigFlags) Compress() bool {
	return viper.GetBool(Compress)
}

func (f WarcWriterConfigFlags) CompressionLevel() int {
	return viper.GetInt(CompressionLevel)
}

func (f WarcWriterConfigFlags) FilePrefix() string {
	return viper.GetString(FilePrefix)
}

func (f WarcWriterConfigFlags) SubdirPattern() string {
	return viper.GetString(SubdirPattern)
}

func (f WarcWriterConfigFlags) NameGenerator() string {
	return viper.GetString(NameGenerator)
}

func (f WarcWriterConfigFlags) Flush() bool {
	return viper.GetBool(Flush)
}

func (f WarcWriterConfigFlags) WarcVersion() string {
	return viper.GetString(WarcVersion)
}

func (f WarcWriterConfigFlags) DefaultDate() string {
	return viper.GetString(DefaultDate)
}

func (f WarcWriterConfigFlags) OutputDir() string {
	return viper.GetString(OutputDir)
}

func (f *WarcWriterConfigFlags) OneToOne() bool {
	return viper.GetBool(OneToOne)
}

func (f WarcWriterConfigFlags) ToWarcWriterConfig() (*warcwriterconfig.WarcWriterConfig, error) {
	return warcwriterconfig.New(f.name,
		warcwriterconfig.WithOneToOneWriter(f.OneToOne()),
		warcwriterconfig.WithOutDir(f.OutputDir()),
		warcwriterconfig.WithConcurrentWriters(f.ConcurrentWriters()),
		warcwriterconfig.WithMaxFileSize(f.FileSize()),
		warcwriterconfig.WithCompress(f.Compress()),
		warcwriterconfig.WithCompressionLevel(f.CompressionLevel()),
		warcwriterconfig.WithFilePrefix(f.FilePrefix()),
		warcwriterconfig.WithSubDirPattern(f.SubdirPattern()),
		warcwriterconfig.WithWarcFileNameGenerator(f.NameGenerator()),
		warcwriterconfig.WithFlush(f.Flush()),
		warcwriterconfig.WithWarcVersion(f.WarcVersion()),
		warcwriterconfig.WithDefaultTime(f.DefaultDate()),
		warcwriterconfig.WithOpenOutputFileHook(viper.GetString(OpenOutputFileHook)),
		warcwriterconfig.WithCloseOutputFileHook(viper.GetString(CloseOutputFileHook)),
	)
}
