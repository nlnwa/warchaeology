---
title: "warc dedup"
---
## warc dedup

Deduplicate WARC files

### Synopsis

Deduplicate WARC files.

NOTE: The filtering options only decides which records are candidates for deduplication.
The remaining records are written as is.

```
warc dedup FILE/DIR ... [flags]
```

### Options

```
      --close-input-file-hook string    command to run after closing each input file; the command receives these environment variables:
                                        	WARC_COMMAND contains the subcommand name
                                        	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
                                        	WARC_FILE_NAME contains the file name of the input file
                                        	WARC_ERROR_COUNT contains the number of errors found if the file was validated and the validation failed
      --close-output-file-hook string   command to run after closing each output file; the command receives these environment variables:
                                        	WARC_COMMAND contains the subcommand name
                                        	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
                                        	WARC_FILE_NAME contains the file name of the output file
                                        	WARC_SIZE contains the size of the output file
                                        	WARC_INFO_ID contains the ID of the output file's WARCInfo-record if created
                                        	WARC_SRC_FILE_NAME contains the file name of the input file if the output file is generated from an input file
                                        	WARC_HASH contains the hash of the output file if computed
                                        	WARC_ERROR_COUNT contains the number of errors found if the file was validated and the validation failed
  -z, --compress                        enable gzip compression for WARC output files (default true)
      --compression-level int           gzip compression level (1-9, -1 uses the gzip library default) (default -1)
  -c, --concurrency int                 number of input files to process in parallel (default 21)
  -C, --concurrent-writers int          maximum number of WARC files written concurrently.
                                        This may create at least this many output files even with a single input file. (default 16)
      --continue-on-error               continue processing remaining files and directories after errors
      --default-date string             fallback date used when records are missing WARC-Date metadata (time is set to 12:00 UTC) (default "2026-3-24")
      --deterministic                   force deterministic execution order (single worker and sorted input paths)
      --file-size string                maximum size of each WARC output file (default "1GB")
      --flush                           sync each WARC file to disk after every record
  -f, --force                           continue iterating even when record read errors occur
      --ftp-pool-size int32             size of the FTP connection pool (default 1)
  -h, --help                            help for dedup
      --index-dir string                directory used to store index data (default "/home/vscode/.cache/warchaeology/dedup")
  -i, --input-file string               input filesystem source; default is the local OS filesystem
                                        Legal values:
                                        	/path/to/archive.( tar | tar.gz | tgz | zip | wacz )
                                        	ftp://user/pass@host:port
                                        
  -k, --keep-index                      keep index files in --index-dir after the run so later runs can continue from them
      --lax-host-parsing                allow lenient host parsing in URL parsing
      --lenient                         minimize validation for faster, more permissive parsing
  -l, --limit int                       maximum number of records to process; ignored when --nth is set
      --max-buffer-mem string           the maximum bytes of memory allowed for each buffer before overflowing to disk (default "1MB")
      --min-disk-free string            minimum free disk space required to continue writing WARC output (default "1GB")
      --min-index-disk-free string      minimum free space on disk to allow index writing (default "1MB")
  -g, --min-size-gain string            minimum bytes one must earn to perform a deduplication (default "2KB")
      --name-generator string           name generator strategy.
                                        With 'identity', the input filename is reused for output (prefix/suffix may still change),
                                        and exactly one output file is created per input file. (default "default")
  -K, --new-index                       start with a fresh index by deleting any existing index in --index-dir at startup
  -n, --nth int                         process only the n-th record after filtering
  -o, --offset int                      start processing at this byte offset in the input file (default: 0)
      --one-to-one                      write each input file to exactly one output file.
                                        Equivalent to: --concurrent-writers=1 --file-size=0 --name-generator=identity
      --open-input-file-hook string     command to run before opening each input file; the command receives these environment variables:
                                        	WARC_COMMAND contains the subcommand name
                                        	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
                                        	WARC_FILE_NAME contains the file name of the input file
      --open-output-file-hook string    command to run before opening each output file; the command receives these environment variables:
                                        	WARC_COMMAND contains the subcommand name
                                        	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
                                        	WARC_FILE_NAME contains the file name of the output file
                                        	WARC_SRC_FILE_NAME contains the file name of the input file if the output file is generated from an input file
  -w, --output-dir string               output directory for generated WARC files (must already exist) (default ".")
  -p, --prefix string                   filename prefix for generated WARC files
      --record-types strings            comma separated list of record types to deduplicate. Other record types are written as is. (default [response,resource])
  -r, --recursive                       walk input directories recursively
  -R, --repair                          attempt to repair malformed records when possible
      --source-file-list string         path to a file listing input paths, one per line
      --strict                          fail on the first validation error
      --subdir-pattern string           pattern used to create output subdirectories.
                                        Use '/' to separate subdirectories on all platforms.
                                        Supported tokens: {YYYY}, {YY}, {MM}, {DD}.
                                        The WARC-Date of each record is used, so one input file may be split across subdirectories.
                                        With --name-generator=identity, only the first record date is used per input file.
      --suffixes strings                only process files with these suffixes (default [.warc,.warc.gz])
  -s, --symlinks                        follow symbolic links while walking
      --tmp-dir string                  directory used for temporary files (default "/tmp")
      --warc-version string             WARC version used for generated files (default "1.1")
```

### Options inherited from parent commands

```
      --config string       path to config file; if unset, searches standard XDG config locations and the current directory for config.yaml
  -O, --log-file string     log output destination ('-' for stderr) (default "-")
      --log-format string   log output format (text or json) (default "text")
      --log-level string    minimum log level (debug, info, warn, error) (default "info")
```

### SEE ALSO

* [warc](warc.md)	 - A tool for handling warc files

