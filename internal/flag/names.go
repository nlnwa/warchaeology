package flag

const (
	LogFileName         = "log-file-name"
	LogFile             = "log-file"
	LogConsole          = "log-console"
	RecordId            = "id"
	RecordType          = "record-type"
	ResponseCode        = "response-code"
	MimeType            = "mime-type"
	Offset              = "offset"
	RecordNum           = "num"
	RecordCount         = "record-count"
	Strict              = "strict"
	Delimiter           = "delimiter"
	Fields              = "fields"
	ShowWarcHeader      = "header"
	ShowProtocolHeader  = "protocol-header"
	ShowPayload         = "payload"
	WarcDir             = "warc-dir"
	NewIndex            = "new-index"
	KeepIndex           = "keep-index"
	IndexDir            = "index-dir"
	Recursive           = "recursive"
	FollowSymlinks      = "symlinks"
	Suffixes            = "suffixes"
	Concurrency         = "concurrency"
	ConcurrentWriters   = "concurrent-writers"
	FileSize            = "file-size"
	Compress            = "compress"
	CompressionLevel    = "compression-level"
	FilePrefix          = "prefix"
	SubdirPattern       = "subdir-pattern"
	NameGenerator       = "name-generator"
	Flush               = "flush"
	WarcVersion         = "warc-version"
	DefaultDate         = "default-date"
	TmpDir              = "tmpdir"
	BufferMaxMem        = "max-buffer-mem"
	DedupSizeGain       = "min-size-gain"
	MinFreeDisk         = "min-free-disk"
	Repair              = "repair"
	SrcFilesystem       = "source-filesystem"
	SrcFileList         = "source-file-list"
	OpenInputFileHook   = "open-input-file-hook"
	CloseInputFileHook  = "close-input-file-hook"
	OpenOutputFileHook  = "open-output-file-hook"
	CloseOutputFileHook = "close-output-file-hook"
	CalculateHash       = "calculate-hash"

	LogFileNameHelp = `a file to write log output. Empty for no log file`
	LogFileHelp     = `the kind of log output to write to file. Valid values: info, error, summary`
	LogConsoleHelp  = `the kind of log output to write to console. Valid values: info, error, summary, progress`
	RecordIdHelp    = `filter record ID's. For more than one, repeat flag or comma separated list.`
	RecordTypeHelp  = `filter record types. For more than one, repeat flag or comma separated list.
Legal values: warcinfo,request,response,metadata,revisit,resource,continuation,conversion`
	ResponseCodeHelp = `filter records with given http response codes. Format is 'from-to' where from is inclusive and to is exclusive.
Examples:
'200': only records with 200 response
'200-300': all records with response code between 200(inclusive) and 300(exclusive)
'-400': all response codes below 400
'500-': all response codes from 500 and above`
	MimeTypeHelp           = `filter records with given mime-types. For more than one, repeat flag or comma separated list.`
	OffsetHelp             = `record offset`
	RecordNumHelp          = `print the n'th record (zero based). This is applied after records are filtered out by other options`
	RecordCountHelp        = `The maximum number of records to show`
	StrictHelp             = `strict parsing`
	DelimiterHelp          = `use string instead of SPACE for field delimiter`
	FieldsHelp             = `which fields to include. See 'warc help ls' for a description`
	ShowWarcHeaderHelp     = `show WARC header`
	ShowProtocolHeaderHelp = `show protocol header`
	ShowPayloadHelp        = `show payload`
	WarcDirHelp            = `output directory for generated warc files. Directory must exist.`
	NewIndexHelp           = `true to start from a fresh index, deleting eventual index from last run`
	KeepIndexHelp          = `true to keep index on disk so that the next run will continue where the previous run left off`
	IndexDirHelp           = `directory to store indexes`
	RecursiveHelp          = `walk directories recursively`
	FollowSymlinksHelp     = `follow symlinks`
	SuffixesHelp           = `filter files by suffixes`
	ConcurrencyHelp        = `number of input files to process simultaneously. The default value is 1.5 x <number of cpu cores>`
	ConcurrentWritersHelp  = `maximum concurrent WARC writers. This is the number of WARC-files simultaneously written to.
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
{DD} is replaced with a 2 digit day
The date used is the WARC date of each record. Therefore a input file might be split into 
WARC files in different subdirectories. If NameGenerator is 'identity' only the first record
of each file's date is used to keep the file as one.`
	NameGeneratorHelp = `the name generator to use. By setting this to 'identity', the input filename will also be used as
output file name (prefix and suffix might still change). In this mode exactly one file is generated for every input file`
	FlushHelp         = `if true, sync WARC file to disk after writing each record`
	WarcVersionHelp   = `the WARC version to use for created files`
	DefaultDateHelp   = `fetch date to use for records missing date metadata. Fetchtime is set to 12:00 UTC for the date`
	TmpDirHelp        = `directory to use for temporary files`
	BufferMaxMemHelp  = "the maximum bytes of memory allowed for each buffer before overflowing to disk"
	DedupSizeGainHelp = `minimum bytes one must earn to perform a deduplication`
	MinFreeDiskHelp   = `minimum free space on disk to allow WARC writing`
	RepairHelp        = `try to fix errors in records`
	SrcFilesystemHelp = `the source filesystem to use for input files. Default is to use OS file system. Legal values:
  ftp://user/pass@host:port
  tar://path/to/archive.tar
  tgz://path/to/archive.tar.gz
`
	SrcFileListHelp       = `a file containing a list of files to process, one file per line`
	OpenInputFileHookHelp = `a command to run before opening each input file. The command has access to data as environment variables.
WARC_COMMAND contains the subcommand name
WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
WARC_FILE_NAME contains the file name of the input file`
	CloseInputFileHookHelp = `a command to run after closing each input file. The command has access to data as environment variables.
WARC_COMMAND contains the subcommand name
WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
WARC_FILE_NAME contains the file name of the input file
WARC_ERROR_COUNT contains the number of errors found if the file was validated and the validation failed`
	OpenOutputFileHookHelp = `a command to run before opening each output file. The command has access to data as environment variables.
WARC_COMMAND contains the subcommand name
WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
WARC_FILE_NAME contains the file name of the output file
WARC_SRC_FILE_NAME contains the file name of the input file if the output file is generated from an input file`
	CloseOutputFileHookHelp = `a command to run after closing each output file. The command has access to data as environment variables.
WARC_COMMAND contains the subcommand name
WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
WARC_FILE_NAME contains the file name of the output file
WARC_SIZE contains the size of the output file
WARC_INFO_ID contains the ID of the output file's WARCInfo-record if created
WARC_SRC_FILE_NAME contains the file name of the input file if the output file is generated from an input file
WARC_HASH contains the hash of the output file if computed
WARC_ERROR_COUNT contains the number of errors found if the file was validated and the validation failed`
	CalculateHashHelp = `calculate hash of output file. The hash is made available to the close output file hook as WARC_HASH. Valid values: md5, sha1, sha256, sha512`
)
