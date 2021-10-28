package flag

const (
	LogFileName       = "log-file-name"
	LogFile           = "log-file"
	LogConsole        = "log-console"
	RecordType        = "record-type"
	WarcDir           = "warc-dir"
	KeepIndex         = "keep-index"
	IndexDir          = "index-dir"
	Recursive         = "recursive"
	FollowSymlinks    = "symlinks"
	Suffixes          = "suffixes"
	Concurrency       = "concurrency"
	ConcurrentWriters = "concurrent-writers"
	FileSize          = "file-size"
	Compress          = "compress"
	CompressionLevel  = "compression-level"
	FilePrefix        = "prefix"
	SubdirPattern     = "subdir-pattern"
	NameGenerator     = "name-generator"
	Flush             = "flush"
	WarcVersion       = "warc-version"
	DefaultDate       = "default-date"

	LogFileNameHelp = `a file to write log output. Empty for no log file`
	LogFileHelp     = `The kind of log output to write to file. Valid values: info, error, summary`
	LogConsoleHelp  = `The kind of log output to write to console. Valid values: info, error, summary, progress`
	RecordTypeHelp  = `record types to dedup. For more than one, repeat flag or comma separated list.
Legal values: warcinfo,request,response,metadata,revisit,resource,continuation,conversion`
	WarcDirHelp           = `output directory for generated warc files. Directory must exist.`
	KeepIndexHelp         = `true to keep index on disk after dedup, so that it can be used for the next run`
	IndexDirHelp          = `directory to store deduplication index`
	RecursiveHelp         = `walk directories recursively`
	FollowSymlinksHelp    = `follow symlinks`
	SuffixesHelp          = `filter by suffixes`
	ConcurrencyHelp       = `number of input files to process simultaneously. The default value is 1.5 x <number of cpu cores>`
	ConcurrentWritersHelp = `maximum concurrent WARC writers. This is the number of WARC-files simultaneously written to.
A consequence is that at least this many WARC files are created even if there is only one input file.`
	FileSizeHelp         = `The maximum size for WARC files`
	CompressHelp         = `use gzip compression for WARC files`
	CompressionLevelHelp = `the gzip compression level to use (value between 1 and 9)`
	FilePrefixHelp       = `filename prefix for WARC files`
	SubdirPatternHelp    = `a pattern to use for generating subdirectories.
/ in pattern separates subdirectories on all platforms
{YYYY} is replaced with a 4 digit year
{YY} is replaced with a 2 digit year
{MM} is replaced with a 2 digit month
{DD} is replaced with a 2 day
The date used is the WARC date of each record. Therefore a input file might be split into 
WARC files in different subdirectories. If NameGenerator is 'identity' only the first record
of each file's date is used to keep the file as one.`
	NameGeneratorHelp = `the name generator to use. By setting this to 'identity', the input filename will also be used as
output file name (preix and suffix might still change). In this mode exactly one file is generated for every input file`
	FlushHelp       = `if true, sync WARC file to disk after writing each record`
	WarcVersionHelp = `the WARC version to use for created files`
	DefaultDateHelp = `fetch date to use for records missing date metadata. Fetchtime is set to 12:00 UTC for the date`
)
