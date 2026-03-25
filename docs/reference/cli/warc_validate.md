---
title: "warc validate"
---
## warc validate

Validate WARC files

```
warc validate FILE/DIR ... [flags]
```

### Options

```
      --calculate-hash string          calculate hash of output file. The hash is made available to the close output file hook as WARC_HASH. Valid values: md5, sha1, sha256, sha512
      --close-input-file-hook string   command to run after closing each input file; the command receives these environment variables:
                                       	WARC_COMMAND contains the subcommand name
                                       	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
                                       	WARC_FILE_NAME contains the file name of the input file
                                       	WARC_ERROR_COUNT contains the number of errors found if the file was validated and the validation failed
  -c, --concurrency int                number of input files to process in parallel (default 21)
      --continue-on-error              continue processing remaining files and directories after errors
  -f, --force                          continue iterating even when record read errors occur
      --ftp-pool-size int32            size of the FTP connection pool (default 1)
  -h, --help                           help for validate
      --id strings                     only process records with these record IDs; repeat the flag or use a comma-separated list
      --index-dir string               directory used to store index data (default "/home/vscode/.cache/warchaeology/validate")
  -i, --input-file string              input filesystem source; default is the local OS filesystem
                                       Legal values:
                                       	/path/to/archive.( tar | tar.gz | tgz | zip | wacz )
                                       	ftp://user/pass@host:port
                                       
  -k, --keep-index                     keep index files in --index-dir after the run so later runs can continue from them
      --lax-host-parsing               allow lenient host parsing in URL parsing
      --lenient                        minimize validation for faster, more permissive parsing
  -l, --limit int                      maximum number of records to process; ignored when --nth is set
  -m, --mime-type strings              only process records with these MIME types; repeat the flag or use a comma-separated list
  -K, --new-index                      start with a fresh index by deleting any existing index in --index-dir at startup
  -n, --nth int                        process only the n-th record after filtering
  -o, --offset int                     start processing at this byte offset in the input file (default: 0)
      --open-input-file-hook string    command to run before opening each input file; the command receives these environment variables:
                                       	WARC_COMMAND contains the subcommand name
                                       	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
                                       	WARC_FILE_NAME contains the file name of the input file
  -t, --record-type strings            only process records with these record types; repeat the flag or use a comma-separated list.
                                       Legal values:
                                       	warcinfo, request, response, metadata, revisit, resource, continuation and conversion
  -r, --recursive                      walk input directories recursively
  -S, --response-code string           only process records by HTTP response code
                                       Example:
                                       	200	- only records with a 200 response
                                       	200-300	- records with response codes between 200 (inclusive) and 300 (exclusive)
                                       	500-	- response codes from 500 and above
                                       	-400	- all response codes below 400
      --source-file-list string        path to a file listing input paths, one per line
      --strict                         fail on the first validation error
      --suffixes strings               only process files with these suffixes (default [.warc,.warc.gz])
  -s, --symlinks                       follow symbolic links while walking
      --tmp-dir string                 directory used for temporary files (default "/tmp")
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

