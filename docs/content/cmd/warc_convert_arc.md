---
date: 2025-02-26T20:24:03+01:00
title: "warc convert arc"
slug: warc_convert_arc
url: /cmd/warc_convert_arc/
---
## warc convert arc

Convert ARC to WARC

```
warc convert arc FILE/DIR ... [flags]
```

### Options

```
      --close-input-file-hook string    a command to run after closing each input file. The command has access to data as environment variables.
                                        	WARC_COMMAND contains the subcommand name
                                        	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
                                        	WARC_FILE_NAME contains the file name of the input file
                                        	WARC_ERROR_COUNT contains the number of errors found if the file was validated and the validation failed
      --close-output-file-hook string   a command to run after closing each output file. The command has access to data as environment variables.
                                        	WARC_COMMAND contains the subcommand name
                                        	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
                                        	WARC_FILE_NAME contains the file name of the output file
                                        	WARC_SIZE contains the size of the output file
                                        	WARC_INFO_ID contains the ID of the output file's WARCInfo-record if created
                                        	WARC_SRC_FILE_NAME contains the file name of the input file if the output file is generated from an input file
                                        	WARC_HASH contains the hash of the output file if computed
                                        	WARC_ERROR_COUNT contains the number of errors found if the file was validated and the validation failed
  -z, --compress                        use gzip compression for WARC files (default true)
      --compression-level int           the gzip compression level to use (value between 1 and 9, -1 means the gzip library default level) (default -1)
  -c, --concurrency int                 number of input files to process simultaneously. (default 24)
  -C, --concurrent-writers int          maximum concurrent WARC writers. This is the number of WARC-files simultaneously written to.
                                        	A consequence is that at least this many WARC files are created even if there is only one input file. (default 16)
      --continue-on-error               continue on error. Will continue processing files and directories in spite of errors.
      --default-date string             fetch date to use for records missing date metadata. Fetchtime is set to 12:00 UTC for the date (default "2025-2-26")
      --file-size string                The maximum size for WARC files (default "1GB")
      --flush                           if true, sync WARC file to disk after writing each record
      --ftp-pool-size int32             size of the ftp pool (default 1)
  -h, --help                            help for arc
      --index-dir string                directory to store indexes (default "/home/mariusb/.cache/arc")
  -i, --input-file string               input file (system). Default is to use OS file system.
                                        Legal values:
                                        	/path/to/archive.( tar | tar.gz | tgz | zip | wacz )
                                        	ftp://user/pass@host:port
                                        
  -k, --keep-index                      true to keep index on disk so that the next run will continue where the previous run left off
      --lenient                         sets the parser to do as little validation as possible.
      --min-disk-free string            minimum free space on disk to allow WARC writing (default "1GB")
      --name-generator string           the name generator to use. By setting this to 'identity', the input filename will also be used as
                                        output file name (prefix and suffix might still change). In this mode exactly one file is generated for every input file (default "default")
  -K, --new-index                       true to start from a fresh index, deleting eventual index from last run
      --one-to-one                      write each input file to a separate output file
                                        The same as --concurrent-writers=1, --file-size=0 and --name-generator=identity (default true)
      --open-input-file-hook string     a command to run before opening each input file. The command has access to data as environment variables.
                                        	WARC_COMMAND contains the subcommand name
                                        	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
                                        	WARC_FILE_NAME contains the file name of the input file
      --open-output-file-hook string    a command to run before opening each output file. The command has access to data as environment variables.
                                        	WARC_COMMAND contains the subcommand name
                                        	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
                                        	WARC_FILE_NAME contains the file name of the output file
                                        	WARC_SRC_FILE_NAME contains the file name of the input file if the output file is generated from an input file
  -w, --output-dir string               output directory for generated warc files. Directory must exist. (default ".")
  -p, --prefix string                   filename prefix for WARC files
  -r, --recursive                       walk directories recursively
      --source-file-list string         a file containing a list of files to process, one file per line
      --strict                          sets the parser to fail on first validation error.
      --subdir-pattern string           a pattern to use for generating subdirectories.
                                        	/ in pattern separates subdirectories on all platforms
                                        	{YYYY} is replaced with a 4 digit year
                                        	{YY} is replaced with a 2 digit year
                                        	{MM} is replaced with a 2 digit month
                                        	{DD} is replaced with a 2 digit day
                                        	The date used is the WARC date of each record. Therefore a input file might be split into 
                                        	WARC files in different subdirectories. If NameGenerator is 'identity' only the first record
                                        	of each file's date is used to keep the file as one.
      --suffixes strings                filter files by suffix (default [.arc,.arc.gz])
  -s, --symlinks                        follow symlinks
      --tmp-dir string                  directory to use for temporary files (default "/tmp")
      --warc-version string             the WARC version to use for created files (default "1.1")
```

### Options inherited from parent commands

```
      --config string       config file. If not set $XDG_CONFIG_DIRS, /etc/xdg/warc $XDG_CONFIG_HOME/warc and the current directory will be searched for a file named 'config.yaml'
  -O, --log-file string     log to file (default "-")
      --log-format string   log format. Valid values: text, json (default "text")
      --log-level string    log level. Valid values: debug, info, warn, error (default "info")
```

### SEE ALSO

* [warc convert](../warc_convert/)	 - Convert web archive files to WARC files. Use subcommands for the supported formats

