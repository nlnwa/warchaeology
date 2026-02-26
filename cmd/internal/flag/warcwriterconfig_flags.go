package flag

import (
	"compress/gzip"
	"time"

	"github.com/nationallibraryofnorway/warchaeology/v4/internal/warcwriterconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ConcurrentWriters     = "concurrent-writers"
	ConcurrentWritersHelp = `maximum number of WARC files written concurrently.
This may create at least this many output files even with a single input file.`

	FileSize     = "file-size"
	FileSizeHelp = `maximum size of each WARC output file`

	Compress     = "compress"
	CompressHelp = `enable gzip compression for WARC output files`

	CompressionLevel     = "compression-level"
	CompressionLevelHelp = `gzip compression level (1-9, -1 uses the gzip library default)`

	FilePrefix     = "prefix"
	FilePrefixHelp = `filename prefix for generated WARC files`

	SubdirPattern     = "subdir-pattern"
	SubdirPatternHelp = `pattern used to create output subdirectories.
Use '/' to separate subdirectories on all platforms.
Supported tokens: {YYYY}, {YY}, {MM}, {DD}.
The WARC-Date of each record is used, so one input file may be split across subdirectories.
With --name-generator=identity, only the first record date is used per input file.`

	NameGenerator     = "name-generator"
	NameGeneratorHelp = `name generator strategy.
With 'identity', the input filename is reused for output (prefix/suffix may still change),
and exactly one output file is created per input file.`

	Flush     = "flush"
	FlushHelp = `sync each WARC file to disk after every record`

	WarcVersion     = "warc-version"
	WarcVersionHelp = `WARC version used for generated files`

	DefaultDate     = "default-date"
	DefaultDateHelp = `fallback date used when records are missing WARC-Date metadata (time is set to 12:00 UTC)`

	OutputDir     = "output-dir"
	OutputDirHelp = `output directory for generated WARC files (must already exist)`

	OneToOne     = "one-to-one"
	OneToOneHelp = `write each input file to exactly one output file.
Equivalent to: --concurrent-writers=1 --file-size=0 --name-generator=identity`
)

type WarcWriterConfigFlags struct {
	defaultFilePrefix string
	defaultOneToOne   bool
	name              string
}

func WithDefaultFilePrefix(prefix string) func(*WarcWriterConfigFlags) {
	return func(f *WarcWriterConfigFlags) {
		f.defaultFilePrefix = prefix
	}
}

func WithDefaultOneToOne(oneToOne bool) func(*WarcWriterConfigFlags) {
	return func(f *WarcWriterConfigFlags) {
		f.defaultOneToOne = oneToOne
	}
}

func (f *WarcWriterConfigFlags) AddFlags(cmd *cobra.Command, options ...func(*WarcWriterConfigFlags)) {
	for _, option := range options {
		option(f)
	}
	f.name = cmd.Name()
	flags := cmd.Flags()
	flags.BoolP(Compress, "z", true, CompressHelp)
	flags.Int(CompressionLevel, gzip.DefaultCompression, CompressionLevelHelp)
	flags.IntP(ConcurrentWriters, "C", 16, ConcurrentWritersHelp)
	flags.String(DefaultDate, time.Now().Format(warcwriterconfig.DefaultDateFormat), DefaultDateHelp)
	flags.String(FileSize, "1GB", FileSizeHelp)
	flags.StringP(FilePrefix, "p", f.defaultFilePrefix, FilePrefixHelp)
	flags.Bool(Flush, false, FlushHelp)
	flags.String(NameGenerator, "default", NameGeneratorHelp)
	flags.String(SubdirPattern, "", SubdirPatternHelp)
	flags.StringP(OutputDir, "w", ".", OutputDirHelp)
	flags.String(WarcVersion, "1.1", WarcVersionHelp)
	flags.Bool(OneToOne, f.defaultOneToOne, OneToOneHelp)

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

func (f *WarcWriterConfigFlags) ConcurrentWriters() int {
	return viper.GetInt(ConcurrentWriters)
}

func (f *WarcWriterConfigFlags) FileSize() string {
	return viper.GetString(FileSize)
}

func (f *WarcWriterConfigFlags) Compress() bool {
	return viper.GetBool(Compress)
}

func (f *WarcWriterConfigFlags) CompressionLevel() int {
	return viper.GetInt(CompressionLevel)
}

func (f *WarcWriterConfigFlags) FilePrefix() string {
	return viper.GetString(FilePrefix)
}

func (f *WarcWriterConfigFlags) SubdirPattern() string {
	return viper.GetString(SubdirPattern)
}

func (f *WarcWriterConfigFlags) NameGenerator() string {
	return viper.GetString(NameGenerator)
}

func (f *WarcWriterConfigFlags) Flush() bool {
	return viper.GetBool(Flush)
}

func (f *WarcWriterConfigFlags) WarcVersion() string {
	return viper.GetString(WarcVersion)
}

func (f *WarcWriterConfigFlags) DefaultDate() string {
	return viper.GetString(DefaultDate)
}

func (f *WarcWriterConfigFlags) OutputDir() string {
	return viper.GetString(OutputDir)
}

func (f *WarcWriterConfigFlags) OneToOne() bool {
	return viper.GetBool(OneToOne)
}

func (f *WarcWriterConfigFlags) ToWarcWriterConfig() (*warcwriterconfig.WarcWriterConfig, error) {
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
